package gossip

import (
	"context"
	"errors"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/gossip"
)

type GossipStartCmd struct {
	*base.Base
	*metrics.GossipState
	*metrics.PeerStore

	SeenFilter bool `ask:"--seen-filter" help:"Enable or Disable the Received Message Validation (Default: True)"`
}

func (c *GossipStartCmd) Help() string {
	return "Start GossipSub"
}

func (c *GossipStartCmd) Default() {
	c.SeenFilter = true
}

func (c *GossipStartCmd) Run(ctx context.Context, args ...string) error {
	h, err := c.Host()
	if err != nil {
		return err
	}
	if c.GossipState.GsNode != nil {
		return errors.New("Already started GossipSub")
	}
	c.GossipState.GsNode, err = gossip.NewGossipSub(c.ActorContext, h, c.SeenFilter)
	if err != nil {
		return err
	}
	c.Log.Info("Started GossipSub")
	// Chech if the Validation Filter has been set
	if c.SeenFilter {
		c.Log.Info("Message Seen Filter has been Enabled, Messages WILL be dropped if we saw them before")
	} else {
		c.Log.Info("Message ValidSeenation Filter has been Disabled, Messages WONT be dropped when received")
	}
	// Doesn't really matter the flag, we have to add it the GossipState anyways
	c.GossipState.SeenFilter = c.SeenFilter
	return nil
}
