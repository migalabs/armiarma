package metrics

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeerStoreStorage(t *testing.T) {
	var p PeerStoreStorage

	db, err := OpenDB("test.db", "PeerStore", 0600, nil)
	require.Nil(t, err)
	defer db.Close()
	defer os.Remove("test.db")

	p = &PeerStoreDb{db}
	testStorage(t, p)

	var m sync.Map
	p = &PeerStoreMemory{&m}
	testStorage(t, p)
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
