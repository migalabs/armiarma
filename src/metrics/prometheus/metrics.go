package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// List of metrics that we are going to export
var (
	clientDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_client_distribution",
		Help:      "Number of peers from each of the clients observed",
	},
		[]string{"client"},
	)
	geoDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "gographical_distribution",
		Help:      "Number of peers from each of the crawled countries",
	},
		[]string{"country"},
	)
	totPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "total_crawled_peers",
		Help:      "The number of discovered peers with the crawler",
	})
	connectedPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "connected_peers",
		Help:      "The number of connected peers with the crawler",
	})
	receivedTotalMessages = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "total_received_messages_psec",
		Help:      "The number of messages received in the last second",
	})
	receivedMessages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "received_messages_psec",
		Help:      "Number of messages received per second on each topic",
	},
		[]string{"topic"},
	)
	// GossipSub Topics
)
