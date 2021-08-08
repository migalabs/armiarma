package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/rumor/metrics/utils"
	pgossip "github.com/protolambda/rumor/p2p/gossip"
	"github.com/protolambda/rumor/p2p/gossip/database"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/libp2p/go-libp2p-core/host"
)

type GossipMetrics struct {
	GossipMetrics   sync.Map
	MessageDatabase *database.MessageDatabase
	StartTime       int64 // milliseconds
	MsgNotChannels  map[string](chan bool)
}

func NewGossipMetrics() GossipMetrics {
	gm := GossipMetrics{
		StartTime:      GetTimeMiliseconds(),
		MsgNotChannels: make(map[string](chan bool)),
	}
	return gm
}

// TODO: quick copy paste
func GetTimeMiliseconds() int64 {
	now := time.Now()
	//secs := now.Unix()
	nanos := now.UnixNano()
	millis := nanos / 1000000

	return millis
}

// Connection event model
type ConnectionEvents struct {
	ConnectionType string
	TimeMili       int64
}

// Exists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

type Checkpoint struct {
	Checkpoint string `json:"checkpoint"`
}

// Import an old GossipMetrics from given file
// return: - return error if there was error while reading the file
//         - return bool for existing file (true if there was a file to read, return false if there wasn't a file to read)
func (c *GossipMetrics) ImportMetrics(importFolder string) (error, bool) {
	// Check if there is any checkpoint file in the folder
	cpFile := importFolder + "/metrics/checkpoint-folder.json"
	if FileExists(cpFile) { // if exists, read it
		// get the json of the file
		fmt.Println("Importing Checkpoint Json:", cpFile)
		cp, err := os.Open(cpFile)
		if err != nil {
			return err, true
		}
		cpBytes, err := ioutil.ReadAll(cp)
		if err != nil {
			return err, true
		}
		cpFolder := Checkpoint{}
		json.Unmarshal(cpBytes, &cpFolder)
		importFile := importFolder + "/metrics/" + cpFolder.Checkpoint + "/gossip-metrics.json"
		// Check if file exist
		if FileExists(importFile) { // if exists, read it
			// get the json of the file
			fmt.Println("Importing Gossip-Metrics Json:", importFile)
			jsonFile, err := os.Open(importFile)
			if err != nil {
				return err, true
			}
			byteValue, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				return err, true
			}
			tempMap := make(map[peer.ID]Peer)
			json.Unmarshal(byteValue, &tempMap)
			// iterate to add the metrics from the json to the the GossipMetrics
			for k, v := range tempMap {
				if v.ConnFlag {
					v.TotDisconnections += 1
					v.LastDisconn = v.LastExport
					v.TotConnTime += (v.LastExport - v.LastConn)
					v.ConnFlag = false
				}
				c.GossipMetrics.Store(k, v)
			}
			return nil, true
		} else {
			return nil, false
		}
	} else {
		fmt.Println("NO previous Checkpoint")
		return nil, false
	}
}

type GossipState struct {
	GsNode  pgossip.GossipSub
	CloseGS context.CancelFunc
	// string -> *pubsub.Topic
	Topics sync.Map
	// Validation Filter Flag
	SeenFilter bool
}

// Function that Wraps/Marshals the content of the sync.Map to be exported as a json
func (c *GossipMetrics) MarshalMetrics() ([]byte, error) {
	exportTime := GetTimeMiliseconds()
	tmpMap := make(map[string]Peer)
	c.GossipMetrics.Range(func(k, v interface{}) bool {
		pm := v.(Peer)
		if pm.ConnFlag {
			pm.LastExport = exportTime
		}
		tmpMap[k.(peer.ID).String()] = pm
		return true
	})

	return json.Marshal(tmpMap)
}

// Function that Wraps/Marshals the content of the Entire Peerstore into a json
func (c *GossipMetrics) MarshalPeerStore(ep track.ExtendedPeerstore) ([]byte, error) {
	peers := ep.Peers()
	peerData := make(map[string]*track.PeerAllData)
	for _, p := range peers {
		peerData[p.String()] = ep.GetAllData(p)
	}
	return json.Marshal(peerData)
}

// Get the Real Ip Address from the multi Address list
// TODO: Implement the Private IP filter in a better way
func GetFullAddress(multiAddrs []string) string {
	var address string
	if len(multiAddrs) > 0 {
		for _, element := range multiAddrs {
			if strings.Contains(element, "/ip4/192.168.") || strings.Contains(element, "/ip4/127.0.") || strings.Contains(element, "/ip6/") || strings.Contains(element, "/ip4/172.") || strings.Contains(element, "0.0.0.0") {
				continue
			} else {
				address = element
				break
			}
		}
	} else {
		address = "/ip4/127.0.0.1/tcp/9000"
	}
	return address
}

