package peering

import (
	"github.com/prometheus/client_golang/prometheus"
)

// List of metrics that we are going to export
var (
	PrunedErrorDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "peering",
		Name:      "pruned_error_distribution",
		Help:      "Filter peers in Peer Queue by errors that were tracked",
	},
		[]string{"controldist"},
	)
	ErrorAttemptDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "peering",
		Name:      "iteration_attempts_by_category_distribution",
		Help:      "Filter attempts in Peer Queue by errors that were tracked",
	},
		[]string{"controlAttemptdist"},
	)
	PeersAttemptedInLastIteration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "peering",
		Name:      "peers_attempted_last_iteration",
		Help:      "The number of discovered peers with the crawler",
	})
	PeerstoreIterTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "peering",
		Name:      "peerstore_iteration_time_secs",
		Help:      "The time that the crawler takes to connect the entire peerstore in secs",
	})
	IterForcingNextConnTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "peering",
		Name:      "iteration_forcing_next_conn_time",
		Help:      "The time reported by the peer that forced the new peerstore iteration",
	})
)
