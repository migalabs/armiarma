package boltdb

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/pkg/errors"
)

// Stores a Peer with the given key.
// @param key: the key to access the object.
// @param Peer: the value to store.
func (b BoltDB) StorePeer(key string, value models.Peer) {
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
func (b BoltDB) LoadPeer(key string) (models.Peer, bool) {
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
func (p BoltDB) DeletePeer(key string) {
	p.db.delete([]byte(key))
}

func (b BoltDB) Range(f func(key string, value models.Peer) bool) {

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
func (b BoltDB) Type() string {
	return BoltDBType
}

// TODO: pending return / print some kind of error if it was the case
// Resturns a list of peerIDs existing in the db
// @return the list of peerID in peer.ID format
func (b BoltDB) GetPeers() []peer.ID {
	peers := make([]peer.ID, 0)
	b.Range(func(key string, value models.Peer) bool {
		peerID_obj, err := peer.Decode(string(key))

		if err != nil {
			fmt.Println("error decoding peer id", err)
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
func (b BoltDB) GetPeerENR(peerID string) (*enode.Node, error) {
	p, ok := b.LoadPeer(peerID)
	if !ok {
		return nil, fmt.Errorf("No peer was found under ID %s", peerID)
	}
	return p.GetBlockchainNode()
}

// ExportToCSV
// This method will export the whole peerstore into a CSV file.
// @param filePath file where to dump the CSV lines (create if it does not exist).
// @return an error if there was.
func (b BoltDB) ExportToCSV(filePath string) error {
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

func (b BoltDB) Close() {
	b.db.close()
}