// Function that resets to 0 the connections/disconnections, and message counters
// this way the Ram Usage gets limited (up to ~10k nodes for a 12h-24h )
// NOTE: Keep in mind that the peers that we ended up connected to, will experience a weid connection time
// TODO: Fix peers that stayed connected to the tool
func (c *GossipMetrics) ResetDynamicMetrics() {
	fmt.Println("Reseting Dynamic Metrics in Peer")
	// Iterate throught the peers in the metrics, restarting connection events and messages
	c.GossipMetrics.Range(func(key interface{}, value interface{}) bool {
		p := value.(Peer)
		p.ResetDynamicMetrics()
		c.GossipMetrics.Store(key, p)
		return true
	})
	fmt.Println("Finished Reseting Dynamic Metrics")
}

// Function that adds a notification channel to the message gossip topic
func (c *GossipMetrics) AddNotChannel(topicName string) {
	c.MsgNotChannels[topicName] = make(chan bool, 100)
}

// TODO: This module shouldnt know about rumor at all. So ExtendedPeerstore
// should be replaced
// Function that iterates through the received peers and fills the missing information
func (c *GossipMetrics) FillMetrics(ep track.ExtendedPeerstore) {
	// to prevent the Filler from crashing (the url-service only accepts 45req/s)
	requestCounter := 0
	// Loop over the Peers on the GossipMetrics
	c.GossipMetrics.Range(func(key interface{}, value interface{}) bool {
		// Read the info that we have from him
		p, ok := c.GossipMetrics.Load(key)
		if ok {
			Peer := p.(Peer)
			peerData := ep.GetAllData(Peer.PeerId)
			if len(Peer.NodeId) == 0 {
				Peer.NodeId = peerData.NodeID.String()
			}

			if len(Peer.UserAgent) == 0 {
				Peer.UserAgent = peerData.UserAgent
				client, version := utils.FilterClientType(Peer.UserAgent)
				// OperatingSystem TODO:
				Peer.ClientName = client
				Peer.ClientVersion = version
			}

			if len(Peer.Pubkey) == 0 {
				Peer.Pubkey = peerData.Pubkey
			}

			if len(Peer.Addrs) == 0 {
				address := GetFullAddress(peerData.Addrs)
				Peer.Addrs = address
			}
			if len(Peer.Country) == 0 {
				if len(Peer.Addrs) == 0 {
					//fmt.Println("No Addrs on the Peer to request the Location")
				} else {
					//fmt.Println("Requesting the Location based on the addrs:", Peer.Addrs)
					ip, country, city := utils.GetIpAndLocationFromAddrs(Peer.Addrs)
					requestCounter = requestCounter + 1
					Peer.Ip = ip
					Peer.Country = country
					Peer.City = city
				}
			}

			// Since we want to have the latest Latency, we update it only when it is different from 0
			// latency in seconds
			if peerData.Latency != 0 {
				Peer.Latency = float64(peerData.Latency/time.Millisecond) / 1000
			}

			// After check that all the info is ready, save the item back into the Sync.Map
			c.GossipMetrics.Store(key, Peer)

			/*
				if requestCounter >= 40 { // Reminder 45 req/s
					time.Sleep(70 * time.Second)
					requestCounter = 0
				}
			*/
		}
		// Keep with the loop on the Range function
		return true
	})

}

// Function that Exports the entire Metrics to a .json file (lets see if in the future we can add websockets or other implementations)
func (c *GossipMetrics) ExportMetrics(filePath string, peerstorePath string, csvPath string, ep track.ExtendedPeerstore) error {
	// Generate the MetricsDataFrame of the Current Metrics
	// Export the metrics to the given CSV file
	err := c.ExportToCSV(csvPath)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	metrics, err := c.MarshalMetrics()
	if err != nil {
		fmt.Println("Error Marshalling the metrics")
	}
	peerstore, err := c.MarshalPeerStore(ep)
	if err != nil {
		fmt.Println("Error Marshalling the peerstore")
	}

	err = ioutil.WriteFile(filePath, metrics, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", filePath)
		return err
	}
	err = ioutil.WriteFile(peerstorePath, peerstore, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", peerstorePath)
		return err
	}
	return nil
}

