package exporters

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	NewExporterCheckInterval = 2 * time.Minute
	ModuleName               = "EXPORTER"
	log                      = logrus.WithField(
		"module", ModuleName,
	)
)

type ExporterService struct {
	// TODO: adding new Exporters to the Service might bring up race conditions
	// control variables
	ctx    context.Context
	cancel context.CancelFunc

	PrometheusRunner PrometheusRunner

	ExporterRoutines map[string]Exporter
	// TODO: Check if we need anything else
}

// Basic unit of metrics export
type Exporter interface {
	Name() string
	Init()
	Run()
	Close()
	Status() string
	Details() string
}

// Creates and initialized the exporter service
// where any module can create a dedicated exporter
func NewExporterService(ctx context.Context) *ExporterService {
	mainctx, cancel := context.WithCancel(ctx)

	exporterSrv := &ExporterService{
		ctx:              mainctx,
		cancel:           cancel,
		PrometheusRunner: NewPrometheusRunner(),
		ExporterRoutines: make(map[string]Exporter),
	}
	return exporterSrv
}

func (s *ExporterService) Run() {
	log.Info("running the exporter service")
	// start the prometheus endpoint
	s.PrometheusRunner.Start()
	// run the exporters check in a go routine (until shutdown)
	ticker := time.NewTicker(NewExporterCheckInterval)
	go func() {
		// iteration loop
		for {
			select {
			case <-s.ctx.Done():
				log.Info("context of exporter was close, closing exporting service")
				return
			case <-ticker.C:
				for _, export := range s.ExporterRoutines {
					// check if the metrics are up exporting
					// run them otherwise
					status := export.Status()
					// Check if the exporter is in ready state to initialize it
					// and run it
					if status == "ready" {
						// get details of the
						log.Debug(export.Details())
						export.Init()
						export.Run()
						log.Debug(export.Status())
					}
				}
			}
		}
	}()

}

// add new exporter to the list of exporters
func (s *ExporterService) AddNewExporter(exptr Exporter) {
	log.Infof("adding exporter %s to the exporter service", exptr.Name())
	// TODO: - might generate a race condition
	s.ExporterRoutines[exptr.Name()] = exptr
}

// Close all the running metrics exposers
func (s *ExporterService) Close() {
	log.Infof("closing %d metrics exporters", len(s.ExporterRoutines))
	for name, export := range s.ExporterRoutines {
		log.Infof("closing exporter %s", name)
		export.Close()
	}
}
