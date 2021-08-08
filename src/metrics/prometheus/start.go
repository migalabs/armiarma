package prometheus

import (
	"fmt"
	"time"
	"context"
	"net/http"

	"github.com/protolambda/rumor/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type PrometheusRunner struct {
	GossipMetrics *metrics.GossipMetrics

	ExposePort      string
	EndpointUrl     string
	RefreshInterval time.Duration
}

func NewPrometheusRunner(gm *metrics.GossipMetrics) PrometheusRunner {
	return PrometheusRunner {
		GossipMetrics: gm,
		ExposePort: "9080",
		EndpointUrl: "metrics",
		RefreshInterval: 10 * time.Second,
	}
}

func (c *PrometheusRunner) Run(ctx context.Context) error {
	go func() {
		log.Info("Exposing prometheus metrics at: ", c.ExposePort)
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", c.ExposePort), nil)
	}()

	// Register the metrics in the prometheus exporter
	prometheus.MustRegister(clientDistribution)
	prometheus.MustRegister(connectedPeers)
	prometheus.MustRegister(receivedTotalMessages)
	prometheus.MustRegister(receivedMessages)
	prometheus.MustRegister(totPeers)
	prometheus.MustRegister(geoDistribution)

	log.Info("debug, done registering")

	go func() {
		for {

			clients := metrics.NewClients()

			// TODO: Use the Gossip Metrics to populate the metrics
			c.GossipMetrics.GossipMetrics.Range(func(k, val interface{}) bool {
				peerData := val.(metrics.Peer)
				//log.Info(v.ToCsvLine())

				// TODO: Rethink this criteria
				if (peerData.ClientName != "Unknown" && peerData.ClientName != "") {
					clients.AddClientVersion(peerData.ClientName, peerData.ClientVersion)
				}

				for _, clientName := range clients.GetClientNames() {
					count := clients.GetCountOfClient(clientName)
					log.Info("found clients: ", clientName, " ", count)
					clientDistribution.WithLabelValues(clientName).Set(float64(count))
				}

				log.Info(clients)


				return true
			})

			/* TODO: All the data to populate the metrics is here in GossipMetrics
			clientDistribution.WithLabelValues("xxx").Set(lig)
			clientDistribution.WithLabelValues("yyy").Set(tek)
			clientDistribution.WithLabelValues("nimbus").Set(nim)
			connectedPeers.Set(conPeers)
			receivedTotalMessages.Set(tot)
			receivedMessages.WithLabelValues("beacon_blocks").Set(bb)
			receivedMessages.WithLabelValues("beacon_aggregate_and_proof").Set(ba)

			*/

			//totPeers.Set(totdisc)

			/*
			for k, v := range geoDist {
				geoDistribution.WithLabelValues(k).Set(v)
			}*/

			time.Sleep(c.RefreshInterval)
		}
	}()

	// launch the collector go routine
	//stopping := make(chan struct{})

	// generate reset channel
	//resetChan := make(chan bool, 2)
	// message counters
	//beacBlock := 0
	//beacAttestation := 0
	//totalMsg := 0
	// go routine to keep track of the received messages
	/*
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
	}()*/
/*
	go func() {
		for {
			fmt.Println("TODO: exporting metrics to prometheus")
			fmt.Println("hej", c.GossipMetrics.GossipMetrics)
			fmt.Println("addres gosmet", &c.GossipMetrics)
			time.Sleep(c.RefreshInterval)
		}
	}()
*/
	//t := time.Now()
	// go routine to export the metrics to prometheus
	/*
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
			var totdisc float64 = 0

			// read the connected peers from the
			h, hostErr := c.Host()
			if hostErr != nil {
				fmt.Println("No host available")
			}
			peers := h.Network().Peers()
			conPeers = float64(len(peers))
			geoDist := make(map[string]float64)
			// iterate through the client types in the metrics
			c.GossipMetrics.GossipMetrics.Range(func(k interface{}, v interface{}) bool {
				totdisc += 1
				p := v.(metrics.Peer)
				if p.MetadataRequest {
					if strings.Contains(strings.ToLower(p.GetClientType()), "lighthouse") {
						lig += 1
					} else if strings.Contains(strings.ToLower(p.GetClientType()), "teku") {
						tek += 1
					} else if strings.Contains(strings.ToLower(p.GetClientType()), "nimbus") {
						nim += 1
					} else if strings.Contains(strings.ToLower(p.GetClientType()), "prysm") {
						pry += 1
					} else if strings.Contains(strings.ToLower(p.GetClientType()), "js-libp2p") {
						lod += 1
					} else if strings.Contains(strings.ToLower(p.GetClientType()), "unknown") {
						unk += 1
					} else {
						unk += 1
					}
				}
				_, ok := geoDist[p.Country]
				if ok {
					geoDist[p.Country] += 1
				} else {
					geoDist[p.Country] = 1
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
			// Country distribution
			for k, v := range geoDist {
				geoDistribution.WithLabelValues(k).Set(v)
			}
			connectedPeers.Set(conPeers)
			receivedMessages.WithLabelValues("beacon_blocks").Set(bb)
			receivedMessages.WithLabelValues("beacon_aggregate_and_proof").Set(ba)
			receivedTotalMessages.Set(tot)
			totPeers.Set(totdisc)
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
	}()*/

	return nil
}
