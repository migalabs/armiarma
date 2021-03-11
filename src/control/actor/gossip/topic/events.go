package topic

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/gossip"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/sirupsen/logrus"
)

type TopicEventsCmd struct {
	*base.Base
	GossipState   *metrics.GossipState
	GossipMetrics *metrics.GossipMetrics
	Store         track.ExtendedPeerstore
	//TopicName string `ask:"<topic>" help:"The name of the topic to track events of"`
	Eth2TopicName string `ask:"--eth-topic" help:"The name of the eth2 topics"`
	ForkDigest    string `ask:"--fork-version" help:"The fork digest value of the network we want to join to (Default Mainnet)"`
	Encoding      string `ask:"--encoding" help:"Encoding that is getting used"`
}

func (c *TopicEventsCmd) Default() {
	c.ForkDigest = "b5303f2a" // Mainnet Fork Digest
	c.Encoding = "ssz_snappy"
}

func (c *TopicEventsCmd) Help() string {
	return "Listen for events (not messages) on this topic. Events: 'join=<peer-ID>', 'leave=<peer-ID>'"
}

func (c *TopicEventsCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}
	if c.Eth2TopicName != "" {
		// Generate the full name of the Eth2 topic
		topicName := gossip.GenerateEth2Topics(c.ForkDigest, c.Eth2TopicName, c.Encoding)
		// Check if the generated topic is on the list of joined topics
		top, ok := c.GossipState.Topics.Load(topicName)
		if !ok {
			return fmt.Errorf("not on gossip topic %s", topicName)
		}
		evHandler, err := top.(*pubsub.Topic).EventHandler()
		if err != nil {
			return err
		}
		ctx, cancelEvs := context.WithCancel(ctx)
		go func() {
			c.Log.Infof("Started listening for peer join/leave events for topic %s", topicName)
			for {
				ev, err := evHandler.NextPeerEvent(ctx)
				if err != nil {
					c.Log.Infof("Stopped listening for peer join/leave events for topic %s", topicName)
					return
				}
				switch ev.Type {
				case pubsub.PeerJoin:
					c.Log.WithFields(logrus.Fields{"peer_id": ev.Peer, "topic": topicName}).Info("topic joined")
				case pubsub.PeerLeave:
					c.Log.WithFields(logrus.Fields{"peer_id": ev.Peer, "topic": topicName}).Info("topic left")
				}
			}
		}()
		c.Control.RegisterStop(func(ctx context.Context) error {
			cancelEvs()
			return nil
		})
	} else {
		return fmt.Errorf("ERROR: No topic was given")
	}
	return nil
}
