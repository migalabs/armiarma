package db

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	m_utils "github.com/migalabs/armiarma/src/db/utils"
	bc_topics "github.com/migalabs/armiarma/src/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/src/utils"
	all_utils "github.com/migalabs/armiarma/src/utils"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
	log "github.com/sirupsen/logrus"
)

// Stores all the information related to a peer
type Peer struct {

	// PeerBASICS
	PeerId string
	Pubkey string

	// PeerEth2Node
	NodeId        string
	UserAgent     string
	ClientName    string
	ClientOS      string //TODO:
	ClientVersion string
	// TODO: Store Enr
	// Latest ENR
	BlockchainNodeENR string

	// PeerNetwork
	Ip              string
	Country         string
	CountryCode     string
	City            string
	Latency         float64
	MAddrs          []ma.Multiaddr
	Protocols       []string
	ProtocolVersion string

	// PeerMetrics
	ConnectedDirection string
	IsConnected        bool
	Attempted          bool   // If the peer has been attempted to stablish a connection
	Succeed            bool   // If the connection attempt has been successful
	Attempts           uint64 // Number of attempts done
	Error              string // Type of error that we detected. TODO: We are just storing the last one
	Deprecated         bool   // Flag to rummarize whether the peer is longer valid for statistics or not. If false, the peer is not exported in db.

	NegativeConnAttempts []time.Time // List of dates when the peer retreived a negative connection attempt (if there is a possitive one, clean the struct)
	ConnectionTimes      []time.Time
	DisconnectionTimes   []time.Time
	MetadataRequest      bool // If the peer has been attempted to request its metadata
	MetadataSucceed      bool // If the peer has been successfully requested its metadata

	LastExport int64 //(timestamp in seconds of the last exported time (backup for when we are loading the Peer)

	// Beacon
	BeaconStatus   BeaconStatusStamped
	BeaconMetadata BeaconMetadataStamped

	// Message
	// Counters for the different topics
	MessageMetrics map[string]MessageMetric
}

// **********************************************************
// *						PEER_BASICS						*
//***********************************************************

func NewPeer(peerId string) Peer {
	pm := Peer{
		PeerId:               peerId,
		Error:                "None",
		MAddrs:               make([]ma.Multiaddr, 0),
		Protocols:            make([]string, 0),
		NegativeConnAttempts: make([]time.Time, 0),
		ConnectionTimes:      make([]time.Time, 0),
		DisconnectionTimes:   make([]time.Time, 0),
		MessageMetrics:       make(map[string]MessageMetric),
	}
	return pm
}

// FetchPeerInfoFromPeer
// * This method will read information from a new peer,
// * and update the self (pm) peer with the new peer's attributes.
// * Keep in mind that only fields which are not empty in the new peer
// * will be overwriten in the old peer.
// @param newPeer: the peer from where to get new information and update
// the old one.
func (pm *Peer) FetchPeerInfoFromPeer(newPeer Peer) {
	// Somehow weird to update the peerID, since it is going to be the same one
	pm.PeerId = getNonEmpty(pm.PeerId, newPeer.PeerId)
	pm.NodeId = getNonEmpty(pm.NodeId, newPeer.NodeId)
	// Check User Agent and derivated client type/version/OS
	pm.UserAgent = getNonEmpty(pm.UserAgent, newPeer.UserAgent)
	pm.ClientOS = getNonEmpty(pm.ClientOS, newPeer.ClientOS)
	if newPeer.ClientName != "" || pm.ClientName == "" {
		pm.ClientName = newPeer.ClientName
		pm.ClientVersion = newPeer.ClientVersion
	}
	pm.Pubkey = getNonEmpty(pm.Pubkey, newPeer.Pubkey)
	pm.MAddrs = getNonEmptyMAddrArray(pm.MAddrs, newPeer.MAddrs)
	pm.ProtocolVersion = getNonEmpty(pm.ProtocolVersion, newPeer.ProtocolVersion)
	if len(newPeer.Protocols) != 0 {
		pm.Protocols = newPeer.Protocols
	}
	pm.Ip = getNonEmpty(pm.Ip, newPeer.Ip)
	if pm.City == "" || newPeer.City != "" {
		pm.City = newPeer.City
		pm.Country = newPeer.Country
	}
	if newPeer.Latency > 0 {
		pm.Latency = newPeer.Latency
	}
	// Metadata requested
	if !pm.MetadataRequest {
		pm.MetadataRequest = newPeer.MetadataRequest
	}
	if !pm.MetadataSucceed {
		pm.MetadataSucceed = newPeer.MetadataSucceed
	}
	// Beacon Metadata and Status
	if newPeer.BeaconMetadata != (BeaconMetadataStamped{}) {
		pm.BeaconMetadata = newPeer.BeaconMetadata
	}
	if newPeer.BeaconStatus != (BeaconStatusStamped{}) {
		pm.BeaconStatus = newPeer.BeaconStatus
	}
	// Check that we dont fetch old peer into old Peer
	// Edgy case that makes the memory increase exponentially after several hours of run
	if len(newPeer.ConnectionTimes) > 1 {
		log.Warnf("careful! peer with %d connections is getting fetched into peer with %d ones. This might end up in an exponential Heap-Memory increase.", len(newPeer.ConnectionTimes), len(pm.ConnectionTimes))
	}

	// Aggregate connections and disconnections
	for _, time := range newPeer.ConnectionTimes {
		pm.ConnectionEvent(newPeer.ConnectedDirection, time)
	}
	for _, time := range newPeer.DisconnectionTimes {
		pm.DisconnectionEvent(time)
	}
}

