package db

import (
	"fmt"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	MetricLoopInterval time.Duration = 15
)

// ServeMetrics
// * This method will Set Metirc values to the
// *local prometheus instance
func (ps *PeerStore) ServeMetrics() {

	// register variables

	prometheus.MustRegister(ClientDistribution)
	prometheus.MustRegister(ConnectedPeers)
	prometheus.MustRegister(ReceivedTotalMessages)
	prometheus.MustRegister(ReceivedMessages)
	prometheus.MustRegister(PeerstoreIterTime)
	prometheus.MustRegister(DeprecatedPeers)
	prometheus.MustRegister(ClientVersionDistribution)
	prometheus.MustRegister(IpDistribution)
	prometheus.MustRegister(TotPeers)
	prometheus.MustRegister(GeoDistribution)
	prometheus.MustRegister(RttDistribution)
	prometheus.MustRegister(TotcontimeDistribution)

	// routine to loop
	go func() {
		for {
			clients := NewClientDist()

			nOfDiscoveredPeers := 0
			nOfConnectedPeers := 0
			nOfDeprecatedPeers := 0
			geoDist := NewStringMapMetric()
			ipDist := NewStringMapMetric()
			rttDis := NewStringMapMetric()
			tctDis := NewStringMapMetric()

			ps.PeerStore.Range(func(k string, peerData Peer) bool {
				if !peerData.IsDeprecated() {
					if peerData.MetadataRequest {
						if peerData.ClientName != "" {
							//fmt.Println(peerData.ClientName)
							clients.AddClientVersion(peerData.ClientName, peerData.ClientVersion)
						}
						if peerData.IsConnected {
							nOfConnectedPeers++
						}

						// Generate the Country Code distribution
						countrycode := peerData.CountryCode
						if countrycode == "" {
							countrycode = "--"
						}
						geoDist.AddOneorCreate(countrycode)

						// Generate the IP Address distribution
						ipDist.AddOneorCreate(peerData.Ip)

						// Generate RTT distribution
						rtt := math.Round(peerData.Latency*2) / 2
						rttDis.AddOneorCreate(fmt.Sprintf("%.1f", rtt))

						// Generate Total connected Time Distribution
						tc := peerData.GetConnectedTime()
						// Round up to multiples of 5
						tc = math.Round(tc*2) / 2
						tctDis.AddOneorCreate(fmt.Sprintf("%.0f", tc))

					}
				} else {
					nOfDeprecatedPeers++
				}
				nOfDiscoveredPeers++

				return true
			})

			TotPeers.Set(float64(nOfDiscoveredPeers))
			ConnectedPeers.Set(float64(nOfConnectedPeers))
			DeprecatedPeers.Set(float64(nOfDeprecatedPeers))

			// Register Clients and Version values
			for clientName, clientObj := range clients.Clients {
				count := clientObj.ReturnTotalCount()
				// TODO: Add also version and OS
				ClientDistribution.WithLabelValues(clientName).Set(float64(count))

				for _, versionObj := range clientObj.Versions {
					clientVersionName := clientName + "_" + versionObj.Name
					ClientVersionDistribution.WithLabelValues(clientVersionName).Set(float64(versionObj.Count))
				}
			}

			// Country distribution
			geoDist.SetValues(GeoDistribution)

			// IP distribution
			// count how many ips host the same nodess
			// key: number of nodes, value: number of ips
			auxIpDist := ipDist.ObtainDistribution()

			auxIpDist.SetValues(IpDistribution)

			rttDis.SetValues(RttDistribution)

			tctDis.SetValues(TotcontimeDistribution)

			//allLastErrors := ps.GetErrorCounter()

			PeerstoreIterTime.Set(float64(ps.PeerstoreIterTime) / (60 * 1000000000))

			log.WithFields(log.Fields{
				"ClientsDist":        clients.GetClientDistribution(),
				"GeoDist":            geoDist,
				"NOfDiscoveredPeers": nOfDiscoveredPeers,
				"NOfConnectedPeers":  nOfConnectedPeers,
				"NOfDeprecatedPeers": nOfDeprecatedPeers,
				"ClientVersionDist":  clients.GetClientVersionDistribution(),
				//"LastErrors":         allLastErrors,
				//"BeaconBlocks":       bb,
				//"BeaconAttestations": ba,
			}).Info("Metrics summary")

			time.Sleep(MetricLoopInterval * time.Second)
		}
	}()
}
