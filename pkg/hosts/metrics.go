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
	SupportedProtocols = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "supported_protocols",
		Help:      "The list of supported procols that the crawler's libp2p host supports",
	},
		[]string{"protocol"},
	)
)

func (bh *BasicLibp2pHost) GetMetrics() *metrics.MetricsModule {
	metricsMod := metrics.NewMetricsModule(
		moduleName,
		moduleDetails,
	)
	metricsMod.AddIndvMetric(bh.connectedPeers())
	metricsMod.AddIndvMetric(bh.supportedProtocols())
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
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(err)
		return nil
	}
	return peersTop
}

func (bh *BasicLibp2pHost) supportedProtocols() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.Register(SupportedProtocols)
		return nil
	}
	updateFn := func() (interface{}, error) {
		protocols := bh.host.Mux().Protocols()
		summary := make(map[string]int8)
		SupportedProtocols.Reset()
		for _, prot := range protocols {
			summary[prot] = int8(1)
			SupportedProtocols.WithLabelValues(prot).Set(float64(1))
		}
		return summary, nil
	}
	peersTop, err := metrics.NewIndvMetrics(
		"suported protocols",
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(err)
		return nil
	}
	return peersTop
}
