package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"runtime"
	"time"

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
func NewBoltPeerDB(path string) BoltPeerDB {
	// Generate a new one
	db, err := OpenBoltDB(path, "peerstore", 0600, nil)
	if err != nil {
		log.Error(err)
		db = nil
	}
	pbdb := BoltPeerDB{
		db:        db,
		startTime: time.Now(),
	}
	// check if there is something inside the bolt database
	peercnt := 0
	pbdb.Range(func(key string, value Peer) bool {
		peercnt++
		return true
	})
	fmt.Println("Peers in DB", peercnt)
	if peercnt > 0 {
		log.Infof("loaded BoltDB with %d peer on it", peercnt)
	} else {
		log.Infof("generated new BoltDB")
	}
	return pbdb
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
	err := json.Unmarshal(value_marshalled, &value)
	if err != nil {
		log.Error(err)
		return Peer{}, false
	}
	return value, true
}

func (p BoltPeerDB) Delete(key string) {
	p.db.Delete([]byte(key))
}

func (p BoltPeerDB) Range(f func(key string, value Peer) bool) {
	p.db.Range(func(key, value []byte) bool {
		var value_unmarshalled Peer
		err := json.Unmarshal(value, &value_unmarshalled)
		if err != nil {
			log.Error(err)
			return false
		}
		ok := f(string(key), value_unmarshalled)
		return ok
	})
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
