package ethereum

import (
	"bytes"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"time"

	// bls "github.com/phoreproject/github.com/bls/g1pubs"

	"github.com/migalabs/armiarma/pkg/gossipsub"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/codec"

	"github.com/golang/snappy"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// EthMessageBaseHandler extracts that basing message data from the
// entire Pubsub.Message message
func EthMessageBaseHandler(topic string, msg *pubsub.Message) ([]byte, error) {
	var msgData []byte
	msgData, err := snappy.Decode(nil, msg.Data)
	if err != nil {
		return msgData, errors.Wrap(err, "cannot decompress snappy message")
	}
	return msgData, nil
}

type EthMessageHandler struct {
	genesisTime time.Time
	pubkeys     []*common.BLSPubkey // pubkeys of those validators we want to track

	attestationCallbacks []func(event *AttestationReceievedEvent)
}

func NewEthMessageHandler(genesis time.Time, pubkeysStr []string) (*EthMessageHandler, error) {
	subHandler := &EthMessageHandler{
		genesisTime: genesis,
		pubkeys:     make([]*common.BLSPubkey, 0, len(pubkeysStr)),
	}
	// parse pubkeys
	for _, pubkeyStr := range pubkeysStr {
		blsKey := &common.BLSPubkey{}
		err := blsKey.UnmarshalText([]byte(pubkeyStr))
		if err != nil {
			return subHandler, err
		}
		if blsKey.String() != pubkeyStr {
			return subHandler, fmt.Errorf("blsKey (%s) and given-pubkey (%s) missmatch", blsKey.String(), pubkeyStr)
		}
		subHandler.pubkeys = append(subHandler.pubkeys, blsKey)
	}
	return subHandler, nil
}

func (s *EthMessageHandler) OnAttestation(fn func(event *AttestationReceievedEvent)) {
	s.attestationCallbacks = append(s.attestationCallbacks, fn)
}

// as reference https://github.com/protolambda/zrnt/blob/4ecaadfe0cb3c0a90d85e6a6dddcd3ebed0411b9/eth2/beacon/phase0/indexed.go#L99
func (s *EthMessageHandler) SubnetMessageHandler(msg *pubsub.Message) (gossipsub.PersistableMsg, error) {
	t := time.Now()
	defer log.Trace("total time to handle msg:", time.Since(t))

	topic := *msg.Topic

	// extract the data from the raw message
	msgBytes, err := EthMessageBaseHandler(topic, msg)
	if err != nil {
		return nil, err
	}
	msgBuf := bytes.NewBuffer(msgBytes)

	// once we have the data, get Attestation from it
	var attestation phase0.Attestation
	err = attestation.Deserialize(configs.Mainnet, codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
	if err != nil {
		return nil, err
	}

	// ----- TODO: attestation ownership still missing
	// // get Signing Root of the Attestation
	// signingRoot := common.ComputeSigningRoot(attestation.Data.HashTreeRoot(tree.GetHashFn()), dom)
	// attSignature, err := attestation.Signature.Signature()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to deserialize and sub-group check indexed attestation signature: %v", err)
	// }
	// for _, pubkey := range s.pubkeys {
	// 	pubk, err := pubkey.Pubkey()
	// 	if err != nil {
	// 		log.WithError(err).Warn("unable to get blsu.Pubkey from BLS.Pubkey")
	// 	}
	// 	if blsu.Verify(pubk, signingRoot[:], attSignature) {
	// 		log.Info("--------->>>>>>Attestation for a known validator found")

	// verify if the hash of the message, the signature and the pubkeys of the list of validators match

	subnet, err := GetSubnetFromTopic(*msg.Topic)
	if err != nil {
		return nil, err
	}

	arrivalTime := time.Now()
	trackedAttestation := &TrackedAttestation{
		MsgID:       msg.ID,
		ArrivalTime: arrivalTime,
		Subnet:      subnet,
		Slot:        int64(attestation.Data.Slot),
		TimeInSlot:  GetTimeInSlot(s.genesisTime, arrivalTime, int64(attestation.Data.Slot)),
		Sender:      msg.ReceivedFrom,
		ValPubkey:   "",
	}

	// Publish the event
	for _, fn := range s.attestationCallbacks {
		// Warning: blocking call, but the only consumers of these "internal" events should be the "events" forwarder which will throw it
		// in to a buffered channel.
		fn(&AttestationReceievedEvent{
			Attestation:        &attestation,
			TrackedAttestation: trackedAttestation,
			PeerID:             msg.ReceivedFrom,
		})
	}

	return trackedAttestation, nil
}

func (mh *EthMessageHandler) BeaconBlockMessageHandler(msg *pubsub.Message) (gossipsub.PersistableMsg, error) {
	t := time.Now()
	defer log.Trace("total time to handle msg:", time.Since(t))
	topic := *msg.Topic

	// extract the data from the raw message
	msgBytes, err := EthMessageBaseHandler(topic, msg)
	if err != nil {
		return nil, err
	}
	msgBuf := bytes.NewBuffer(msgBytes)
	bblock := new(deneb.SignedBeaconBlock)

	err = bblock.Deserialize(configs.Mainnet, codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
	if err != nil {
		return nil, err
	}

	arrivalTime := time.Now()
	trackedBlock := &TrackedBeaconBlock{
		MsgID:       msg.ID,
		Sender:      msg.ReceivedFrom,
		ArrivalTime: arrivalTime,
		TimeInSlot:  GetTimeInSlot(mh.genesisTime, arrivalTime, int64(bblock.Message.Slot)),
		ValIndex:    int64(bblock.Message.ProposerIndex),
		Slot:        int64(bblock.Message.Slot),
	}

	return trackedBlock, nil
}
