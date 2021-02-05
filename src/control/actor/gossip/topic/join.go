package topic

import (
	"context"
	"fmt"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/gossip"
)

type TopicJoinCmd struct {
	*base.Base
	GossipState   *metrics.GossipState
	Eth2TopicName string `ask:"--eth-topic" help:"The name of the eth2 topics"`
	ForkDigest    string `ask:"--fork-version" help:"The fork digest value of the network we want to join to (Default Mainnet)"`
	Encoding      string `ask:"--encoding" help:"Encoding that is getting used"`
}

func (c *TopicJoinCmd) Default() {
	c.ForkDigest = "b5303f2a" // Mainnet Fork Digest
	c.Encoding = "ssz_snappy"
}

func (c *TopicJoinCmd) Help() string {
	return "Join a gossip topic. This only sets up the topic, it does not actively find peers. See `gossip log start` and `gossip publish`."
}

func (c *TopicJoinCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}
	if c.Eth2TopicName != "" {
		topicName := gossip.GenerateEth2Topics(c.ForkDigest, c.Eth2TopicName, c.Encoding)
		_, ok := c.GossipState.Topics.Load(topicName)
		if ok {
			return fmt.Errorf("already on gossip topic %s", topicName)
		}
		top, err := c.GossipState.GsNode.Join(topicName)
		if err != nil {
			return err
		}
		c.GossipState.Topics.Store(topicName, top)
		c.Log.Infof("joined topic %s", topicName)
	} else {
		return fmt.Errorf("ERROR: No topic was given")
	}

	return nil
}
