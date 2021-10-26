package prometheus

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/db/utils"
	"github.com/migalabs/armiarma/src/gossipsub"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const GOSSIP_BEACON_BLOCK string = "32"
const GOSSIP_AGGREGATION_PROOF string = "32"

type PrometheusRunner struct {
	PeerStore *db.PeerStore

	ExposePort      string
	EndpointUrl     string
	RefreshInterval time.Duration
}

func NewPrometheusRunner(gm *db.PeerStore) PrometheusRunner {
	return PrometheusRunner{
		PeerStore:       gm,
		ExposePort:      "9080",
		EndpointUrl:     "metrics",
		RefreshInterval: 15 * time.Second,
	}
}

func (c *PrometheusRunner) Run(ctx context.Context) error {
	http.Handle("/metrics", promhttp.Handler())

	prometheus.MustRegister(clientDistribution)
	prometheus.MustRegister(connectedPeers)
	prometheus.MustRegister(receivedTotalMessages)
	prometheus.MustRegister(receivedMessages)
	prometheus.MustRegister(peerstoreIterTime)
	prometheus.MustRegister(deprecatedPeers)
	prometheus.MustRegister(clientVersionDistribution)
	prometheus.MustRegister(ipDistribution)
	prometheus.MustRegister(totPeers)
	prometheus.MustRegister(geoDistribution)
	prometheus.MustRegister(rttDistribution)
	prometheus.MustRegister(totcontimeDistribution)

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

	go func() {
		for {
			clients := db.NewClients()

			// TODO: Use the Gossip Metrics to populate the metrics
			nOfDiscoveredPeers := 0
			nOfConnectedPeers := 0
			nOfDeprecatedPeers := 0
			geoDist := make(map[string]float64)
			clientVerDist := make(map[string]float64)
			ipDist := make(map[string]float64)
			rttDis := make(map[string]float64)
			tctDis := make(map[string]float64)
			c.PeerStore.PeerStore.Range(func(k string, peerData db.Peer) bool {
				if !peerData.IsDeprecated() {
					if peerData.MetadataRequest {
						if peerData.ClientName != "" {
							clients.AddClientVersion(peerData.ClientName, peerData.ClientVersion)
						}
						if peerData.IsConnected {
							nOfConnectedPeers++
						}
						/*
							// TODO: Expose also the city, swap it for Country code exportage
							_, ok := geoDist[peerData.Country]
							if ok {
								geoDist[peerData.Country]++
							} else {
								geoDist[peerData.Country] = 1
							}
						*/
						// Generate the Country Code distribution
						countrycode := peerData.CountryCode
						if countrycode == "" {
							countrycode = "--"
						}
						_, ok := geoDist[countrycode]
						if ok {
							geoDist[countrycode] += 1
						} else {
							geoDist[countrycode] = 1
						}
						// Client Version Distribution
						client, version := utils.FilterClientType(peerData.UserAgent)
						clientVer := fmt.Sprintf("%v_%v", client, version)
						_, ok = clientVerDist[clientVer]
						if ok {
							clientVerDist[clientVer] += 1
						} else {
							clientVerDist[clientVer] = 1
						}
						// Generate the IP Address distribution
						_, ok = ipDist[peerData.Ip]
						if ok {
							ipDist[peerData.Ip] += 1
						} else {
							ipDist[peerData.Ip] = 1
						}
						// Generate RTT distribution
						rtt := math.Round(peerData.Latency*2) / 2
						_, ok = rttDis[fmt.Sprintf("%.1f", rtt)]

						if ok {
							rttDis[fmt.Sprintf("%.1f", rtt)] += 1
						} else {
							rttDis[fmt.Sprintf("%.1f", rtt)] = 1
						}
						// Generate Total connected Time Distribution
						tc := peerData.GetConnectedTime()
						// Round up to multiples of 5
						tc = math.Round(tc*2) / 2
						tct := fmt.Sprintf("%.0f", tc)
						_, ok = tctDis[tct]
						if ok {
							tctDis[tct] += 1
						} else {
							tctDis[tct] = 1
						}
					}
				} else {
					nOfDeprecatedPeers++
				}
				nOfDiscoveredPeers++

				return true
			})

			totPeers.Set(float64(nOfDiscoveredPeers))
			connectedPeers.Set(float64(nOfConnectedPeers))
			deprecatedPeers.Set(float64(nOfDeprecatedPeers))

			for _, clientName := range clients.GetClientNames() {
				count := clients.GetCountOfClient(clientName)
				// TODO: Add also version and OS
				clientDistribution.WithLabelValues(clientName).Set(float64(count))
			}
			// Country distribution
			for k, v := range geoDist {
				geoDistribution.WithLabelValues(k).Set(v)
			}
			// Client Version distribution
			for k, v := range clientVerDist {
				clientVersionDistribution.WithLabelValues(k).Set(v)
			}
			// IP distribution
			// count how many ips host the same nodess
			auxIpDist := make(map[float64]float64)
			for _, v := range ipDist {
				_, ok := auxIpDist[v]
				if ok {
					auxIpDist[v] = auxIpDist[v] + 1.0
				} else {
					auxIpDist[v] = 1.0
				}
			}

			// Reset previous distributions
			ipDistribution.Reset()
			rttDistribution.Reset()
			totcontimeDistribution.Reset()
			for k, v := range auxIpDist {
				ipDistribution.WithLabelValues(fmt.Sprintf("%.0f", v)).Set(k)
			}
			for k, v := range rttDis {
				rttDistribution.WithLabelValues(k).Set(v)
			}
			for k, v := range tctDis {
				totcontimeDistribution.WithLabelValues(k).Set(v)
			}
			allLastErrors := c.PeerStore.GetErrorCounter()

			peerstoreIterTime.Set(float64(c.PeerStore.PeerstoreIterTime) / (60 * 1000000000))

			// get the message counter per minutes
			secs := c.RefreshInterval.Seconds()
			bb := (float64(beacBlock) / secs) * 60
			ba := (float64(beacAttestation) / secs) * 60
			tot := float64(totalMsg)

			receivedMessages.WithLabelValues("beacon_blocks").Set(bb)
			receivedMessages.WithLabelValues("beacon_aggregate_and_proof").Set(ba)
			receivedTotalMessages.Set(tot)

			resetChan <- true

			log.WithFields(log.Fields{
				"ClientsDist":        clients,
				"GeoDist":            geoDist,
				"NOfDiscoveredPeers": nOfDiscoveredPeers,
				"NOfConnectedPeers":  nOfConnectedPeers,
				"NOfDeprecatedPeers": nOfDeprecatedPeers,
				"LastErrors":         allLastErrors,
				"BeaconBlocks":       bb,
				"BeaconAttestations": ba,
			}).Info("Metrics summary")

			time.Sleep(c.RefreshInterval)
		}
	}()

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", c.ExposePort), nil))
	}()

	return nil
}
