package metrics

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
	// TODO: -Implement statistics directly from the PeerStoreStorage module?
}
