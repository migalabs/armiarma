package peering

import (
	"github.com/migalabs/armiarma/pkg/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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
	ConnectionErrorDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "peering",
		Name:      "conn_error_distribution",
		Help:      "The error distribtuion of the attempted to connect peers since the last iteration",
	},
		[]string{"error_type"},
	)
	TotalConnectionErrorDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "peering",
		Name:      "total_conn_error_distribution",
		Help:      "The total error distribtuion the active peers",
	},
		[]string{"error_type"},
	)
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
	metricsMod.AddIndvMetric(p.getConnErrorDistribution())
	metricsMod.AddIndvMetric(p.getTotalConnErrorDistribution())

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
		for key, val := range controlDist {
			PrunedErrorDistribution.WithLabelValues(key).Set(float64(val))
			summary[key] = val
		}
		return summary, nil
	}

	indvMetr, err := metrics.NewIndvMetrics(
		"pruned_error_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(errors.Wrap(err, "unable to init pruned_error_distribution"))
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
		for key, val := range errorAttemptDist {
			summary[key] = val
			ErrorAttemptDistribution.WithLabelValues(key).Set(float64(val))
		}
		return summary, nil
	}

	indvMetr, err := metrics.NewIndvMetrics(
		"iteration_attempts_by_category_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(errors.Wrap(err, "unable to init iteration_attempts_by_category_distribution"))
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
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(errors.Wrap(err, "unable to init peers_attempted_last_iteration"))
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
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(errors.Wrap(err, "unable to init peerstore_iteration_time_secs"))
		return nil
	}

	return indvMetr
}

func (p *PeeringService) getConnErrorDistribution() *metrics.IndvMetrics {

	initFn := func() error {
		prometheus.MustRegister(ConnectionErrorDistribution)
		return nil
	}

	updateFn := func() (interface{}, error) {
		summary := make(map[string]interface{})
		errDist := p.strategy.GetConnErrorDistribution()
		for key, val := range errDist {
			summary[key] = val
			ConnectionErrorDistribution.WithLabelValues(key).Set(float64(val))
		}
		return summary, nil
	}
	IndvMetr, err := metrics.NewIndvMetrics(
		"conn_error_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(errors.Wrap(err, "unable to init conn_error_distribution"))
		return nil
	}

	return IndvMetr
}

func (p *PeeringService) getTotalConnErrorDistribution() *metrics.IndvMetrics {

	initFn := func() error {
		prometheus.MustRegister(TotalConnectionErrorDistribution)
		return nil
	}

	updateFn := func() (interface{}, error) {
		summary := make(map[string]interface{})
		errDist := p.strategy.GetTotalConnErrorDistribution()
		for key, val := range errDist {
			summary[key] = val
			TotalConnectionErrorDistribution.WithLabelValues(key).Set(float64(val))
		}
		return summary, nil
	}
	IndvMetr, err := metrics.NewIndvMetrics(
		"total_conn_error_distribution",
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(errors.Wrap(err, "unable to init total_conn_error_distribution"))
		return nil
	}

	return IndvMetr
}
