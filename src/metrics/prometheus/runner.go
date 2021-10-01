package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/protolambda/rumor/metrics"

	//"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type PrometheusRunner struct {
	PeerStore *metrics.PeerStore

	ExposePort      string
	EndpointUrl     string
	RefreshInterval time.Duration
}

func NewPrometheusRunner(gm *metrics.PeerStore) PrometheusRunner {
	return PrometheusRunner{
		PeerStore:       gm,
		ExposePort:      "9080",
		EndpointUrl:     "metrics",
		RefreshInterval: 30 * time.Second,
	}
}

func (c *PrometheusRunner) Run(ctx context.Context) error {
	http.Handle("/metrics", promhttp.Handler())

	prometheus.MustRegister(clientDistribution)
	prometheus.MustRegister(connectedPeers)
	prometheus.MustRegister(receivedTotalMessages)
	prometheus.MustRegister(receivedMessages)
	prometheus.MustRegister(totPeers)
	prometheus.MustRegister(geoDistribution)

	go func() {
		for {
			clients := metrics.NewClients()

			// TODO: Use the Gossip Metrics to populate the metrics
			nOfDiscoveredPeers := 0
			nOfConnectedPeers := 0
			geoDist := make(map[string]float64)

			c.PeerStore.PeerStore.Range(func(key string, value metrics.Peer) bool {
				if value.ClientName != "" {
					clients.AddClientVersion(value.ClientName, value.ClientVersion)
				}

				if value.IsConnected {
					nOfConnectedPeers++
				}

				// TODO: Expose also the city
				_, ok := geoDist[value.Country]
				if ok {
					geoDist[value.Country]++
				} else {
					geoDist[value.Country] = 1
				}

				nOfDiscoveredPeers++

				return true
			})

			totPeers.Set(float64(nOfDiscoveredPeers))
			connectedPeers.Set(float64(nOfConnectedPeers))
			//receivedTotalMessages.Set(TODO)

			//receivedMessages.WithLabelValues("beacon_blocks").Set(TODO)
			//receivedMessages.WithLabelValues("beacon_aggregate_and_proof").Set(TODO)

			for _, clientName := range clients.GetClientNames() {
				count := clients.GetCountOfClient(clientName)
				// TODO: Add also version and OS
				clientDistribution.WithLabelValues(clientName).Set(float64(count))
			}

			// Country distribution
			for k, v := range geoDist {
				geoDistribution.WithLabelValues(k).Set(v)
			}

			allLastErrors := c.PeerStore.GetErrorCounter()

			log.WithFields(log.Fields{
				"ClientsDist":        clients,
				"GeoDist":            geoDist,
				"NOfDiscoveredPeers": nOfDiscoveredPeers,
				"NOfConnectedPeers":  nOfConnectedPeers,
				"LastErrors":         allLastErrors,
			}).Info("Metrics summary")

			time.Sleep(c.RefreshInterval)
		}
	}()

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", c.ExposePort), nil))
	}()

	return nil
}
