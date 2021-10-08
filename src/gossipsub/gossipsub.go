/**
This packages describes the implementation of the GossipSub
protocol.
It also provides all needed functions and methods todeploy and interact
with a GossipSub service


*/

package gossipsub

import (
	"context"
	"encoding/base64"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	"github.com/minio/sha256-simd"
)

const PKG_NAME string = "GOSSIP_SUB"

type GossipSub struct {
	*base.Base
	PubsubService *pubsub.PubSub

	TopicArray []*TopicHandler
	InfoObj    *info.InfoData
}

func NewEmptyGossipSub() *GossipSub {
	return &GossipSub{}
}

// constructor
func NewGossipSub(ctx context.Context, i_host hosts.BasicLibp2pHost, stdOpts base.LogOpts) *GossipSub {

	localLogger := gossipsubLoggerOpts(stdOpts)

	// instance base
	new_base, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(localLogger),
	)

	if err != nil {
		new_base.Log.Errorf(err.Error())
	}

	// define gossipsub option
	// Signature is not used in Eth2, therefore it is needed
	// to specify this options to false
	// Otherwise, messages are discarded
	psOptions := []pubsub.Option{
		pubsub.WithMessageSigning(false),
		pubsub.WithStrictSignatureVerification(false),
		pubsub.WithMessageIdFn(MsgIDFunction),
	}
	ps, err := pubsub.NewGossipSub(ctx, i_host.Host(), psOptions...)

	// return the GossipSub object
	return &GossipSub{
		Base:          new_base,
		PubsubService: ps,
		InfoObj:       i_host.GetInfoObj(),
	}
}

// WithMessageIdFn is an option to customize the way a message ID is computed for a pubsub message
func MsgIDFunction(pmsg *pubsub_pb.Message) string {
	h := sha256.New()
	// never errors, see crypto/sha256 Go doc

	_, _ = h.Write(pmsg.Data)
	id := h.Sum(nil)
	return base64.URLEncoding.EncodeToString(id)
}

// this method allows the GossipSub service to join and
// subscribe to a topic
func (gs *GossipSub) JoinAndSubscribe(i_topic string) {

	topic, err := gs.PubsubService.Join(i_topic)

	if err != nil {
		gs.Log.Errorf("Could not join topic: %s", i_topic)
		gs.Log.Errorf(err.Error())
	}

	sub, err := topic.Subscribe()

	if err != nil {
		gs.Log.Errorf("Could not subscribe to topic: %s", i_topic)
		gs.Log.Errorf(err.Error())
	}

	topicLogOpts := base.LogOpts{
		Output:    "terminal",
		Formatter: "text",
		Level:     gs.InfoObj.GetLogLevel(),
	}

	new_topic_handler := NewTopicHandler(gs.Ctx(), topic, *sub, topicLogOpts)

	gs.TopicArray = append(gs.TopicArray, new_topic_handler)

	go gs.TopicArray[len(gs.TopicArray)-1].readLoop()

}

func gossipsubLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME

	return input_opts
}
