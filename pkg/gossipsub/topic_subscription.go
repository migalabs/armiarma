package gossipsub

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
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
	psub        *pubsub.PubSub
	messages    chan []byte
	topic       *pubsub.Topic
	sub         *pubsub.Subscription
	handlerFn   MessageHandler
	persistMsgs bool
}

// NewTopicSubscription sumarizes the control fields necesary to manage and
// govern over a joined and subscribed topic.
func NewTopicSubscription(
	ctx context.Context,
	topic *pubsub.Topic,
	sub pubsub.Subscription,
	msgHandlerFn MessageHandler,
	persistMsgs bool) *TopicSubscription {
	return &TopicSubscription{
		ctx:         ctx,
		topic:       topic,
		sub:         &sub,
		messages:    make(chan []byte),
		handlerFn:   msgHandlerFn,
		persistMsgs: persistMsgs,
	}
}

// MessageReadingLoop pulls messages from the pubsub topic and pushes them onto the Messages channel
// and the underlaying msg metrics.
func (c *TopicSubscription) MessageReadingLoop(selfId peer.ID, dbClient database) {
	log.Debugf("topic subscription %s reading loop", c.sub.Topic())
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
			if msg.ReceivedFrom != selfId {
				log.Debugf("new message on %s from %s", c.sub.Topic(), msg.ReceivedFrom)
				// use the msg handler for that specific topic that we have
				content, err := c.handlerFn(msg)
				if err != nil {
					log.Error(errors.Wrap(err, "unable to unwrap message on topic "+c.sub.Topic()))
					continue
				}
				if !content.IsZero() && c.persistMsgs {
					log.Debugf("msg on %s content: %+v", c.sub.Topic(), content)
					dbClient.PersistToDB(content)
				}
			} else {
				log.Debugf("message sent by ourselfs received on %s", c.sub.Topic())
			}
		}
	}
	<-subsCtx.Done()
	log.Debugf("ending %s reading loop", c.sub.Topic())
}
