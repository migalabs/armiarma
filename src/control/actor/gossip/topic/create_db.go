package topic

import (
	"context"
	"fmt"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/gossip"
)

type TopicCreateDBCmd struct {
	*base.Base
	GossipState *metrics.GossipState
    GossipMetrics *metrics.GossipMetrics
	// Variables might be usable to see if it already exists a db for the given topic
	//TopicName   string `ask:"--topic-name" help:"The name of the topic to join"`
	Eth2TopicName    string `ask:"--eth-topic" help:"The name of the eth2 topics"`
	ForkDigest       string `ask:"--fork-version" help:"The fork digest value of the network we want to join to (Default Mainnet)"`
	Encoding         string `ask:"--encoding" help:"Encoding that is getting used"`
	TempMessageLimit int    `ask:"--temp-msg-limit" help:"The number of Messages that will be kept on the Temporary Database (Like a Cache of messages)"`
	// For later, when the real database is stored
	//StoreType string `ask:"--store-type" help:"The type of datastore to use. Options: 'mem', 'leveldb', 'badger'"`
	//StorePath string `ask:"--store-path" help:"The path of the datastore, must be empty for memory store."`
}

func (c *TopicCreateDBCmd) Default() {
	c.ForkDigest = "b5303f2a" // Mainnet Fork Digest
	c.TempMessageLimit = 20
	c.Encoding = "ssz_snappy"
}

func (c *TopicCreateDBCmd) Help() string {
	return "Creates a Database where all the received messages on the given topics will be stored. The topic has to be already Joined (Recomended to do it before subscribing the topic)"
}

func (c *TopicCreateDBCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}
	if c.Eth2TopicName != "" {
		// Genereate the full name of the Eth2 topic
		topicName := gossip.GenerateEth2Topics(c.ForkDigest, c.Eth2TopicName, c.Encoding)
		// Check if the topic has been joined
		_, ok := c.GossipState.Topics.Load(topicName)
		if !ok {
			return fmt.Errorf("not on gossip topic %s", topicName)
		}
		// Add the the topic database to the list
		err := c.GossipMetrics.TopicDatabase.NewTopic(topicName, c.TempMessageLimit)
		if err != nil {
			return err
		}
		// TODO: Implement the Real database to be exported every certain time
		return nil
	} else {
		return fmt.Errorf("ERROR: No topic was given")
	}
}
