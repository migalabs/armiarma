package db

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// logging variables
	ModuleName = "PEERSTORE"
	Log        = logrus.WithField(
		"module", ModuleName,
	)

	// TODO: put it into a config-variable
	ExportLoopTime time.Duration = 1 * time.Minute
	// DB config-options (TODO: unnecessary so far, we just have 2 of them)
	BoltDBKey string            = "BoltDB"
	MemoryKey string            = "Memory"
	DBTypes   map[string]string = map[string]string{
		BoltDBKey: "bolt",
		MemoryKey: "memory",
	}
)

type ErrorHandling func(*Peer)

type PeerStore struct {
	// control variables for the exporting routines
	ctx    context.Context
	cancel context.CancelFunc

	PeerStore PeerStoreStorage
	// MessageDatabase   *database.MessageDatabase // TODO: Discuss
	StartTime      time.Time
	MsgNotChannels map[string](chan bool) // TODO: Unused?
}

func NewPeerStore(ctx context.Context, dbtype string, path string) PeerStore {
	mainCtx, cancel := context.WithCancel(ctx)
	var db PeerStoreStorage

	switch dbtype {
	case DBTypes[BoltDBKey]:
		if len(path) <= 0 {
			path = default_db_path
		}
		db = NewBoltPeerDB(path)
	case DBTypes[MemoryKey]:
		db = NewMemoryDB()
	default:
		if len(path) <= 0 {
			path = default_db_path
		}
		db = NewBoltPeerDB(path)
	}
	ps := PeerStore{
		ctx:            mainCtx,
		cancel:         cancel,
		PeerStore:      db,
		StartTime:      time.Now(),
		MsgNotChannels: make(map[string](chan bool)),
	}
	return ps
}

// Ctx
// Retreives the context asigned to the Peerstore
func (c *PeerStore) Ctx() context.Context {
	return c.ctx
}

// GetCtxCancel
// Retreives the cancel function to kill the Peerstore ctx
func (c *PeerStore) GetCtxCancel() context.CancelFunc {
	return c.cancel
}

// TODO: ImportPeerStoreMetrics
func (c *PeerStore) ImportPeerStoreMetrics(importFolder string) error {
	// TODO: Load to memory an existing csv
	// Perhaps not needed since we are migrating to a database
	// See: https://github.com/migalabs/armiarma/pull/18

	return nil
}

// TODO: Review if needed ResetDynamicMetrics
// Function that resets to 0 the connections/disconnections, and message counters
// this way the Ram Usage gets limited (up to ~10k nodes for a 12h-24h )
// NOTE: Keep in mind that the peers that we ended up connected to, will experience a weid connection time
// TODO: Fix peers that stayed connected to the tool
func (c *PeerStore) ResetDynamicMetrics() {
	Log.Info("Reseting Dynamic Metrics in Peer")

	// Iterate throught the peers in the metrics, restarting connection events and messages
	c.PeerStore.Range(func(key string, value Peer) bool {
		value.ResetDynamicMetrics()
		c.PeerStore.Store(key, value)
		return true
	})
	Log.Info("Finished Reseting Dynamic Metrics")
}

// AddNotChannel:
// Function that adds a notification channel to the message gossip topic.
// @param topicName: the name of the topic used to identify the channel in the map.
func (c *PeerStore) AddNotChannel(topicName string) {
	c.MsgNotChannels[topicName] = make(chan bool, 100)
}

// StoreOrUpdatePeer:
// Updates the peer without overwritting all its content.
// If peer exists, aggregate data to the existing peer.
// Otherwise, store the peer.
// @param peer: the peer to store or update
func (c *PeerStore) StoreOrUpdatePeer(peer Peer) {

	oldPeer, err := c.GetPeerData(peer.PeerId)
	// if error means not found, just store it
	if err != nil {
		c.PeerStore.Store(peer.PeerId, peer)
	} else {
		// Fetch the new info of a peer directly from the new peer struct
		oldPeer.FetchPeerInfoFromNewPeer(peer)
		c.PeerStore.Store(peer.PeerId, oldPeer)
	}
	// Force Garbage collector
	runtime.GC()
}

