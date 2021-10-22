package db

import "github.com/libp2p/go-libp2p-core/peer"

const DEFAULT_DB_PATH string = "peerstore.db"

var (
	default_db_path = "peerstore.db"
)

// PeerStoreStorage provides the necessary functions
// to interact with peers data.
// It is either saved in a Db or in Memory.
type PeerStoreStorage interface {
	Store(key string, value Peer)
	Load(key string) (value Peer, ok bool)
	Delete(key string)
	Range(f func(key string, value Peer) bool)
	Close()
	Peers() []peer.ID
	// TODO: -Implement statistics directly from the PeerStoreStorage module?
}
