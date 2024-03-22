package models

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

type AttemptStatus string

const (
	PossitiveAttempt AttemptStatus = "positive"
	NegativeAttempt  AttemptStatus = "negative"
)

func NewConnAttempt(remotePeer peer.ID, connStatus AttemptStatus, err string, dep, leftNet bool) *ConnectionAttempt {
	return &ConnectionAttempt{
		RemotePeer:  remotePeer,
		Timestamp:   time.Now(),
		Status:      connStatus,
		Error:       err,
		Deprecable:  dep,
		LeftNetwork: leftNet,
	}
}

// ConnectionAttempt is the basic struct that tracks the status of any proactive-attempt to connect any peer in the network
type ConnectionAttempt struct {
	RemotePeer  peer.ID
	Timestamp   time.Time
	Status      AttemptStatus
	Error       string
	Deprecable  bool
	LeftNetwork bool
}
