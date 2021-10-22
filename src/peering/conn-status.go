package peering

/**
Connection Status is the struct that an active connection attempt done by the host will return to the
peering strategy.
*/

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

// Connection Status is the struct that an active connection
// attempt done by the host will return to the peering strategy.
type ConnectionStatus struct {
	peerID     peer.ID
	timestamp  time.Time // Timestamp of when was the attempt done
	successful bool      // Whether the connection attempt was successfully done or not
	err        error     // if the connection attempt reported any error, nil otherwise
}
