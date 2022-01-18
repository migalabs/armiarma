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
	Store(key string, value models.Peer)
	Load(key string) (value models.Peer, ok bool)
	Delete(key string)
	Range(f func(key string, value models.Peer) bool)
	Close()
	// New calls
	Type() string
	Peers() []peer.ID
	GetENR(string) (*enode.Node, error)
	ExportToCSV(string) error
	// TODO: -Implement statistics directly from the PeerStoreStorage module?
}
