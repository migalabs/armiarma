package peering

import (
	"sync"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/hosts"
)

// Strategy is the common interface the any desired Peering Strategy should follow
// TODO:  -Still waiting to be defined to make it official
type PeeringStrategy interface {
	// one channel to give the next peer, one to request the second one
	Run() chan *models.HostInfo
	Type() string
	// Peering Strategy interaction
	NextPeer()
	NewConnectionAttempt(*models.ConnectionAttempt)
	NewConnectionEvent(*models.EventTrace)
	NewIdentificationEvent(hosts.IdentificationEvent)
	// Prometheus Export Calls
	LastIterTime() float64
	IterForcingNextConnTime() string
	AttemptedPeersSinceLastIter() int64
	ControlDistribution() map[string]int
	GetErrorAttemptDistribution() sync.Map
}
