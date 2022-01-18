package db

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/utils"
	bolt "go.etcd.io/bbolt"
)

var (
	BoltDBType = "BoltDB"
)

// PeerStoreDb save the peer's data in a persistant Db.
type BoltPeerDB struct {
	db        *boltDB
	startTime time.Time
}

// New BoltPeerDB creates an new MemoryDB ready to accept new peers.
// Fulfills PeerStoreStorage interface
// @param folderpath: folder where to open / create the db (always named peerstore)
func NewBoltPeerDB(folderpath string) BoltPeerDB {
	// Generate a new one
	db, err := openBoltDB(folderpath+"/peerstore.db", "peerstore", 0600, nil)
	if err != nil {
		Log.Panicf(err.Error())
	}
	dbObj := BoltPeerDB{
		db:        db,
		startTime: time.Now(),
	}

	// check if there is something inside the bolt database
	// if so, fill the disconnection for these peers that were connected
	// when the crawler was shutdown
	connectedPeers := make([]models.Peer, 0) // keep the Peers that were still connected, meaning that Disconnection was not registered.
	peercnt := 0
	lastCrawlerActivity := time.Time{} // representation of least possible time
	dbReadingError := false
	dbObj.Range(func(key string, value models.Peer) bool {
		// TEMPORARY FIX - check if the peer is empty (error reading the existing DB)
		if value.IsEmpty() {
			// the peer is empty, therefore, there was an error reading the DB
			dbReadingError = true
			return false
		}
		// check if there was an open connection
		if value.IsConnected {
			// it "remains" connected
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
	if dbReadingError {
		// If there has been any issue reading the previously existing DB, generte a new one on the
		// user given new name
		Log.Errorf("Unable to read existing DB at %s", folderpath+"/peerstore.db")
		Log.Error("It could be originated from a non-compatible DB version or a corrupted DB")
		Log.Error("Please, introduce the name you want to assign to the backup-file of the previously existing DB (press ENTER to confirm)")
		var newName string
		// Taking input from user
		n, err := fmt.Scanln(&newName)
		if err != nil {
			Log.Error(n, err)
		}
		if !strings.Contains(newName, ".db") {
			newName = newName + ".db"
		}
		Log.Info("making backup to ", newName)
		// Rename the old db
		err = utils.CopyFileToNewPath(folderpath+"/peerstore.db", folderpath+"/"+newName)
		if err != nil {
			Log.Errorf("Unable to copy existing DB to %s .", folderpath+"/"+newName)
			os.Exit(0)
		}
		// Generate new file for the new Bolt DB
		db, err = openBoltDB(folderpath+"/peerstore.db", "peerstore", 0600, nil)
		if err != nil {
			Log.Panicf(err.Error())
		}
		// Fill the previous existing DB Obj with the new db
		dbObj.db = db

	} else {
		// if there hasn't been any error, proceed to fill the connect peers with the needed disconnections
		if peercnt > 0 {
			Log.Infof("loaded BoltDB with %d peer on it (%d connected)", peercnt, len(connectedPeers))
		} else {
			Log.Infof("generated new BoltDB")
		}

		// last, lets add the disconnection event to those peers that remained connected
		for _, connectedPeerTmp := range connectedPeers {
			connectedPeerTmp.DisconnectionEvent(lastCrawlerActivity)
			dbObj.Store(connectedPeerTmp.PeerId, connectedPeerTmp)
		}
	}

	return dbObj
}

// Stores a Peer with the given key.
// @param key: the key to access the object.
// @param Peer: the value to store.
func (b BoltPeerDB) Store(key string, value models.Peer) {
	value_marshalled, err := json.Marshal(value)
	if err != nil {
		Log.Error(err)
		return
	}
	b.db.store([]byte(key), value_marshalled)

}

// Retrieves an object from the db using a key.
// @param key: the string to use to get the object.
// @return Peer: the resulting object.
// @return ok: whether the operation was successful or not.
func (b BoltPeerDB) Load(key string) (models.Peer, bool) {
	value_marshalled, ok := b.db.load([]byte(key))
	if !ok {
		return models.Peer{}, false
	}
	var obj map[string]interface{}
	err := json.Unmarshal(value_marshalled, &obj)

	if err != nil {
		Log.Error(err)
		return models.Peer{}, false
	}
	pObj, err := models.PeerUnMarshal(obj)
	if err != nil {
		return models.Peer{}, false
	}
	return pObj, true
}

// Deletes the object for the given key in the db.
// @param key: the string to access the desired object.
func (p BoltPeerDB) Delete(key string) {
	p.db.delete([]byte(key))
}

func (b BoltPeerDB) Range(f func(key string, value models.Peer) bool) {

	b.db.Range(func(key, value []byte) bool {

		var obj map[string]interface{}

		err := json.Unmarshal(value, &obj)
		if err != nil {
			Log.Error(err)

			return false
		}
		// unmarshal the peer
		// If the peer wasn't able to be unmarshalled, we will return an empty peer to the given func
		// handle in the fn that empty peers as it requires
		pObj, _ := models.PeerUnMarshal(obj)
		ok := f(string(key), pObj)
		return ok

	})

}

// Type
func (b BoltPeerDB) Type() string {
	return BoltDBType
}

// TODO: pending return / print some kind of error if it was the case
// Resturns a list of peerIDs existing in the db
// @return the list of peerID in peer.ID format
func (b BoltPeerDB) Peers() []peer.ID {
	peers := make([]peer.ID, 0)
	b.Range(func(key string, value models.Peer) bool {
		peerID_obj, err := peer.Decode(string(key))

		if err != nil {
			return false
		}
		peers = append(peers, peerID_obj)
		return true
	})

	return peers
}

// GetENR returns the Node after parsing the ENR.
// @param peerID: the peerID of which to get the Node.
// @return the resulting Node.
// @return error if applicable, nil in any other case.
func (b BoltPeerDB) GetENR(peerID string) (*enode.Node, error) {
	p, ok := b.Load(peerID)
	if !ok {
		return nil, fmt.Errorf("No peer was found under ID %s", peerID)
	}
	return p.GetBlockchainNode()
}

// ExportToCSV
// This method will export the whole peerstore into a CSV file.
// @param filePath file where to dump the CSV lines (create if it does not exist).
// @return an error if there was.
func (b BoltPeerDB) ExportToCSV(filePath string) error {
	Log.Info("Exporting metrics to csv: ", filePath)
	csvFile, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "error opening the file "+filePath)
	}
	defer csvFile.Close()

	// First raw of the file will be the Titles of the columns
	_, err = csvFile.WriteString("Peer Id,Node Id,Fork Digest,User Agent,Client,Version,Pubkey,Address,Ip,Country,City,ENR,Request Metadata,Success Metadata,Attempted,Succeed,Deprecated,ConnStablished,IsConnected,Attempts,Error,Last Error Timestamp,Last Identify Timestamp,Latency,Connections,Disconnections,Last Connection,Conn Direction,Connected Time,Beacon Blocks,Beacon Aggregations,Voluntary Exits,Proposer Slashings,Attester Slashings,Total Messages\n")
	if err != nil {
		errors.Wrap(err, "error while writing the titles on the csv "+filePath)
	}

	err = nil
	b.Range(func(key string, value models.Peer) bool {
		_, err = csvFile.WriteString(value.ToCsvLine())
		return true
	})

	if err != nil {
		return errors.Wrap(err, "could not export peer metrics")
	}
	return nil
}

func (b BoltPeerDB) Close() {
	b.db.close()
}

// --- Low level interaction with boltdb

// BoltDB implements basic operations to provide a key-value DB for any kind of
type boltDB struct {
	db     *bolt.DB
	bucket string
}

// Opens the existing db and creates a bucket is not existing. The busket is where we will store the information.
// @param path: path to db to open.
//@param bucketName: the bucket we are opening / creating (in our case, we always use the same)
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
// @param key: the key in byte format to access the data.
// @return the bytes as a result, marshaled.
// @return a boolean whether the operation was successful or not.
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

// Stores a given data in the db.
// @param key: the key to access the desired object in the db.
// @param value: the data to store.
func (db *boltDB) store(key, value []byte) {

	db.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(db.bucket))
		err := b.Put(key, value)
		return err
	})
}

// Deletes a given object from the db.
// @param key: the key to locate the object in the db.
func (db *boltDB) delete(key []byte) {

	db.db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(db.bucket))
		err := b.Delete(key)
		return err
	})
}

// Iterates and executes a function over the db for each value.
// @param f: the function to apply to each item in the db.
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
