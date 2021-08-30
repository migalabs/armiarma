package metrics

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_StoreOrUpdatePeer(t *testing.T) {
	// stores a peer
	peerStore := NewPeerStore()
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
