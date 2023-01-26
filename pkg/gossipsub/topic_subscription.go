package gossipsub

import (
	"context"
	"strings"

	"github.com/golang/snappy"
	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	log "github.com/sirupsen/logrus"
)

// TopicSubscription
// Sumarizes the control fields necesary to manage and
// govern over a joined and subscribed topic like
// message logging or record.
// Serves as a server for a singe topic subscription.
type TopicSubscription struct {
	ctx context.Context

	// Messages is a channel of messages received from other peers in the chat room
	Messages       chan []byte
	Topic          *pubsub.Topic
	Sub            *pubsub.Subscription
	MessageMetrics *MessageMetrics
}

// NewTopicSubscription:
// Sumarizes the control fields necesary to manage and
// govern over a joined and subscribed topic.
// @param ctx: parent context of the topic subscription, generally gossipsub context.
// @param topic: the libp2p.PubSub topic of the joined topic.
// @param sub: the libp2p.PubSub subscription of the subscribed topic.
// @param msgMetrics: underlaying message metrics regarding each of the joined topics.
// @param stdOpts: list of options to generate the base of the topic subscription service.
// @return: pointer to TopicSubscription.
func NewTopicSubscription(ctx context.Context, topic *pubsub.Topic, sub pubsub.Subscription, msgMetrics *MessageMetrics) *TopicSubscription {
	return &TopicSubscription{
		ctx:            ctx,
		Topic:          topic,
		Sub:            &sub,
		MessageMetrics: msgMetrics,
		Messages:       make(chan []byte),
	}
}

// MessageReadingLoop:
// Pulls messages from the pubsub topic and pushes them onto the Messages channel
// and the underlaying msg metrics.
// @param h: libp2p host.
// @param peerstore: peerstore of the crawler app.
func (c *TopicSubscription) MessageReadingLoop(h host.Host, dbClient *psql.DBClient) {
	log.Infof("topic subscription %s reading loop", c.Sub.Topic())
	subsCtx := c.ctx
	for {
		msg, err := c.Sub.Next(subsCtx)
		if err != nil {
			if err == subsCtx.Err() {
				log.Errorf("context of the subsciption %s has been canceled", c.Sub.Topic())
				break
			}
			log.Errorf("error reading next message in topic %s. %slol", c.Sub.Topic(), err.Error())
		} else {
			var msgData []byte
			if strings.HasSuffix(c.Sub.Topic(), "_snappy") {
				msgData, err = snappy.Decode(nil, msg.Data)
				if err != nil {
					log.WithError(err).WithField("topic", c.Sub.Topic()).Error("Cannot decompress snappy message")
					continue
				}
			}
			// To avoid getting track of our own messages, check if we are the senders
			if msg.ReceivedFrom != h.ID() {
				log.Debugf("new message on %s from %s", c.Sub.Topic(), msg.ReceivedFrom)
				// HUGE TODO: missing message track for analytics
				// newPeer := models.NewPeer(msg.ReceivedFrom.String())
				// newPeer.MessageEvent(c.Sub.Topic(), time.Now())
				// peerstore.StoreOrUpdatePeer(newPeer)
				// Add message to msg metrics counter
				_ = c.MessageMetrics.AddMessgeToTopic(c.Sub.Topic())

				log.Debugf("msg content %s", msgData)
			} else {
				log.Debugf("message sent by ourselfs received on %s", c.Sub.Topic())
			}
		}
	}
	<-subsCtx.Done()
	log.Debugf("ending %s reading loop", c.Sub.Topic())
}
