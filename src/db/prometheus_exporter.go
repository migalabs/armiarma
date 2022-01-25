package db

import (
	"fmt"
	"math"
	"time"

	promth "github.com/migalabs/armiarma/src/exporters"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ServeMetrics
// * This method will serve the global peerstore values to the
// * local prometheus instance
func (ps *PeerStore) ServePrometheusMetrics() {
	ctx := ps.Ctx()

	// Generate new ticker
	ticker := time.NewTicker(promth.MetricLoopInterval)
	// register variables
	prometheus.MustRegister(ClientDistribution)
	prometheus.MustRegister(ConnectedPeers)
	prometheus.MustRegister(DeprecatedPeers)
	prometheus.MustRegister(ClientVersionDistribution)
	prometheus.MustRegister(IpDistribution)
	prometheus.MustRegister(TotPeers)
	prometheus.MustRegister(GeoDistribution)
	prometheus.MustRegister(RttDistribution)
	prometheus.MustRegister(TotcontimeDistribution)

	// routine to loop
	go func() {
		//t := time.Now()
		for {
			select {
			case <-ticker.C:
				// auxuliar variables
				clients := NewClientDist()
				nOfDiscoveredPeers := 0
				nOfConnectedPeers := 0
				nOfDeprecatedPeers := 0
				geoDist := NewStringMapMetric()
				ipDist := NewStringMapMetric()
				rttDis := NewStringMapMetric()
				tctDis := NewStringMapMetric()

				// TODO:	-remove the Storage.Range from the PrometheusExport workflow
				//			for loop over the PeerList might not be the best idea, but should work for now
				// Iterate the peerstore to generate the exporting metrics
				peerList := ps.GetPeerList()
				for _, pID := range peerList {
					peerData, err := ps.GetPeerData(pID.String())
					if err != nil {
						continue
					}
					if !peerData.IsDeprecated() {
						//if t.Sub(peerData.LastIdentifyTimestamp) < 1024*time.Minute {
						if peerData.MetadataRequest {
							if peerData.ClientName != "" {
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
				}
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
				// count how many ips host the same nodes
				// key: number of nodes, value: number of ips
				auxIpDist := ipDist.ObtainDistribution()
				auxIpDist.SetValues(IpDistribution)
				rttDis.SetValues(RttDistribution)
				tctDis.SetValues(TotcontimeDistribution)
				Log.WithFields(logrus.Fields{
					"NOfDiscoveredPeers": nOfDiscoveredPeers,
					"NOfConnectedPeers":  nOfConnectedPeers,
					"NOfDeprecatedPeers": nOfDeprecatedPeers,
				}).Info("peerstore metrics summary")

			case <-ctx.Done():
				// closing the routine in a ordened way
				ticker.Stop()
				Log.Info("Closing DB prometheus exporter")
				return
			}
		}
	}()
}
