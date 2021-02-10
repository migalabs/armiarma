package metrics

import (
    "sync"
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
    ConnectedTimes    ConnectedTimeList

    RBeaconBlocks   RBeaconBlockList
    RBeaconAggregations RBeaconAggregationList
    RVoluntaryExits RVoluntaryExitList
    RProposerSlashings  RProposerSlashingList
    RAttesterSlashings  RAttesterSlashingList

    // Aux
    Len         int
    ExportTime  time.Time
}


// Generate New DataFrame out of the GossipMetrics sync.Map copy 
func NewMetricsDataFrame(metricsCopy sync.Map) *MetricsDataFrame {
    // Initialize the DataFrame with the expoting time
    mdf := &MetricsDataFrame{
        Len:        0,
        ExportTime: time.Now(),
    }
    fmt.Println("Exproting, leng of the sync.Map:")
    // Generate the loop over each peer of the Metrics
    metricsCopy.Range(func (k, val interface{}) bool{
        var v PeerMetrics
        v = val.(PeerMetrics)
        fmt.Println("Copying the od metrics to the dataframe")
        mdf.PeerIds.AddItem(v.PeerId)
        mdf.NodeIds.AddItem(v.NodeId)
        // Parse the client and version type from the UserAgent/ClientType
        client, version := FilterClientType(v.ClientType)
        mdf.ClientTypes.AddItem(client)
        mdf.ClientVersions.AddItem(version)
        mdf.PubKeys.AddItem(v.Pubkey)
//        mdf.Addresses.AddItem(v.Addrs)
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
        
        mdf.Len = mdf.Len+1
        fmt.Println(mdf.Addresses)
        return true
    })
    // return the new generated dataframe
    fmt.Println("Len of the dataframe", len(mdf.PeerIds), mdf.Len)
    return mdf
}

// export MetricsDataFrame into a CSV format
func (mdf MetricsDataFrame) ExportToCSV(filePath string) error {
    fmt.Println("Exporting the metrics")
    fmt.Println(mdf)
    csvFile, err := os.Create(filePath) // Create, New file, if exist overwrite
    if err != nil{
        return fmt.Errorf("Error Opening the file:", filePath)
    }
    defer csvFile.Close()
    // First raw of the file will be the Titles of the columns
    _, err = csvFile.WriteString("Peer Id,Node Id,Client,Version,Pubkey,Ip,Country,City,Latency,Connections,Disconnections,Connected Time,Beacon Blocks,Beacon Aggregations,Voluntary Exits,Proposer Slashings,Attester Slashings")
    if err != nil{
        return fmt.Errorf("Error while Writing the Titles on the csv")
    }
    fmt.Println("len of the dataframe when exporting:", len(mdf.PeerIds))
    fmt.Println("List of peers", mdf.Addresses)
//    for idx, _ := range mdf.PeerIds{ // all the fields must have the same length,
    for idx := 0; idx < mdf.Len; idx++{
        var csvRow string
        fmt.Println("Item Number:", idx)
        // write the csv row format on a string
        fmt.Println(mdf.PeerIds[idx].String())
        fmt.Println(mdf.NodeIds[idx])
        fmt.Println(mdf.ClientTypes[idx])
        fmt.Println(mdf.ClientVersions[idx])
        fmt.Println(mdf.PubKeys[idx])
//        fmt.Println(mdf.Addresses[idx])
        fmt.Println(mdf.Ips[idx])
        fmt.Println(mdf.Countries[idx])
        fmt.Println(mdf.Cities[idx])
        fmt.Println(string(mdf.Latencies[idx]))
        fmt.Println(string(mdf.Connections[idx]))
        fmt.Println(string(mdf.Disconnections[idx]))
        fmt.Println(string(mdf.ConnectedTimes[idx]))
        fmt.Println(string(mdf.RBeaconBlocks[idx]))
        fmt.Println(string(mdf.RBeaconAggregations[idx]))
        fmt.Println(string(mdf.RVoluntaryExits[idx]))
        fmt.Println(string(mdf.RProposerSlashings[idx]))
        fmt.Println(string(mdf.RAttesterSlashings[idx]))

        csvRow =  mdf.PeerIds[idx].String() +  "," +  mdf.NodeIds[idx] +  "," +  mdf.ClientTypes[idx] +  "," +
                    mdf.ClientVersions[idx] +  "," +  mdf.PubKeys[idx] +  "," +  mdf.Ips[idx] +  "," + 
                    mdf.Countries[idx] +  "," +  mdf.Cities[idx] +  "," +  string(mdf.Latencies[idx]) +  "," +  string(mdf.Connections[idx]) +  "," + 
                    string(mdf.Disconnections[idx]) +  "," +  string(mdf.ConnectedTimes[idx]) +  "," +  string(mdf.RBeaconBlocks[idx]) +  "," +  string(mdf.RBeaconAggregations[idx]) +  "," + 
                    string(mdf.RVoluntaryExits[idx]) +  "," +  string(mdf.RProposerSlashings[idx]) +  "," +  string(mdf.RAttesterSlashings[idx])
        fmt.Println("New Row on the csv" , csvRow)
         _, err = csvFile.WriteString(csvRow)
        if err != nil{
            return fmt.Errorf("Error while Writing the Titles on the csv")
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
        cp, ok := v.(PeerMetrics)
        if ok {
            newMap.Store(k, cp)
        }
        return true
    })
    return newMap
}
