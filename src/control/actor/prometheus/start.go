package prometheus

import (
	"fmt"
	//"time"
	"context"
	"net/http"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"

	"github.com/prometheus/client_golang/prometheus"
 	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusStartCmd struct {
	*base.Base
	*metrics.GossipMetrics

	ExposePort string `ask:"--expose-port" help:"port that will be used to offer the metrics to prometheus"`
	EndpointUrl string `ask:"--endpoint-url" help:"url path where the metrics will be offered"`
	UserName string `ask:"--user" help:"username for the prometheus server"`
	PassWord string `ask:"--password" help:"password to access the prometheus server with the given user name"`
}

func (c *PrometheusStartCmd) Default() {
	c.ExposePort = "9080"
	c.EndpointUrl = "metrics"
}

func (c *PrometheusStartCmd) Run(ctx context.Context, args ...string) error {
	// generate the endpoint where the metrics will be offered for prometheus
	path := "/" + c.EndpointUrl
	port := ":" + c.ExposePort 
	fmt.Println("Exporting metrics at:", path, port)
	http.Handle(path, promhttp.Handler())

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error while generating m")
	}
	return nil
}

func (c *PrometheusStartCmd) Help() string {
	return "Start the service to offer the Metrics with prometheus"
}

// Necesary Code to export/offer the metrics to prometheus

// List of metrics that we are going to export
var (
	// Metrics
	up = prometheus.NewDesc(
		prometheus.BuildFQName("", "", "crawler_up"),
		"Is the crawler up.",
		nil, nil,
	)
	clientDistribution = NewDesc(prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "crawler_observed_client_distribution",
			Help: "Number of peers from each of the clients observed",
		},
		[]string{"client"},)
	)
)

type Exporter struct {
	mirthEndpoint, mirthUsername, mirthPassword string
}

func NewExporter(mirthEndpoint string, mirthUsername string, mirthPassword string) *Exporter {
	ex := &Exporter{
		mirthEndpoint: mirthEndpoint,
		mirthUsername: mirthUsername,
		mirthPassword: mirthPassword,
	}
	return ex 
}
   
 
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- clientDistribution
}


func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	
}


// Initialization of the prometheus metrics (Registration)
func PrometheusInit() {
	prometheus.MustRegister(clientDistribution)
}