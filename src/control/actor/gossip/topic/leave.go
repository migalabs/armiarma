package topic

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"
	"github.com/protolambda/rumor/control/actor/gossipimport"
)

type TopicLeaveCmd struct {
	*base.Base
	GossipState   *gossipimport.GossipState
	Eth2TopicName string `ask:"--eth-topic" help:"The name of the eth2 topics"`
	ForkDigest    string `ask:"--fork-version" help:"The fork digest value of the network we want to join to (Default Mainnet)"`
	Encoding      string `ask:"--encoding" help:"Encoding that is getting used"`
}

func (c *TopicLeaveCmd) Default() {
	c.ForkDigest = "b5303f2a" // Mainnet Fork Digest
	c.Encoding = "ssz_snappy"
}

func (c *TopicLeaveCmd) Help() string {
	return "Leave a gossip topic."
}

func (c *TopicLeaveCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}
	if c.Eth2TopicName != "" {
		// Genereate the full name of the Eth2 topic
		topicName := gossip.GenerateEth2Topics(c.ForkDigest, c.Eth2TopicName, c.Encoding)
		if top, ok := c.GossipState.Topics.Load(topicName); !ok {
			return fmt.Errorf("not on gossip topic %s", topicName)
		} else {
			err := top.(*pubsub.Topic).Close()
			if err != nil {
				return err
			}
			c.GossipState.Topics.Delete(topicName)
			return nil
		}
	} else {
		return fmt.Errorf("ERROR: No topic was given")
	}
}
