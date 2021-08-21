package metrics

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_PeerStore(t *testing.T) {
	peerStore := NewPeerStore()
	peerStore.AddPeer(Peer{
		PeerId:     "Peer1",
		ClientName: "Client1",
	})
	peerStore.AddPeer(Peer{
		PeerId:     "Peer2",
		ClientName: "Client2",
	})

	p, ok := peerStore.GetPeerData("Peer1")
	require.Equal(t, ok, true)
	require.Equal(t, p.ClientName, "Client1")

	p, ok = peerStore.GetPeerData("Peer2")
	require.Equal(t, ok, true)
	require.Equal(t, p.ClientName, "Client2")

	peerStore.AddPeer(Peer{
		PeerId:     "Peer1",
		ClientName: "Client3",
	})

	p, ok = peerStore.GetPeerData("Peer1")
	require.Equal(t, ok, true)
	require.Equal(t, p.ClientName, "Client3")
}
