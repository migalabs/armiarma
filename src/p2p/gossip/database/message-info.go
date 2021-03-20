package database

import (
//	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/zrnt/eth2/beacon"
)

type MessageInfo struct {
	MessageID        string
	MessageType      string // block / attestation
	Slot             beacon.Slot
	ValidatorIndex   beacon.ValidatorIndex
	GotFromList      map[peer.ID]time.Time
	FirstSender      peer.ID
	FirstArrivalTime time.Time
	// currently also de content is getting recorded, let's see later
	Content interface{}
}

func NewMessageInfo(msg *ReceivedMessage) *MessageInfo {
	// generate the MessageInfo
	msgInfo := &MessageInfo{
		MessageID:        msg.GetMessageID(),
		MessageType:      msg.GetMessageType(),
		Slot:             msg.GetSlot(),
		ValidatorIndex:   msg.GetValidatorIndex(),
		GotFromList:      make(map[peer.ID]time.Time, 0),
		FirstSender:      msg.GetSender(),
		FirstArrivalTime: msg.GetArrivalTime(),
		Content:          msg.GetContent(),
	}
	// add the Sender info to the List
	msgInfo.AddNewMsgSender(msg)
	return msgInfo
}

func (mi *MessageInfo) GetMessageID() string {
	return mi.MessageID
}

func (mi *MessageInfo) GetMessageType() string {
	return mi.MessageType
}

func (mi *MessageInfo) GetSlot() beacon.Slot {
	return mi.Slot
}

func (mi *MessageInfo) GetGotFromList() map[peer.ID]time.Time {
	return mi.GotFromList
}

func (mi *MessageInfo) AddNewMsgSender(msg *ReceivedMessage) {
	mi.GotFromList[msg.GetSender()] = msg.GetArrivalTime()
}

func (mi *MessageInfo) GetProposerIndex() beacon.ValidatorIndex {
	return mi.ValidatorIndex
}

func (mi *MessageInfo) GetFirstSender() peer.ID {
	return mi.FirstSender
}

func (mi *MessageInfo) GetFirstArrivalTime() time.Time {
	return mi.FirstArrivalTime
}

func (mi *MessageInfo) GetContent() interface{} {
	return mi.Content
}
