package gossipsub

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
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
)
