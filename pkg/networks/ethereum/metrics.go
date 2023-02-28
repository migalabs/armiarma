package ethereum

import (
	"github.com/migalabs/armiarma/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	modName    = "local_node"
	modDetails = "general metrics about the local ethereum node"

	// List of metrics that we are going to export

	localHeadSlot = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "local_head_slot",
		Help:      "Number of the last slot that the crawler saw and that it advertises as his last slot",
	})
)

func (c *LocalEthereumNode) GetMetrics() *metrics.MetricsModule {
	metricsMod := metrics.NewMetricsModule(
		modName,
		modDetails,
	)
	// compose all the metrics
	metricsMod.AddIndvMetric(c.localHeadSlot())
	return metricsMod
}

func (c *LocalEthereumNode) localHeadSlot() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(localHeadSlot)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary := make(map[string]interface{}, 0)
		summary["local_head_slot"] = c.LocalStatus.HeadSlot
		localHeadSlot.Set(float64(c.LocalStatus.HeadSlot))
		return summary, nil
	}
	indvMetr, err := metrics.NewIndvMetrics(
		"local_head_slot",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return indvMetr
}
