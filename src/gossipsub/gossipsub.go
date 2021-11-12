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
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	"github.com/minio/sha256-simd"
)

const PKG_NAME string = "GOSSIP_SUB"

// GossipSub
// * Sumarizes the control fields necesary to manage and
// * govern the GossipSub internal service
type GossipSub struct {
	*base.Base
	InfoObj       *info.InfoData
	BasicHost     *hosts.BasicLibp2pHost
	PeerStore     *db.PeerStore
	PubsubService *pubsub.PubSub
	// map where the key are the topic names in string, and the values are the TopicSubscription
	TopicArray     map[string]*TopicSubscription
	MessageMetrics *MessageMetrics
}

// NewEmptyGossipSub
// * Sumarizes the control fields necesary to manage and
// * govern over a joined and subscribed topic like
// * @return: gossipsub struct
func NewEmptyGossipSub() *GossipSub {
	return &GossipSub{}
}

// NewGossipSub
// * Sumarizes the control fields necesary to manage and
// * govern over a joined and subscribed topic like
// * @param ctx: parent context for the gossip service
// * @param h: the libp2p.PubSub topic of the joined topic
// * @param peerstore: the peerstore of the
// * @param stdOpts: list of options to generate the base of the gossipsub service
// * @return: pointer to GossipSub struct
func NewGossipSub(ctx context.Context, h *hosts.BasicLibp2pHost, peerstore *db.PeerStore, stdOpts base.LogOpts) *GossipSub {
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
	ps, err := pubsub.NewGossipSub(ctx, h.Host(), psOptions...)
	msgMetrics := NewMessageMetrics()
	// return the GossipSub object
	return &GossipSub{
		Base:           new_base,
		InfoObj:        h.GetInfoObj(),
		BasicHost:      h,
		PeerStore:      peerstore,
		PubsubService:  ps,
		TopicArray:     make(map[string]*TopicSubscription),
		MessageMetrics: &msgMetrics,
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

// JoinAndSubscribe
// * this method allows the GossipSub service to join and
// * subscribe to a topic
// * @param topicName: name of the topic to subscribe
// * @return: pointer to GossipSub struct
func (gs *GossipSub) JoinAndSubscribe(topicName string) {
	// Join topic
	topic, err := gs.PubsubService.Join(topicName)
	if err != nil {
		gs.Log.Errorf("Could not join topic: %s", topicName)
		gs.Log.Errorf(err.Error())
	}
	// Subscribe to the topic
	sub, err := topic.Subscribe()
	if err != nil {
		gs.Log.Errorf("Could not subscribe to topic: %s", topicName)
		gs.Log.Errorf(err.Error())
	}
	// Add the topic to the metrics list
	_ = gs.MessageMetrics.NewTopic(topicName)

	topicLogOpts := base.LogOpts{
		Output:    "terminal",
		Formatter: "text",
		Level:     gs.InfoObj.GetLogLevel(),
	}

	new_topic_handler := NewTopicSubscription(gs.Ctx(), topic, *sub, gs.MessageMetrics, topicLogOpts)
	// Add the new Topic to the list of supported/subscribed topics in GossipSub
	gs.TopicArray[topicName] = new_topic_handler

	go gs.TopicArray[topicName].MessageReadingLoop(gs.BasicHost.Host(), gs.PeerStore)

}

func gossipsubLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME

	return input_opts
}
