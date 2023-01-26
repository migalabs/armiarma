package peerstore

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
	"runtime"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	PeerstoreType      = "BoltDB"
	PeerNotInAddrsBook = errors.New("peer not in addr book")
)

const (
	peerstoreFile = "peerstore.db"
)

// New BoltPeerDB creates an new local Peertore ready to accept new peers.
func NewPeerstore(folderpath string) *Peerstore {
	// Generate a new one
	db, err := openBoltDB(folderpath+"/"+peerstoreFile, "address_book", 0600, nil)
	if err != nil {
		log.Panicf(err.Error())
	}
	dbObj := &Peerstore{
		db:        db,
		startTime: time.Now(),
	}

	// check if there is something inside the bolt database
	peers := dbObj.GetAllPeers()
	// if there hasn't been any error, proceed to fill the connect peers with the needed disconnections
	if len(peers) > 0 {
		log.Infof("loaded Peerstore with %d peers", len(peers))
	} else {
		log.Infof("generated new Peerstore")
	}

	return dbObj
}

// PeerStoreDb save the peer's data in a persistant Db.
type Peerstore struct {
	db        *boltDB
	startTime time.Time
}

// Stores a Peer with the given key.
func (b *Peerstore) StorePeer(value PersistablePeer) error {
	value_marshalled, err := json.Marshal(value)
	if err != nil {
		return err
	}
	b.db.store([]byte(value.ID.String()), value_marshalled)
	return nil
}

// Retrieves an object from the db using a key.
func (b *Peerstore) LoadPeer(key string) (PersistablePeer, bool) {
	value_marshalled, ok := b.db.load([]byte(key))
	if !ok {
		return PersistablePeer{}, false
	}
	// Unmarshal content into new obj
	var obj PersistablePeer
	err := json.Unmarshal(value_marshalled, &obj)
	if err != nil {
		log.Error(err)
		return PersistablePeer{}, false
	}
	return obj, true
}

// Deletes the object for the given key in the db.
func (b *Peerstore) DeletePeer(key string) {
	b.db.delete([]byte(key))
}

// range itereates through all the records found in the Peerstore, applying
// the given function as argument
func (b *Peerstore) Range(f func(value PersistablePeer) bool) {
	b.db.Range(func(key, value []byte) bool {
		var obj PersistablePeer
		err := json.Unmarshal(value, &obj)
		if err != nil {
			log.Errorf("unable to unmarshal peer.ID stored in the local peerstore - %s", err)
			return false
		}
		ok := f(obj)
		return ok
	})
}

// GetAllPeers returns the entire list of PersistablePeers found in the Peerstore
func (b *Peerstore) GetAllPeers() []PersistablePeer {
	persistables := make([]PersistablePeer, 0)

	copyFunc := func(value PersistablePeer) bool {
		persistables = append(persistables, value)
		return true
	}

	b.Range(copyFunc)
	return persistables
}

// GetPeersIDs resturns a list of peerIDs existing in the db
func (b *Peerstore) GetPeerIDs() []peer.ID {
	peers := make([]peer.ID, 0)

	copyID := func(value PersistablePeer) bool {
		peers = append(peers, value.ID)
		return true
	}
	b.Range(copyID)
	return peers
}

func (b *Peerstore) Close() {
	b.db.close()
}

// --- Low level interaction with boltdb

// BoltDB implements basic operations to provide a key-value DB for any kind of
type boltDB struct {
	db     *bolt.DB
	bucket string
}

// Opens the existing db and creates a bucket is not existing. The busket is where we will store the information.
func openBoltDB(path string, bucketName string, mode fs.FileMode, options *bolt.Options) (*boltDB, error) {
	bDB, err := bolt.Open(path, mode, options)
	if err != nil {
		return &boltDB{}, err
	}

	err = bDB.Update(func(tx *bolt.Tx) error {

		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})

	db := &boltDB{bDB, bucketName}

	return db, err
}

// close closes the Boltdb database in a secure way
func (db *boltDB) close() {
	db.db.Close()
}

// Loads data from the db.
func (db *boltDB) load(key []byte) ([]byte, bool) {
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

// store stores a given data in the db.
func (db *boltDB) store(key, value []byte) {
	db.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(db.bucket))
		err := b.Put(key, value)
		return err
	})
}

// delete deletes a given object from the db.
func (db *boltDB) delete(key []byte) {
	db.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(db.bucket))
		err := b.Delete(key)
		return err
	})
}

// Range Iterates and executes a function over the db for each value.
func (db *boltDB) Range(f func(key, value []byte) bool) {
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
