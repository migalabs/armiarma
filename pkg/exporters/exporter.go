package exporters

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

var (
	// Assiciated with the Status of the exporter
	ReadyStatus       = "ready"
	InitializedStatus = "initialized"
	RunningStatus     = "running"
	ClosedStatus      = "closed"
)

type MetricsExporter struct {
	// control variables
	ctx    context.Context
	cancel context.CancelFunc
	// TODO: add metrics to the export (time¿?)
	name    string // name the defines the exporter (Example Peer-Prometheus-Exporter)
	status  string
	details string

	interval time.Duration

	initFunc  func() // Initialization of the exporter
	runFunc   func() // function that will be executed in the running loop (the func needs to run a go routine)
	closeFunc func() // might be empty if nothing is required
}

func NewMetricsExporter(ctx context.Context,
	name string,
	details string,
	initf func(),
	runf func(),
	closef func(),
	interval time.Duration) (*MetricsExporter, error) {
	// check if all the necesaty parameters where given
	if len(name) <= 0 {
		return nil, errors.New("no name was provided" + name)
	}
	// Extract context from the given one
	mainctx, cancel := context.WithCancel(ctx)

	metricsExporter := &MetricsExporter{
		ctx:       mainctx,
		cancel:    cancel,
		name:      name,
		status:    ReadyStatus, // Hardcoded to the ready status, Everytime a new metrics is created Run it
		details:   details,
		interval:  interval,
		initFunc:  initf,
		runFunc:   runf,
		closeFunc: closef,
	}
	return metricsExporter, nil
}

func (e MetricsExporter) Name() string {
	return e.name
}

func (e MetricsExporter) Details() string {
	return e.details
}

func (e MetricsExporter) Status() string {
	return e.status
}

func (e *MetricsExporter) Init() {
	// Init loop for each of the Exporters
	log.Infof("initializing exporter %s", e.name)
	e.initFunc()
	e.status = InitializedStatus
}

func (e *MetricsExporter) Run() {
	// Init loop for each of the Exporters
	log.Infof("running exporter %s", e.name)
	e.status = RunningStatus
	// generate the ticker for the periodic metrics export
	ticker := time.NewTicker(e.interval)
	// run the export loop in a go routine (until ctx dies)
	go func() {
		for {
			select {
			case <-ticker.C:
				// execute the routine designed
				e.runFunc()
			case <-e.ctx.Done():
				e.closeFunc()
				return
			}
		}
	}()
}

func (e *MetricsExporter) Close() {
	// Init loop for each of the Exporters
	log.Infof("closing exporter %s", e.name)
	e.cancel()
	e.status = ClosedStatus
}
