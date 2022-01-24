package db

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/migalabs/armiarma/src/db/models"
	postgresql "github.com/migalabs/armiarma/src/db/postgresql"

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
	BoltDBKey string            = "bolt"
	MemoryKey string            = "memory"
	PsqlKey   string            = "postgresql"
	DBTypes   map[string]string = map[string]string{
		BoltDBKey: "bolt",
		MemoryKey: "memory",
		PsqlKey:   "postgresql",
	}
)

type ErrorHandling func(*models.Peer)

type PeerStore struct {
	// control variables for the exporting routines
	ctx    context.Context
	cancel context.CancelFunc

	Storage PeerStoreStorage
}

func NewPeerStore(ctx context.Context, dbtype string, path string, endpoint string) PeerStore {
	mainCtx, cancel := context.WithCancel(ctx)
	var db PeerStoreStorage
	var err error

	switch dbtype {
	case DBTypes[BoltDBKey]:
		if len(path) <= 0 {
			path = default_db_path
		}
		db = NewBoltPeerDB(path)
	case DBTypes[MemoryKey]:
		db = NewMemoryDB()
	case DBTypes[PsqlKey]:
		db, err = postgresql.ConnectToDB(mainCtx, endpoint)
		if err != nil {
			Log.Panic(err.Error())
		}
	default:
		if len(path) <= 0 {
			path = default_db_path
		}
		db = NewBoltPeerDB(path)
	}
	ps := PeerStore{
		ctx:     mainCtx,
		cancel:  cancel,
		Storage: db,
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
func (c *PeerStore) Close() {
	Log.Info("Closing Peerstore")
	c.Storage.Close()
	c.cancel()
}

// StoreOrUpdatePeer:
// Updates the peer without overwritting all its content.
// If peer exists, aggregate data to the existing peer.
// Otherwise, store the peer.
// @param peer: the peer to store or update
func (c *PeerStore) StoreOrUpdatePeer(peer models.Peer) {

	oldPeer, err := c.GetPeerData(peer.PeerId)
	// if error means not found, just store it
	if err != nil {
		c.Storage.StorePeer(peer.PeerId, peer)
	} else {
		// Fetch the new info of a peer directly from the new peer struct
		oldPeer.FetchPeerInfoFromNewPeer(peer)
		c.Storage.StorePeer(peer.PeerId, oldPeer)
	}
	// Force Garbage collector
	runtime.GC()
}

// StorePeer:
// This method stores a single peer in the peerstore.
// It will use the peerID as key.
// @param peer: the peer object to store.
func (c *PeerStore) StorePeer(peer models.Peer) {
	c.Storage.StorePeer(peer.PeerId, peer)
}

// GetPeerData:
// This method return a Peer object from the peerstore
// using the given peerID.
// @param peerID: the peerID to look for in string format.
// @return the found Peer object and an error if there was.
func (c *PeerStore) GetPeerData(peerId string) (models.Peer, error) {
	peerData, ok := c.Storage.LoadPeer(peerId)
	if !ok {
		return models.Peer{}, errors.New("could not find peer in peerstore or peer was unable to unmarshal: " + peerId)
	}

	return peerData, nil
}

// GetPeerList:
// This method returns the list of PeerIDs in the DB.
// @return the list of PeerIDs in string format.
func (c *PeerStore) GetPeerList() []peer.ID {
	return c.Storage.GetPeers()
}

func (c *PeerStore) GetPeerENR(peerID string) (*enode.Node, error) {
	return c.Storage.GetPeerENR(peerID)
}

// ExportCsvService will export to csv regularly, therefoe this service will execute the export every X seconds (ExportLoopTime)
// @param folderpath: the folder to export the csv file (always named metrics.csv)
func (ps *PeerStore) ExportCsvService(folderpath string) {
	fmt.Println("------")
	Log.Info("Peerstore CSV exporting service launched")
	go func() {
		ctx := ps.Ctx()
		ticker := time.NewTicker(ExportLoopTime)
		for {
			select {
			case <-ticker.C:
				_ = ps.Storage.ExportToCSV(folderpath + "/metrics.csv")
			case <-ctx.Done():
				ticker.Stop()
				_ = ps.Storage.ExportToCSV(folderpath + "/metrics.csv")
				Log.Info("Closing DB CSV exporter")
				return
			}
		}
	}()
}
