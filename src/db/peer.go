package db

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	bc_topics "github.com/migalabs/armiarma/src/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/src/utils"
	all_utils "github.com/migalabs/armiarma/src/utils"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/sirupsen/logrus"
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
	ConnectedDirection    []string  // The directions of each connection event.
	IsConnected           bool      // If the peer is connected (CheckIfRealConnect).
	Attempted             bool      // If the peer has been attempted to stablish a connection.
	Succeed               bool      // If the connection attempt (outbound) has been successful.
	Attempts              uint64    // Number of attempts done.
	Error                 []string  // List of errors (also adding "None" errors), aligned with connection events.
	LastErrorTimestamp    time.Time // Timestamp of the last error reported for this peer.
	Deprecated            bool      // Flag to rummarize whether the peer is still valid for statistics or not. If true, the peer is not exported CSV / metrics.
	LastIdentifyTimestamp time.Time // Timestamp of the last time the peer was identified (get user agent...).

	NegativeConnAttempts []time.Time // List of dates when the peer retrieved a negative connection attempt (outbound) (if there is a possitive one, clean the array).
	ConnectionTimes      []time.Time // List of registered connections events.
	DisconnectionTimes   []time.Time // List of Disconnection Events.
	MetadataRequest      bool        // If the peer has been attempted to request its metadata.
	MetadataSucceed      bool        // If the peer has been successfully requested its metadata.

	LastExport int64 // Timestamp in seconds of the last exported time (backup for when we are loading the Peer).

	// Beacon
	BeaconStatus   BeaconStatusStamped
	BeaconMetadata BeaconMetadataStamped

	// Message
	// Counters for the different topics
	MessageMetrics map[string]MessageMetric
}

// Default constructor
func NewPeer(peerId string) Peer {
	pm := Peer{
		PeerId:               peerId,
		Error:                make([]string, 0),
		MAddrs:               make([]ma.Multiaddr, 0),
		Protocols:            make([]string, 0),
		ConnectedDirection:   make([]string, 0),
		NegativeConnAttempts: make([]time.Time, 0),
		ConnectionTimes:      make([]time.Time, 0),
		DisconnectionTimes:   make([]time.Time, 0),
		MessageMetrics:       make(map[string]MessageMetric),
	}
	return pm
}

// FetchPeerInfoFromPeer:
// This method will read information from a new peer,
// and update the self (pm) peer with the new peer's attributes.
// Keep in mind that only fields which are not empty in the new peer
// will be overwriten in the old peer.
// @param newPeer: the peer from where to get new information and update
// the old one.
func (pm *Peer) FetchPeerInfoFromNewPeer(newPeer Peer) {

	pm.FetchBasicHostInfoFromNewPeer(newPeer)
	pm.FetchConnectionsFromNewPeer(newPeer)
	pm.FetchChainNodeFromNewPeer(newPeer)

}

// FetchBasicHostInfoFromNewPeer:
// This method will read basic host info from the new peer and import it
// into our peer (pm).
// @param newPeer: the peer where to extract new information.
func (pm *Peer) FetchBasicHostInfoFromNewPeer(newPeer Peer) {

	// Somehow weird to update the peerID, since it is going to be the same one
	pm.PeerId = getNonEmpty(pm.PeerId, newPeer.PeerId)
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

	// Check User Agent and derivated client type/version/OS
	pm.UserAgent = getNonEmpty(pm.UserAgent, newPeer.UserAgent)
	pm.ClientOS = getNonEmpty(pm.ClientOS, newPeer.ClientOS)
	if newPeer.ClientName != "" || pm.ClientName == "" {
		pm.ClientName = newPeer.ClientName
		pm.ClientVersion = newPeer.ClientVersion
	}

}

