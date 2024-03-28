package ethereum

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/codec"
	attdeneb "github.com/attestantio/go-eth2-client/spec/deneb"

	// bls "github.com/phoreproject/github.com/bls/g1pubs"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/migalabs/armiarma/pkg/gossipsub"

	"github.com/golang/snappy"
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

	// SubHanndlers for each gossipsub topic (Oriented mostly for HTTP SSE)
	callbackM        sync.RWMutex
	messageCallbacks map[EthereumGossipTopic][]func(event interface{})
}

func NewEthMessageHandler(genesis time.Time, pubkeysStr []string) (*EthMessageHandler, error) {
	subHandler := &EthMessageHandler{
		genesisTime:      genesis,
		pubkeys:          make([]*common.BLSPubkey, 0, len(pubkeysStr)),
		messageCallbacks: make(map[EthereumGossipTopic][]func(event interface{})),
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

// AddCallBack makes the message sub-handling independent of the message read and agnostic to the topic
func (s *EthMessageHandler) AddCallback(topic EthereumGossipTopic, fn func(event interface{})) {
	// check if there are already any existing callback
	s.callbackM.Lock()
	defer s.callbackM.Unlock()
	callbacks, ok := s.messageCallbacks[topic]
	if !ok {
		callbacks = make([]func(event interface{}), 0)
	}
	callbacks = append(callbacks, fn)
	s.messageCallbacks[topic] = callbacks
}

func (mh *EthMessageHandler) getCallBacks(topic EthereumGossipTopic) ([]func(event interface{}), bool) {
	mh.callbackM.RLock()
	defer mh.callbackM.RUnlock()
	callbacks, ok := mh.messageCallbacks[topic]
	return callbacks, ok
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

	trackedBlock := &TrackedBeaconBlock{
		TrackedMessage: TrackedMessage{
			Msg:    bblock,
			MsgID:  msg.ID,
			Time:   msg.ArrivalTime,
			Sender: msg.ReceivedFrom,
		},
		TimeInSlot: GetTimeInSlot(mh.genesisTime, msg.ArrivalTime, int64(bblock.Message.Slot)),
		ValIndex:   uint64(bblock.Message.ProposerIndex),
		Slot:       uint64(bblock.Message.Slot),
	}
	// check if there is any callback
	callbacks, ok := mh.getCallBacks(BeaconBlockTopic)
	if ok {
		for _, callback := range callbacks {
			callback(&trackedBlock) // TODO: update to submite the
		}
	}
	return trackedBlock, nil
}

// as reference https://github.com/protolambda/zrnt/blob/4ecaadfe0cb3c0a90d85e6a6dddcd3ebed0411b9/eth2/beacon/phase0/indexed.go#L99
func (s *EthMessageHandler) AttestationSubnetMessageHandler(msg *pubsub.Message) (gossipsub.PersistableMsg, error) {
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

	subnet, err := GetSubnetFromTopic("attestation", *msg.Topic)
	if err != nil {
		return nil, err
	}

	trackedAttestation := &TrackedAttestation{
		TrackedMessage: TrackedMessage{
			Msg:    attestation,
			MsgID:  msg.ID,
			Time:   msg.ArrivalTime,
			Sender: msg.ReceivedFrom,
		},
		Subnet:     subnet,
		Slot:       uint64(attestation.Data.Slot),
		TimeInSlot: GetTimeInSlot(s.genesisTime, msg.ArrivalTime, int64(attestation.Data.Slot)),
		ValPubkey:  "",
	}
	// Publish the event
	callbacks, ok := s.getCallBacks(BeaconSubnetAttestationTopic)
	if ok {
		for _, callback := range callbacks {
			// Warning: blocking call, but the only consumers of these "internal" events should be the "events" forwarder which will throw it
			// in to a buffered channel.
			callback(trackedAttestation)
		}
	}
	return trackedAttestation, nil
}

func (mh *EthMessageHandler) BeaconAggregationAndProofMessageHandler(msg *pubsub.Message) (gossipsub.PersistableMsg, error) {
	t := time.Now()
	defer log.Trace("total time to handle msg:", time.Since(t))
	topic := *msg.Topic

	// extract the data from the raw message
	msgBytes, err := EthMessageBaseHandler(topic, msg)
	if err != nil {
		return nil, err
	}
	msgBuf := bytes.NewBuffer(msgBytes)
	aggregation := new(altair.SignedContributionAndProof)

	err = aggregation.Deserialize(configs.Mainnet, codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
	if err != nil {
		return nil, err
	}

	trackedAggragation := &TrackedAggregateAndProof{
		TrackedMessage: TrackedMessage{
			Msg:    aggregation,
			MsgID:  msg.ID,
			Time:   msg.ArrivalTime,
			Sender: msg.ReceivedFrom,
		},
		TimeInSlot: GetTimeInSlot(mh.genesisTime, msg.ArrivalTime, int64(aggregation.Message.Contribution.Slot)),
		Slot:       uint64(aggregation.Message.Contribution.Slot),
	}
	// check if there is any callback
	callbacks, ok := mh.getCallBacks(BeaconAggregateAndProofTopic)
	if ok {
		for _, callback := range callbacks {
			callback(trackedAggragation) // TODO: update to submite the event
		}
	}
	return trackedAggragation, nil
}

func (mh *EthMessageHandler) VoluntaryExitMessageHandler(msg *pubsub.Message) (gossipsub.PersistableMsg, error) {
	t := time.Now()
	defer log.Trace("total time to handle msg:", time.Since(t))
	topic := *msg.Topic

	// extract the data from the raw message
	msgBytes, err := EthMessageBaseHandler(topic, msg)
	if err != nil {
		return nil, err
	}
	msgBuf := bytes.NewBuffer(msgBytes)
	voluntaryExit := new(phase0.SignedVoluntaryExit)

	err = voluntaryExit.Deserialize(codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
	if err != nil {
		return nil, err
	}

	trackedVoluntaryExit := &TrackedVoluntaryExit{
		TrackedMessage: TrackedMessage{
			Msg:    voluntaryExit,
			MsgID:  msg.ID,
			Time:   msg.ArrivalTime,
			Sender: msg.ReceivedFrom,
		},
		ValIndex: uint64(voluntaryExit.Message.ValidatorIndex),
		Epoch:    uint64(voluntaryExit.Message.Epoch),
	}
	// check if there is any callback
	callbacks, ok := mh.getCallBacks(BeaconVoluntaryExitTopic)
	if ok {
		for _, callback := range callbacks {
			callback(trackedVoluntaryExit) // TODO: update to submite the event
		}
	}
	return trackedVoluntaryExit, nil
}

func (mh *EthMessageHandler) ProposerSlashingMessageHandler(msg *pubsub.Message) (gossipsub.PersistableMsg, error) {
	t := time.Now()
	defer log.Trace("total time to handle msg:", time.Since(t))
	topic := *msg.Topic

	// extract the data from the raw message
	msgBytes, err := EthMessageBaseHandler(topic, msg)
	if err != nil {
		return nil, err
	}
	msgBuf := bytes.NewBuffer(msgBytes)
	proposerSlashing := new(phase0.ProposerSlashing)

	err = proposerSlashing.Deserialize(codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
	if err != nil {
		return nil, err
	}

	trackedProposerSlashing := &TrackedProposerSlashing{
		TrackedMessage: TrackedMessage{
			Msg:    proposerSlashing,
			MsgID:  msg.ID,
			Time:   msg.ArrivalTime,
			Sender: msg.ReceivedFrom,
		},
		ProposerIndex: uint64(proposerSlashing.SignedHeader1.Message.ProposerIndex),
		Slot:          uint64(proposerSlashing.SignedHeader1.Message.Slot),
	}
	// check if there is any callback
	callbacks, ok := mh.getCallBacks(BeaconVoluntaryExitTopic)
	if ok {
		for _, callback := range callbacks {
			callback(trackedProposerSlashing) // TODO: update to submite the event
		}
	}
	return trackedProposerSlashing, nil
}

func (mh *EthMessageHandler) TrackedSyncAggregate(msg *pubsub.Message) (gossipsub.PersistableMsg, error) {
	t := time.Now()
	defer log.Trace("total time to handle msg:", time.Since(t))
	topic := *msg.Topic

	// extract the data from the raw message
	msgBytes, err := EthMessageBaseHandler(topic, msg)
	if err != nil {
		return nil, err
	}
	msgBuf := bytes.NewBuffer(msgBytes)
	syncContribution := new(altair.ContributionAndProof)

	err = syncContribution.Deserialize(configs.Mainnet, codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
	if err != nil {
		return nil, err
	}
	trackedSyncAggregate := &TrackedSyncAggregate{
		TrackedMessage: TrackedMessage{
			Msg: 	syncContribution,
			MsgID:  msg.ID,
			Time:   msg.ArrivalTime,
			Sender: msg.ReceivedFrom,
		},
		AggragatorIndex: uint64(syncContribution.AggregatorIndex),
		TimeInSlot:      GetTimeInSlot(mh.genesisTime, msg.ArrivalTime, int64(syncContribution.Contribution.Slot)),
		Slot:            uint64(syncContribution.Contribution.Slot),
	}
	// check if there is any callback
	callbacks, ok := mh.getCallBacks(BeaconSyncCommitteeAggregationTopic)
	if ok {
		for _, callback := range callbacks {
			callback(trackedSyncAggregate) // TODO: update to submite the event
		}
	}
	return trackedSyncAggregate, nil
}

func (mh *EthMessageHandler) TrackedSyncVotes(msg *pubsub.Message) (gossipsub.PersistableMsg, error) {
	t := time.Now()
	defer log.Trace("total time to handle msg:", time.Since(t))
	topic := *msg.Topic

	// extract the data from the raw message
	msgBytes, err := EthMessageBaseHandler(topic, msg)
	if err != nil {
		return nil, err
	}
	msgBuf := bytes.NewBuffer(msgBytes)
	syncVote := new(altair.SyncCommitteeMessage)

	err = syncVote.Deserialize(codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
	if err != nil {
		return nil, err
	}
	trackedSyncMsg := &TrackedSyncMessage{
		TrackedMessage: TrackedMessage{
			Msg: 	syncVote,
			MsgID:  msg.ID,
			Time:   msg.ArrivalTime,
			Sender: msg.ReceivedFrom,
		},
		ValIndex:   uint64(syncVote.ValidatorIndex),
		TimeInSlot: GetTimeInSlot(mh.genesisTime, msg.ArrivalTime, int64(syncVote.Slot)),
		Slot:       uint64(syncVote.Slot),
	}
	// check if there is any callback
	callbacks, ok := mh.getCallBacks(BeaconSubnetSyncCommitteeVoteTopic)
	if ok {
		for _, callback := range callbacks {
			callback(trackedSyncMsg) // TODO: update to submite the event
		}
	}
	return trackedSyncMsg, nil
}


func (mh *EthMessageHandler) TrackedBlobSidecars(msg *pubsub.Message) (gossipsub.PersistableMsg, error) {
	t := time.Now()
	defer log.Trace("total time to handle msg:", time.Since(t))
	topic := *msg.Topic

	// extract the data from the raw message
	msgBytes, err := EthMessageBaseHandler(topic, msg)
	if err != nil {
		return nil, err
	}
	msgBuf := bytes.NewBuffer(msgBytes)
	blobSidecar := new(attdeneb.BlobSidecar)

	err = blobSidecar.UnmarshalSSZ(msgBuf.Bytes())
	if err != nil {
		return nil, err
	}
	trackedSyncMsg := &TrackedBlobSidecards{
		TrackedMessage: TrackedMessage{
			Msg: 	blobSidecar,
			MsgID:  msg.ID,
			Time:   msg.ArrivalTime,
			Sender: msg.ReceivedFrom,
		},
		BlobIndex:   uint64(blobSidecar.Index),
		BeaconBlockRoot: blobSidecar.SignedBlockHeader.Message.Root.String(), 

	}
	// check if there is any callback
	callbacks, ok := mh.getCallBacks(BeaconSubnetSyncCommitteeVoteTopic)
	if ok {
		for _, callback := range callbacks {
			callback(trackedSyncMsg) // TODO: update to submite the event
		}
	}
	return trackedSyncMsg, nil
}