package models

import (
	"time"
)

type PeerActivity int8

const (
	PossitiveActivity PeerActivity = iota
	NegativeActivity
)

type IdentificationState int8

const (
	NonIdentified IdentificationState = iota
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
	// major variables
	Deprecated  bool
	LeftNetwork bool

	// state of the peer
	IdentState           IdentificationState
	LastConnAttemptError error

	// control timestamps
	LastActivity    time.Time
	LastConnAttempt time.Time
	NextConnDelay   time.Duration
}

func NewControlInfo() ControlInfo {
	return ControlInfo{}
}

func (c *ControlInfo) Deprecate() {
	c.Deprecated = true
	// check if the peer has passed time to be consider as network leaver
	if c.checkIfLeftNetwork() {
		c.MarkAsLeftNetwork()
	}
}

func (c ControlInfo) IsDeprecated() bool {
	return c.Deprecated
}

func (c ControlInfo) checkIfLeftNetwork() bool {
	return time.Since(c.LastActivity) >= LeftNetworkTime
}

func (c *ControlInfo) MarkAsLeftNetwork() {
	c.LeftNetwork = true
}

func (c ControlInfo) HasLeftNetwork() bool {
	return c.LeftNetwork
}

func (c *ControlInfo) UpdateStateOfRemotePeer(newState IdentificationState) {
	c.IdentState = newState
}
