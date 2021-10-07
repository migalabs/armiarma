package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var path string = "./test_db"

func TestPeerStoreStorage(t *testing.T) {
	// Test the BoltDB to store the information about a peer
	db := NewBoltPeerDB(path)
	defer db.Close()
	defer os.Remove("test.db")
	testStorage(t, db)
	m := NewMemoryDB()
	// Test the Memory DB to store the information about a peer
	testStorage(t, m)
}

func testStorage(t *testing.T, p PeerStoreStorage) {
	peer1 := Peer{
		PeerId:     "1",
		ClientName: "Client1",
	}

	peer2 := Peer{
		PeerId:     "2",
		ClientName: "Client2",
	}

	peer3 := Peer{
		PeerId:     "3",
		ClientName: "Client3",
	}

	p.Store("1", peer1)
	p.Store("2", peer2)
	p.Store("3", peer3)

	peer_test, ok := p.Load("1")
	require.True(t, ok)
	require.Equal(t, "Client1", peer_test.ClientName)

	peer_test, ok = p.Load("2")
	require.True(t, ok)
	require.Equal(t, "Client2", peer_test.ClientName)

	var peerStoreTest []Peer

	p.Range(func(key string, value Peer) bool {
		peerStoreTest = append(peerStoreTest, value)
		return true
	})
	require.Equal(t, 3, len(peerStoreTest))
	require.Equal(t, peer1, peerStoreTest[0])
	require.Equal(t, peer2, peerStoreTest[1])
	require.Equal(t, peer3, peerStoreTest[2])

	p.Delete("2")
	_, ok = p.Load("2")
	require.False(t, ok)
}
