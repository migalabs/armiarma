package gossip

import (
	"context"
	"fmt"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/gossip/database"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/rumor/control/actor/gossipimport"
)

type GossipMessageDBCmd struct {
	*base.Base
	GossipState   *gossipimport.GossipState
	PeerStore *metrics.PeerStore

	MessageLimit int `ask:"--temp-msg-limit" help:"The number of Messages that will be kept on the Temporary Database (Like a Cache of messages)"`
	// For later, when the real database is stored`
	StorePath string `ask:"--store-path" help:"The path of the datastore, If the folder isn't empty, the current msgs will be added to the Seen List."`
}

func (c *GossipMessageDBCmd) Default() {
	c.StorePath = "message-db" // Mainnet Fork Digest
	c.MessageLimit = 3000 // hardcoded to 3000
}

func (c *GossipMessageDBCmd) Help() string {
	return "Creates a Database where all the received messages on the given topics will be stored. The topic has to be already Joined (Recomended to do it before subscribing the topic)"
}

func (c *GossipMessageDBCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}
	// config has been harcoded to the Mainnet Network
	c.PeerStore.MessageDatabase = database.NewMessageDatabase(configs.Mainnet, c.MessageLimit, c.StorePath)
	if c.PeerStore.MessageDatabase == nil {
		return fmt.Errorf("The Message Database failed to initialize")
	}
	// TODO: Implement the Real database to be exported every certain time
	// Currently the Messages in Memory gets exported when the defined limit is given
	return nil
}
