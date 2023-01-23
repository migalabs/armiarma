package models

import (
	"sync"
	"time"
)

type IdentificationState int8

const (
	NotIdentified IdentificationState = iota
	AttemptedAndSucceed
	AttemptedNotSucceed
)

const (
	DeprecableTime = 24 * time.Hour
	// if in 2 months we didn't connect the peer,
	// say that it left the network
	// unless discv5 says the opposite
	LeftNetworkTime = 24 * time.Hour * 60
)

type ControlInfo struct {
	m sync.RWMutex
	// major variables
	Deprecated  bool
	LeftNetwork bool

	// state of the peer
	IdentState IdentificationState

	// control timestamps
	LastActivity    time.Time
	LastConnAttempt time.Time
	LastError       string
	NextConnDelay   time.Duration
}

func NewControlInfo() ControlInfo {
	return ControlInfo{
		IdentState: NotIdentified,
		LastError:  "",
	}
}

// THIS MIGHT BE OUTDATED due to the constant info-sharing with the DB

// // Deprecate sets the deprecation flag to true while checking if the remote peer looks like it has left the network
// func (c *ControlInfo) Deprecate() {
// 	c.m.Lock()
// 	defer c.m.Unlock()

// 	c.Deprecated = true
// 	// check if the peer has passed time to be consider as network leaver
// 	if c.checkIfLeftNetwork() {
// 		c.markAsLeftNetwork()
// 	}
// }

// // IsDeprecated checks whether the peer has been previoulsy deprecated
// func (c ControlInfo) IsDeprecated() bool {
// 	c.m.RLock()
// 	defer c.m.RUnlock()

// 	return c.Deprecated
// }

// // checkIfLeftNetwork check whether the peer has been unreachable for more thatn LeftNetworkTime
// func (c ControlInfo) checkIfLeftNetwork() bool {
// 	c.m.RLock()
// 	defer c.m.RUnlock()

// 	return time.Since(c.LastActivity) >= LeftNetworkTime
// }

// // markAsLeftNetwork sets the LeftNetwork flag to true
// func (c *ControlInfo) markAsLeftNetwork() {
// 	c.m.Lock()
// 	defer c.m.Unlock()

// 	c.LeftNetwork = true
// }

// // HasLeftNetwork checks whether the peer has been inactive for more than LeftNetworkTime
// // updates the LeftNetwork flag and returns the
// func (c ControlInfo) HasLeftNetwork() bool {
// 	c.m.Lock()
// 	defer c.m.Unlock()

// 	if c.checkIfLeftNetwork() {
// 		c.markAsLeftNetwork()
// 	}
// 	return c.LeftNetwork
// }

// // updateStateOfRemotePeer modifies the state of a remote peer depending on their identification state
// func (c *ControlInfo) updateStateOfRemotePeer(newState IdentificationState) {
// 	c.m.Lock()
// 	defer c.m.Unlock()

// 	c.IdentState = newState
// }
