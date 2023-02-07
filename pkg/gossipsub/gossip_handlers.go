package gossipsub

import (
	"bytes"
	"fmt"
	"time"

	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"

	blsu "github.com/protolambda/bls12-381-util"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"

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

type SubnetsHandler struct {
	pubkeys []*common.BLSPubkey // pubkeys of those validators we want to track

}

func NewSubnetsHandler(pubkeysStr []string) (*SubnetsHandler, error) {
	subHandler := &SubnetsHandler{
		pubkeys: make([]*common.BLSPubkey, 0, len(pubkeysStr)),
	}
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

// as reference https://github.com/protolambda/zrnt/blob/4ecaadfe0cb3c0a90d85e6a6dddcd3ebed0411b9/eth2/beacon/phase0/indexed.go#L99
func (s *SubnetsHandler) SubnetMessageHandler(msg *pubsub.Message) (PersistableMsg, error) {
	t := time.Now()
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
	fmt.Println("time to deserialize Msg", time.Since(t))
	t1 := time.Now()
	// get the hashTree of the Attestation
	attHash := attestation.Data.HashTreeRoot(tree.GetHashFn())

	// get the signature of the Attestation Data
	attSignature, err := attestation.Signature.Signature()
	if err != nil {
		return nil, err
	}
	fmt.Println("time to get hash and signature Msg", time.Since(t1))
	t2 := time.Now()
	// verify if the hash of the message, the signature and the pubkeys of the list of validators match
	trackedAttestation := &eth.TrackedAttestation{}
	defer func() {
		fmt.Println("time to check if our validators are the origin of the msg", time.Since(t2))
		fmt.Println("total operation time", time.Since(t))
	}()
	for _, pubkey := range s.pubkeys {
		pubk, err := pubkey.Pubkey()
		if err != nil {
			log.WithError(err).Warn("unable to get blsu.Pubkey from BLS.Pubkey")
		}
		if blsu.Verify(pubk, attHash[:], attSignature) {
			log.Info("Attestation for a known validator found")
			trackedAttestation.ArrivalTime = msg.ArrivalTime
			trackedAttestation.Sender = msg.ReceivedFrom
			trackedAttestation.Slot = int64(attestation.Data.Slot)
			trackedAttestation.Pubkey = pubkey.String()
			break
		}
	}

	return trackedAttestation, nil
}