// FetchConnectionsFromNewPeer:
// This method will read connections from the new peer and import it
// into our peer (pm).
// @param newPeer: the peer where to extract new information.
func (pm *Peer) FetchConnectionsFromNewPeer(newPeer Peer) {

	// only change these when old one was false, otherwise, leave as t is
	if !pm.MetadataRequest {
		pm.MetadataRequest = newPeer.MetadataRequest
	}
	if !pm.MetadataSucceed {
		pm.MetadataSucceed = newPeer.MetadataSucceed
	}

	pm.Attempts += newPeer.Attempts

	if newPeer.Attempted {
		pm.Attempted = newPeer.Attempted
		// as it was attempted, check if there were negative connections
		// if not, then clean our negative connections
		if len(newPeer.NegativeConnAttempts) == 0 {
			//there are no negative attempts here
			pm.NegativeConnAttempts = make([]time.Time, 0)
		} else {
			// copy the negative attempts
			for _, negAttTmp := range newPeer.NegativeConnAttempts {
				pm.NegativeConnAttempts = append(pm.NegativeConnAttempts, negAttTmp)
			}
		}
	}

	if !pm.Succeed {
		pm.Succeed = newPeer.Succeed
	}

	// Check that we dont fetch old peer into old Peer
	// Edgy case that makes the memory increase exponentially after several hours of run
	if len(newPeer.ConnectionTimes) > 1 {
		Log.Warnf("careful! peer with %d connections is getting fetched into peer with %d ones. This might end up in an exponential Heap-Memory increase.", len(newPeer.ConnectionTimes), len(pm.ConnectionTimes))
	}

	if len(newPeer.ConnectionTimes) != len(newPeer.ConnectedDirection) {
		Log.Warnf("Attention, fetching peer with different number of directions and connections")
		Log.Warnf("ConnectionTimes: %d, ConnectedDirection: %d", len(newPeer.ConnectionTimes), len(newPeer.ConnectedDirection))
	}

	connectedDirectionindex := 0
	// Aggregate connections with directions and disconnections
	for _, time := range newPeer.ConnectionTimes {

		// check if the connection has an associated direction
		newConnectedDirection := ""
		// if we have exceeded the length of the array, default
		if connectedDirectionindex >= len(newPeer.ConnectedDirection) {
			// this should never happen, all connections must have a direction
			newConnectedDirection = "unknown"
		} else {
			newConnectedDirection = newPeer.ConnectedDirection[connectedDirectionindex]
			connectedDirectionindex++
		}

		pm.ConnectionEvent(newConnectedDirection, time)
	}
	for _, time := range newPeer.DisconnectionTimes {
		pm.DisconnectionEvent(time)
	}

	for _, errorTmp := range newPeer.Error {
		pm.Error = append(pm.Error, errorTmp)
	}

	if newPeer.LastErrorTimestamp.After(pm.LastErrorTimestamp) {
		pm.LastErrorTimestamp = newPeer.LastErrorTimestamp
	}

	if newPeer.LastIdentifyTimestamp.After(pm.LastIdentifyTimestamp) {
		pm.LastIdentifyTimestamp = newPeer.LastIdentifyTimestamp
	}

}

// FetchChainNodeFromNewPeer:
// This method will read chain node information from the new peer and import it
// into our peer (pm).
// @param newPeer: the peer where to extract new information.

func (pm *Peer) FetchChainNodeFromNewPeer(newPeer Peer) {
	pm.NodeId = getNonEmpty(pm.NodeId, newPeer.NodeId)

	// Beacon Metadata and Status
	if newPeer.BeaconMetadata != (BeaconMetadataStamped{}) {
		pm.BeaconMetadata = newPeer.BeaconMetadata
	}
	if newPeer.BeaconStatus != (BeaconStatusStamped{}) {
		pm.BeaconStatus = newPeer.BeaconStatus
	}
}

// GetBlockchainNode:
// Parses and returns the stored BlockchainNode. It uses the stored ENR to get the data.
// @return Node: the resulting node of parsing the ENR.
// @return any error if it was the case.
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

// AddAddr:
// This method adds a new multiaddress in string format to the MAddrs array.
// @return Any error. Otherwise nil.
func (pm *Peer) AddMAddr(input_addr string) error {
	new_ma, err := ma.NewMultiaddr(input_addr) // parse and format

	if err != nil {
		return err
	}
	pm.MAddrs = append(pm.MAddrs, new_ma)
	return nil
}

// ExtractPublicMAddr:
// This method loops over all multiaddress and extract the first one that has
// a public IP.
// @return the found multiaddress, nil if error.
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

// ResetDynamicMetrics:
// It will reset the metrics by reinstancing the map.
func (pm *Peer) ResetDynamicMetrics() {
	pm.MessageMetrics = make(map[string]MessageMetric)
}

// IsDeprecated:
// This method return the deprecated attirbute of the peer.
// @return true or false.
func (pm Peer) IsDeprecated() bool {
	return pm.Deprecated
}