func (c *GossipMetrics) ExportToCSV(filePath string) error {
	csvFile, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "Error Opening the file")
	}
	defer csvFile.Close()

	// First raw of the file will be the Titles of the columns
	_, err = csvFile.WriteString("Peer Id,Node Id,User Agent,Client,Version,Pubkey,Address,Ip,Country,City,Request Metadata,Success Metadata,Attempted,Succeed,Connected,Attempts,Error,Latency,Connections,Disconnections,Connected Time,Beacon Blocks,Beacon Aggregations,Voluntary Exits,Proposer Slashings,Attester Slashings,Total Messages\n")
	if err != nil {
		errors.Wrap(err, "Error while Writing the Titles on the csv")
	}

	err = nil
	c.GossipMetrics.Range(func(k, val interface{}) bool {
		v := val.(Peer)
		_, err = csvFile.WriteString(v.ToCsvLine())
		return true
	})

	if err != nil {
		return errors.Wrap(err, "could not export peer metrics")
	}

	return nil
}

// Add new peer with all the information from the peerstore to the metrics db
// returns: Alredy (Bool)
func (c *GossipMetrics) AddNewPeer(peerId peer.ID) bool {
	_, ok := c.GossipMetrics.Load(peerId)
	if !ok {
		// We will just add the info that we have (the peerId)
		Peer := NewPeer(peerId)
		// Include it to the Peer DB
		c.GossipMetrics.Store(peerId, Peer)
		// return that wasn't already on the peerstore
		return false
	}
	return true
}

// Add a connection Event to the given peer
func (c *GossipMetrics) AddConnectionEvent(peerId peer.ID, connectionType string) {
	pMetrics, ok := c.GossipMetrics.Load(peerId)
	if ok {
		currTime := GetTimeMiliseconds()
		newConnection := ConnectionEvents{
			ConnectionType: connectionType,
			TimeMili:       currTime,
		}
		Peer := pMetrics.(Peer)
		Peer.ConnectionEvents = append(Peer.ConnectionEvents, newConnection)
		if connectionType == "Connection" {
			Peer.Connected = true
			// add the connections
			Peer.TotConnections += 1
			Peer.LastConn = currTime // Current time in SECs
			Peer.ConnFlag = true
		} else {
			// add the disconnections and sum the time
			Peer.TotDisconnections += 1
			Peer.LastDisconn = currTime // Current time in SECs
			Peer.TotConnTime += Peer.LastDisconn - Peer.LastConn
			Peer.ConnFlag = false
		}
		c.GossipMetrics.Store(peerId, Peer)
	} else {
		// Might be possible to add
		fmt.Println("Counld't add Event, Peer is not in the list")
	}
}

// Add a connection Event to the given peer
func (c *GossipMetrics) AddMetadataEvent(peerId peer.ID, success bool) {
	pMetrics, ok := c.GossipMetrics.Load(peerId)
	if ok {
		Peer := pMetrics.(Peer)
		Peer.MetadataRequest = true
		if success {
			Peer.MetadataSucceed = true
		}
		c.GossipMetrics.Store(peerId, Peer)
	} else {
		// Might be possible to add
		fmt.Println("Counld't add Event, Peer is not in the list")
	}
}

// Function that Manages the metrics updates for the incoming messages
func (c *GossipMetrics) IncomingMessageManager(peerId peer.ID, topicName string) error {
	pMetrics, _ := c.GossipMetrics.Load(peerId)
	Peer := pMetrics.(Peer)
	messageMetrics, err := GetMessageMetrics(&Peer, topicName)
	if err != nil {
		return errors.New("Topic Name no supported")
	}
	if messageMetrics.Cnt == 0 {
		messageMetrics.StampTime("first")
	}

	messageMetrics.IncrementCnt()
	messageMetrics.StampTime("last")

	// Store back the Loaded/Modified Variable
	c.GossipMetrics.Store(peerId, Peer)

	return nil
}

// AddNewAttempts adds the resuts of a new attempt over an existing peer
// increasing the attempt counter and the respective fields
func (gm *GossipMetrics) AddNewConnectionAttempt(id peer.ID, succeed bool, err string) error {
	v, ok := gm.GossipMetrics.Load(id)
	if !ok { // the peer was already in the sync.Map return true
		return fmt.Errorf("Not peer found with that ID %s", id.String())
	}
	// Update the counter and connection status
	p := v.(Peer)

	if !p.Attempted {
		p.Attempted = true
		//fmt.Println("Original ", err)
		// MIGHT be nice to try if we can change the uncertain errors for the dial backoff
		if err != "" || err != "dial backoff" {
			p.Error = FilterError(err)
		}
	}
	if succeed {
		p.Succeed = succeed
		p.Error = "None"
	}
	p.Attempts += 1

	// Store the new struct in the sync.Map
	gm.GossipMetrics.Store(id, p)
	return nil
}

