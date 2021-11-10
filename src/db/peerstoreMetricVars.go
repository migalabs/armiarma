package db

import (
	"github.com/prometheus/client_golang/prometheus"
)

// List of metrics that we are going to export
var (
	ClientDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_client_distribution",
		Help:      "Number of peers from each of the clients observed",
	},
		[]string{"client"},
	)
	GeoDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "geographical_distribution",
		Help:      "Number of peers from each of the crawled countries",
	},
		[]string{"country"},
	)
	TotPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "total_crawled_peers",
		Help:      "The number of discovered peers with the crawler",
	})
	ConnectedPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "connected_peers",
		Help:      "The number of connected peers with the crawler",
	})
	DeprecatedPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "deprecated_peers",
		Help:      "The number of peers deprecated by the crawler",
	})
	PeerstoreIterTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "peerstore_iteration_time_mins",
		Help:      "The time that the crawler takes to connect the entire peerstore in mins",
	})
	ClientVersionDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_client_version_distribution",
		Help:      "Number of peers from each of the clients versions observed",
	},
		[]string{"client_version"},
	)
	IpDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_ip_distribution",
		Help:      "Number of Ips hosting number of nodes",
	},
		[]string{"numberips"},
	)
	ReceivedTotalMessages = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "total_received_messages_psec",
		Help:      "The number of messages received in the last second",
	})
	ReceivedMessages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "received_messages_psec",
		Help:      "Number of messages received per second on each topic",
	},
		[]string{"topic"},
	)
	RttDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_rtt_distribution",
		Help:      "RTT distribution for the active discovered peers",
	},
		[]string{"secs"},
	)
	TotcontimeDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_total_connected_time_distribution",
		Help:      "Distribution of the connected time for each active discovered peer",
	},
		[]string{"mins"},
	)
)