// LastNegAttempt:
// This method will calculate the last negative attempt time.
// @return the time of the last negative connection attempt with this peer and and error if applicable.
func (pm Peer) LastNegAttempt() (t time.Time, err error) {
	if len(pm.NegativeConnAttempts) == 0 {
		err = errors.New("no negative connections for the peer")
		return
	}
	t = pm.NegativeConnAttempts[len(pm.NegativeConnAttempts)-1]
	err = nil
	return
}

// FirstNegAttempt:
// This method will calculate the last negative attempt time.
// @return the time of the first negative connection attempt with this peer and and error if applicable.
func (pm Peer) FirstNegAttempt() (t time.Time, err error) {
	if len(pm.NegativeConnAttempts) == 0 {
		err = errors.New("no negative connections for the peer")
		return
	}
	t = pm.NegativeConnAttempts[0]
	err = nil
	return
}

// AddNegConnAtt:
// This method will register a new negative connection attempt in the peer (outbound).
// @param deprecated: in case we want to activate the deprecation flag.
// @param err: error string to add to the peer error list.
func (pm *Peer) AddNegConnAtt(deprecated bool, err string) {

	t := time.Now()
	// add a new time to the array of negative attempts
	pm.NegativeConnAttempts = append(pm.NegativeConnAttempts, t)
	pm.Deprecated = deprecated // set deprecated to the indicated by the param
	pm.Attempts += 1
	pm.Attempted = true
	pm.Error = append(pm.Error, err)
	pm.LastErrorTimestamp = t

}

// TODO: these two methods could possibly be merged with parameters

// AddPositiveConnAttempt:
// This method will register a new positive connection attempt in the peer (outbound).
func (pm *Peer) AddPositiveConnAttempt() {
	// as we have a positive attempt, flush the negative attempts
	pm.NegativeConnAttempts = make([]time.Time, 0)
	pm.Deprecated = false // not deprecated anymore
	pm.Attempted = true
	pm.Attempts += 1
	pm.Succeed = true                   // this peer counts now as succeeded
	pm.Error = append(pm.Error, "None") // append the no error

}

// ConnectionEvent:
// Register when a new connection was detected and the direction.
// @param direction: whether inbound or outbound.
// @param time: when the event happenned.
func (pm *Peer) ConnectionEvent(direction string, time time.Time) {
	pm.ConnectionTimes = append(pm.ConnectionTimes, time)
	pm.ConnectedDirection = append(pm.ConnectedDirection, direction)
	// update isconnected flag based on the last connection / disconnection
	pm.IsConnected = pm.CheckIfPeerRealConnect()
}

// DisconnectionEvent:
// Register when a new disconnection was detected.
// @param time: when the event happenned.
func (pm *Peer) DisconnectionEvent(time time.Time) {
	pm.DisconnectionTimes = append(pm.DisconnectionTimes, time)
	// update isconnected flag based on the last connection / disconnection
	pm.IsConnected = pm.CheckIfPeerRealConnect()
}

// CheckIfPeerRealConnect:
// This method will return whether the peer is currently connected or not.
// @return true if connected, false if not.
func (pm *Peer) CheckIfPeerRealConnect() bool {
	if len(pm.ConnectionTimes) == 0 {
		return false
	}
	lastConn := pm.ConnectionTimes[len(pm.ConnectionTimes)-1]

	if len(pm.DisconnectionTimes) == 0 {
		return true
	}
	lastDisconn := pm.DisconnectionTimes[len(pm.DisconnectionTimes)-1]
	// if the last disconnection is before the last connection,
	// then the connection is not closed, therefore still connected
	return lastDisconn.Before(lastConn)
}

// GetLastActivityTime:
// Calculates the last activity recorded for the peer.
// @return last activity recorded for the peer.
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

// GetLastErrors:
// Returns the error in last position. The array also contains the same error if it was repeated consecutively.
// @return and array with the same error.
func (pm *Peer) GetLastErrors() []string {
	errorList := make([]string, 0)
	lastError := ""

	if len(pm.Error) > 0 { // get the last error
		lastError = pm.Error[len(pm.Error)-1]
	}

	for i := range pm.Error {
		tmpError := pm.Error[len(pm.Error)-i-1] // range backwards
		if tmpError == lastError {              // if the error was the same as the lastone, add it to the list
			errorList = append(errorList, tmpError)
		} else { // once we find a different error, then it is not consecutive anymore: break
			break
		}
	}
	return errorList

}

