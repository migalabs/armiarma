package hosts

import (
	"github.com/migalabs/armiarma/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	moduleName    = "host"
	moduleDetails = "All the metrics around the libp2p host"

	ConnectedPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "connected_peers",
		Help:      "The number of connected peers to our host",
	})
)

func (bh *BasicLibp2pHost) GetMetrics() *metrics.MetricsModule {

	metricsMod := metrics.NewMetricsModule(
		moduleName,
		moduleDetails,
	)

	metricsMod.AddIndvMetric(bh.connectedPeers())

	return metricsMod
}

func (bh *BasicLibp2pHost) connectedPeers() *metrics.IndvMetrics {

	initFn := func() error {
		prometheus.Register(ConnectedPeers)
		return nil
	}

	updateFn := func() (interface{}, error) {
		peers := bh.host.Network().Peers()
		ConnectedPeers.Set(float64(len(peers)))
		return len(peers), nil
	}

	peersTop, err := metrics.NewIndvMetrics(
		"connected_peers",
		"total_connected_peers",
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(err)
		return nil
	}
	return peersTop
}
