package metrics

import (
	"strconv"
	"fmt"
)

// Base Struct for the topic name and the received messages on the different topics
// TODO: In the future we might reuse the Rumor struct and add the missing fields
type Peer struct {
	PeerId     string
	NodeId     string
	UserAgent  string
	ClientName string
	ClientOS   string //TODO:
	ClientVersion    string
	Pubkey     string
	Addrs      string
	Ip         string
	Country    string
	City       string
	Latency    float64

	Attempted bool   // If the peer has been attempted to stablish a connection
	Succeed   bool   // If the connection attempt has been successful
	Connected bool   // If the peer was at any point connected by the crawler (keep count of incoming dials)
	Attempts  int    // Number of attempts done
	Error     string // Type of error that we detected

	MetadataRequest bool // If the peer has been attempted to request its metadata
	MetadataSucceed bool // If the peer has been successfully requested its metadata

	ConnectionEvents []ConnectionEvents

	TotConnections    int64
	LastConn          int64 //(timestamp in seconds of the last exported time (backup for when we are loading the Peer)
	TotDisconnections int64
	LastDisconn       int64 //(timestamp in seconds of the last exported time (backup for when we are loading the Peer)
	TotConnTime       int64
	ConnFlag          bool  // Flag that points if the peer was connected (for re-loading purposes)
	LastExport        int64 //(timestamp in seconds of the last exported time (backup for when we are loading the Peer)

	// Counters for the different topics
	BeaconBlock          MessageMetrics
	BeaconAggregateProof MessageMetrics
	VoluntaryExit        MessageMetrics
	ProposerSlashing     MessageMetrics
	AttesterSlashing     MessageMetrics
	// Variables related to the SubNets (only needed for when Shards will be implemented)
}

func NewPeer(peerId string) Peer {
	pm := Peer{
		// TODO Check. What is the difference between Unknown and "" empty.
		PeerId:     peerId,
		NodeId:     "",
		UserAgent:  "",
		Pubkey:     "",
		Addrs:      "",
		Ip:         "",
		Country:    "",
		City:       "",
		Latency:    0,

		Attempted: false,
		Succeed:   false,
		Connected: false,
		Attempts:  0,
		Error:     "None",

		MetadataRequest: false,
		MetadataSucceed: false,

		ConnectionEvents:  make([]ConnectionEvents, 0),
		TotConnections:    0,
		LastConn:          0,
		TotDisconnections: 0,
		LastDisconn:       0,
		TotConnTime:       0,
		ConnFlag:          false,
		LastExport:        0,

		// Counters for the different topics
		BeaconBlock:          NewMessageMetrics(),
		BeaconAggregateProof: NewMessageMetrics(),
		VoluntaryExit:        NewMessageMetrics(),
		ProposerSlashing:     NewMessageMetrics(),
		AttesterSlashing:     NewMessageMetrics(),
	}
	return pm
}

func (pm *Peer) ResetDynamicMetrics() {
	pm.Attempts = 0
	pm.ConnectionEvents = make([]ConnectionEvents, 0)
	pm.BeaconBlock = NewMessageMetrics()
	pm.BeaconAggregateProof = NewMessageMetrics()
	pm.VoluntaryExit = NewMessageMetrics()
	pm.ProposerSlashing = NewMessageMetrics()
	pm.AttesterSlashing = NewMessageMetrics()
}

func (pm *Peer) GetAllMessagesCount() uint64 {
	return (pm.BeaconBlock.Cnt +
		pm.BeaconAggregateProof.Cnt +
		pm.VoluntaryExit.Cnt +
		pm.AttesterSlashing.Cnt +
		pm.ProposerSlashing.Cnt)
}

// TODO: quick copy paste
// filter the received Connection/Disconnection events generating a counter and the connected time
func AnalyzeConnDisconnTime(pm *Peer, currentTime int64) (int64, int64, float64) {
	var connTime int64
	// Use the counters in the PeerMetrics to check if the peer is still connected
	if pm.ConnFlag {
		connTime = pm.TotConnTime + (currentTime - pm.LastConn)
	} else {
		connTime = pm.TotConnTime
	}
	pm.LastExport = currentTime
	return pm.TotConnections, pm.TotDisconnections, float64(connTime) / 60000 // return the connection time in minutes ( / 60)
}

func (pm *Peer) ToCsvLine() string {
	// TODO: Perhaps move the following three lines somewhere else
	expTime := GetTimeMiliseconds()
	connections, disconnections, connTime := AnalyzeConnDisconnTime(pm, expTime)

	csvRow := pm.PeerId + "," +
		pm.NodeId + "," +
		pm.UserAgent + "," +
		pm.ClientName + "," +
		pm.ClientVersion + "," +
		pm.Pubkey + "," +
		pm.Addrs + "," +
		pm.Ip + "," +
		pm.Country + "," +
		pm.City + "," +
		strconv.FormatBool(pm.MetadataRequest) + "," +
		strconv.FormatBool(pm.MetadataSucceed) + "," +
		strconv.FormatBool(pm.Attempted) + "," +
		strconv.FormatBool(pm.Succeed) + "," +
		strconv.FormatBool(pm.Connected) + "," +
		strconv.Itoa(pm.Attempts) + "," +
		pm.Error + "," +
		fmt.Sprint(pm.Latency) + "," +
		strconv.FormatInt(connections, 10) + "," +
		strconv.FormatInt(disconnections, 10) + "," +
		fmt.Sprintf("%.3f", connTime) + "," +
		strconv.FormatUint(pm.BeaconBlock.Cnt, 10) + "," +
		strconv.FormatUint(pm.BeaconAggregateProof.Cnt, 10) + "," +
		strconv.FormatUint(pm.VoluntaryExit.Cnt, 10) + "," +
		strconv.FormatUint(pm.ProposerSlashing.Cnt, 10) + "," +
		strconv.FormatUint(pm.AttesterSlashing.Cnt, 10) + "," +
		strconv.FormatUint(pm.GetAllMessagesCount(), 10) + "\n"

		return csvRow

}

func (pm *Peer) LogPeer() {
	// TODO
}
