package metrics

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_PeerStore(t *testing.T) {
	peerStore := NewPeerStore()
	peerStore.StorePeer(Peer{
		PeerId:     "Peer1",
		ClientName: "Client1",
	})
	peerStore.StorePeer(Peer{
		PeerId:     "Peer2",
		ClientName: "Client2",
	})

	p, err := peerStore.GetPeerData("Peer1")
	require.NoError(t, err)
	require.Equal(t, p.ClientName, "Client1")

	p, err = peerStore.GetPeerData("Peer2")
	require.NoError(t, err)
	require.Equal(t, p.ClientName, "Client2")

	peerStore.StorePeer(Peer{
		PeerId:     "Peer1",
		ClientName: "Client3",
	})

	p, err = peerStore.GetPeerData("Peer1")
	require.NoError(t, err)
	require.Equal(t, p.ClientName, "Client3")
}
