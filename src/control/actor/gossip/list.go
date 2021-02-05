package gossip

import (
	"context"
	"github.com/protolambda/rumor/control/actor/base"
    "github.com/protolambda/rumor/metrics"
)

type GossipListCmd struct {
	*base.Base
	*metrics.GossipState
}

func (c *GossipListCmd) Help() string {
	return "List joined gossip topics"
}

func (c *GossipListCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}
	var topics []string
    //topics := make([]string, 0)
	c.GossipState.Topics.Range(func(key, value interface{}) bool {
		topics = append(topics, key.(string))
		return true
	})
	c.Log.WithField("topics", topics).Infof("On %d topics.", len(topics))
	return nil
}
