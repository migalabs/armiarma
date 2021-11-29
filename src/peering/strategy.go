package peering

import (
	"time"

	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/hosts"
)

// Strategy is the common interface the any desired Peering Strategy should follow
// TODO:  -Still waiting to be defined to make it official
type PeeringStrategy interface {
	// one channel to give the next peer, one to request the second one
	Run() chan db.Peer
	Close()
	Type() string
	// Peering Strategy interaction
	NextPeer()
	NewConnectionAttempt(ConnectionAttemptStatus)
	NewConnectionEvent(hosts.ConnectionEvent)
	NewIdentificationEvent(hosts.IdentificationEvent)
	// Prometheus Export Calls
	LastIterTime() float64
	IterForcingNextConnTime() string
	AttemptedPeersSinceLastIter() int64
	ControlDistribution() map[string]int64
}

// ConnectionAttemptStatus
// * It is the struct that compiles the data of an active connection attempt done by the host
// * The struct will be shared between peering and strategy.
type ConnectionAttemptStatus struct {
	Peer       db.Peer   // TODO: right now just sending the entire info about the peer, (recheck after Peer struct subdivision)
	Attempts   int32     // attemps tried on the given peer
	Timestamp  time.Time // Timestamp of when was the attempt done
	Successful bool      // Whether the connection attempt was successfully done or not
	RecError   error     // if the connection attempt reported any error, nil otherwise
	// TODO: More things to add in te future
}
