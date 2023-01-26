package gossipsub

import (
	"fmt"
	"sync/atomic"

	"github.com/migalabs/armiarma/pkg/exporters"
	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// MessageMetrics
// fgdgdfgdfgSummarizes all the metrics that could be obtained from the received msgs.
// Right now divided by topic and containing only the local counter between server ticker.
type MessageMetrics struct {
	topicList map[string]*int32
}

// NewMessageMetrics:
// @return intialized MessageMetrics struct
func NewMessageMetrics() MessageMetrics {
	return MessageMetrics{
		topicList: make(map[string]*int32, 0),
	}
}

// NewTopic:
// @param name of the topic
// @return a possitive boolean if the topic was
// already in Metrics, negative one otherwise
func (c *MessageMetrics) NewTopic(topic string) bool {
	var counter int32
	atomic.StoreInt32(&counter, 0)
	_, exists := c.topicList[topic]
	if exists {
		return true
	}
	c.topicList[topic] = &counter
	return false
}

// AddMessgeToTopic:
// @param gossipsub topic name.
// @return curren message counter, or -1 if there was an error (non-existing topic).
func (c *MessageMetrics) AddMessgeToTopic(topic string) int32 {
	v, exists := c.topicList[topic]
	if !exists {
		return int32(-1)
	}
	return atomic.AddInt32(v, 1)
}

// ResetTopic:
// @param gossipsub topic name.
// @return curren message counter, or -1 if there was an error (non-existing topic).
func (c *MessageMetrics) ResetTopic(topic string) int32 {
	v, exists := c.topicList[topic]
	if !exists {
		return int32(-1)
	}
	return atomic.SwapInt32(v, int32(0))
}

// ResetAllTopics:
// Resets all the topic counters to 0.
// @return current message counter, or -1 if there was an error (non-existing topic).
func (c *MessageMetrics) ResetAllTopics() error {
	for k, _ := range c.topicList {
		r := c.ResetTopic(k)
		if r < int32(0) {
			return fmt.Errorf("non existing topic %s in list", k)
		}
	}
	return nil
}

// GetTopicMsgs:
// Obtain the counter of messages from last ticker of given topic.
// @return current message counter, or -1 if there was an error (non-existing topic).
func (c *MessageMetrics) GetTopicMsgs(topic string) int32 {
	v, exists := c.topicList[topic]
	if !exists {
		return int32(-1)
	}
	return atomic.LoadInt32(v)
}

// GetTotalMessages:
// Obtain the total of messages received from last ticker from all the topics.
// @return total message counter, or -1 if there was an error (non-existing topic).
func (c *MessageMetrics) GetTotalMessages() int64 {
	var total int64
	for k, _ := range c.topicList {
		r := c.ResetTopic(k)
		if r < int32(0) {
			continue
		}
		total += int64(r)
	}
	return total
}

// ServePrometheusMetrics:
// This method will generate the metrics from GossipSub msg Metrics
// and serve the values to the local prometheus instance.
func (gs *GossipSub) ServeMetrics() {
	exptr, _ := exporters.NewMetricsExporter(
		gs.ctx,
		"Gossip-Metrics-Prometheus",
		"Expose in Prometheus the gossip metrics of the tools DB",
		gs.initGossipPrometheusMetrics,
		gs.runGossipPrometheusMetrics,
		func() {},
		exporters.MetricLoopInterval,
	)
	gs.ExporterService.AddNewExporter(exptr)
}

func (gs *GossipSub) initGossipPrometheusMetrics() {
	// register variables
	prometheus.MustRegister(ReceivedTotalMessages)
	prometheus.MustRegister(ReceivedMessages)
}

func (gs *GossipSub) runGossipPrometheusMetrics() {
	var totMsg int64
	msgPerMin := make(map[string]float64, 0)
	// get the total of the messages
	for k, _ := range gs.MessageMetrics.topicList {
		r := gs.MessageMetrics.GetTopicMsgs(k)
		if r < int32(0) {
			log.Warnf("Unable to get message count for topic %s", k)
			continue
		}
		msgC := (float64(r) / (exporters.MetricLoopInterval.Seconds())) * 60 // messages per minute
		totMsg += int64(r)
		ReceivedMessages.WithLabelValues(eth.Eth2TopicPretty(k)).Set(msgC)
		msgPerMin[eth.Eth2TopicPretty(k)] = msgC
	}
	// get total of msgs
	tot := (float64(totMsg) / (exporters.MetricLoopInterval.Seconds())) * 60 // messages per minute
	ReceivedTotalMessages.Set(tot)
	// reset the values
	err := gs.MessageMetrics.ResetAllTopics()
	if err != nil {
		log.Warnf("Unable to reset the gossip topic metrics. ", err.Error())
	}
	log.WithFields(log.Fields{
		"TopicMsg/min": msgPerMin,
		"TotalMsg/min": tot,
	}).Info("gossip metrics summary")
}
