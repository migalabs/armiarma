package topic

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/snappy"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/gossip"
	"github.com/protolambda/rumor/control/actor/gossipimport"
)

type TopicPublishCmd struct {
	*base.Base
	GossipState   *gossipimport.GossipState
	Eth2TopicName string `ask:"--eth-topic" help:"The name of the eth2 topics"`
	ForkDigest    string `ask:"--fork-version" help:"The fork digest value of the network we want to join to (Default Mainnet)"`
	Encoding      string `ask:"--encoding" help:"Encoding that is getting used"`
	Message       []byte `ask:"<message>" help:"The uncompressed message bytes, hex-encoded"`
}

func (c *TopicPublishCmd) Default() {
	c.ForkDigest = "b5303f2a" // Mainnet Fork Digest
	c.Encoding = "ssz_snappy"
}

func (c *TopicPublishCmd) Help() string {
	return "Publish a message to the topic. The message should be hex-encoded."
}

func (c *TopicPublishCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}
	if c.Eth2TopicName != "" {
		// Genereate the full name of the Eth2 topic
		topicName := gossip.GenerateEth2Topics(c.ForkDigest, c.Eth2TopicName, c.Encoding)
		if top, ok := c.GossipState.Topics.Load(topicName); !ok {
			return fmt.Errorf("not on gossip topic %s", topicName)
		} else {
			data := c.Message
			if strings.HasSuffix(topicName, "_snappy") {
				data = snappy.Encode(nil, data)
			}
			if err := top.(*pubsub.Topic).Publish(ctx, data); err != nil {
				return fmt.Errorf("failed to publish message, err: %v", err)
			}
			return nil
		}
	} else {
		return fmt.Errorf("ERROR: No topic was given")
	}
}
