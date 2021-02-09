package metrics

import (
    "fmt"
    "os"
    "time"
)

// Main Data Structure that will be used to analyze and plot the metrics
type MetricsDataFrame struct {
    // Peer Related 
    PeerIds     PeerIdList
    NodeIds     NodeIdList
    ClientTypes ClientTypeList
    ClientVersions  ClientVersionList
    PubKeys     PubKeyList
    Addresses   AddressList
    Ips         IpList
    Countries   CountryList
    Cities      CityList
    Latencies   LatencyList

    // Metrics Related
    Connections     ConnectionList
    Disconnections  DisconnectionList
    ConnectedTimes    ConnectTimeList

    RBeaconBlocks   RBeaconBlockList
    RBeaconAggregations RBeaconAggregationList
    RVoluntaryExits RVoluntaryExitList
    RProposerSlashings  RProposerSlashingList
    RAttesterSlashings  RAttesterSlashingList

    ExportTime  time.Duration
}


// Generate New DataFrame out of the GossipMetrics sync.Map copy 
func NewMetricsDataFrame(metricsCopy sync.Map) *MetricsDataFrame {
    // Initialize the DataFrame with the expoting time
    mdf := &MetricsDataFrame{
        ExportTime: time.Now(),
    }

    // Generate the loop over each peer of the Metrics
    metricsCopy.Range(func (k, v interface{}){
        v = v.(PeerMetrics)
        mdf.PeerIds.AddNew(v.PeerId)
        mdf.NodeIds.AddNew(v.NodeId)
        // Parse the client and version type from the UserAgent/ClientType
        client, version := FilterClientType(v.ClientType)
        mdf.ClientTypes.AddNew(client)
        mdf.ClientVersions.AddNew(version)
        mdf.PubKeys.AddNew(v.Pubkey)
        mdf.Addresses.AddNew(v.Addrs)
        mdf.Ips.AddNew(v.Ip)
        mdf.Countries.AddNew(v.Country)
        mdf.Cities.AddNew(v.City)
        mdf.Latency.AddNew(v.Lantency) // in milliseconds
        // Analyze the connections from the events
        connections, disconnections, connTime := AnalyzeConnectionEvents(v.ConnectionEvents, mdf.ExportTime)
        mdf.ConnectionsList.AddNew(connections)
        mdf.DisconnectionsList.AddNew(disconnections)
        mdf.ConnectedTimes.AddNew(connTime)
        // Gossip Messages
        mdf.RBeaconBlocks.AddNew(v.BeaconBlock.Cnt)
        mdf.RBeaconAggregations.AddNew(v.BeaconAggregateProof.Cnt)
        mdf.RVoluntaryExit.AddNew(v.VoluntaryExit.Cnt)
        mdf.RAttesterSlashing.AddNew(v.AttesterSlashing.Cnt)
        mdf.RProposerSlashing.AddNew(v.ProposerSlashing.Cnt)
    })
    // return the new generated dataframe
    return mdf
}

// export MetricsDataFrame into a CSV format
func (mdf *MetricsDataFrame) ExportToCSV(filePath string) error {
    csvFile, err := os.Create(filePath) // Create, New file, if exist overwrite
    if err != nil{
        return fmt.Errorf("Error Opening the file:", filePath)
    }
    defer csvFile.Close()
    // First raw of the file will be the Titles of the columns
    _, err := csvFile.WriteString("Peer Id, Node Id, Client, Version, Pubkey, Address, Ip, 
                                    Country, City, Latency, Connections, Disconnections, 
                                    Connected Time, Beacon Blocks, Beacon Aggregations, 
                                    Voluntary Exits, Proposer Slashings, Attester Slashings\n")
    if err != nil{
        return fmt.Errorf("Error while Writing the Titles on the csv")
    }
    for idx, item in range mdf.PeerIds{ // all the fields must have the same length,
        var csvRow string
        // write the csv row format on a string
        fmt.Fprintln(&csvRow, mdf.PeerIds[idx], ",", mdf.PeerIds[idx], ",", mdf.NodeIds[idx], ",", mdf.ClientTypes[idx], ",",
                    mdf.ClientVersions[idx], ",", mdf.PubKeys[idx], ",", mdf.Addresses[idx], ",", mdf.Ips[idx], ",",
                    mdf.Countries[idx], ",", mdf.Cities[idx], ",", mdf.Latencies[idx], ",", mdf.Connections[idx], ",",
                    mdf.Disconnections[idx], ",", mdf.ConnectedTimes[idx], ",", mdf.RBeaconBlocks[idx], ",", mdf.RBeaconAggregateProof[idx], ",",
                    mdf.RVoruntaryExits[idx], ",", mdf.RProposerSlashings[idx], ",", mdf.RAttesterSlashings[idx])
        _, err := csvFile.WriteString(csvRow)
        if err != nil {
            return fmt.Errorf("Error writing the row:" idx, "on the CSV file. Row:", csvRow)
        }
    }
    return nil
}

// Copy the sync.Map into the local DataFrame
// (every given interval the main plotter loop will update the information)
func GetMetricsDuplicate(original sync.Map) sync.Map{
    var newMap sync.Map
    // Iterate through the items on the original Map
    original.Range(func(k, v interface{})bool{
        cp, ok := v.(metrics.PeerMetrics)
        if ok {
            newMap.Store(k, cp)
        }
        return true
    })
    return newMap
}