// GetConnectedTime:
// This method will calculate the total connected time
// based on con/disc timestamps. This means the total time that
// the peer has been connected.
// Shifted some calculus to nanoseconds. Millisecons were
// leaving fields empty when exporting (less that 3 decimals).
// @return the resulting time in float64 format.
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

// MetadataEvent:
// Add a Metadata Event to the given peer (successful or not).
// @param success: whether successful (to identify the peer) or not.
func (pm *Peer) MetadataEvent(success bool) {
	pm.MetadataRequest = true
	pm.MetadataSucceed = success
}

// UpdateBeaconMetadata:
// Update beacon Metadata of the peer.
// @param bMetadata: the Metadata object used to update the data
func (pm *Peer) UpdateBeaconMetadata(bMetadata common.MetaData) {
	pm.BeaconMetadata = BeaconMetadataStamped{
		Timestamp: time.Now(),
		Metadata:  bMetadata,
	}
}

// UpdateBeaconStatus:
// Update beacon Status of the peer.
// @param bStatus: the Status object used to update the data
func (pm *Peer) UpdateBeaconStatus(bStatus common.Status) {
	pm.BeaconStatus = BeaconStatusStamped{
		Timestamp: time.Now(),
		Status:    bStatus,
	}
}

// Basic BeaconMetadata struct that includes the timestamp of the received beacon metadata
type BeaconMetadataStamped struct {
	Timestamp time.Time
	Metadata  common.MetaData
}

//  Basic BeaconMetadata struct that includes The timestamp of the received beacon Status
type BeaconStatusStamped struct {
	Timestamp time.Time
	Status    common.Status
}

// Information regarding the messages received on a given topic
type MessageMetric struct {
	Count            uint64
	FirstMessageTime time.Time
	LastMessageTime  time.Time
}

// MessageEvent:
// Add one to the message count for the given topic.
// Also update the LastMessageTime.
// @param topicName: the topic to add a message on.
// @param time: when it was received.
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

// GetNumOfMsgFromTopic:
// Get the number of messages that we got for a given topic. Note that
// the topic name is the shortened name i.e. BeaconBlock
// @param shortTopic: the topic to get count from.
// @return a uint64 with the total count.
func (pm *Peer) GetNumOfMsgFromTopic(shortTopic string) uint64 {
	msgMetric, ok := pm.MessageMetrics[bc_topics.GenerateEth2Topics(bc_topics.MainnetKey, shortTopic)]
	if ok {
		return msgMetric.Count
	}
	return uint64(0)
}

// GetAllMessagesCount:
// Get total of messages received from that peer.
// @return a unit64 with the count.
func (pm *Peer) GetAllMessagesCount() uint64 {
	totalMessages := uint64(0)
	for _, messageMetric := range pm.MessageMetrics {
		totalMessages += messageMetric.Count
	}
	return totalMessages
}

// getNonEmpty:
// This method will compare two strings and return one of them.
// @param old: the old string.
// @param new: the new string.
// @return the new string in case not empty, the old one in any other case.
func getNonEmpty(old string, new string) string {
	if new != "" {
		return new
	}
	return old
}

// getNonEmptyMAddrArray:
// This method will compare two multiaddresses arrays and return one of them.
// @param old: the old (already in db) multiaddress array.
// @param new: the new (just found) multiadddress array.
// @return the new ma array in case not empty, the old one in any other case.
func getNonEmptyMAddrArray(old []ma.Multiaddr, new []ma.Multiaddr) []ma.Multiaddr {
	if len(new) != 0 {
		return new
	}
	return old
}

