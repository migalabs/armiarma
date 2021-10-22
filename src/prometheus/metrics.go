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
		Name:      "geographical_distribution",
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
	deprecatedPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "deprecated_peers",
		Help:      "The number of peers deprecated by the crawler",
	})
	peerstoreIterTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "peerstore_iteration_time_mins",
		Help:      "The time that the crawler takes to connect the entire peerstore in mins",
	})
	clientVersionDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_client_version_distribution",
		Help:      "Number of peers from each of the clients versions observed",
	},
		[]string{"client_version"},
	)
	ipDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_ip_distribution",
		Help:      "Number of Ips hosting number of nodes",
	},
		[]string{"numberips"},
	)
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
	rttDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_rtt_distribution",
		Help:      "RTT distribution for the active discovered peers",
	},
		[]string{"secs"},
	)
	totcontimeDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_total_connected_time_distribution",
		Help:      "Distribution of the connected time for each active discovered peer",
	},
		[]string{"mins"},
	)
)
