package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	pb "github.com/protolambda/rumor/proto"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/rumor/metrics/export"
	server "github.com/protolambda/rumor/metrics/export/serverexporter"
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
	ArmiarmaServer  *server.ServerEndpoint
}

func NewGossipMetrics() GossipMetrics {
	// TODO: Just a proof of concept. This has to be moved
	// to a more appropriate place
	// TODO Handle if dial doesnt work, time out

	gm := GossipMetrics{
		StartTime:      utils.GetTimeMiliseconds(),
		MsgNotChannels: make(map[string](chan bool)),
		ArmiarmaServer: nil,
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
			tempMap := make(map[peer.ID]utils.PeerMetrics)
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
	exportTime := utils.GetTimeMiliseconds()
	tmpMap := make(map[string]utils.PeerMetrics)
	c.GossipMetrics.Range(func(k, v interface{}) bool {
		pm := v.(utils.PeerMetrics)
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
		p := value.(utils.PeerMetrics)
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
func (c *GossipMetrics) FillMetrics(ep track.ExtendedPeerstore, h host.Host) {
	// to prevent the Filler from crashing (the url-service only accepts 45req/s)
	requestCounter := 0
	hostID := h.ID()
	// Loop over the Peers on the GossipMetrics
	c.GossipMetrics.Range(func(key interface{}, value interface{}) bool {
		newMetaFlag := false
		msg := &pb.NewPeerMetadataRequest{
			CrawlerId:  hostID.String(),
			PeerId:     key.(peer.ID).String(),
			NodeId:     "",
			ClientType: "",
			PubKey:     "",
			Address:    "",
			Country:    "",
			City:       "",
			Latency:    "",
		}
		// Read the info that we have from him
		p, ok := c.GossipMetrics.Load(key)
		if ok {
			peerMetrics := p.(utils.PeerMetrics)
			peerData := ep.GetAllData(peerMetrics.PeerId)
			//fmt.Println("Filling Metrics of Peer:", peerMetrics.PeerId.String())
			if len(peerMetrics.NodeId) == 0 {
				//fmt.Println("NodeID empty", peerMetrics.NodeId, "Adding NodeId:", peerData.NodeID.String())
				peerMetrics.NodeId = peerData.NodeID.String()
				newMetaFlag = true
				msg.NodeId = peerData.NodeID.String()
			}
			if len(peerMetrics.ClientType) == 0 || peerMetrics.ClientType == "Unknown" {
				//fmt.Println("ClientType empty", peerMetrics.ClientType, "Adding ClientType:", peerData.UserAgent)
				peerMetrics.ClientType = peerData.UserAgent
				newMetaFlag = true
				msg.ClientType = peerData.UserAgent
			}
			if len(peerMetrics.Pubkey) == 0 {
				//fmt.Println("Pubkey empty", peerMetrics.Pubkey, "Adding Pubkey:", peerData.Pubkey)
				peerMetrics.Pubkey = peerData.Pubkey
				newMetaFlag = true
				msg.PubKey = peerData.Pubkey
			}
			if len(peerMetrics.Addrs) == 0 || peerMetrics.Addrs == "/ip4/127.0.0.1/0000" || peerMetrics.Addrs == "/ip4/127.0.0.1/9000" {
				address := GetFullAddress(peerData.Addrs)
				//fmt.Println("Addrs empty", peerMetrics.Addrs, "Adding Addrs:", address)
				peerMetrics.Addrs = address
				newMetaFlag = true
				msg.Address = address
			}
			if len(peerMetrics.Country) == 0 || peerMetrics.Country == "Unknown" {
				if len(peerMetrics.Addrs) == 0 {
					//fmt.Println("No Addrs on the PeerMetrics to request the Location")
				} else {
					//fmt.Println("Requesting the Location based on the addrs:", peerMetrics.Addrs)
					ip, country, city := getIpAndLocationFromAddrs(peerMetrics.Addrs)
					requestCounter = requestCounter + 1
					peerMetrics.Ip = ip
					peerMetrics.Country = country
					peerMetrics.City = city
					newMetaFlag = true
					msg.Country = country
					msg.City = city
				}
			}
			// Since we want to have the latest Latency, we update it only when it is different from 0
			// latency in seconds
			if peerData.Latency != 0 {
				peerMetrics.Latency = float64(peerData.Latency/time.Millisecond) / 1000
				newMetaFlag = true
				msg.Latency = fmt.Sprintf("%f", float64(peerData.Latency/time.Millisecond)/1000)
			}

			// Check if there is new metadata to report to the Armiarma Server
			if newMetaFlag {
				fmt.Println("New Msg to send")
				// add the metadata fields to the msg
				// msg.MetadataRequest = peerData.MetadataRequest
				// msg.MetadataSucceed = peerData.MetadataSucced
				// notify of new message to send
				c.ArmiarmaServer.NewPeerMetadata <- msg
			}
			// After check that all the info is ready, save the item back into the Sync.Map
			c.GossipMetrics.Store(key, peerMetrics)
		}
		// Keep with the loop on the Range function
		return true
	})

}

// Function that Exports the entire Metrics to a .json file (lets see if in the future we can add websockets or other implementations)
func (c *GossipMetrics) ExportMetrics(filePath string, peerstorePath string, csvPath string, ep track.ExtendedPeerstore) error {
	// Generate the MetricsDataFrame of the Current Metrics
	// Export the metrics to the given CSV file
	mdf := export.NewMetricsDataFrame(&c.GossipMetrics)
	err := mdf.ExportToCSV(csvPath)
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

// IP-API message structure
type IpApiMessage struct {
	Query       string `json:"query"`
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	City        string `json:"city"`
	Zip         string `json:"zip"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Timezone    string `json:"timezone"`
	Isp         string `json:"isp"`
	Org         string `json:"org"`
	As          string `json:"as"`
}

// get IP, location country and City from the multiaddress of the peer on the peerstore
func getIpAndLocationFromAddrs(multiAddrs string) (ip string, country string, city string) {
	ip = strings.TrimPrefix(multiAddrs, "/ip4/")
	ipSlices := strings.Split(ip, "/")
	ip = ipSlices[0]
	url := "http://ip-api.com/json/" + ip
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		country = "Unknown"
		city = "Unknown"
		return ip, country, city
	}

	attemptsLeft, _ := strconv.Atoi(resp.Header["X-Rl"][0])
	timeLeft, _ := strconv.Atoi(resp.Header["X-Ttl"][0])

	if attemptsLeft == 0 { // We have exceeded the limit of requests 45req/min
		time.Sleep(time.Duration(timeLeft) * time.Second)
		resp, err = http.Get(url)
		if err != nil {
			fmt.Println(err)
			country = "Unknown"
			city = "Unknown"
			return ip, country, city
		}
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to Todo struct
	var ipApiResp IpApiMessage
	json.Unmarshal(bodyBytes, &ipApiResp)

	// Check if the status of the request has been succesful
	if ipApiResp.Status != "success" {
		country = "Unknown"
		city = "Unknown"
		return ip, country, city
	}

	country = ipApiResp.Country
	city = ipApiResp.City

	// check if country and city are correctly imported
	if len(country) == 0 || len(city) == 0 {
		country = "Unknown"
		city = "Unknown"
		return ip, country, city
	}

	// return the received values from the received message
	return ip, country, city

}

// Add new peer with all the information from the peerstore to the metrics db
// returns: Alredy (Bool)
func (c *GossipMetrics) AddNewPeer(peerId peer.ID) bool {
	_, ok := c.GossipMetrics.Load(peerId)
	if !ok {
		// We will just add the info that we have (the peerId)
		peerMetrics := utils.NewPeerMetrics(peerId)
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
		currTime := utils.GetTimeMiliseconds()
		newConnection := utils.ConnectionEvents{
			ConnectionType: connectionType,
			TimeMili:       currTime,
		}
		peerMetrics := pMetrics.(utils.PeerMetrics)
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
		peerMetrics := pMetrics.(utils.PeerMetrics)
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
	peerMetrics := pMetrics.(utils.PeerMetrics)
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

func GetMessageMetrics(c *utils.PeerMetrics, topicName string) (mesMetr *utils.MessageMetrics, err error) {
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
