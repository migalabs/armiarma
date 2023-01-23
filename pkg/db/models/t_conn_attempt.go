package models

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

type AttemptStatus string

const (
	PossitiveAttempt AttemptStatus = "positive"
	NegativeAttempt  AttemptStatus = "negative"
)

func NewConnAttempt(remotePeer peer.ID, connStatus AttemptStatus, err string) *ConnectionAttempt {
	return &ConnectionAttempt{
		RemotePeer: remotePeer,
		Timestamp:  time.Now(),
		Status:     connStatus,
		Error:      err,
	}
}

// ConnectionAttempt is the basic struct that tracks the status of any proactive-attempt to connect any peer in the network
type ConnectionAttempt struct {
	RemotePeer peer.ID
	Timestamp  time.Time
	Status     AttemptStatus
	Error      string
}

// TODO: this should be handled by the PSQL database model

// // AddNewConnAttempt track any new connection attempt to the remote peer
// // TODO: - Check whether this should be better done in the Deprecation part
// // performs the logic to determine whether the peer needs to be deprecated or not
// func (c *ControlInfo) AddNewConnAttemtp(attempt ConnectionAttempt) {
// 	c.m.Lock()
// 	defer c.m.Unlock()

// 	// update always last ConnAttempt
// 	c.LastConnAttempt = attempt.Timestamp

// 	// Update to the new error
// 	c.LastError = attempt.Error

// 	// check if the connection was successful
// 	if attempt.Status == PossitiveAttempt {
// 		c.LastActivity = attempt.Timestamp
// 	}

// }
