package gossipsub

import (
	"github.com/migalabs/armiarma/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	moduleName    = "gossipsub"
	moduleDetails = "All the metrics around the gossipsub host"

	ReceivedTotalMessages = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "total_received_messages_psec",
		Help:      "The number of messages received in the last second",
	})
	ReceivedMessages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "received_messages_psec",
		Help:      "Number of messages received per second on each topic",
	},
		[]string{"topic"},
	)
	PeersPerTopic = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "peers_per_topic",
		Help:      "Number of peers that we have connected per topic",
	},
		[]string{"topic"},
	)
)

func (gs *GossipSub) GetMetrics() *metrics.MetricsModule {

	metricsMod := metrics.NewMetricsModule(
		moduleName,
		moduleDetails,
	)

	metricsMod.AddIndvMetric(gs.peersPerTopic())

	return metricsMod
}

func (gs *GossipSub) peersPerTopic() *metrics.IndvMetrics {

	initFn := func() error {
		prometheus.Register(PeersPerTopic)
		return nil
	}

	updateFn := func() (interface{}, error) {
		summary := gs.PubsubService.TopicsPerPeer()
		for top, value := range summary {
			PeersPerTopic.WithLabelValues(top).Set(float64(value))
		}
		return summary, nil
	}

	peersTop, err := metrics.NewIndvMetrics(
		"peers_per_topic",
		"Number of peers that we have connected per topic",
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(err)
		return nil
	}
	return peersTop
}