// AddNewAttempts adds the resuts of a new attempt over an existing peer
// increasing the attempt counter and the respective fields
func (gm *GossipMetrics) AddNewConnection(id peer.ID) error {
	v, ok := gm.GossipMetrics.Load(id)
	if !ok { // the peer was already in the sync.Map return true
		return fmt.Errorf("Not peer found with that ID %s", id.String())
	}
	// Update the counter and connection status
	p := v.(Peer)

	p.Connected = true

	// Store the new struct in the sync.Map
	gm.GossipMetrics.Store(id, p)
	return nil
}

// CheckIdConnected check if the given peer was already connected
// returning true if it was connected before or false if wasn't
func (gm *GossipMetrics) CheckIfConnected(id peer.ID) bool {
	v, ok := gm.GossipMetrics.Load(id)
	if !ok { // the peer was already in the sync.Map we didn't connect the peer -> false
		return false
	}
	// Check if the peer was connected
	p := v.(Peer)
	if p.Succeed {
		return true
	} else {
		return false
	}
}

// GetConnectionsMetrics returns the analysis over the peers found in the
// ExtraMetrics. Return Values = (0)->succeed | (1)->failed | (2)->notattempted
func (gm *GossipMetrics) GetConnectionMetrics(h host.Host) (int, int, int) {
	totalrecorded := 0
	succeed := 0
	failed := 0
	notattempted := 0
	// Read from the recorded ExtraMetrics the status of each peer connections
	gm.GossipMetrics.Range(func(key interface{}, value interface{}) bool {
		p := value.(Peer)
		totalrecorded += 1
		// Catalog each of the peers for the experienced status
		if p.Attempted {
			if p.Succeed {
				succeed += 1
			} else {
				failed += 1
			}
		} else {
			notattempted += 1
		}
		return true
	})
	// get the len of the peerstore to complete the number of notattempted peers
	peerList := h.Peerstore().Peers()
	peerstoreLen := len(peerList)
	notattempted = notattempted + (peerstoreLen - totalrecorded)
	// MAYBE -> include here the error reader?
	return succeed, failed, notattempted
}

// GetConnectionsMetrics returns the analysis over the peers found in the ExtraMetrics.
// Return Values = (0)->resetbypeer | (1)->timeout | (2)->dialtoself | (3)->dialbackoff | (4)->uncertain
func (gm *GossipMetrics) GetErrorCounter(h host.Host) (int, int, int, int, int) {
	totalfailed := 0
	dialbackoff := 0
	timeout := 0
	resetbypeer := 0
	dialtoself := 0
	uncertain := 0
	// Read from the recorded ExtraMetrics the status of each peer connections
	gm.GossipMetrics.Range(func(key interface{}, value interface{}) bool {
		p := value.(Peer)
		// Catalog each of the peers for the experienced status
		if p.Attempted && !p.Succeed { // atempted and failed should have generated an error
			erro := p.Error
			totalfailed += 1
			switch erro {
			case "Connection reset by peer":
				resetbypeer += 1
			case "i/o timeout":
				timeout += 1
			case "dial to self attempted":
				dialtoself += 1
			case "dial backoff":
				dialbackoff += 1
			case "Uncertain":
				uncertain += 1
			default:
				fmt.Println("The recorded error type doesn't match any of the error on the list", erro)
			}
		}
		return true
	})
	return resetbypeer, timeout, dialtoself, dialbackoff, uncertain
}

// funtion that formats the error into a Pretty understandable (standard) way
// also important to cohesionate the extra-metrics output csv
func FilterError(err string) string {
	errorPretty := "Uncertain"
	// filter the error type
	if strings.Contains(err, "connection reset by peer") {
		errorPretty = "Connection reset by peer"
	} else if strings.Contains(err, "i/o timeout") {
		errorPretty = "i/o timeout"
	} else if strings.Contains(err, "dial to self attempted") {
		errorPretty = "dial to self attempted"
	} else if strings.Contains(err, "dial backoff") {
		errorPretty = "dial backoff"
	}

	return errorPretty
}

func GetMessageMetrics(c *Peer, topicName string) (mesMetr *MessageMetrics, err error) {
	// All this could be inside a different function
	switch topicName {
	case pgossip.BeaconBlock:
		return &c.BeaconBlock, nil
	case pgossip.BeaconAggregateProof:
		return &c.BeaconAggregateProof, nil
	case pgossip.VoluntaryExit:
		return &c.VoluntaryExit, nil
	case pgossip.ProposerSlashing:
		return &c.ProposerSlashing, nil
	case pgossip.AttesterSlashing:
		return &c.AttesterSlashing, nil
	default: //TODO: - Not returning BeaconBlock as Default
		return &c.BeaconBlock, err
	}
}
