package topic

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/golang/snappy"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/gossip"
	"github.com/protolambda/rumor/p2p/gossip/database"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/ztyp/codec"
	"github.com/sirupsen/logrus"
)

type TopicLogCmd struct {
	*base.Base
	GossipMetrics *metrics.GossipMetrics
	GossipState   *metrics.GossipState
	Eth2TopicName string `ask:"--eth-topic" help:"The name of the eth2 topics"`
	ForkDigest    string `ask:"--fork-version" help:"The fork digest value of the network we want to join to (Default Mainnet)"`
	Encoding      string `ask:"--encoding" help:"Encoding that is getting used"`
}

func (c *TopicLogCmd) Default() {
	c.ForkDigest = "b5303f2a" // Mainnet Fork Digest
	c.Encoding = "ssz_snappy"
}

func (c *TopicLogCmd) Help() string {
	return "Log the messages of a gossip topic. Messages are hex-encoded. Join a topic first."
}

func (c *TopicLogCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}
	if c.Eth2TopicName != "" {
		// Genereate the full name of the Eth2 topic
		topicName := gossip.GenerateEth2Topics(c.ForkDigest, c.Eth2TopicName, c.Encoding)
		top, ok := c.GossipState.Topics.Load(topicName)
		if !ok {
			return fmt.Errorf("not on gossip topic %s", topicName)
		} else {
			sub, err := top.(*pubsub.Topic).Subscribe()
			if err != nil {
				return fmt.Errorf("cannot open subscription on topic %s: %v", topicName, err)
			}
			h, err := c.Host()
			if err != nil {
				return fmt.Errorf("cannot get host from base, %v", err)
			}
			ctx, cancelLog := context.WithCancel(ctx)
			go func() {
				defer sub.Cancel()
				for {
					msg, err := sub.Next(ctx)
					if err != nil {
						if err == ctx.Err() { // expected quit, context stopped.
							break
						}
						c.Log.WithError(err).WithField("topic", topicName).Error("Gossip logging encountered error")
						return
					} else {
						var msgData []byte
						if strings.HasSuffix(topicName, "_snappy") {
							msgData, err = snappy.Decode(nil, msg.Data)
							if err != nil {
								c.Log.WithError(err).WithField("topic", topicName).Error("Cannot decompress snappy message")
								continue
							}
						} else {
							msgData = msg.Data
						}

						// To avoid getting track of our own messages, check if we are the senders
						if msg.ReceivedFrom != h.ID() {
							c.Log.WithFields(logrus.Fields{
								"from":      msg.ReceivedFrom.String(),
								"data":      hex.EncodeToString(msgData),
								"signature": hex.EncodeToString(msg.Signature),
								"seq_no":    hex.EncodeToString(msg.Seqno),
							}).Infof("new message on %s", topicName)
							c.GossipMetrics.IncomingMessageManager(msg.ReceivedFrom, topicName)
							// Deserialize the message depending on the topic name
							// generate a new ReceivedMessage on the Temp Database
							// check if the topic has a db asiciated
							if _, ok := c.GossipMetrics.TopicDatabase.TopicDB.Load(topicName); ok {
								err := AddMsgToTopicDB(&c.GossipMetrics.TopicDatabase, msg.ReceivedFrom, topicName, msgData)
								if err != nil {
									c.Log.WithError(err).WithField("topic", topicName).Error("Error saving message on temp database")
								}
							}
						} else {
							c.Log.WithFields(logrus.Fields{
								"from":      msg.ReceivedFrom.String(),
								"data":      hex.EncodeToString(msgData),
								"signature": hex.EncodeToString(msg.Signature),
								"seq_no":    hex.EncodeToString(msg.Seqno),
							}).Infof("message sent by ourselfs received on %s", topicName)
						}

					}
				}
			}()

			c.Control.RegisterStop(func(ctx context.Context) error {
				cancelLog()
				c.Log.Info("Stopped gossip logger")
				return nil
			})
			return nil
		}
	} else {
		return fmt.Errorf("ERROR: No topic was given")
	}
}

//
func AddMsgToTopicDB(topicDB *database.TopicDatabase, from peer.ID, topic string, rawMsg []byte) error {
	if _, ok := topicDB.TopicDB.Load(topic); !ok {
		return fmt.Errorf("not on gossip topic %s", topic)
	} else {
		// Pass from []byte to *bytes.Buffer so that we can Deserialize it and get the real content of the message
		msgBuf := bytes.NewBuffer(rawMsg)
		// classify the topic on the message types that we support
		switch topic {
		case gossip.BeaconBlock: // Hardcoded to Mainnet ForkDigest
			var signedBB beacon.SignedBeaconBlock
			err := signedBB.Deserialize(topicDB.Spec, codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
			if err != nil {
				return err
			}
			msg := database.NewReceivedBeaconBlock(from, signedBB)
			// after generating the msg struct (doesn't matter which message type it was) we save the msg on the corresponding topicDB
			err = topicDB.WriteMessage(msg, topic)
			if err != nil {
				return err
			}
			return nil
		case gossip.BeaconAggregateProof: // Hardcoded to Mainnet ForkDigest
			fmt.Println("New Attestation")
			return nil
		default:
			return fmt.Errorf("Message Struct for this topic was not defined")
		}
	}
}
