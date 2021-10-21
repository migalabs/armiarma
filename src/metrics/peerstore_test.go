package metrics

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStoreOrUpdatePeer(t *testing.T) {
	// stores a peer
	peerStore := NewPeerStore("memory", "")
	p1 := NewPeer("Peer1")
	p1.ClientName = "Client1"
	peerStore.StoreOrUpdatePeer(p1)

	// rx some messages of topic1
	err := peerStore.MessageEvent("Peer1", "topic1")
	require.NoError(t, err)
	err = peerStore.MessageEvent("Peer1", "topic1")
	require.NoError(t, err)

	// check that its data is correct
	p, err := peerStore.GetPeerData("Peer1")
	require.NoError(t, err)
	require.Equal(t, p.ClientName, "Client1")
	require.Equal(t, p.MessageMetrics["topic1"].Count, uint64(2))

	// add a new version of the same peer
	p1Mod := NewPeer("Peer1")
	p1Mod.ClientName = "NewClient"
	peerStore.StoreOrUpdatePeer(p1Mod)
	p, err = peerStore.GetPeerData("Peer1")
	require.NoError(t, err)

	// check that the new information was updated
	require.Equal(t, p.ClientName, "NewClient")
	require.Equal(t, p.MessageMetrics["topic1"].Count, uint64(2))
}

func TestPeerStoreConnectionFetchingTest(t *testing.T) {
	// generate the peerstore in memory
	peerStore := NewPeerStore("memory", "")
	// generate the sample peer
	p1 := NewPeer("Peer1")
	peerStore.StoreOrUpdatePeer(p1)

	// Start the fetching test
	testIterations := 10
	for i := 0; i < testIterations; i++ {

		// check that its data is correct
		ptest, err := peerStore.GetPeerData("Peer1")
		require.NoError(t, err)
		require.Equal(t, "Peer1", ptest.PeerId)
		// check if the number of connections matches the number of testIterations
		require.Equal(t, i, len(ptest.ConnectionTimes))

		// generate the new peer with the new connection events
		newp := NewPeer("Peer1")
		// aggregate a new connection to the list
		newp.ConnectionEvent("inbound", parseTime("2021-08-23T01:00:10.000Z", t))
		peerStore.StoreOrUpdatePeer(newp)
		time.Sleep(300 * time.Millisecond)
	}
}

/*
// Error test screnario where we try to fetch a copy of oldPeer to the oldPeer itseld
// which should cause the array of connections to increase and therefore the heap
// RECREATION OF A BUG
func TestReproduceAccumulativeHeapFailure(t *testing.T) {
	// generate the peerstore in memory
	peerStore := NewPeerStore("memory", "")
	// generate the sample peer
	p1 := NewPeer("Peer1")
	peerStore.StoreOrUpdatePeer(p1)

	// Start the fetching test
	testIterations := 10
	for i := 0; i < testIterations; i++ {

		// check that its data is correct
		ptest, err := peerStore.GetPeerData("Peer1")
		require.NoError(t, err)

		// aggregate a new connection to the previously existing peer (causing to double count all the connection events)
		ptest.ConnectionEvent("inbound", parseTime("2021-08-23T01:00:10.000Z", t))
		peerStore.StoreOrUpdatePeer(ptest)

		time.Sleep(300 * time.Millisecond)
	}
	// check that its data is correct
	p, err := peerStore.GetPeerData("Peer1")
	require.NoError(t, err)
	require.Equal(t, "Peer1", p.PeerId)
	require.Equal(t, testIterations, len(p.ConnectionTimes))
}
*/
