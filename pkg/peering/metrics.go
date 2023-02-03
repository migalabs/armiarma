package peering

import (
	"github.com/migalabs/armiarma/pkg/metrics"
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
)

// ServeMetrics:
// This method will serve the global peerstore values to the
// local prometheus instance.
func (p *PeeringService) GetMetrics() *metrics.MetricsModule {
	// register variables

	metricsMod := metrics.NewMetricsModule(
		"peering",
		"internal module of the crawler used for peering and prunning peers",
	)

	metricsMod.AddIndvMetric(p.getPrunedErrorDistribtuion())
	metricsMod.AddIndvMetric(p.getErrorAttemptDistribtuion())
	metricsMod.AddIndvMetric(p.getAttemptedPeersInLastIteration())
	metricsMod.AddIndvMetric(p.getPeerstoreIterTime())

	return metricsMod

}

func (p *PeeringService) getPrunedErrorDistribtuion() *metrics.IndvMetrics {

	initFn := func() error {
		prometheus.MustRegister(PrunedErrorDistribution)
		return nil
	}

	updateFn := func() (interface{}, error) {
		summary := make(map[string]interface{}, 0)
		controlDist := p.strategy.ControlDistribution()
		controlDist.Range(func(key, value interface{}) bool {
			summary[key.(string)] = value
			PrunedErrorDistribution.WithLabelValues(key.(string)).Set(float64(value.(int)))
			return true
		})
		return summary, nil
	}

	indvMetr, err := metrics.NewIndvMetrics(
		"pruned_error_distribution",
		"Filter peers in Peer Queue by errors that were tracked",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}

	return indvMetr
}

func (p *PeeringService) getErrorAttemptDistribtuion() *metrics.IndvMetrics {

	initFn := func() error {
		prometheus.MustRegister(ErrorAttemptDistribution)
		return nil
	}

	updateFn := func() (interface{}, error) {
		summary := make(map[string]interface{})
		errorAttemptDist := p.strategy.GetErrorAttemptDistribution()
		errorAttemptDist.Range(func(key, value interface{}) bool {
			summary[key.(string)] = value
			ErrorAttemptDistribution.WithLabelValues(key.(string)).Set(float64(value.(int)))
			return true
		})
		return summary, nil
	}

	indvMetr, err := metrics.NewIndvMetrics(
		"iteration_attempts_by_category_distribution",
		"Filter attempts in Peer Queue by errors that were tracked",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}

	return indvMetr
}

func (p *PeeringService) getAttemptedPeersInLastIteration() *metrics.IndvMetrics {

	initFn := func() error {
		prometheus.MustRegister(PeersAttemptedInLastIteration)
		return nil
	}

	updateFn := func() (interface{}, error) {
		peersAttemtpted := p.strategy.AttemptedPeersSinceLastIter()
		PeersAttemptedInLastIteration.Set(float64(peersAttemtpted))
		return peersAttemtpted, nil
	}

	indvMetr, err := metrics.NewIndvMetrics(
		"peers_attempted_last_iteration",
		"The number of discovered peers with the crawler",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}

	return indvMetr
}

func (p *PeeringService) getPeerstoreIterTime() *metrics.IndvMetrics {

	initFn := func() error {
		prometheus.MustRegister(PeerstoreIterTime)
		return nil
	}

	updateFn := func() (interface{}, error) {
		iterTime := p.strategy.LastIterTime()
		PeerstoreIterTime.Set(float64(iterTime))
		return iterTime, nil
	}

	indvMetr, err := metrics.NewIndvMetrics(
		"peerstore_iteration_time_secs",
		"Time that the crawler took to connect the entire peerstore in secs",
		initFn,
		updateFn,
	)
	if err != nil {
		return nil
	}

	return indvMetr
}
