package gossipsub

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/gossipsub/message"
)

type TopicHandler struct {
	*base.Base

	// Messages is a channel of messages received from other peers in the chat room
	Messages chan *message.TopicMessage
	Topic    *pubsub.Topic
	Sub      *pubsub.Subscription
}

func NewTopicHandler(ctx context.Context, i_topic *pubsub.Topic, i_sub pubsub.Subscription, stdOpts base.LogOpts) *TopicHandler {
	localLogger := createTopicLoggerOpts(stdOpts)

	// instance base
	new_base, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(localLogger),
	)

	if err != nil {
		new_base.Log.Errorf(err.Error())
	}

	return &TopicHandler{
		Base:  new_base,
		Topic: i_topic,
		Sub:   &i_sub,
	}

}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (th *TopicHandler) readLoop() {

	for {
		th.Log.Infof("Read Loop: %s", th.Sub.Topic())

		msg, err := th.Sub.Next(th.Ctx())
		if err != nil {
			th.Log.Infof("Exit Read Loop:")
			th.Log.Errorf(err.Error())

			return
		}
		th.Log.Infof(msg.String())

	}
}

func createTopicLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME

	return input_opts
}
