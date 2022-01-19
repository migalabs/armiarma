package db

import (
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/db/models"
)

const DEFAULT_DB_PATH string = "peerstore.db"

var (
	default_db_path = "peerstore.db"
)

// PeerStoreStorage provides the necessary functions
// to interact with peers data.
// It is either saved in a Db or in Memory.
type PeerStoreStorage interface {
	StorePeer(string, models.Peer)
	LoadPeer(string) (models.Peer, bool)
	DeletePeer(string)
	Close()
	// New calls
	Type() string
	GetPeers() []peer.ID
	GetPeerENR(string) (*enode.Node, error)
	ExportToCSV(string) error
	// TODO: -Implement statistics directly from the PeerStoreStorage module?
}
