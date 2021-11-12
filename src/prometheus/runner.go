package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	MetricLoopInterval time.Duration = 15 * time.Second
)

type PrometheusRunner struct {
	ExposePort      string
	EndpointUrl     string
	RefreshInterval time.Duration
}

func NewPrometheusRunner() PrometheusRunner {
	return PrometheusRunner{
		ExposePort:  "9080",
		EndpointUrl: "metrics",
	}
}

func (c *PrometheusRunner) Start(ctx context.Context) error {
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", c.ExposePort), nil))
	}()

	return nil
}
