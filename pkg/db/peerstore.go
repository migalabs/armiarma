package db

import (
	"context"
	"runtime"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
	postgresql "github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/migalabs/armiarma/pkg/exporters"

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
	ExportLoopTime time.Duration = 30 * time.Minute
)

type ErrorHandling func(*models.Peer)

type PeerStore struct {
	// control variables for the exporting routines
	ctx context.Context

	Storage         *postgresql.PostgresDBService
	ExporterService *exporters.ExporterService
}

func NewPeerStore(ctx context.Context, exporter *exporters.ExporterService, path string, endpoint string, netModel postgresql.NetworkModel) PeerStore {
	var db *postgresql.PostgresDBService
	var err error

	db, err = postgresql.ConnectToDB(ctx, endpoint, netModel)
	if err != nil {
		Log.Panic(err.Error())
	}
	ps := PeerStore{
		ctx:             ctx,
		Storage:         db,
		ExporterService: exporter,
	}
	return ps
}

// Ctx
// Retreives the context asigned to the Peerstore
func (c *PeerStore) Ctx() context.Context {
	return c.ctx
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
