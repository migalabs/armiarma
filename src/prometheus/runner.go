package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/migalabs/armiarma/src/db"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	MetricLoopInterval time.Duration = 15 * time.Second
)

type PrometheusRunner struct {
	PeerStore *db.PeerStore

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

	/*	// launch the collector go routine
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
					select { // TODO: change the constants
					case <-c.PeerStore.MsgNotChannels[gossipsub.BeaconBlock]:
						beacBlock += 1
						totalMsg += 1 // TODO: change the constants
					case <-c.PeerStore.MsgNotChannels[gossipsub.BeaconAggregateProof]:
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
				log.Info("End Message tracker")
			}()


		/*
			// get the message counter per minutes
			secs := c.RefreshInterval.Seconds()
			bb := (float64(beacBlock) / secs) * 60
			ba := (float64(beacAttestation) / secs) * 60
			tot := float64(totalMsg)

			receivedMessages.WithLabelValues("beacon_blocks").Set(bb)
			receivedMessages.WithLabelValues("beacon_aggregate_and_proof").Set(ba)
			receivedTotalMessages.Set(tot)

			resetChan <- true
	*/

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", c.ExposePort), nil))
	}()

	return nil
}
