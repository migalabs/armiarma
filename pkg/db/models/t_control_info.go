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
	LastError       string
	NextConnDelay   time.Duration
}

func NewControlInfo() ControlInfo {
	return ControlInfo{}
}

// Deprecate sets the deprecation flag to true while checking if the remote peer looks like it has left the network
func (c *ControlInfo) Deprecate() {
	c.Deprecated = true
	// check if the peer has passed time to be consider as network leaver
	if c.checkIfLeftNetwork() {
		c.markAsLeftNetwork()
	}
}

// IsDeprecated checks whether the peer has been previoulsy deprecated
func (c ControlInfo) IsDeprecated() bool {
	return c.Deprecated
}

// checkIfLeftNetwork check whether the peer has been unreachable for more thatn LeftNetworkTime
func (c ControlInfo) checkIfLeftNetwork() bool {
	return time.Since(c.LastActivity) >= LeftNetworkTime
}

// markAsLeftNetwork sets the LeftNetwork flag to true
func (c *ControlInfo) markAsLeftNetwork() {
	c.LeftNetwork = true
}

// HasLeftNetwork checks whether the peer has been inactive for more than LeftNetworkTime
// updates the LeftNetwork flag and returns the
func (c ControlInfo) HasLeftNetwork() bool {

	return c.LeftNetwork
}

// updateStateOfRemotePeer modifies the state of a remote peer depending on their identification state
func (c *ControlInfo) updateStateOfRemotePeer(newState IdentificationState) {
	c.IdentState = newState
}

// ConnectionAttempt is the basic struct that tracks the status of any proactive-attempt to connect any peer in the network
type ConnectionAttempt struct {
	RemotePeer peer.ID
	Timestamp  time.Time
	Activity   AttemptStatus
	Error      string
}

// AddNewConnAttempt track any new connection attempt to the remote peer
// TODO: - Check whether this should be better done in the Deprecation part
// performs the logic to determine whether the peer needs to be deprecated or not
func (c *ControlInfo) AddNewConnAttemtp(attempt ConnectionAttempt) {

}
