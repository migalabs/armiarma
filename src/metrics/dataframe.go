package metrics

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Main Data Structure that will be used to analyze and plot the metrics
type MetricsDataFrame struct {
	// Peer Related
	PeerIds        PeerIdList
	NodeIds        NodeIdList
	UserAgent      UserAgentList
	ClientTypes    ClientTypeList
	ClientVersions ClientVersionList
	PubKeys        PubKeyList
	Addresses      AddressList
	Ips            IpList
	Countries      CountryList
	Cities         CityList
	Latencies      LatencyList

	// Metrics Related
	Connections    ConnectionList
	Disconnections DisconnectionList
	ConnectedTimes ConnectedTimeList

	RBeaconBlocks       RBeaconBlockList
	RBeaconAggregations RBeaconAggregationList
	RVoluntaryExits     RVoluntaryExitList
	RProposerSlashings  RProposerSlashingList
	RAttesterSlashings  RAttesterSlashingList

	RTotalMessages TotalMessagesList

	// Aux
	Len        int
	ExportTime time.Time
}

// Generate New DataFrame out of the GossipMetrics sync.Map copy
func NewMetricsDataFrame(metricsCopy sync.Map) *MetricsDataFrame {
	// Initialize the DataFrame with the expoting time
	mdf := &MetricsDataFrame{
		Len:        0,
		ExportTime: time.Now(),
	}
	// Generate the loop over each peer of the Metrics
	metricsCopy.Range(func(k, val interface{}) bool {
		var v PeerMetrics
		v = val.(PeerMetrics)
		mdf.PeerIds.AddItem(v.PeerId)
		mdf.NodeIds.AddItem(v.NodeId)
		mdf.UserAgent.AddItem(v.ClientType)
		// Parse the client and version type from the UserAgent/ClientType
		client, version := FilterClientType(v.ClientType)
		mdf.ClientTypes.AddItem(client)
		mdf.ClientVersions.AddItem(version)
		mdf.PubKeys.AddItem(v.Pubkey)
		mdf.Addresses.AddItem(v.Addrs)
		mdf.Ips.AddItem(v.Ip)
		mdf.Countries.AddItem(v.Country)
		mdf.Cities.AddItem(v.City)
		mdf.Latencies.AddItem(v.Latency) // in milliseconds
		// Analyze the connections from the events
		connections, disconnections, connTime := AnalyzeConnectionEvents(v.ConnectionEvents, mdf.ExportTime)
		mdf.Connections.AddItem(connections)
		mdf.Disconnections.AddItem(disconnections)
		mdf.ConnectedTimes.AddItem(connTime)
		// Gossip Messages
		mdf.RBeaconBlocks.AddItem(v.BeaconBlock.Cnt)
		mdf.RBeaconAggregations.AddItem(v.BeaconAggregateProof.Cnt)
		mdf.RVoluntaryExits.AddItem(v.VoluntaryExit.Cnt)
		mdf.RAttesterSlashings.AddItem(v.AttesterSlashing.Cnt)
		mdf.RProposerSlashings.AddItem(v.ProposerSlashing.Cnt)
		tm := v.BeaconBlock.Cnt + v.BeaconAggregateProof.Cnt + v.VoluntaryExit.Cnt +
			v.AttesterSlashing.Cnt + v.ProposerSlashing.Cnt
		mdf.RTotalMessages.AddItem(tm)

		mdf.Len = mdf.Len + 1
		return true
	})
	// return the new generated dataframe
	return mdf
}

// export MetricsDataFrame into a CSV format
func (mdf MetricsDataFrame) ExportToCSV(filePath string) error {
	csvFile, err := os.Create(filePath) // Create, New file, if exist overwrite
	if err != nil {
		return fmt.Errorf("Error Opening the file:", filePath)
	}
	defer csvFile.Close()
	// First raw of the file will be the Titles of the columns
	_, err = csvFile.WriteString("Peer Id,Node Id,User Agent,Client,Version,Pubkey,Address,Ip,Country,City,Latency,Connections,Disconnections,Connected Time,Beacon Blocks,Beacon Aggregations,Voluntary Exits,Proposer Slashings,Attester Slashings,Total Messages\n")
	if err != nil {
		return fmt.Errorf("Error while Writing the Titles on the csv")
	}
	//    for idx, _ := range mdf.PeerIds{ // all the fields must have the same length,
	for idx := 0; idx < mdf.Len; idx++ {
		var csvRow string
		// special case for the latency
		lat := fmt.Sprint(mdf.Latencies.GetByIndex(idx))
        conTime := fmt.Sprintf("%.3f", mdf.ConnectedTimes.GetByIndex(idx))
		csvRow = mdf.PeerIds.GetByIndex(idx).String() + "," + mdf.NodeIds.GetByIndex(idx) + "," + mdf.UserAgent.GetByIndex(idx) + "," + mdf.ClientTypes.GetByIndex(idx) + "," +
			mdf.ClientVersions.GetByIndex(idx) + "," + mdf.PubKeys.GetByIndex(idx) + "," + mdf.Addresses.GetByIndex(idx) + "," + mdf.Ips.GetByIndex(idx) + "," +
			mdf.Countries.GetByIndex(idx) + "," + mdf.Cities.GetByIndex(idx) + "," + lat + "," + strconv.Itoa(int(mdf.Connections.GetByIndex(idx))) + "," +
			strconv.Itoa(int(mdf.Disconnections.GetByIndex(idx))) + "," + conTime + "," + strconv.Itoa(int(mdf.RBeaconBlocks.GetByIndex(idx))) + "," + strconv.Itoa(int(mdf.RBeaconAggregations.GetByIndex(idx))) + "," +
			strconv.Itoa(int(mdf.RVoluntaryExits.GetByIndex(idx))) + "," + strconv.Itoa(int(mdf.RProposerSlashings.GetByIndex(idx))) + "," + strconv.Itoa(int(mdf.RAttesterSlashings.GetByIndex(idx))) + "," + strconv.Itoa(int(mdf.RTotalMessages.GetByIndex(idx))) + "\n"
		_, err = csvFile.WriteString(csvRow)
		if err != nil {
			return fmt.Errorf("Error while Writing the Titles on the csv")
		}
	}
	return nil
}

// Copy the sync.Map into the local DataFrame
// (every given interval the main plotter loop will update the information)
func GetMetricsDuplicate(original sync.Map) sync.Map {
	var newMap sync.Map
	// Iterate through the items on the original Map
	original.Range(func(k, v interface{}) bool {
		cp, ok := v.(PeerMetrics)
		if ok {
			newMap.Store(k, cp)
		}
		return true
	})
	return newMap
}
