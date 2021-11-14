package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"runtime"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// PeerStoreDb save the peer's data in a persistant Db.
type BoltPeerDB struct {
	db        *BoltDB
	startTime time.Time
}

// New BoltPeerDB creates an new MemoryDB ready to accept new peers
// fulfills PeerStoreStorage interface
func NewBoltPeerDB(folderpath string) BoltPeerDB {
	// Generate a new one
	db, err := OpenBoltDB(folderpath+"/peerstore.db", "peerstore", 0600, nil)
	if err != nil {
		log.Panicf(err.Error())
	}
	db_obj := BoltPeerDB{
		db:        db,
		startTime: time.Now(),
	}

	// check if there is something inside the bolt database
	// if so, fill the disconnectin for these peers that were connected
	// when the crawler was shutdown
	connectedPeers := make([]Peer, 0) // store the Peers that were connected, so Disconnection was not registered.
	// so they remain connected
	peercnt := 0
	lastCrawlerActivity := time.Time{} // representation of least possible time
	db_obj.Range(func(key string, value Peer) bool {
		// check if there was an Open connection
		if value.IsConnected {
			// it remains connected
			connectedPeers = append(connectedPeers, value)
		}

		// we also need to figure out the last activity of the crawler
		// this way we can set the disconnection time for the remained connected
		peerLastActivity := value.GetLastActivityTime()
		if peerLastActivity.After(lastCrawlerActivity) {
			lastCrawlerActivity = peerLastActivity
		}
		peercnt++
		return true
	})
	if peercnt > 0 {
		log.Infof("loaded BoltDB with %d peer on it (%d connected)", peercnt, len(connectedPeers))
	} else {
		log.Infof("generated new BoltDB")
	}

	// last, lets add the disconnection event to those peers that remained connected
	for _, connectedPeerTmp := range connectedPeers {
		connectedPeerTmp.DisconnectionEvent(lastCrawlerActivity)
	}

	return db_obj
}

func (p BoltPeerDB) Store(key string, value Peer) {
	value_marshalled, err := json.Marshal(value)
	if err != nil {
		log.Error(err)
		return
	}
	p.db.Store([]byte(key), value_marshalled)

}

func (p BoltPeerDB) Load(key string) (value Peer, ok bool) {

	value_marshalled, ok := p.db.Load([]byte(key))
	if !ok {
		return Peer{}, false
	}
	var obj map[string]interface{}
	err := json.Unmarshal(value_marshalled, &obj)

	if err != nil {
		log.Error(err)
		return Peer{}, false
	}
	return PeerUnMarshal(obj), true
}

func (p BoltPeerDB) Delete(key string) {
	p.db.Delete([]byte(key))
}

func (p BoltPeerDB) Range(f func(key string, value Peer) bool) {

	p.db.Range(func(key, value []byte) bool {

		var obj map[string]interface{}

		err := json.Unmarshal(value, &obj)
		if err != nil {
			log.Error(err)

			return false
		}
		value_unmarshalled := PeerUnMarshal(obj)

		ok := f(string(key), value_unmarshalled)
		return ok
	})

}

// TODO: pending return / print some kind of error if it was the case
func (p BoltPeerDB) Peers() []peer.ID {
	peers := make([]peer.ID, 0)
	p.Range(func(key string, value Peer) bool {
		peerID_obj, err := peer.Decode(string(key))

		if err != nil {
			return false
		}
		peers = append(peers, peerID_obj)
		return true
	})

	return peers
}

func (p BoltPeerDB) Close() {
	p.db.Close()
}

// BoltDB implements basic operations to provide a key-value DB for any kind of
type BoltDB struct {
	db     *bolt.DB
	bucket string
}

func OpenBoltDB(path string, bucketName string, mode fs.FileMode, options *bolt.Options) (*BoltDB, error) {
	boltDB, err := bolt.Open(path, mode, options)
	if err != nil {
		return &BoltDB{}, err
	}

	err = boltDB.Update(func(tx *bolt.Tx) error {

		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})

	db := &BoltDB{boltDB, bucketName}

	return db, err
}

func (db *BoltDB) Close() {
	db.db.Close()
}

func (db *BoltDB) Load(key []byte) ([]byte, bool) {

	var got []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.bucket))
		if b == nil {
			return fmt.Errorf("bucket is nil")
		}
		got = b.Get([]byte(key))
		return nil
	})
	if err != nil || got == nil {
		return got, false
	}
	value := make([]byte, len(got))
	copy(value, got)
	return value, true
}

func (db *BoltDB) Store(key, value []byte) {

	db.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(db.bucket))
		err := b.Put(key, value)
		return err
	})
}

func (db *BoltDB) Delete(key []byte) {

	db.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(db.bucket))
		err := b.Delete(key)
		return err
	})
}

func (db *BoltDB) Range(f func(key, value []byte) bool) {

	db.db.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(db.bucket))
		err := b.ForEach(func(k, v []byte) error {
			if ok := f(k, v); !ok {
				func_name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
				err := fmt.Sprintf("Error while executing the function %v while on key %v", func_name, string(k))
				return errors.New(err)
			}
			return nil
		})

		return err
	})
}
