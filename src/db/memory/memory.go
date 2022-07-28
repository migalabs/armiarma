package db

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	MemoryType = "MemoryDB"
	log        = logrus.WithField(
		"module", MemoryType,
	)
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
	log.Infof("generated new MemoryDB")
	return mdb
}

// Store keeps adds key and Peer values into a sync.Map in memory.
// @param key: used as key in the map.
// @param value: the object to store with the given key.
func (m MemoryDB) StorePeer(key string, value models.Peer) {
	m.m.Store(key, value)
}

// Loads peer value of given key from sync.Map in memory.
func (m MemoryDB) LoadPeer(key string) (value models.Peer, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return models.Peer{}, ok
	}
	value = v.(models.Peer)
	return
}

// Delete removes key and value from sync.Map.
// @param key: the string to locate the value to delete.
func (m MemoryDB) DeletePeer(key string) {
	m.m.Delete(key)
}

// Range iterates through the key and values of the sync.Map
// @param f: the function to apply to each item in the db
func (m MemoryDB) Range(f func(key string, value models.Peer) bool) {
	m.m.Range(func(key, value interface{}) bool {
		ok := f(key.(string), value.(models.Peer))
		return ok
	})
}

// Close and free the space from the memory relative to the DB
// So far just initialize the
func (m MemoryDB) Close() {
	m.m.Range(func(key, _ interface{}) bool {
		m.m.Delete(key.(string))
		return true
	})
}

// Type
func (m MemoryDB) Type() string {
	return MemoryType
}

// Peers
// This method returns a string array with the list of PeerIDs
// existing in the DB.
// These would be the keys of each entry in the map
// @return the string array containisssng the PeerIDs
func (m MemoryDB) GetPeers() []peer.ID {
	result := make([]peer.ID, 0)
	m.Range(func(key string, value models.Peer) bool {
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

// GetENR returns the Node after parsing the ENR.
// @param peerID: the peerID of which to get the Node.
// @return the resulting Node.
// @return error if applicable, nil in any other case.
func (m MemoryDB) GetPeerENR(peerID string) (*enode.Node, error) {
	p, ok := m.LoadPeer(peerID)
	if !ok {
		return nil, fmt.Errorf("No peer was found under ID %s", peerID)
	}
	enr, ok := p.GetAtt("enr")
	if !ok {
		return nil, fmt.Errorf("No ENR was found for peer %s", peerID)
	}
	return enode.MustParse(enr.(string)), nil
}

// ExportToCSV
// This method will export the whole peerstore into a CSV file.
// @param filePath file where to dump the CSV lines (create if it does not exist).
// @return an error if there was.
func (m MemoryDB) ExportToCSV(filePath string) error {
	log.Info("Exporting metrics to csv: ", filePath)
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
	m.Range(func(key string, value models.Peer) bool {
		_, err = csvFile.WriteString(value.ToCsvLine())
		return true
	})

	if err != nil {
		return errors.Wrap(err, "could not export peer metrics")
	}
	return nil
}
