package gossipsub

import (
	"context"
	"strings"
	"time"

	"github.com/golang/snappy"
	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/db"
)

// TopicSubscription
// * Sumarizes the control fields necesary to manage and
// * govern over a joined and subscribed topic like
// * message logging or record
// * Serves as a server for a singe topic subscription
type TopicSubscription struct {
	*base.Base

	// Messages is a channel of messages received from other peers in the chat room
	Messages       chan []byte
	Topic          *pubsub.Topic
	Sub            *pubsub.Subscription
	MessageMetrics *MessageMetrics
}

// NewTopicSubscription
// * Sumarizes the control fields necesary to manage and
// * govern over a joined and subscribed topic like
// * @param ctx: parent context of the topic subscription, generaly gossipsub context
// * @param topic: the libp2p.PubSub topic of the joined topic
// * @param sub: the libp2p.PubSub subscription of the subscribed topic
// * @param msgMetrics: underlaying message metrics regarding each of the joined topics
// * @param stdOpts: list of options to generate the base of the topic subscription service
// * @return: pointer to TopicSubscription
func NewTopicSubscription(ctx context.Context, topic *pubsub.Topic, sub pubsub.Subscription, msgMetrics *MessageMetrics, stdOpts base.LogOpts) *TopicSubscription {
	localLogger := createTopicLoggerOpts(stdOpts)
	//localLogger.Level = "debug"

	// instance base
	new_base, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(localLogger),
	)

	if err != nil {
		new_base.Log.Errorf(err.Error())
	}

	return &TopicSubscription{
		Base:           new_base,
		Topic:          topic,
		Sub:            &sub,
		MessageMetrics: msgMetrics,
		Messages:       make(chan []byte, 10),
	}
}

// MessageReadingLoop
// * pulls messages from the pubsub topic and pushes them onto the Messages channel
// * and the underlaying msg metrics
// * @param h: libp2p host
// * @param peerstore: peerstore of the crawler app
func (c *TopicSubscription) MessageReadingLoop(h host.Host, peerstore *db.PeerStore) {
	c.Log.Infof("topic subscription %s reading loop", c.Sub.Topic())
	subsCtx := c.Ctx()
	for {
		msg, err := c.Sub.Next(subsCtx)
		if err != nil {
			if err == subsCtx.Err() {
				c.Log.Errorf("context of the subsciption %s has been canceled", c.Sub.Topic())
				break
			}
			c.Log.Errorf("error reading next message in topic %s. %slol", c.Sub.Topic(), err.Error())
		} else {
			var msgData []byte
			if strings.HasSuffix(c.Sub.Topic(), "_snappy") {
				msgData, err = snappy.Decode(nil, msg.Data)
				if err != nil {
					c.Log.WithError(err).WithField("topic", c.Sub.Topic()).Error("Cannot decompress snappy message")
					continue
				}
			}
			// To avoid getting track of our own messages, check if we are the senders
			if msg.ReceivedFrom != h.ID() {
				c.Log.Debugf("new message on %s from %s", c.Sub.Topic(), msg.ReceivedFrom)
				newPeer := db.NewPeer(msg.ReceivedFrom.String())
				newPeer.MessageEvent(c.Sub.Topic(), time.Now())
				peerstore.StoreOrUpdatePeer(newPeer)
				// peerstore.MessageEvent(msg.ReceivedFrom.String(), c.Sub.Topic())
				// Add notification on the notification channel
				c.Messages <- msgData
				// Add message to msg metrics counter
				_ = c.MessageMetrics.AddMessgeToTopic(c.Sub.Topic())

				// Deserialize the message depending on the topic name
				// generate a new ReceivedMessage on the Temp Database
				// check if the topic has a db asiciated
				/*
					if peerstore.MessageDatabase != nil {
						err = AddMsgToMsgDB(peerstore.MessageDatabase, msgData, msg.ReceivedFrom, msg.ArrivalTime, msg.MessageID, topicName)
						if err != nil {
							c.Log.WithError(err).WithField("topic", topicName).Error("Error saving message on temp database")
						}
					}
				*/
			} else {
				c.Log.Debugf("message sent by ourselfs received on %s", c.Sub.Topic())
			}
		}
	}
	<-subsCtx.Done()
	c.Log.Debugf("ending %s reading loop", c.Sub.Topic())
}

func createTopicLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME

	return input_opts
}