// **********************************************************
// *						PEER_ETH2_NODE					*
//***********************************************************

func (pm *Peer) GetBlockchainNode() (*enode.Node, error) {
	if pm.BlockchainNodeENR == "" {
		return nil, fmt.Errorf("unable to get ENODE for peer, no ENR was recorded")
	}
	pointer := enode.MustParse(pm.BlockchainNodeENR)
	if pointer == nil {
		return nil, fmt.Errorf("pointer to ENR was nil")
	}
	return pointer, nil
}

// **********************************************************
// *						PEER_NETWORK					*
//***********************************************************

// AddAddr
// * This method adds a new multiaddress in string format to the
// * Addrs array.
// @return Any error. Otherwise nil.
func (pm *Peer) AddMAddr(input_addr string) error {
	new_ma, err := ma.NewMultiaddr(input_addr) // parse and format

	if err != nil {
		return err
	}
	pm.MAddrs = append(pm.MAddrs, new_ma)
	return nil
}

// ExtractPublicMAddr
// * This method loops over all multiaddress and extract the one that has
// * a public IP. There must be only one.
// @return the found multiaddress, nil if error
func (pm *Peer) ExtractPublicAddr() ma.Multiaddr {

	// loop over all multiaddresses in the array
	for _, temp_addr := range pm.MAddrs {
		temp_extracted_ip := utils.ExtractIPFromMAddr(temp_addr)

		// check if IP is public
		if utils.IsIPPublic(temp_extracted_ip) == true {
			// the IP is public
			return temp_addr
		}
	}
	return nil // ended loop without returning a public address

}

// **********************************************************
// *						PEER_METRICS					*
//***********************************************************

func (pm *Peer) ResetDynamicMetrics() {
	pm.MessageMetrics = make(map[string]MessageMetric)
}

// IsDeprecated
// This method return the deprecated attirbute of the peer
// @return true or false
func (pm Peer) IsDeprecated() bool {
	return pm.Deprecated
}

// return the time of the last connection with this peer
func (pm Peer) LastNegAttempt() (t time.Time, err error) {
	if len(pm.NegativeConnAttempts) == 0 {
		err = errors.New("no negative connections for the peer")
		return
	}
	t = pm.NegativeConnAttempts[len(pm.NegativeConnAttempts)-1]
	err = nil
	return
}

// return the time of the last connection with this peer
func (pm Peer) FirstNegAttempt() (t time.Time, err error) {
	if len(pm.NegativeConnAttempts) == 0 {
		err = errors.New("no negative connections for the peer")
		return
	}
	t = pm.NegativeConnAttempts[0]
	err = nil
	return
}

func (pm *Peer) AddNegConnAtt(deprecated bool) {
	t := time.Now()
	pm.NegativeConnAttempts = append(pm.NegativeConnAttempts, t)
	fmt.Println(deprecated)
	if deprecated {
		pm.Deprecated = true
	}
}

