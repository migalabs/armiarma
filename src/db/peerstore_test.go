package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_StoreOrUpdatePeer(t *testing.T) {
	// stores a peer
	peerStore := NewPeerStore("memory", "")
	p1 := NewPeer("Peer1")
	p1.ClientName = "Client1"
	peerStore.StoreOrUpdatePeer(p1)

	// rx some messages of topic1
	p1.MessageEvent("topic1", time.Now())
	peerStore.StoreOrUpdatePeer(p1)

	// check that its data is correct
	p, err := peerStore.GetPeerData("Peer1")
	require.NoError(t, err)
	require.Equal(t, p.ClientName, "Client1")
	require.Equal(t, p.MessageMetrics["topic1"].Count, uint64(1))

	// add a new version of the same peer
	p1Mod := NewPeer("Peer1")
	p1Mod.ClientName = "NewClient"
	peerStore.StoreOrUpdatePeer(p1Mod)
	p, err = peerStore.GetPeerData("Peer1")
	require.NoError(t, err)

	// check that the new information was updated
	require.Equal(t, p.ClientName, "NewClient")
	require.Equal(t, p.MessageMetrics["topic1"].Count, uint64(1))

}

func Test_GetPeerList(t *testing.T) {
	// stores a peer
	peerStore := NewPeerStore("memory", "")
	p1 := NewPeer("16Uiu2HAmCWwVV2qaLpEpPqkQHyX3ozazQs4sasXtFmVex8qzDqRG")
	peerStore.StoreOrUpdatePeer(p1)
	peerIDList := peerStore.GetPeerList()
	require.Equal(t, peerIDList[0].String(), "16Uiu2HAmCWwVV2qaLpEpPqkQHyX3ozazQs4sasXtFmVex8qzDqRG")
}
