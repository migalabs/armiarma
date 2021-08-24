package metrics

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	pgossip "github.com/protolambda/rumor/p2p/gossip"
	"github.com/protolambda/rumor/p2p/gossip/database"
	log "github.com/sirupsen/logrus"
)

type PeerStore struct {
	PeerStore       sync.Map
	MessageDatabase *database.MessageDatabase // TODO: Discuss
	StartTime       time.Time
	MsgNotChannels  map[string](chan bool) // TODO: Unused?
}

// TODO: Remove from here?
type GossipState struct {
	GsNode  pgossip.GossipSub
	CloseGS context.CancelFunc
	// string -> *pubsub.Topic
	Topics sync.Map
	// Validation Filter Flag
	SeenFilter bool
}

func NewPeerStore() PeerStore {
	gm := PeerStore{
		StartTime:      time.Now(),
		MsgNotChannels: make(map[string](chan bool)),
	}
	return gm
}

func (c *PeerStore) ImportPeerStoreMetrics(importFolder string) error {
	// TODO: Load to memory an existing csv
	// Perhaps not needed since we are migrating to a database
	// See: https://github.com/migalabs/armiarma/pull/18

	return nil
}

// Function that resets to 0 the connections/disconnections, and message counters
// this way the Ram Usage gets limited (up to ~10k nodes for a 12h-24h )
// NOTE: Keep in mind that the peers that we ended up connected to, will experience a weid connection time
// TODO: Fix peers that stayed connected to the tool
func (c *PeerStore) ResetDynamicMetrics() {
	log.Info("Reseting Dynamic Metrics in Peer")

	// Iterate throught the peers in the metrics, restarting connection events and messages
	c.PeerStore.Range(func(key interface{}, value interface{}) bool {
		p := value.(Peer)
		p.ResetDynamicMetrics()
		c.PeerStore.Store(key, p)
		return true
	})
	log.Info("Finished Reseting Dynamic Metrics")
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
}

func (c *PeerStore) GetPeerData(peerId string) (Peer, bool) {
	peerData, ok := c.PeerStore.Load(peerId)
	if !ok {
		return Peer{}, ok
	}
	return peerData.(Peer), ok
}

// Add a connection Event to the given peer
func (c *PeerStore) AddConnectionEvent(peerId string, direction string) error {
	pMetrics, ok := c.PeerStore.Load(peerId)
	if ok {
		peer := pMetrics.(Peer)
		peer.AddConnectionEvent(direction, time.Now())
		c.PeerStore.Store(peerId, peer)
		return nil
	}
	return errors.New("could not add event, peer is not in the list")
}

// Add a connection Event to the given peer
func (c *PeerStore) AddDisconnectionEvent(peerId string) error {
	pMetrics, ok := c.PeerStore.Load(peerId)
	if ok {
		peer := pMetrics.(Peer)
		peer.AddDisconnectionEvent(time.Now())
		c.PeerStore.Store(peerId, peer)
		return nil
	}
	return errors.New("could not add connection event, peer is not in the list")
}

// Add a connection Event to the given peer
func (c *PeerStore) AddMetadataEvent(peerId string, success bool) error {
	pMetrics, ok := c.PeerStore.Load(peerId)
	if ok {
		Peer := pMetrics.(Peer)
		Peer.MetadataRequest = true
		if success {
			Peer.MetadataSucceed = true
		}
		c.PeerStore.Store(peerId, Peer)
		return nil
	}
	return errors.New("counld't add Event, Peer is not in the list: " + peerId)
}

// AddNewAttempts adds the resuts of a new attempt over an existing peer
// increasing the attempt counter and the respective fields
func (gm *PeerStore) AddNewConnectionAttempt(peerId string, succeed bool, err string) error {
	pMetrics, ok := gm.PeerStore.Load(peerId)
	if ok {
		peer := pMetrics.(Peer)
		peer.AddNewConnectionAttempt(succeed, err)
		gm.PeerStore.Store(peerId, peer)
		return nil
	}
	return errors.New("could not add connection attempt, peer is not in the list")
}

// Function that Manages the metrics updates for the incoming messages
func (c *PeerStore) IncomingMessageManager(peerId string, topicName string) error {
	pMetrics, ok := c.PeerStore.Load(peerId)
	if ok {
		peer := pMetrics.(Peer)
		messageMetrics, err := peer.GetMessageMetrics(topicName)
		if err != nil {
			return errors.Wrap(err, "could not not get message metrics struct")
		}

		if messageMetrics.Count == 0 {
			messageMetrics.StampTime("first")
		}

		messageMetrics.IncrementCnt()
		messageMetrics.StampTime("last")

		c.PeerStore.Store(peerId, peer)
	} else {
		return errors.New("could not add incomming message to topics list")
	}
	return nil
}

// GetConnectionsMetrics returns the analysis over the peers found in the
// ExtraMetrics. Return Values = (0)->succeed | (1)->failed | (2)->notattempted

/* TODO: Rethink this function
func (gm *PeerStore) GetConnectionMetrics() (int, int, int) {
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
	//peerList := h.Peerstore().Peers()
	//peerstoreLen := len(peerList)
	//notattempted = notattempted + (peerstoreLen - totalrecorded)
	// MAYBE -> include here the error reader?
	return succeed, failed, notattempted
}
*/

// GetConnectionsMetrics returns the analysis over the peers found in the ExtraMetrics.
// Return Values = (0)->resetbypeer | (1)->timeout | (2)->dialtoself | (3)->dialbackoff | (4)->uncertain
func (gm *PeerStore) GetErrorCounter() map[string]uint64 {
	errorsAndAmount := make(map[string]uint64)
	gm.PeerStore.Range(func(key interface{}, value interface{}) bool {
		peer := value.(Peer)
		errorsAndAmount[peer.Error]++
		return true
	})

	return errorsAndAmount
}

// Exports to a csv, useful for debug
func (c *PeerStore) ExportToCSV(filePath string) error {
	log.Info("Exporting metrics to csv: ", filePath)
	csvFile, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "error opening the file "+filePath)
	}
	defer csvFile.Close()

	// First raw of the file will be the Titles of the columns
	_, err = csvFile.WriteString("Peer Id,Node Id,User Agent,Client,Version,Pubkey,Address,Ip,Country,City,Request Metadata,Success Metadata,Attempted,Succeed,Connected,Attempts,Error,Latency,Connections,Disconnections,Connected Time,Beacon Blocks,Beacon Aggregations,Voluntary Exits,Proposer Slashings,Attester Slashings,Total Messages\n")
	if err != nil {
		errors.Wrap(err, "error while writing the titles on the csv "+filePath)
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
