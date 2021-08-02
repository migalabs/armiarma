package metrics

import (
	"github.com/libp2p/go-libp2p-core/peer"
	//"github.com/protolambda/rumor/metrics/export"
	"strconv"
	"fmt"
)

// Base Struct for the topic name and the received messages on the different topics
// TODO: In the future we might reuse the Rumor struct and add the missing fields
type PeerMetrics struct {
	PeerId     peer.ID
	NodeId     string
	UserAgent  string
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
	LastConn          int64 //(timestamp in seconds of the last exported time (backup for when we are loading the peermetrics)
	TotDisconnections int64
	LastDisconn       int64 //(timestamp in seconds of the last exported time (backup for when we are loading the peermetrics)
	TotConnTime       int64
	ConnFlag          bool  // Flag that points if the peer was connected (for re-loading purposes)
	LastExport        int64 //(timestamp in seconds of the last exported time (backup for when we are loading the peermetrics)

	// Counters for the different topics
	BeaconBlock          MessageMetrics
	BeaconAggregateProof MessageMetrics
	VoluntaryExit        MessageMetrics
	ProposerSlashing     MessageMetrics
	AttesterSlashing     MessageMetrics
	// Variables related to the SubNets (only needed for when Shards will be implemented)
}

func NewPeerMetrics(peerId peer.ID) PeerMetrics {
	pm := PeerMetrics{
		// TODO Check. What is the difference between Unknown and "" empty.
		PeerId:     peerId,
		NodeId:     "",
		UserAgent: "Unknown", // TODO: why no just using ""
		Pubkey:     "",
		Addrs:      "/ip4/127.0.0.1/0000", // TODO: why not just using ""
		Ip:         "127.0.0.1",
		Country:    "Unknown",
		City:       "Unknown",
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

func (pm *PeerMetrics) ResetDynamicMetrics() {
	pm.Attempts = 0
	pm.ConnectionEvents = make([]ConnectionEvents, 0)
	pm.BeaconBlock = NewMessageMetrics()
	pm.BeaconAggregateProof = NewMessageMetrics()
	pm.VoluntaryExit = NewMessageMetrics()
	pm.ProposerSlashing = NewMessageMetrics()
	pm.AttesterSlashing = NewMessageMetrics()
}

func (pm *PeerMetrics) GetAllMessagesCount() uint64 {
	return (pm.BeaconBlock.Cnt +
		pm.BeaconAggregateProof.Cnt +
		pm.VoluntaryExit.Cnt +
		pm.AttesterSlashing.Cnt +
		pm.ProposerSlashing.Cnt)
}

func (pm *PeerMetrics) ToCsvLine() string {
	// TODO: Perhaps move the following three lines somewhere else
	expTime := GetTimeMiliseconds()
	connections, disconnections, connTime := AnalyzeConnDisconnTime(pm, expTime)

	csvRow := pm.PeerId.String() + "," +
		pm.NodeId + "," +
		pm.UserAgent + "," +
		pm.GetClientType() + "," +
		pm.GetClientVersion() + "," +
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

func (pm *PeerMetrics) LogPeerMetrics() {
	// TODO
}

func (pm *PeerMetrics) GetClientType() string {

	// TODO: Rethink, just reusing this by now
	client, _ := FilterClientType(pm.UserAgent)

	return client

}

func (pm *PeerMetrics) GetClientVersion() string {
	// TODO: Rethink, just reusing this by now
	_, version := FilterClientType(pm.UserAgent)
	return version

}

func (pm *PeerMetrics) GetClientOS() string {
	return "TODO"

}
