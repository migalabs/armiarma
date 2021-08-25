package topic

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

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
	PeerStore *metrics.PeerStore
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
					//fmt.Println("New Message, ID:", msg.MessageID)
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
						}
						// To avoid getting track of our own messages, check if we are the senders
						if msg.ReceivedFrom != h.ID() {
							c.Log.WithFields(logrus.Fields{
								"from":      msg.ReceivedFrom.String(),
								"data":      hex.EncodeToString(msg.Data),
								"signature": hex.EncodeToString(msg.Signature),
								"seq_no":    hex.EncodeToString(msg.Seqno),
							}).Infof("new message on %s", topicName)
							c.PeerStore.AddMessageEvent(msg.ReceivedFrom.String(), topicName)
							// Add notification on the notification channel
							c.PeerStore.MsgNotChannels[topicName] <- true
							// Deserialize the message depending on the topic name
							// generate a new ReceivedMessage on the Temp Database
							// check if the topic has a db asiciated
							if c.PeerStore.MessageDatabase != nil {
								err = AddMsgToMsgDB(c.PeerStore.MessageDatabase, msgData, msg.ReceivedFrom, msg.ArrivalTime, msg.MessageID, topicName)
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
func AddMsgToMsgDB(msgDB *database.MessageDatabase, rawMsg []byte, sender peer.ID, arrTime time.Time, msgID string, topic string) error {
	// Pass from []byte to *bytes.Buffer so that we can Deserialize it and get the real content of the message
	msgBuf := bytes.NewBuffer(rawMsg)
	// classify the topic on the message types that we support
	switch topic {
	case gossip.BeaconBlock: // Hardcoded to Mainnet ForkDigest
		var signedBB beacon.SignedBeaconBlock
		err := signedBB.Deserialize(msgDB.Spec, codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
		if err != nil {
			return err
		}
		rm := &database.ReceivedMessage{
			MessageID:      msgID,
			MessageType:    "beacon-block",
			Slot:           signedBB.Message.Slot,
			ValidatorIndex: signedBB.Message.ProposerIndex,
			Sender:         sender,
			ArrivalTime:    arrTime,
			Content:        signedBB,
		}
		// after generating the msg struct (doesn't matter which message type it was) we save the msg on the msg DB
		seen, err := msgDB.AddMessage(rm)
		if err != nil {
			return err
		}
		// notify of a new block arrival if we didn't see the msg before
		if !seen {
			msgDB.BlockNotChan <- &signedBB
		}
		return nil

	case gossip.BeaconAggregateProof: // Hardcoded to Mainnet ForkDigest
		var signedBAggr beacon.SignedAggregateAndProof
		err := signedBAggr.Deserialize(msgDB.Spec, codec.NewDecodingReader(msgBuf, uint64(len(msgBuf.Bytes()))))
		if err != nil {
			return err
		}
		rm := &database.ReceivedMessage{
			MessageID:      msgID,
			MessageType:    "beacon-aggregation",
			Slot:           signedBAggr.Message.Aggregate.Data.Slot,
			ValidatorIndex: signedBAggr.Message.AggregatorIndex,
			Sender:         sender,
			ArrivalTime:    arrTime,
			Content:        signedBAggr,
		}
		// after generating the msg struct (doesn't matter which message type it was) we save the msg on the msg DB
		_, err = msgDB.AddMessage(rm)
		if err != nil {
			return err
		}
		// notify of a new block arrival if we didn't see the msg before
		// currently not needed for the attestations
		return nil
	default:
		return fmt.Errorf("Message Struct for this topic was not defined")
	}
}
