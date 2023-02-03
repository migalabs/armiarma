package metrics

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type MetricsModule struct {
	// TODO: add metrics to the export (time¿?)
	name    string // name the defines the exporter (Example Peer-Prometheus-Exporter)
	details string

	IndvMetrics []*IndvMetrics
}

func NewMetricsModule(
	name string,
	details string) *MetricsModule {

	module := &MetricsModule{
		name:    name,
		details: details,
	}
	return module
}

//
func (m *MetricsModule) AddIndvMetric(indvMetric *IndvMetrics) error {
	m.IndvMetrics = append(m.IndvMetrics, indvMetric)
	return nil
}

func (m *MetricsModule) Init() error {
	log.Debug("ini")
	for _, metric := range m.IndvMetrics {
		err := metric.Init()
		if err != nil {
			return errors.Wrap(err, "error registering metric "+metric.Name())
		}
	}
	return nil
}

func (m *MetricsModule) UpdateSummary() map[string]interface{} {
	summary := make(map[string]interface{}, 0)
	// iter over indvModules
	for _, metric := range m.IndvMetrics {
		// add metric status to summary
		indvSum, err := metric.UpdateMetrics()
		if err != nil {
			log.Error("unable update metrics for indv metric " + metric.Name())
		}
		summary[metric.Name()] = indvSum
	}
	return summary
}

func (m *MetricsModule) Name() string {
	return m.name
}

func (m *MetricsModule) Details() string {
	return m.details
}

type IndvMetrics struct {
	// TODO: add metrics to the export (time¿?)
	name    string // name the defines the exporter (Example Peer-Prometheus-Exporter)
	details string

	initFn   func() error                // Initialization of the exporter
	updateFn func() (interface{}, error) // function that will be executed in the running loop (the func needs to run a go routine)
}

// NewIndvMetrics
func NewIndvMetrics(
	name string,
	details string,
	initFn func() error,
	updateFn func() (interface{}, error)) (*IndvMetrics, error) {

	// check if all the necesaty parameters where given
	if len(name) <= 0 {
		return nil, errors.New("no name was provided" + name)
	}

	module := &IndvMetrics{
		name:     name,
		details:  details,
		initFn:   initFn,
		updateFn: updateFn,
	}
	return module, nil
}

func (m *IndvMetrics) Init() error {
	// Init loop for each of the Exporters
	log.Infof("initializing exporter %s", m.name)
	return m.initFn()
}

func (m *IndvMetrics) UpdateMetrics() (interface{}, error) {
	// generate the ticker for the periodic metrics export
	return m.updateFn()
}

func (m *IndvMetrics) Name() string {
	return m.name
}

func (m *IndvMetrics) Details() string {
	return m.details
}
