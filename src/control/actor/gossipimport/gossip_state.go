package gossipimport

import (
  "context"
  "sync"
  pgossip "github.com/protolambda/rumor/p2p/gossip"
)


// Moving this here temporally from /metrics
type GossipState struct {
	GsNode  pgossip.GossipSub
	CloseGS context.CancelFunc
	// string -> *pubsub.Topic
	Topics sync.Map
	// Validation Filter Flag
	SeenFilter bool
}
