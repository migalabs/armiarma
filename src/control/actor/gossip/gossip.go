package gossip

import (
	"errors"

	"github.com/protolambda/ask"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/control/actor/gossip/topic"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/protolambda/rumor/control/actor/gossipimport"
)

type GossipCmd struct {
	*base.Base
	*gossipimport.GossipState
	*metrics.PeerStore
	Store track.ExtendedPeerstore
}

func (c *GossipCmd) Cmd(route string) (cmd interface{}, err error) {
	switch route {
	case "start":
		cmd = &GossipStartCmd{Base: c.Base, GossipState: c.GossipState}
	case "list":
		cmd = &GossipListCmd{Base: c.Base, GossipState: c.GossipState}
	case "message-db":
		cmd = &GossipMessageDBCmd{Base: c.Base, GossipState: c.GossipState, PeerStore: c.PeerStore}
	case "blacklist":
		cmd = &GossipBlacklistCmd{Base: c.Base, GossipState: c.GossipState}
	case "topic":
		cmd = &topic.TopicCmd{Base: c.Base, GossipState: c.GossipState, PeerStore: c.PeerStore, Store: c.Store}
	default:
		return nil, ask.UnrecognizedErr
	}
	return cmd, nil
}

func (c *GossipCmd) Routes() []string {
	return []string{"start", "list", "message-db", "blacklist", "topic"}
}

func (c *GossipCmd) Help() string {
	return "Manage Libp2p GossipSub"
}

var NoGossipErr = errors.New("Must start gossip-sub first. Try 'gossip start'")
