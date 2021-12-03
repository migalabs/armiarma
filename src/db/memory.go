package db

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

// PeerStoreMemory save the peer's data in RAM.
// Unless exported, data is lost after execution.
type MemoryDB struct {
	m         *sync.Map
	startTime time.Time
}

// New NewMemoryDB creates an new MemoryDB ready to accept new peers
// fulfills PeerStoreStorage interface
func NewMemoryDB() MemoryDB {
	var m sync.Map
	mdb := MemoryDB{
		m:         &m,
		startTime: time.Now(),
	}
	Log.Infof("generated new MemoryDB")
	return mdb
}

// Store keeps adds key and Peer values into a sync.Map in memory.
// @param key: used as key in the map.
// @param value: the object to store with the given key.
func (p MemoryDB) Store(key string, value Peer) {
	p.m.Store(key, value)
}

// Loads peer value of given key from sync.Map in memory.
func (p MemoryDB) Load(key string) (value Peer, ok bool) {
	v, ok := p.m.Load(key)
	if !ok {
		return Peer{}, ok
	}
	value = v.(Peer)
	return
}

// Delete removes key and value from sync.Map.
// @param key: the string to locate the value to delete.
func (p MemoryDB) Delete(key string) {
	p.m.Delete(key)
}

// Range iterates through the key and values of the sync.Map
// @param f: the function to apply to each item in the db
func (p MemoryDB) Range(f func(key string, value Peer) bool) {
	p.m.Range(func(key, value interface{}) bool {
		ok := f(key.(string), value.(Peer))
		return ok
	})
}

// Close and free the space from the memory relative to the DB
// So far just initialize the
func (p MemoryDB) Close() {
	p.m.Range(func(key, _ interface{}) bool {
		p.m.Delete(key.(string))
		return true
	})
}

// Peers
// This method returns a string array with the list of PeerIDs
// existing in the DB.
// These would be the keys of each entry in the map
// @return the string array containisssng the PeerIDs
func (p MemoryDB) Peers() []peer.ID {
	result := make([]peer.ID, 0)
	p.Range(func(key string, value Peer) bool {
		peerID_obj, err := peer.Decode(key)
		if err != nil {
			//return false
			// TODO: print warning: peer was not read
		}
		result = append(result, peerID_obj)
		return true
	})
	return result
}
