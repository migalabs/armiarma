package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"runtime"
	"sync"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

// PeerStoreStorage provides the necessary functions
// to interact with peers data.
// It is either saved in a Db or in Memory.
type PeerStoreStorage interface {
	Store(key string, value Peer)
	Load(key string) (value Peer, ok bool)
	Delete(key string)
	Range(f func(key string, value Peer) bool)
}

// PeerStoreDb save the peer's data in a persistant Db.
type PeerStoreDb struct {
	db *BoltDB
}

// PeerStoreMemory save the peer's data in RAM.
// Unless exported, data is lost after execution.
type PeerStoreMemory struct {
	m *sync.Map
}

func (p *PeerStoreDb) Store(key string, value Peer) {
	value_marshalled, err := json.Marshal(value)
	if err != nil {
		log.Error(err)
		return
	}
	p.db.Store([]byte(key), value_marshalled)
}

func (p *PeerStoreDb) Load(key string) (value Peer, ok bool) {
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

func (p *PeerStoreDb) Delete(key string) {
	p.db.Delete([]byte(key))
}

func (p *PeerStoreDb) Range(f func(key string, value Peer) bool) {
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

func (p *PeerStoreMemory) Store(key string, value Peer) {
	p.m.Store(key, value)
}

func (p *PeerStoreMemory) Load(key string) (value Peer, ok bool) {
	v, ok := p.m.Load(key)
	if !ok {
		return Peer{}, ok
	}
	value = v.(Peer)
	return
}

func (p *PeerStoreMemory) Delete(key string) {
	p.m.Delete(key)
}

func (p *PeerStoreMemory) Range(f func(key string, value Peer) bool) {
	p.m.Range(func(key, value interface{}) bool {
		ok := f(key.(string), value.(Peer))
		return ok
	})
}

// BoltDB implements basic operations to provide a key-value DB.
type BoltDB struct {
	db     *bolt.DB
	bucket string
}

func OpenDB(path string, bucketName string, mode fs.FileMode, options *bolt.Options) (*BoltDB, error) {
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

func (db *BoltDB) Load(key []byte) (value []byte, ok bool) {
	db.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(db.bucket))
		value = b.Get(key)
		return nil
	})

	if value == nil {
		ok = false
	} else {
		ok = true
	}

	return
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
