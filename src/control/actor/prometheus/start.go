package prometheus

import (
	"fmt"
	"time"

	//"time"
	"context"
	"net/http"
	"strings"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/metrics/utils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	pgossip "github.com/protolambda/rumor/p2p/gossip"
)

// List of metrics that we are going to export
var (
	// Metrics
	clientDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "observed_client_distribution",
		Help:      "Number of peers from each of the clients observed",
	},
		[]string{"client"},
	)
	connectedPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "connected_peers",
		Help:      "The number of connected peers with the crawler",
	})
	receivedTotalMessages = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "total_received_messages_psec",
		Help:      "The number of messages received in the last second",
	})
	receivedMessages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "crawler",
		Name:      "received_messages_psec",
		Help:      "Number of messages received per second on each topic",
	},
		[]string{"topic"},
	)
	// GossipSub Topics
)

type PrometheusStartCmd struct {
	*base.Base
	GossipMetrics *metrics.GossipMetrics

	ExposePort      string        `ask:"--expose-port" help:"port that will be used to offer the metrics to prometheus"`
	EndpointUrl     string        `ask:"--endpoint-url" help:"url path where the metrics will be offered"`
	RefreshInterval time.Duration `ask:"--refresh-interval" help:"Time duration between metrics updates"`
}

func (c *PrometheusStartCmd) Default() {
	c.ExposePort = "9080"
	c.EndpointUrl = "metrics"
	c.RefreshInterval = 10 * time.Second
}

func (c *PrometheusStartCmd) Run(ctx context.Context, args ...string) error {
	// generate the endpoint where the metrics will be offered for prometheus
	path := "/" + c.EndpointUrl
	port := ":" + c.ExposePort
	fmt.Println("Exposing prometheus metrics at:", path, port)
	http.Handle(path, promhttp.Handler())

	// Register the metrics in the prometheus exporter
	prometheus.MustRegister(clientDistribution)
	prometheus.MustRegister(connectedPeers)
	prometheus.MustRegister(receivedTotalMessages)
	prometheus.MustRegister(receivedMessages)
	// launch the collector go routine
	stopping := make(chan struct{})

	// generate reset channel
	resetChan := make(chan bool, 2)
	// message counters
	beacBlock := 0
	beacAttestation := 0
	totalMsg := 0
	// go routine to keep track of the received messages
	go func() {
		for {
			select {
			case <-c.GossipMetrics.MsgNotChannels[pgossip.BeaconBlock]:
				beacBlock += 1
				totalMsg += 1
			case <-c.GossipMetrics.MsgNotChannels[pgossip.BeaconAggregateProof]:
				beacAttestation += 1
				totalMsg += 1
			case <-resetChan:
				// reset the counters
				beacBlock = 0
				beacAttestation = 0
				totalMsg = 0
			case <-stopping:
				fmt.Println("Stopping the go prometheus go routine")
				return
			}
		}
		fmt.Println("End Message tracker")
	}()

	t := time.Now()
	// go routine to export the metrics to prometheus
	go func() {
		for {
			// variable definitions
			var lig float64 = 0
			var tek float64 = 0
			var nim float64 = 0
			var pry float64 = 0
			var lod float64 = 0
			var unk float64 = 0

			var conPeers float64 = 0

			// read the connected peers from the
			h, hostErr := c.Host()
			if hostErr != nil {
				fmt.Println("No host available")
			}
			peers := h.Network().Peers()
			conPeers = float64(len(peers))
			// iterate through the client types in the metrics
			c.GossipMetrics.GossipMetrics.Range(func(k interface{}, v interface{}) bool {
				p := v.(utils.PeerMetrics)
				if p.MetadataRequest {
					if strings.Contains(strings.ToLower(p.ClientType), "lighthouse") {
						lig += 1
					} else if strings.Contains(strings.ToLower(p.ClientType), "teku") {
						tek += 1
					} else if strings.Contains(strings.ToLower(p.ClientType), "nimbus") {
						nim += 1
					} else if strings.Contains(strings.ToLower(p.ClientType), "prysm") {
						pry += 1
					} else if strings.Contains(strings.ToLower(p.ClientType), "js-libp2p") {
						lod += 1
					} else if strings.Contains(strings.ToLower(p.ClientType), "unknown") {
						unk += 1
					} else {
						unk += 1
					}
				}
				return true
			})
			// get the message counter
			secs := c.RefreshInterval.Seconds()
			bb := float64(beacBlock) / secs
			//fmt.Println("Beacon_blocks", beacBlock, "m/ps", bb)
			ba := float64(beacAttestation) / secs
			//fmt.Println("Beacon_Attestation", beacAttestation, "m/ps", ba)
			tot := float64(totalMsg)

			// Add the metrics to the exporter
			clientDistribution.WithLabelValues("lighthouse").Set(lig)
			clientDistribution.WithLabelValues("teku").Set(tek)
			clientDistribution.WithLabelValues("nimbus").Set(nim)
			clientDistribution.WithLabelValues("prysm").Set(pry)
			clientDistribution.WithLabelValues("lodestar").Set(lod)
			clientDistribution.WithLabelValues("unknown").Set(unk)
			connectedPeers.Set(conPeers)
			receivedMessages.WithLabelValues("beacon_blocks").Set(bb)
			receivedMessages.WithLabelValues("beacon_aggregate_and_proof").Set(ba)
			receivedTotalMessages.Set(tot)
			// reset the counters for the message statistics
			resetChan <- true

			tr := time.Since(t)
			if tr < c.RefreshInterval { // sleep necessary time between iterations
				s := c.RefreshInterval - tr
				time.Sleep(s)
				t = time.Now()
			}
			if stopping == nil {
				return
			}
		}
	}()

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error while generating m")
	}

	c.Control.RegisterStop(func(ctx context.Context) error {
		// stop the message reading go routine
		close(stopping)

		c.Log.Infof("Stoped Prometheus Metrics Exporter")
		return nil
	})
	return nil
}

func (c *PrometheusStartCmd) Help() string {
	return "Start the service to offer the Metrics with prometheus"
}

// Necesary Code to export/offer the metrics to prometheus
