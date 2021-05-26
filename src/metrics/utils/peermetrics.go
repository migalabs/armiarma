package utils

import (
	"github.com/libp2p/go-libp2p-core/peer"
)

// Base Struct for the topic name and the received messages on the different topics
type PeerMetrics struct {
	PeerId     peer.ID
	NodeId     string
	ClientType string
	Pubkey     string
	Addrs      string
	Ip         string
	Country    string
	City       string
	Latency    float64

	Attempted bool   // If the peer has been attempted to stablish a connection
	Succeed   bool   // If the connection attempt has been successful
	Connected bool   // If the peer was at any point connected by the crawler (keep count of incoming dials)
	Attempts  int    // Number of attempts done
	Error     string // Type of error that we detected

	ConnectionEvents []ConnectionEvents
	// Counters for the different topics
	BeaconBlock          MessageMetrics
	BeaconAggregateProof MessageMetrics
	VoluntaryExit        MessageMetrics
	ProposerSlashing     MessageMetrics
	AttesterSlashing     MessageMetrics
	// Variables related to the SubNets (only needed for when Shards will be implemented)
}

func NewPeerMetrics(peerId peer.ID) PeerMetrics {
	pm := PeerMetrics{
		PeerId:     peerId,
		NodeId:     "",
		ClientType: "Unknown",
		Pubkey:     "",
		Addrs:      "/ip4/127.0.0.1/0000",
		Ip:         "127.0.0.1",
		Country:    "Unknown",
		City:       "Unknown",
		Latency:    0,

		Attempted: false,
		Succeed:   false,
		Connected: false,
		Attempts:  0,
		Error:     "None",

		ConnectionEvents: make([]ConnectionEvents, 0),
		// Counters for the different topics
		BeaconBlock:          NewMessageMetrics(),
		BeaconAggregateProof: NewMessageMetrics(),
		VoluntaryExit:        NewMessageMetrics(),
		ProposerSlashing:     NewMessageMetrics(),
		AttesterSlashing:     NewMessageMetrics(),
	}
	return pm
}

func (pm *PeerMetrics) ResetDynamicMetrics() {
	pm.Attempts = 0
	pm.ConnectionEvents = make([]ConnectionEvents, 0)
	pm.BeaconBlock = NewMessageMetrics()
	pm.BeaconAggregateProof = NewMessageMetrics()
	pm.VoluntaryExit = NewMessageMetrics()
	pm.ProposerSlashing = NewMessageMetrics()
	pm.AttesterSlashing = NewMessageMetrics()
}