// ToCsvLine:
// This method will export all peer attributes into a single string
// in CSV format.
// @return the resulting string.
func (pm *Peer) ToCsvLine() string {
	// register if the peer was conected
	connStablished := "false"
	if len(pm.ConnectionTimes) > 0 {
		connStablished = "true"
	}
	// get the multiaddress of the peers
	uniqueAddr := ""
	if len(pm.MAddrs) != 0 {
		mAddrss := pm.ExtractPublicAddr()
		if mAddrss != nil {
			uniqueAddr = mAddrss.String()
		} else {
			uniqueAddr = pm.MAddrs[0].String()
		}
	}

	node, err := pm.GetBlockchainNode()
	forkDigest := ""
	if err != nil {
		Log.Errorf("Could not parse ENR to CSV")

	} else {
		eth2Dat, _, err := all_utils.ParseNodeEth2Data(*node)

		if err == nil {
			forkDigest = eth2Dat.ForkDigest.String()
		}
	}
	lastConnectionTime := ""
	if len(pm.ConnectionTimes) > 0 {
		lastConnectionTime = pm.ConnectionTimes[len(pm.ConnectionTimes)-1].String()
	}

	csvRow := pm.PeerId + "," +
		pm.NodeId + "," +
		forkDigest + "," +
		pm.UserAgent + "," +
		pm.ClientName + "," +
		pm.ClientVersion + "," +
		pm.Pubkey + "," +
		uniqueAddr + "," +
		pm.Ip + "," +
		pm.Country + "," +
		pm.City + "," +
		strconv.FormatBool(pm.MetadataRequest) + "," +
		strconv.FormatBool(pm.MetadataSucceed) + "," +
		strconv.FormatBool(pm.Attempted) + "," +
		strconv.FormatBool(pm.Succeed) + "," +
		strconv.FormatBool(pm.Deprecated) + "," +
		connStablished + "," +
		strconv.FormatBool(pm.IsConnected) + "," +
		strconv.FormatUint(pm.Attempts, 10) + "," +
		strings.Join(pm.Error, "|") + "," +
		pm.LastErrorTimestamp.String() + "," +
		pm.LastIdentifyTimestamp.String() + "," +
		fmt.Sprintf("%.6f", pm.Latency) + "," +
		fmt.Sprintf("%d", len(pm.ConnectionTimes)) + "," +
		fmt.Sprintf("%d", len(pm.DisconnectionTimes)) + "," +
		lastConnectionTime + "," +
		strings.Join(pm.ConnectedDirection, "|") + "," +
		fmt.Sprintf("%.6f", pm.GetConnectedTime()) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("BeaconBlock"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("BeaconAggregateProof"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("VoluntaryExit"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("ProposerSlashing"), 10) + "," +
		strconv.FormatUint(pm.GetNumOfMsgFromTopic("AttesterSlashing"), 10) + "," +
		strconv.FormatUint(pm.GetAllMessagesCount(), 10) + "\n"

	return csvRow
}

// LogPeer:
// Log peer information
func (pm *Peer) LogPeer() {
	Log.WithFields(logrus.Fields{
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

// PeerUnMarshal:
// This method will create a Peer object reading a map of (string -> interface).
// @return the resulting Peer.
func PeerUnMarshal(m map[string]interface{}) Peer {

	// for some fields we need to perform a check and parsing
	m_addrs := make([]ma.Multiaddr, 0) // where to store the unmarshalled
	err := errors.New("")
	if m["MAddrs"] != nil {
		m_addrs, err = utils.ParseInterfaceAddrArray(m["MAddrs"].([]interface{}))
		if err != nil {
			Log.Errorf(err.Error())
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

	directionList := make([]string, 0)
	if m["ConnectedDirection"] != nil {
		directionList = utils.ParseInterfaceStringArray(m["ConnectedDirection"].([]interface{}))
	}

	errorList := make([]string, 0)
	if m["Error"] != nil {
		errorList = utils.ParseInterfaceStringArray(m["Error"].([]interface{}))
	}

	protocolVersionNew := ""
	if m["ProtocolVersion"] != nil {
		protocolVersionNew = m["ProtocolVersion"].(string)
	}

	msgMetrics := make(map[string]MessageMetric)
	if m["MessageMetrics"] != nil {
		msgMetrics, err = ParseInterfaceMapMessageMetrics(m["MessageMetrics"].(map[string]interface{}))
		if err != nil {
			Log.Warnf("unable to cast full gossip msg metrics while unmarshaling. %s", err.Error())
		}
	}

	beaconStatus := BeaconStatusStamped{}
	if m["BeaconStatus"] != nil {
		beaconStatus, err = ParseBeaconStatusFromInterface(m["BeaconStatus"])
		if err != nil {
			Log.Warnf("unable to cast beaconStatus while unmarshaling. %s", err.Error())
		}
	}

	lastError, err := time.Parse(time.RFC3339, m["LastErrorTimestamp"].(string))
	if err != nil {
		lastError = time.Time{}
	}

	lastIdentify, err := time.Parse(time.RFC3339, m["LastIdentifyTimestamp"].(string))
	if err != nil {
		lastIdentify = time.Time{}
	}

	// TODO: use constants for names
	return Peer{
		PeerId:                m["PeerId"].(string),
		Pubkey:                m["Pubkey"].(string),
		NodeId:                m["NodeId"].(string),
		UserAgent:             m["UserAgent"].(string),
		ClientName:            m["ClientName"].(string),
		ClientOS:              m["ClientOS"].(string),
		ClientVersion:         m["ClientVersion"].(string),
		BlockchainNodeENR:     m["BlockchainNodeENR"].(string),
		Ip:                    m["Ip"].(string),
		Country:               m["Country"].(string),
		CountryCode:           m["CountryCode"].(string),
		City:                  m["City"].(string),
		Latency:               m["Latency"].(float64),
		MAddrs:                m_addrs, // correct
		Protocols:             protocolList,
		ProtocolVersion:       protocolVersionNew,
		ConnectedDirection:    directionList,
		IsConnected:           m["IsConnected"].(bool),
		Attempted:             m["Attempted"].(bool),
		Succeed:               m["Succeed"].(bool),
		Attempts:              uint64(m["Attempts"].(float64)),
		Error:                 errorList,
		LastErrorTimestamp:    lastError,
		LastIdentifyTimestamp: lastIdentify,
		Deprecated:            m["Deprecated"].(bool),
		NegativeConnAttempts:  negConns,
		ConnectionTimes:       connTimes,
		DisconnectionTimes:    disconnTimes,
		MetadataRequest:       m["MetadataRequest"].(bool),
		MetadataSucceed:       m["MetadataSucceed"].(bool),
		LastExport:            int64(m["LastExport"].(float64)),
		MessageMetrics:        msgMetrics,
		BeaconStatus:          beaconStatus,
		// BeaconMetadata:       m["BeaconMetadata"].(BeaconMetadataStamped),
	}
}

// ParseInterfaceMapMessageMetrics:
// Parse the inputMap into the MessageMetric format
// @param inputMap: a map of string interface
// @return a map of string MessageMetric
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

// ParseBeaconStatusFromInterface:
// Parse the inputMap into the BeaconStatusStamped format
// @param inputMap: a map of string interface
// @return a map of string BeaconStatusStamped
func ParseBeaconStatusFromInterface(input interface{}) (BeaconStatusStamped, error) {
	var result BeaconStatusStamped
	var err error

	inputMap := input.(map[string]interface{})

	// timestamp
	result.Timestamp, err = time.Parse(time.RFC3339, inputMap["Timestamp"].(string))
	if err != nil {
		return result, errors.Wrap(err, "unable to compose BeaconStatus.Timestamp from readed interface")
	}
	// BeaconStatus
	status := inputMap["Status"].(map[string]interface{})
	// if the forkdigest field is empty, return empty BeaconStatus
	fd, _ := status["ForkDigest"].(string)
	if len(fd) == 0 {
		return result, nil
	}
	// otherwise, compose the readed beaconStatus
	err = result.Status.ForkDigest.UnmarshalText([]byte(fd))
	if err != nil {
		return result, errors.Wrap(err, "unable to compose BeaconStatus.ForkDigest from readed interface")
	}
	fr, _ := status["FinalizedRoot"].(string)
	var frByte [32]byte
	copy(frByte[:], fr[:32])
	result.Status.FinalizedRoot = common.Root(frByte)
	e, err := strconv.ParseUint(status["Epoch"].(string), 0, 64)
	if err != nil {
		return result, errors.Wrap(err, "unable to compose BeaconStatus.Epoch from readed interface")
	}
	result.Status.FinalizedEpoch = common.Epoch(uint64(e))
	hr, _ := status["HeadRoot"].(string)
	var hrBytes [32]byte
	copy(hrBytes[:], hr[:32])
	result.Status.HeadRoot = common.Root(hrBytes)
	s, err := strconv.ParseUint(status["HeadSlot"].(string), 0, 64)
	if err != nil {
		return result, errors.Wrap(err, "unable to compose BeaconStatus.HeadSlot from readed interface")
	}
	result.Status.HeadSlot = common.Slot(uint64(s))
	return result, nil

}
