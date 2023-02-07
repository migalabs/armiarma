package gossipsub

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type MessageHandler func(*pubsub.Message) (PersistableMsg, error)

type PersistableMsg interface {
	IsZero() bool
}

// TopicSubscription
// Sumarizes the control fields necesary to manage and
// govern over a joined and subscribed topic like
// message logging or record.
// Serves as a server for a singe topic subscription.
type TopicSubscription struct {
	ctx context.Context

	// Messages is a channel of messages received from other peers in the chat room
	messages  chan []byte
	topic     *pubsub.Topic
	sub       *pubsub.Subscription
	handlerFn MessageHandler
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
func NewTopicSubscription(
	ctx context.Context,
	topic *pubsub.Topic,
	sub pubsub.Subscription,
	msgHandlerFn MessageHandler) *TopicSubscription {
	return &TopicSubscription{
		ctx:       ctx,
		topic:     topic,
		sub:       &sub,
		messages:  make(chan []byte),
		handlerFn: msgHandlerFn,
	}
}

// MessageReadingLoop:
// Pulls messages from the pubsub topic and pushes them onto the Messages channel
// and the underlaying msg metrics.
// @param h: libp2p host.
// @param peerstore: peerstore of the crawler app.
func (c *TopicSubscription) MessageReadingLoop(h host.Host, dbClient *psql.DBClient) {
	log.Infof("topic subscription %s reading loop", c.sub.Topic())
	subsCtx := c.ctx
	for {
		msg, err := c.sub.Next(subsCtx)
		if err != nil {
			if err == subsCtx.Err() {
				log.Errorf("context of the subsciption %s has been canceled", c.sub.Topic())
				break
			}
			log.Errorf("error reading next message in topic %s. %slol", c.sub.Topic(), err.Error())
		} else {
			// To avoid getting track of our own messages, check if we are the senders
			if msg.ReceivedFrom != h.ID() {
				log.Infof("new message on %s from %s", c.sub.Topic(), msg.ReceivedFrom)
				// use the msg handler for that specific topic that we have
				content, err := c.handlerFn(msg)
				if err != nil {
					log.Error(errors.Wrap(err, "unable to unwrap message"))
				}
				if !content.IsZero() {
					log.Infof("successfully tracked message: %+v", content)
				}
			} else {
				log.Debugf("message sent by ourselfs received on %s", c.sub.Topic())
			}
		}
	}
	<-subsCtx.Done()
	log.Debugf("ending %s reading loop", c.sub.Topic())
}
