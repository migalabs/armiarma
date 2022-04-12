package exporters

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (

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
	http.Handle("/"+c.EndpointUrl, promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", c.ExposePort), nil))
	}()

	return nil
}
