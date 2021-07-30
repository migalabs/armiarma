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
			tempMap := make(map[peer.ID]PeerMetrics)
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
	tmpMap := make(map[string]PeerMetrics)
	c.GossipMetrics.Range(func(k, v interface{}) bool {
		pm := v.(PeerMetrics)
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
		p := value.(PeerMetrics)
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

// Function that iterates through the received peers and fills the missing information
func (c *GossipMetrics) FillMetrics(ep track.ExtendedPeerstore) {
	// to prevent the Filler from crashing (the url-service only accepts 45req/s)
	requestCounter := 0
	// Loop over the Peers on the GossipMetrics
	c.GossipMetrics.Range(func(key interface{}, value interface{}) bool {
		// Read the info that we have from him
		p, ok := c.GossipMetrics.Load(key)
		if ok {
			peerMetrics := p.(PeerMetrics)
			peerData := ep.GetAllData(peerMetrics.PeerId)
			//fmt.Println("Filling Metrics of Peer:", peerMetrics.PeerId.String())
			if len(peerMetrics.NodeId) == 0 {
				//fmt.Println("NodeID empty", peerMetrics.NodeId, "Adding NodeId:", peerData.NodeID.String())
				peerMetrics.NodeId = peerData.NodeID.String()
			}

 			// TODO: Why does it matter that is "" or "Unknown". What is the difference.
			if len(peerMetrics.UserAgent) == 0 || peerMetrics.UserAgent == "Unknown" {
				//fmt.Println("ClientType empty", peerMetrics.ClientType, "Adding ClientType:", peerData.UserAgent)
				peerMetrics.UserAgent = peerData.UserAgent
			}

			if len(peerMetrics.Pubkey) == 0 {
				//fmt.Println("Pubkey empty", peerMetrics.Pubkey, "Adding Pubkey:", peerData.Pubkey)
				peerMetrics.Pubkey = peerData.Pubkey
			}

			if len(peerMetrics.Addrs) == 0 || peerMetrics.Addrs == "/ip4/127.0.0.1/0000" || peerMetrics.Addrs == "/ip4/127.0.0.1/9000" {
				address := GetFullAddress(peerData.Addrs)
				//fmt.Println("Addrs empty", peerMetrics.Addrs, "Adding Addrs:", address)
				peerMetrics.Addrs = address
			}

			if len(peerMetrics.Country) == 0 || peerMetrics.Country == "Unknown" {
				if len(peerMetrics.Addrs) == 0 {
					//fmt.Println("No Addrs on the PeerMetrics to request the Location")
				} else {
					//fmt.Println("Requesting the Location based on the addrs:", peerMetrics.Addrs)
					ip, country, city := utils.GetIpAndLocationFromAddrs(peerMetrics.Addrs)
					requestCounter = requestCounter + 1
					peerMetrics.Ip = ip
					peerMetrics.Country = country
					peerMetrics.City = city
				}
			}

			// Since we want to have the latest Latency, we update it only when it is different from 0
			// latency in seconds
			if peerData.Latency != 0 {
				peerMetrics.Latency = float64(peerData.Latency/time.Millisecond) / 1000
			}

			// After check that all the info is ready, save the item back into the Sync.Map
			c.GossipMetrics.Store(key, peerMetrics)

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
	fmt.Println("---- exporting metrics")
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
		v := val.(PeerMetrics)
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
		peerMetrics := NewPeerMetrics(peerId)
		// Include it to the Peer DB
		c.GossipMetrics.Store(peerId, peerMetrics)
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
		peerMetrics := pMetrics.(PeerMetrics)
		peerMetrics.ConnectionEvents = append(peerMetrics.ConnectionEvents, newConnection)
		if connectionType == "Connection" {
			peerMetrics.Connected = true
			// add the connections
			peerMetrics.TotConnections += 1
			peerMetrics.LastConn = currTime // Current time in SECs
			peerMetrics.ConnFlag = true
		} else {
			// add the disconnections and sum the time
			peerMetrics.TotDisconnections += 1
			peerMetrics.LastDisconn = currTime // Current time in SECs
			peerMetrics.TotConnTime += peerMetrics.LastDisconn - peerMetrics.LastConn
			peerMetrics.ConnFlag = false
		}
		c.GossipMetrics.Store(peerId, peerMetrics)
	} else {
		// Might be possible to add
		fmt.Println("Counld't add Event, Peer is not in the list")
	}
}

// Add a connection Event to the given peer
func (c *GossipMetrics) AddMetadataEvent(peerId peer.ID, success bool) {
	pMetrics, ok := c.GossipMetrics.Load(peerId)
	if ok {
		peerMetrics := pMetrics.(PeerMetrics)
		peerMetrics.MetadataRequest = true
		if success {
			peerMetrics.MetadataSucceed = true
		}
		c.GossipMetrics.Store(peerId, peerMetrics)
	} else {
		// Might be possible to add
		fmt.Println("Counld't add Event, Peer is not in the list")
	}
}

// Function that Manages the metrics updates for the incoming messages
func (c *GossipMetrics) IncomingMessageManager(peerId peer.ID, topicName string) error {
	pMetrics, _ := c.GossipMetrics.Load(peerId)
	peerMetrics := pMetrics.(PeerMetrics)
	messageMetrics, err := GetMessageMetrics(&peerMetrics, topicName)
	if err != nil {
		return errors.New("Topic Name no supported")
	}
	if messageMetrics.Cnt == 0 {
		messageMetrics.StampTime("first")
	}

	messageMetrics.IncrementCnt()
	messageMetrics.StampTime("last")

	// Store back the Loaded/Modified Variable
	c.GossipMetrics.Store(peerId, peerMetrics)

	return nil
}

func GetMessageMetrics(c *PeerMetrics, topicName string) (mesMetr *MessageMetrics, err error) {
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
