package gossipsub

import (
	"context"
	"strings"

	"github.com/golang/snappy"
	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/db"
)

type TopicSubscription struct {
	*base.Base

	// Messages is a channel of messages received from other peers in the chat room
	Messages chan []byte
	Topic    *pubsub.Topic
	Sub      *pubsub.Subscription
}

func NewTopicSubscription(ctx context.Context, topic *pubsub.Topic, sub pubsub.Subscription, stdOpts base.LogOpts) *TopicSubscription {
	localLogger := createTopicLoggerOpts(stdOpts)

	// instance base
	new_base, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(localLogger),
	)

	if err != nil {
		new_base.Log.Errorf(err.Error())
	}

	return &TopicSubscription{
		Base:     new_base,
		Topic:    topic,
		Sub:      &sub,
		Messages: make(chan []byte, 10),
	}
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
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
				peerstore.MessageEvent(msg.ReceivedFrom.String(), c.Sub.Topic())
				// Add notification on the notification channel
				c.Messages <- msgData
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