func (pm *Peer) AddPositiveConnAttempt() {
	pm.NegativeConnAttempts = make([]time.Time, 0)
	pm.Deprecated = false
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

// GetLastActivityTime
// * Calculates the last activity recorded for the peer
// @return last activity recorded for the peer
func (pm Peer) GetLastActivityTime() time.Time {
	// check len before
	last_negative_activity := time.Time{}
	last_connection_activity := time.Time{}
	last_disconnection_activity := time.Time{}

	if len(pm.NegativeConnAttempts) > 0 {
		last_negative_activity = pm.NegativeConnAttempts[len(pm.NegativeConnAttempts)-1]
	}

	if len(pm.ConnectionTimes) > 0 {
		last_connection_activity = pm.ConnectionTimes[len(pm.ConnectionTimes)-1]
	}
	if len(pm.DisconnectionTimes) > 0 {
		last_disconnection_activity = pm.DisconnectionTimes[len(pm.DisconnectionTimes)-1]
	}

	return utils.ReturnGreatestTime([]time.Time{last_negative_activity,
		last_connection_activity, last_disconnection_activity})

}

// ConnectionAttemptEvent
// TODO: comment
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
		pm.Error = m_utils.FilterError(err)
	}
}

// GetConnectedTime
// * This method will calculate the total connected time
// * based on con/disc timestamps. This means the total time that
// * the peer has been connected.
// * Shifted some calculus to nanoseconds, Millisecons were
// * leaving fields empty when exporting (less that 3 decimals)
// @return the resulting time in float64 format
func (pm *Peer) GetConnectedTime() float64 {
	var totalConnectedTime int64
	for _, conTime := range pm.ConnectionTimes {
		for _, discTime := range pm.DisconnectionTimes {
			singleConnectionTime := discTime.Sub(conTime)
			if singleConnectionTime >= 0 {
				totalConnectedTime += int64(singleConnectionTime * time.Nanosecond)
				break
			} else {

			}
		}
	}
	return float64(totalConnectedTime) / 60000000000
}

// **********************************************************
// *						BEACON							*
//***********************************************************
// Basic BeaconMetadata struct that includes the timestamp of the received beacon metadata
// TODO: comment
type BeaconMetadataStamped struct {
	Timestamp time.Time
	Metadata  beacon.MetaData
}

// Update beacon Status of the peer
func (pm *Peer) UpdateBeaconStatus(bStatus beacon.Status) {
	pm.BeaconStatus = BeaconStatusStamped{
		Timestamp: time.Now(),
		Status:    bStatus,
	}
}

// Update beacon Metadata of the peer
func (pm *Peer) UpdateBeaconMetadata(bMetadata beacon.MetaData) {
	pm.BeaconMetadata = BeaconMetadataStamped{
		Timestamp: time.Now(),
		Metadata:  bMetadata,
	}
}

// BEACON STATUS

//  Basic BeaconMetadata struct that includes The timestamp of the received beacon Status
// TODO: comment
type BeaconStatusStamped struct {
	Timestamp time.Time
	Status    beacon.Status
}

// **********************************************************
// *						MESSAGE							*
//***********************************************************

// Information regarding the messages received on a given topic
type MessageMetric struct {
	Count            uint64
	FirstMessageTime time.Time
	LastMessageTime  time.Time
}

// Count the messages we get per topis and its first/last timestamps
// TODO: comment
func (pm *Peer) MessageEvent(topicName string, time time.Time) {
	m, ok := pm.MessageMetrics[topicName]
	if !ok {
		m = MessageMetric{
			FirstMessageTime: time,
		}
	}
	m.LastMessageTime = time
	m.Count++
	pm.MessageMetrics[topicName] = m
}

// Get the number of messages that we got for a given topic. Note that
// the topic name is the shortened name i.e. BeaconBlock
// TODO: comment
func (pm *Peer) GetNumOfMsgFromTopic(shortTopic string) uint64 {
	msgMetric, ok := pm.MessageMetrics[bc_topics.GenerateEth2Topics(bc_topics.MainnetKey, shortTopic)]
	if ok {
		return msgMetric.Count
	}
	return uint64(0)
}

