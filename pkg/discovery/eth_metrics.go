package discovery

import (
	"github.com/migalabs/armiarma/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	modName    = "discovery"
	modDetails = "general metrics about the crawler"

	// List of metrics that we are going to export
	NodesPerForkDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "node_per_fork_distribution",
		Help:      "Number of non-deprecated nodes per fork in the Ethereum network",
	},
		[]string{"fork"},
	)
	// List of metrics that we are going to export
	AttnetsDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "attnets_distribution",
		Help:      "Distribution of number of nodes subscribed to number of subnets",
	},
		[]string{"att_number"},
	)
)

func (d *Discovery) GetEthereumMetrics() *metrics.MetricsModule {

	metricsMod := metrics.NewMetricsModule(
		modName,
		modDetails,
	)

	// compose all the metrics
	metricsMod.AddIndvMetric(d.nodesPerForkMetrics())
	metricsMod.AddIndvMetric(d.AttnetsDistMetrics())

	return metricsMod
}

func (d *Discovery) nodesPerForkMetrics() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(NodesPerForkDistribution)
		return nil
	}

	updateFn := func() (interface{}, error) {
		summary, err := d.DBClient.GetNodePerForkDistribution()
		if err != nil {
			return nil, err
		}
		for forkName, cnt := range summary {
			NodesPerForkDistribution.WithLabelValues(forkName).Set(float64(cnt.(int)))
		}
		return summary, nil
	}

	nodeDist, err := metrics.NewIndvMetrics(
		"node_per_fork_distribution",
		"Number of non-deprecated nodes per fork in the Ethereum network",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return nodeDist
}

func (c *Discovery) AttnetsDistMetrics() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(AttnetsDistribution)
		return nil
	}

	updateFn := func() (interface{}, error) {
		summary, err := c.DBClient.GetAttnetsDistribution()
		if err != nil {
			return nil, err
		}
		for attnets, cnt := range summary {
			AttnetsDistribution.WithLabelValues(attnets).Set(float64(cnt.(int)))
		}
		return summary, nil
	}

	nodeDist, err := metrics.NewIndvMetrics(
		"attnets_distribution",
		"Distribution of number of nodes subscribed to number of subnets",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return nodeDist
}
