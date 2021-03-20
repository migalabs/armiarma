package database

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/zrnt/eth2/beacon"
)


type ReceivedMessage struct {
	MessageID 		string
	MessageType		string // block / attestation
	Slot			beacon.Slot
	ValidatorIndex	beacon.ValidatorIndex
	Sender			peer.ID
	ArrivalTime		time.Time
 	// currently also de content is getting recorded, let's see later
	Content 		interface{}
}

func (rm *ReceivedMessage) GetMessageID() string{
	return rm.MessageID
}

func (rm *ReceivedMessage) GetMessageType() string{
	return rm.MessageType
}

func (rm *ReceivedMessage) GetSlot() beacon.Slot{
	return rm.Slot
}

func (rm *ReceivedMessage) GetValidatorIndex() beacon.ValidatorIndex{
	return rm.ValidatorIndex
}

func (rm *ReceivedMessage) GetSender() peer.ID{
	return rm.Sender
}

func (rm *ReceivedMessage) GetArrivalTime() time.Time{
	return rm.ArrivalTime
}

func (rm *ReceivedMessage) GetContent() interface{}{
	return rm.Content
}