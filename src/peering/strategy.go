package peering

import (
	"github.com/libp2p/go-libp2p-core/peer"
)

// Strategy is the common interface the any desired Peering Strategy should follow
// TODO:  -Still waiting to be defined to make it official
type PeeringStrategy interface {
	PeerStream() chan NextPeer
	//GetPeerBatch() []peer.ID
	Start()
	Stop()
}

// struct that will be served by the Peering strategy to the Peering Service
// containing the peerID of the next peer to connect, and a channel were to return the status of the connection attempt
type NextPeer struct {
	PeerID       peer.ID
	StatusReport chan ConnectionStatus
}
