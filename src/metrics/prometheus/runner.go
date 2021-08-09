package prometheus

import (
	"time"
	"context"
	"net/http"
	"fmt"

	"github.com/protolambda/rumor/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus"

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
	return PrometheusRunner {
		PeerStore: gm,
		ExposePort: "9080",
		EndpointUrl: "metrics",
		RefreshInterval: 10 * time.Second,
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
			geoDist := make(map[string]float64)

			//log.Info("peerstore entero", c.PeerStore.PeerStore)

			c.PeerStore.PeerStore.Range(func(k, val interface{}) bool {
				peerData := val.(metrics.Peer)

				// TODO: Rethink this criteria
				if (peerData.ClientName != "Unknown" && peerData.ClientName != "") {
					clients.AddClientVersion(peerData.ClientName, peerData.ClientVersion)

				}

				// TODO: Expose also the city
				_, ok := geoDist[peerData.Country]
				if ok {
					geoDist[peerData.Country]++
				} else {
					geoDist[peerData.Country] = 1
				}

				//connectedPeers.Set(TODO)
				//receivedTotalMessages.Set(TODO)

				//receivedMessages.WithLabelValues("beacon_blocks").Set(TODO)
				//receivedMessages.WithLabelValues("beacon_aggregate_and_proof").Set(TODO)

				//log.Info("peer in  metris IS", peerData)


				nOfDiscoveredPeers++

				return true
			})

			log.Info("Debug", clients)

			totPeers.Set(float64(nOfDiscoveredPeers))

			log.Info("discovered peers", nOfDiscoveredPeers)

			for _, clientName := range clients.GetClientNames() {
				count := clients.GetCountOfClient(clientName)
				// TODO: Add also version and OS
				clientDistribution.WithLabelValues(clientName).Set(float64(count))
			}

			// Country distribution
			for k, v := range geoDist {
				geoDistribution.WithLabelValues(k).Set(v)
			}

			time.Sleep(c.RefreshInterval)
		}
	}()

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", c.ExposePort), nil))
	}()

	return nil
}
