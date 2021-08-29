package metrics

import (
	"fmt"
	//"github.com/pkg/errors"
	"github.com/protolambda/rumor/metrics/utils"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

// Stores all the information related to a peer
type Peer struct {
	PeerId        string
	NodeId        string
	UserAgent     string
	ClientName    string
	ClientOS      string //TODO:
	ClientVersion string
	Pubkey        string
	Addrs         string
	Ip            string
	Country       string
	City          string
	Latency       float64
	// TODO: Store Enr

	ConnectedDirection string
	IsConnected        bool
	Attempted          bool   // If the peer has been attempted to stablish a connection
	Succeed            bool   // If the connection attempt has been successful
	Attempts           uint64 // Number of attempts done
	Error              string // Type of error that we detected. TODO: We are just storing the last one
	ConnectionTimes    []time.Time
	DisconnectionTimes []time.Time

	MetadataRequest bool  // If the peer has been attempted to request its metadata
	MetadataSucceed bool  // If the peer has been successfully requested its metadata
	LastExport      int64 //(timestamp in seconds of the last exported time (backup for when we are loading the Peer)

	// Counters for the different topics
	MessageMetrics map[string]*MessageMetric
}

// Information regarding the messages received on a given topic
type MessageMetric struct {
	Count            uint64
	FirstMessageTime time.Time
	LastMessageTime  time.Time
}

func NewPeer(peerId string) Peer {
	pm := Peer{
		PeerId:    peerId,
		Error:     "None",
		ConnectionTimes:    make([]time.Time, 0),
		DisconnectionTimes: make([]time.Time, 0),
		MessageMetrics: make(map[string]*MessageMetric),
	}
	return pm
}

func (pm *Peer) ResetDynamicMetrics() {
	pm.Attempts = 0
	pm.MessageMetrics = make(map[string]*MessageMetric)
}

// Register when a new connection was detected
func (pm *Peer) ConnectionEvent(direction string, time time.Time) {
	pm.ConnectionTimes = append(pm.ConnectionTimes, time)
	pm.IsConnected = true
	pm.ConnectedDirection = direction
}

// Register when a disconnection was detected
func (pm *Peer) DisconnectionEvent(time time.Time) {
	pm.DisconnectionTimes = append(pm.DisconnectionTimes, time)
	pm.IsConnected = false
	pm.ConnectedDirection = ""
}

// Register when a connection attempt was made. Note that there is some
// overlap with ConnectionEvent
func (pm *Peer) ConnectionAttemptEvent(succeed bool, err string) {
	pm.Attempts += 1
	if !pm.Attempted {
		pm.Attempted = true
	}
	if succeed {
		pm.Succeed = true
		pm.Error = "None"
	} else {
		pm.Error = utils.FilterError(err)
	}
}

// Count the messages we get per topis and its first/last timestamps
func (pm *Peer) MessageEvent(topicName string, time time.Time) {
	if pm.MessageMetrics[topicName] == nil {
		pm.MessageMetrics[topicName] = &MessageMetric{}
		pm.MessageMetrics[topicName].FirstMessageTime = time
	}
	pm.MessageMetrics[topicName].LastMessageTime = time
	pm.MessageMetrics[topicName].Count++
}

// Calculate the total connected time based on con/disc timestamps
func (pm *Peer) GetConnectedTime() float64 {
	var totalConnectedTime int64
	for _, conTime := range pm.ConnectionTimes {
		for _, discTime := range pm.DisconnectionTimes {
			singleConnectionTime := discTime.Sub(conTime).Milliseconds()
			if singleConnectionTime >= 0 {
				totalConnectedTime += singleConnectionTime
				break
			}
		}
	}
	return float64(totalConnectedTime) / 60000
}

// Get the number of messages that we got for a given topic. Note that
// the topic name is the shortened name i.e. BeaconBlock
func (pm *Peer) GetNumOfMsgFromTopic(shortTopic string) uint64 {
	msgMetric := pm.MessageMetrics[utils.ShortToFullTopicName(shortTopic)]
	if msgMetric != nil {
		return msgMetric.Count
	}
	return uint64(0)
}

// Get total of message rx from that peer
func (pm *Peer) GetAllMessagesCount() uint64 {
	totalMessages := uint64(0)
	for _, messageMetric := range pm.MessageMetrics {
		totalMessages += messageMetric.Count
	}
	return totalMessages
}

func (pm *Peer) ToCsvLine() string {
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
		strconv.FormatBool(pm.IsConnected) + "," +
		strconv.FormatUint(pm.Attempts, 10) + "," +
		pm.Error + "," +
		fmt.Sprint(pm.Latency) + "," +
		fmt.Sprintf("%d", len(pm.ConnectionTimes)) + "," +
		fmt.Sprintf("%d", len(pm.DisconnectionTimes)) + "," +
		fmt.Sprintf("%.3f", pm.GetConnectedTime()) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("BeaconBlock"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("BeaconAggregateProof"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("VoluntaryExit"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("ProposerSlashing"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("AttesterSlashing"), 10) + "," +
		strconv.FormatUint(pm.GetAllMessagesCount(), 10) + "\n"

	return csvRow
}

func (pm *Peer) LogPeer() {
	log.WithFields(log.Fields{
		"PeerId":        pm.PeerId,
		"NodeId":        pm.NodeId,
		"UserAgent":     pm.UserAgent,
		"ClientName":    pm.ClientName,
		"ClientOS":      pm.ClientOS,
		"ClientVersion": pm.ClientVersion,
		"Pubkey":        pm.Pubkey,
		"Addrs":         pm.Addrs,
		"Ip":            pm.Ip,
		"Country":       pm.Country,
		"City":          pm.City,
		"Latency":       pm.Latency,
	}).Info("Peer Info")
}
