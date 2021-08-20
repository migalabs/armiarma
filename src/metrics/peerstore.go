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
	"github.com/protolambda/rumor/metrics/utils"

	"github.com/libp2p/go-libp2p-core/peer"
	//"github.com/protolambda/rumor/metrics/utils"
	pgossip "github.com/protolambda/rumor/p2p/gossip"
	"github.com/protolambda/rumor/p2p/gossip/database"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/libp2p/go-libp2p-core/host"
	log "github.com/sirupsen/logrus"
)

type PeerStore struct {
	PeerStore   sync.Map
	PeerCount   int
	// TODO: Connected Peers? Directly populated by h.Network().Peers() for the client
	MessageDatabase *database.MessageDatabase
	StartTime       int64 // milliseconds
	MsgNotChannels  map[string](chan bool)
}

func NewPeerStore() PeerStore {
	gm := PeerStore{
		StartTime:      utils.GetTimeMiliseconds(),
		MsgNotChannels: make(map[string](chan bool)),
	}
	return gm
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

// Import an old PeerStore from given file
// return: - return error if there was error while reading the file
//         - return bool for existing file (true if there was a file to read, return false if there wasn't a file to read)
func (c *PeerStore) ImportMetrics(importFolder string) (error, bool) {
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
			// iterate to add the metrics from the json to the the PeerStore
			for k, v := range tempMap {
				if v.ConnFlag {
					v.TotDisconnections += 1
					v.LastDisconn = v.LastExport
					v.TotConnTime += (v.LastExport - v.LastConn)
					v.ConnFlag = false
				}
				c.PeerStore.Store(k, v)
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
func (c *PeerStore) MarshalMetrics() ([]byte, error) {
	/*
	exportTime := utils.GetTimeMiliseconds()
	tmpMap := make(map[string]Peer)
	c.PeerStore.Range(func(k, v interface{}) bool {
		pm := v.(Peer)
		if pm.ConnFlag {
			pm.LastExport = exportTime
		}
		tmpMap[k.(peer.ID).String()] = pm
		return true
	})

	return json.Marshal(tmpMap)
	*/
	return nil, nil
}

// Function that Wraps/Marshals the content of the Entire Peerstore into a json
func (c *PeerStore) MarshalPeerStore(ep track.ExtendedPeerstore) ([]byte, error) {
	peers := ep.Peers()
	peerData := make(map[string]*track.PeerAllData)
	for _, p := range peers {
		peerData[p.String()] = ep.GetAllData(p)
	}
	return json.Marshal(peerData)
}

// Function that resets to 0 the connections/disconnections, and message counters
// this way the Ram Usage gets limited (up to ~10k nodes for a 12h-24h )
// NOTE: Keep in mind that the peers that we ended up connected to, will experience a weid connection time
// TODO: Fix peers that stayed connected to the tool
func (c *PeerStore) ResetDynamicMetrics() {
	fmt.Println("Reseting Dynamic Metrics in Peer")
	// Iterate throught the peers in the metrics, restarting connection events and messages
	c.PeerStore.Range(func(key interface{}, value interface{}) bool {
		p := value.(Peer)
		p.ResetDynamicMetrics()
		c.PeerStore.Store(key, p)
		return true
	})
	fmt.Println("Finished Reseting Dynamic Metrics")
}

// Function that adds a notification channel to the message gossip topic
func (c *PeerStore) AddNotChannel(topicName string) {
	c.MsgNotChannels[topicName] = make(chan bool, 100)
}

// Adds or updates peer
func (c *PeerStore) AddPeer(peer Peer) {
	oldData, loaded := c.PeerStore.LoadOrStore(peer.PeerId, peer)

	// If already present
	if loaded {
		// TODO: We could also store the old data if there was a change. For example
		// if a given client upgrated it version. Use oldData
		// See: https://github.com/migalabs/armiarma/issues/17
		// Currently just overwritting what was before
		_ = oldData // TODO:
		c.PeerStore.Store(peer.PeerId, peer)
	}

	c.PeerCount++
}

func (c *PeerStore) GetPeerData(peerId string) (Peer, bool) {
	peerData, ok := c.PeerStore.Load(peerId)
	return peerData.(Peer), ok
}

// Add new peer with all the information from the peerstore to the metrics db
// returns: Alredy (Bool)
// TODO: Join AddPeer and AddNewPeer?
func (c *PeerStore) AddNewPeer(peerId string) bool {
	_, ok := c.PeerStore.Load(peerId)
	if !ok {
		// We will just add the info that we have (the peerId)
		Peer := NewPeer(peerId)
		// Include it to the Peer DB
		c.PeerStore.Store(peerId, Peer)
		// return that wasn't already on the peerstore
		return false
	}
	return true
}

// Add a connection Event to the given peer
func (c *PeerStore) AddConnectionEvent(peerId string, connectionType string) {
	pMetrics, ok := c.PeerStore.Load(peerId)
	if ok {
		currTime := utils.GetTimeMiliseconds()
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
		c.PeerStore.Store(peerId, Peer)
	} else {
		// Might be possible to add
		fmt.Println("Counld't add Event, Peer is not in the list")
	}
}

// Add a connection Event to the given peer
func (c *PeerStore) AddMetadataEvent(peerId string, success bool) {
	pMetrics, ok := c.PeerStore.Load(peerId)
	if ok {
		Peer := pMetrics.(Peer)
		Peer.MetadataRequest = true
		if success {
			Peer.MetadataSucceed = true
		}
		c.PeerStore.Store(peerId, Peer)
	} else {
		// Might be possible to add
		fmt.Println("Counld't add Event, Peer is not in the list")
	}
}

// AddNewAttempts adds the resuts of a new attempt over an existing peer
// increasing the attempt counter and the respective fields
func (gm *PeerStore) AddNewConnectionAttempt(id peer.ID, succeed bool, err string) error {
	v, ok := gm.PeerStore.Load(id)
	if !ok { // the peer was already in the sync.Map return true
		return fmt.Errorf("Not peer found with that ID %s", id.String())
	}
	// Update the counter and connection status
	p := v.(Peer)

	if !p.Attempted {
		p.Attempted = true
		//fmt.Println("Original ", err)
		// MIGHT be nice to try if we can change the uncertain errors for the dial backoff
		if err != "dial backoff" {
			p.Error = FilterError(err)
		}
	}
	if succeed {
		p.Succeed = succeed
		p.Error = "None"
	}
	p.Attempts += 1

	// Store the new struct in the sync.Map
	gm.PeerStore.Store(id, p)
	return nil
}

// AddNewAttempts adds the resuts of a new attempt over an existing peer
// increasing the attempt counter and the respective fields
func (gm *PeerStore) AddNewConnection(id peer.ID) error {
	v, ok := gm.PeerStore.Load(id)
	if !ok { // the peer was already in the sync.Map return true
		return fmt.Errorf("Not peer found with that ID %s", id.String())
	}
	// Update the counter and connection status
	p := v.(Peer)

	p.Connected = true

	// Store the new struct in the sync.Map
	gm.PeerStore.Store(id, p)
	return nil
}

// Exports to a csv, useful for debug
func (c *PeerStore) ExportToCSV(filePath string) error {
	log.Info("Exporting metrics to csv: ", filePath)
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
	c.PeerStore.Range(func(k, val interface{}) bool {
		v := val.(Peer)
		_, err = csvFile.WriteString(v.ToCsvLine())
		return true
	})

	if err != nil {
		return errors.Wrap(err, "could not export peer metrics")
	}

	return nil
}

// Function that Manages the metrics updates for the incoming messages
func (c *PeerStore) IncomingMessageManager(peerId string, topicName string) error {
	pMetrics, _ := c.PeerStore.Load(peerId)
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
	c.PeerStore.Store(peerId, Peer)

	return nil
}

// CheckIdConnected check if the given peer was already connected
// returning true if it was connected before or false if wasn't
func (gm *PeerStore) CheckIfConnected(id peer.ID) bool {
	v, ok := gm.PeerStore.Load(id)
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
func (gm *PeerStore) GetConnectionMetrics(h host.Host) (int, int, int) {
	totalrecorded := 0
	succeed := 0
	failed := 0
	notattempted := 0
	// Read from the recorded ExtraMetrics the status of each peer connections
	gm.PeerStore.Range(func(key interface{}, value interface{}) bool {
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
func (gm *PeerStore) GetErrorCounter(h host.Host) (int, int, int, int, int) {
	totalfailed := 0
	dialbackoff := 0
	timeout := 0
	resetbypeer := 0
	dialtoself := 0
	uncertain := 0
	// Read from the recorded ExtraMetrics the status of each peer connections
	gm.PeerStore.Range(func(key interface{}, value interface{}) bool {
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
