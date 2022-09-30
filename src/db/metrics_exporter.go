package db

import (
	"fmt"
	"math"
	"time"

	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/exporters"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ServeMetrics
// * This method will serve the global peerstore values to the
// * local prometheus instance
func (ps *PeerStore) ServeMetrics() {
	// generate the Prometheus exporter
	exptr, _ := exporters.NewMetricsExporter(
		ps.ctx,
		"Peer-Metrics-Prometheus",
		"Expose in Prometheus the metrics of the tool's DB",
		ps.initPeerPrometheusMetrics,
		ps.runPeerPrometheusMetrics,
		func() {},
		exporters.MetricLoopInterval,
	)
	// add the new exptr to the ExporterService
	ps.ExporterService.AddNewExporter(exptr)

	// generate the Prometheus exporter
	exptr, _ = exporters.NewMetricsExporter(
		ps.ctx,
		"Client Diversity DB Exporter",
		"Expose Client Distribution to PostgreSQL",
		ps.initClientDistributionSQL,
		ps.runClientDistributionSQL,
		func() {},
		6*time.Second,
	)
	// add the new exptr to the ExporterService
	ps.ExporterService.AddNewExporter(exptr)

}

// ------ Prometheus Peer Metrics Summary ------

func (ps *PeerStore) initPeerPrometheusMetrics() {
	// register variables
	prometheus.MustRegister(ClientDistribution)
	prometheus.MustRegister(ConnectedPeers)
	prometheus.MustRegister(DeprecatedPeers)
	prometheus.MustRegister(ClientVersionDistribution)
	prometheus.MustRegister(IpDistribution)
	prometheus.MustRegister(TotPeers)
	prometheus.MustRegister(GeoDistribution)
	prometheus.MustRegister(RttDistribution)
	//prometheus.MustRegister(TotcontimeDistribution)
}

func (ps *PeerStore) runPeerPrometheusMetrics() {
	// auxuliar variables
	clients := NewClientDist()
	nOfDiscoveredPeers := 0
	nOfConnectedPeers := 0
	nOfDeprecatedPeers := 0
	geoDist := NewStringMapMetric()
	ipDist := NewStringMapMetric()
	rttDis := NewStringMapMetric()
	//tctDis := NewStringMapMetric()

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
				//tc := peerData.GetConnectedTime()
				// Round up to multiples of 5
				//tc = math.Round(tc*2) / 2
				//tctDis.AddOneorCreate(fmt.Sprintf("%.0f", tc))
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
	//tctDis.SetValues(TotcontimeDistribution)
	Log.WithFields(logrus.Fields{
		"NOfDiscoveredPeers": nOfDiscoveredPeers,
		"NOfConnectedPeers":  nOfConnectedPeers,
		"NOfDeprecatedPeers": nOfDeprecatedPeers,
	}).Info("peerstore metrics summary")
}

// ------ SQL Client Diversity Exporter ------

func (ps *PeerStore) initClientDistributionSQL() {
	// register variables

}

func (ps *PeerStore) runClientDistributionSQL() {
	// get client distribution
	clients := NewClientDist()
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
			}
		}
	}
	// Count the client names
	clientCount := make(map[string]int)
	for clientName, clientObj := range clients.Clients {
		count := clientObj.ReturnTotalCount()
		clientCount[clientName] = count
	}
	// create the ClientDiversity obj for adding it to the SQL
	diversity := models.NewClientDiversity()
	diversity.Prysm = clientCount["Prysm"]
	diversity.Lighthouse = clientCount["Lighthouse"]
	diversity.Teku = clientCount["Teku"]
	diversity.Nimbus = clientCount["Nimbus"]
	diversity.Lodestar = clientCount["Lodestar"]
	diversity.Grandine = clientCount["Grandine"]
	diversity.Others = clientCount["Others"]

	ps.Storage.StoreClientDiversitySnapshot(diversity)
}
