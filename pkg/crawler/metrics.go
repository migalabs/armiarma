package crawler

import (
	"fmt"

	"github.com/migalabs/armiarma/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/pkg/errors"
)

var (
	modName    = "crawler"
	modDetails = "general metrics about the crawler"

	// List of metrics that we are going to export
	ClientDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "observed_client_distribution",
		Help:      "Number of peers from each of the clients observed",
	},
		[]string{"client"},
	)
	VersionDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "observed_client_version_distribution",
		Help:      "Number of peers from each of the clients versions observed",
	},
		[]string{"client_version"},
	)
	GeoDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "geographical_distribution",
		Help:      "Number of peers from each of the crawled countries",
	},
		[]string{"country"},
	)
	NodeDistribution = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "node_distribution",
		Help:      "Number of peers from each of the crawled countries",
	})
	DeprecatedCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "deprecated_nodes",
		Help:      "Total number of deprecated peers",
	})
	OsDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "os_distribution",
		Help:      "Distribution of OS used by the connected peers",
	},
		[]string{"os"},
	)
	ArchDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "arch_distribution",
		Help:      "Architecture distribution of the active peers in the network",
	},
		[]string{"arch"},
	)
	HostedPeers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "hosted_peers_distribution",
		Help:      "Distribution of nodes that are hosted on non-residential networks",
	},
		[]string{"ip_host"},
	)
	RttDist = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "observed_rtt_distribution",
		Help:      "Distribution of RTT between our tool and nodes in the network",
	},
		[]string{"rtt_range"},
	)
	IPDist = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: modName,
		Name:      "observed_ip_distribution",
		Help:      "Distribution of IPs hosting nodes in the network",
	},
		[]string{"repeated_ips"},
	)
)

func (c *EthereumCrawler) GetMetrics() *metrics.MetricsModule {
	metricsMod := metrics.NewMetricsModule(
		modName,
		modDetails,
	)
	// compose all the metrics
	metricsMod.AddIndvMetric(c.clientDistributionMetrics())
	metricsMod.AddIndvMetric(c.versionDistributionMetrics())
	metricsMod.AddIndvMetric(c.geoDistributionMetrics())
	metricsMod.AddIndvMetric(c.nodeDistributionMetrics())
	metricsMod.AddIndvMetric(c.deprecatedNodeMetrics())
	metricsMod.AddIndvMetric(c.getPeersOs())
	metricsMod.AddIndvMetric(c.getPeersArch())
	metricsMod.AddIndvMetric(c.getHostedPeers())
	metricsMod.AddIndvMetric(c.getRTTDist())
	metricsMod.AddIndvMetric(c.getIPDist())

	return metricsMod
}

func (c *EthereumCrawler) clientDistributionMetrics() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(ClientDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary, err := c.DB.GetClientDistribution()
		if err != nil {
			return nil, err
		}
		for cliName, cnt := range summary {
			ClientDistribution.WithLabelValues(cliName).Set(float64(cnt.(int)))
		}
		return summary, nil
	}
	cliDist, err := metrics.NewIndvMetrics(
		"client_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return cliDist
}

func (c *EthereumCrawler) versionDistributionMetrics() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(VersionDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary, err := c.DB.GetVersionDistribution()
		if err != nil {
			return nil, err
		}
		for cliVer, cnt := range summary {
			VersionDistribution.WithLabelValues(cliVer).Set(float64(cnt.(int)))
		}
		return summary, nil
	}
	versDist, err := metrics.NewIndvMetrics(
		"client_version_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return versDist
}

func (c *EthereumCrawler) geoDistributionMetrics() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(GeoDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary, err := c.DB.GetGeoDistribution()
		if err != nil {
			fmt.Println(errors.Wrap(err, "unable to get GeoDist"))
			return nil, err
		}
		for country, cnt := range summary {
			GeoDistribution.WithLabelValues(country).Set(float64(cnt.(int)))
		}
		return summary, nil
	}
	versDist, err := metrics.NewIndvMetrics(
		"geographical_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return versDist
}

func (c *EthereumCrawler) nodeDistributionMetrics() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(NodeDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		peerLs, err := c.DB.GetNonDeprecatedPeers()
		if err != nil {
			return nil, err
		}

		NodeDistribution.Set(float64(len(peerLs)))

		return len(peerLs), nil
	}
	nodeDist, err := metrics.NewIndvMetrics(
		"geographical_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return nodeDist
}

func (c *EthereumCrawler) deprecatedNodeMetrics() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(DeprecatedCount)
		return nil
	}
	updateFn := func() (interface{}, error) {
		nodeCnt, err := c.DB.GetDeprecatedNodes()
		if err != nil {
			return nil, err
		}
		DeprecatedCount.Set(float64(nodeCnt))
		return nodeCnt, nil
	}
	depNodes, err := metrics.NewIndvMetrics(
		"deprecated_nodes",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return depNodes
}

func (c *EthereumCrawler) getPeersOs() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(OsDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		osDist, err := c.DB.GetOsDistribution()
		if err != nil {
			return nil, err
		}
		for key, val := range osDist {
			OsDistribution.WithLabelValues(key).Set(float64(val.(int)))
		}
		return osDist, nil
	}
	osMetr, err := metrics.NewIndvMetrics(
		"os_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return osMetr
}

func (c *EthereumCrawler) getPeersArch() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(ArchDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		archDist, err := c.DB.GetArchDistribution()
		if err != nil {
			return nil, err
		}
		for key, val := range archDist {
			ArchDistribution.WithLabelValues(key).Set(float64(val.(int)))
		}
		return archDist, nil
	}
	archMetr, err := metrics.NewIndvMetrics(
		"arch_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return archMetr
}

func (c *EthereumCrawler) getHostedPeers() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(HostedPeers)
		return nil
	}
	updateFn := func() (interface{}, error) {
		ipSummary, err := c.DB.GetHostingDistribution()
		if err != nil {
			return nil, err
		}
		for key, val := range ipSummary {
			HostedPeers.WithLabelValues(key).Set(float64(val.(int)))
		}
		return ipSummary, nil
	}
	ipHosting, err := metrics.NewIndvMetrics(
		"hosted_peer_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return ipHosting
}


func (c *EthereumCrawler) getRTTDist() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(RttDist)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary, err := c.DB.GetRTTDistribution()
		if err != nil {
			return nil, err
		}
		for key, val := range summary {
			RttDist.WithLabelValues(key).Set(float64(val.(int)))
		}
		return summary, nil
	}
	indvMetric, err := metrics.NewIndvMetrics(
		"rtt_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return indvMetric
}

func (c *EthereumCrawler) getIPDist() *metrics.IndvMetrics {
	initFn := func() error {
		prometheus.MustRegister(IPDist)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary, err := c.DB.GetIPDistribution()
		if err != nil {
			return nil, err
		}
		for key, val := range summary {
			IPDist.WithLabelValues(key).Set(float64(val.(int)))
		}
		return summary, nil
	}
	indvMetric, err := metrics.NewIndvMetrics(
		"ip_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}
	return indvMetric
}