// StorePeer:
// This method stores a single peer in the peerstore.
// It will use the peerID as key.
// @param peer: the peer object to store.
func (c *PeerStore) StorePeer(peer Peer) {
	c.PeerStore.Store(peer.PeerId, peer)
}

// GetPeerData:
// This method return a Peer object from the peerstore
// using the given peerID.
// @param peerID: the peerID to look for in string format.
// @return the found Peer object and an error if there was.
func (c *PeerStore) GetPeerData(peerId string) (Peer, error) {
	peerData, ok := c.PeerStore.Load(peerId)
	if !ok {
		return Peer{}, errors.New("could not find peer in peerstore or peer was unable to unmarshal: " + peerId)
	}

	return peerData, nil
}

// GetPeerList:
// This method returns the list of PeerIDs in the DB.
// @return the list of PeerIDs in string format.
func (c *PeerStore) GetPeerList() []peer.ID {
	return c.PeerStore.Peers()
}

// GetENR
// Returns the Node after parsing the ENR.
// @param peerID: the peerID of which to get the Node.
// @return the resulting Node.
// @return error if applicable, nil in any other case.
func (c *PeerStore) GetENR(peerID string) (*enode.Node, error) {
	p, err := c.GetPeerData(peerID)
	if err != nil {
		return nil, err
	}
	return p.GetBlockchainNode()
}

// TODO: review if needed
// GetErrorCounter:
// Get a map with the errors we got when connecting and their amount.
// @return a map where string is the errorName and value is the count.
func (gm *PeerStore) GetErrorCounter() map[string]uint64 {
	errorsAndAmount := make(map[string]uint64)
	gm.PeerStore.Range(func(key string, value Peer) bool {
		for _, errTmp := range value.Error {
			errorsAndAmount[errTmp]++
		}

		return true
	})

	return errorsAndAmount
}

// ExportToCSV
// This method will export the whole peerstore into a CSV file.
// @param filePath file where to dump the CSV lines (create if it does not exist).
// @return an error if there was.
func (c *PeerStore) ExportToCSV(filePath string) error {
	Log.Info("Exporting metrics to csv: ", filePath)
	csvFile, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "error opening the file "+filePath)
	}
	defer csvFile.Close()

	// First raw of the file will be the Titles of the columns
	_, err = csvFile.WriteString("Peer Id,Node Id,Fork Digest,User Agent,Client,Version,Pubkey,Address,Ip,Country,City,ENR,Request Metadata,Success Metadata,Attempted,Succeed,Deprecated,ConnStablished,IsConnected,Attempts,Error,Last Error Timestamp,Last Identify Timestamp,Latency,Connections,Disconnections,Last Connection,Conn Direction,Connected Time,Beacon Blocks,Beacon Aggregations,Voluntary Exits,Proposer Slashings,Attester Slashings,Total Messages\n")
	if err != nil {
		errors.Wrap(err, "error while writing the titles on the csv "+filePath)
	}

	err = nil
	c.PeerStore.Range(func(key string, value Peer) bool {
		_, err = csvFile.WriteString(value.ToCsvLine())
		return true
	})

	if err != nil {
		return errors.Wrap(err, "could not export peer metrics")
	}
	// Force Garbage collector
	runtime.GC()
	return nil
}

// ExportCsvService
// We will export to csv regularly, therefoe this service will execute the export every X seconds (ExportLoopTime)
// @param folderpath: the folder to export the csv file (always named metrics.csv)
func (ps *PeerStore) ExportCsvService(folderpath string) {
	Log.Info("Peerstore CSV exporting service launched")
	go func() {
		ctx := ps.Ctx()
		ticker := time.NewTicker(ExportLoopTime)
		for {
			select {
			case <-ticker.C:
				ps.ExportToCSV(folderpath + "/metrics.csv")
			case <-ctx.Done():
				ticker.Stop()
				Log.Info("Closing DB CSV exporter")
				return
			}
		}
	}()
}

// CloseMetricsExport: closes any go routine related with the DB metrics
func (ps *PeerStore) CloseMetricsExport() {
	Log.Info("closing metrics exporting services")
	cancel := ps.GetCtxCancel()
	cancel()
}