// Get total of message rx from that peer
// GetAllMessagesCount
// TODO: comment
func (pm *Peer) GetAllMessagesCount() uint64 {
	totalMessages := uint64(0)
	for _, messageMetric := range pm.MessageMetrics {
		totalMessages += messageMetric.Count
	}
	return totalMessages
}

// **********************************************************
// *						UTILS							*
//***********************************************************

// getNonEmpty
// * This method will compare two strings and return one of them
// @param old: the old string
// @param new: the new string
// @return the new string in case not empty, the old one in any other case
func getNonEmpty(old string, new string) string {
	if new != "" {
		return new
	}
	return old
}

// getNonEmptyMAddrArray
// * This method will compare two multiaddresses arrays and return one of them
// @param old: the old (already in db) multiaddress array
// @param new: the new (just found) multiadddress array
// @return the new ma array in case not empty, the old one in any other case
func getNonEmptyMAddrArray(old []ma.Multiaddr, new []ma.Multiaddr) []ma.Multiaddr {
	if len(new) != 0 {
		return new
	}
	return old
}

// ToCsvLine
// * This method will export all peer attributes into a single string
// * in CSV format
// @return the resulting string
func (pm *Peer) ToCsvLine() string {
	// register if the peer was conected
	connStablished := "false"
	if len(pm.ConnectionTimes) > 0 {
		connStablished = "true"
	}
	// get the multiaddress of the peers
	mAddrss := ""
	if len(pm.MAddrs) != 0 {
		mAddrss = pm.ExtractPublicAddr().String()
	}

	node, err := pm.GetBlockchainNode()
	forkDigest := ""
	if err != nil {
		fmt.Errorf("Could not parse ENR to CSV")
		eth2Dat, _, err := all_utils.ParseNodeEth2Data(*node)

		if err != nil {
			forkDigest = eth2Dat.ForkDigest.String()
		}
	}

	csvRow := pm.PeerId + "," +
		pm.NodeId + "," +
		forkDigest + "," +
		pm.UserAgent + "," +
		pm.ClientName + "," +
		pm.ClientVersion + "," +
		pm.Pubkey + "," +
		mAddrss + "," +
		pm.Ip + "," +
		pm.Country + "," +
		pm.City + "," +
		strconv.FormatBool(pm.MetadataRequest) + "," +
		strconv.FormatBool(pm.MetadataSucceed) + "," +
		strconv.FormatBool(pm.Attempted) + "," +
		strconv.FormatBool(pm.Succeed) + "," +
		strconv.FormatBool(pm.Deprecated) + "," +
		// right now we would just write TRUE if the peer was connected when exporting the metrics
		// However, we want to know if the peer established a connection with us
		// Measure it, as we said from the length of the connection times
		connStablished + "," +
		strconv.FormatBool(pm.IsConnected) + "," +
		strconv.FormatUint(pm.Attempts, 10) + "," +
		pm.Error + "," +
		fmt.Sprintf("%.6f", pm.Latency) + "," +
		fmt.Sprintf("%d", len(pm.ConnectionTimes)) + "," +
		fmt.Sprintf("%d", len(pm.DisconnectionTimes)) + "," +
		fmt.Sprintf("%.6f", pm.GetConnectedTime()) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("BeaconBlock"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("BeaconAggregateProof"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("VoluntaryExit"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("ProposerSlashing"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("AttesterSlashing"), 10) + "," +
		strconv.FormatUint(pm.GetAllMessagesCount(), 10) + "\n"

	return csvRow
}

// LogPeer
// TODO: comment
func (pm *Peer) LogPeer() {
	log.WithFields(log.Fields{
		"PeerId":        pm.PeerId,
		"NodeId":        pm.NodeId,
		"UserAgent":     pm.UserAgent,
		"ClientName":    pm.ClientName,
		"ClientOS":      pm.ClientOS,
		"ClientVersion": pm.ClientVersion,
		"Pubkey":        pm.Pubkey,
		"Addrs":         pm.MAddrs,
		"Ip":            pm.Ip,
		"Country":       pm.Country,
		"City":          pm.City,
		"Latency":       pm.Latency,
	}).Info("Peer Info")
}

// PeerUnMarshal
// * This method will create a Peer object reading a map of string -> interface
// @return the resulting Peer
func PeerUnMarshal(m map[string]interface{}) Peer {

	// for some fields we need to perform a check and parsing
	m_addrs := make([]ma.Multiaddr, 0) // where to store the unmarshalled
	err := errors.New("")
	if m["MAddrs"] != nil {
		m_addrs, err = utils.ParseInterfaceAddrArray(m["MAddrs"].([]interface{}))
		if err != nil {
			log.Errorf(err.Error())
		}
	}

	negConns := make([]time.Time, 0)
	if m["NegativeConnAttempts"] != nil {
		negConns, err = utils.ParseInterfaceTimeArray(m["NegativeConnAttempts"].([]interface{}))

	}

	connTimes := make([]time.Time, 0)
	if m["ConnectionTimes"] != nil {
		connTimes, err = utils.ParseInterfaceTimeArray(m["ConnectionTimes"].([]interface{}))
	}

	disconnTimes := make([]time.Time, 0)
	if m["DisconnectionTimes"] != nil {
		disconnTimes, err = utils.ParseInterfaceTimeArray(m["DisconnectionTimes"].([]interface{}))
	}

	protocolList := make([]string, 0)
	if m["Protocols"] != nil {
		protocolList = utils.ParseInterfaceStringArray(m["Protocols"].([]interface{}))
	}

	protocolVersionNew := ""
	if m["ProtocolVersion"] != nil {
		protocolVersionNew = m["ProtocolVersion"].(string)
	}

	msgMetrics := make(map[string]MessageMetric)
	if m["MessageMetrics"] != nil {
		msgMetrics, err = ParseInterfaceMapMessageMetrics(m["MessageMetrics"].(map[string]interface{}))
		if err != nil {
			log.Warnf("unable to cast full gossip msg metrics while unmarshaling. %s", err.Error())
		}
	}

	// TODO: use constants for names
	return Peer{
		PeerId:               m["PeerId"].(string),
		Pubkey:               m["Pubkey"].(string),
		NodeId:               m["NodeId"].(string),
		UserAgent:            m["UserAgent"].(string),
		ClientName:           m["ClientName"].(string),
		ClientOS:             m["ClientOS"].(string),
		ClientVersion:        m["ClientVersion"].(string),
		BlockchainNodeENR:    m["BlockchainNodeENR"].(string),
		Ip:                   m["Ip"].(string),
		Country:              m["Country"].(string),
		CountryCode:          m["CountryCode"].(string),
		City:                 m["City"].(string),
		Latency:              m["Latency"].(float64),
		MAddrs:               m_addrs, // correct
		Protocols:            protocolList,
		ProtocolVersion:      protocolVersionNew,
		ConnectedDirection:   m["ConnectedDirection"].(string),
		IsConnected:          m["IsConnected"].(bool),
		Attempted:            m["Attempted"].(bool),
		Succeed:              m["Succeed"].(bool),
		Attempts:             uint64(m["Attempts"].(float64)),
		Error:                m["Error"].(string),
		Deprecated:           m["Deprecated"].(bool),
		NegativeConnAttempts: negConns,
		ConnectionTimes:      connTimes,
		DisconnectionTimes:   disconnTimes,
		MetadataRequest:      m["MetadataRequest"].(bool),
		MetadataSucceed:      m["MetadataSucceed"].(bool),
		LastExport:           int64(m["LastExport"].(float64)),
		MessageMetrics:       msgMetrics,
		//BeaconStatus:         beaconStt,
		//BeaconMetadata:       m["BeaconMetadata"].(BeaconMetadataStamped),
	}
}

func ParseInterfaceMapMessageMetrics(inputMap map[string]interface{}) (map[string]MessageMetric, error) {
	result := make(map[string]MessageMetric)
	// we will range over a slice of interfaces
	for k, v := range inputMap {
		vaux := v.(map[string]interface{})
		ft, err := time.Parse(time.RFC3339, vaux["FirstMessageTime"].(string))
		if err != nil {
			return result, err
		}
		lt, err := time.Parse(time.RFC3339, vaux["LastMessageTime"].(string))
		if err != nil {
			return result, err
		}
		mm := MessageMetric{
			Count:            uint64(vaux["Count"].(float64)),
			FirstMessageTime: ft,
			LastMessageTime:  lt,
		}
		result[k] = mm
	}
	return result, nil

}
