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
	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/migalabs/armiarma/pkg/hosts"
	"github.com/migalabs/armiarma/pkg/metrics"
	"github.com/minio/sha256-simd"
	log "github.com/sirupsen/logrus"
)

// GossipSub
// sumarizes the control fields necesary to manage and
// govern the GossipSub internal service.
type GossipSub struct {
	ctx context.Context

	BasicHost     *hosts.BasicLibp2pHost
	DBClient      *psql.DBClient
	PubsubService *pubsub.PubSub
	Metrics       *metrics.MetricsModule
	// map where the key are the topic names in string, and the values are the TopicSubscription
	TopicArray map[string]*TopicSubscription
}

func NewEmptyGossipSub() *GossipSub {
	return &GossipSub{}
}

// NewGossipSub sumarizes the control fields necesary to manage and govern over a joined and subscribed topic.
func NewGossipSub(ctx context.Context, h *hosts.BasicLibp2pHost, dbClient *psql.DBClient) *GossipSub {

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
	if err != nil {
		log.Panic(err)
	}
	// return the GossipSub object
	return &GossipSub{
		ctx:           ctx,
		BasicHost:     h,
		DBClient:      dbClient,
		PubsubService: ps,
		// Metrics:        metrMod, // TODO: finish this
		TopicArray: make(map[string]*TopicSubscription),
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

// JoinAndSubscribe this method allows the GossipSub service to join and subscribe to a topic.
func (gs *GossipSub) JoinAndSubscribe(topicName string, handlerFn MessageHandler) {
	// Join topic
	topic, err := gs.PubsubService.Join(topicName)
	if err != nil {
		log.Errorf("Could not join topic: %s", topicName)
		log.Errorf(err.Error())
	}
	// Subscribe to the topic
	sub, err := topic.Subscribe()
	if err != nil {
		log.Errorf("Could not subscribe to topic: %s", topicName)
		log.Errorf(err.Error())
	}

	new_topic_handler := NewTopicSubscription(gs.ctx, topic, *sub, handlerFn)
	// Add the new Topic to the list of supported/subscribed topics in GossipSub
	gs.TopicArray[topicName] = new_topic_handler

	go gs.TopicArray[topicName].MessageReadingLoop(gs.BasicHost.Host(), gs.DBClient)
}
