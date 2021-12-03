package prometheus

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	ModuleName = "PROMETHEUS"
	Log        = logrus.WithField(
		"module", ModuleName,
	)

	// TODO: Just hardcoded, move to config
	ExposedPort string = "9080"
	EndpointUrl string = "metrics"

	MetricLoopInterval time.Duration = 15 * time.Second
)

type PrometheusRunner struct {
	ExposePort      string
	EndpointUrl     string
	RefreshInterval time.Duration
}

func NewPrometheusRunner() PrometheusRunner {
	return PrometheusRunner{
		ExposePort:  ExposedPort,
		EndpointUrl: EndpointUrl,
	}
}

func (c *PrometheusRunner) Start() error {
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		Log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", c.ExposePort), nil))
	}()

	return nil
}
