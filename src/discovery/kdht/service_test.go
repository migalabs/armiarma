package kdht

import (
	"context"
	"testing"

	"github.com/migalabs/armiarma/src/utils"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus"
)

func TestDiscoveredPeers(t *testing.T) {
	testMAddrs := []string{
		"/ip4/192.168.0.11/tcp/9000/p2p/12D3KooWLoj95HPXW8omPESoLiEMDLskASha7kK3uGfAvrLS1xtN",
		"/ip4/192.168.0.12/tcp/9000/p2p/12D3KooWLoj95HPXW8omPESoLiEMDLskASha7kK3uGfAvrLS1xtN",
		"/ip4/192.168.0.13/tcp/9000/p2p/12D3KooWQnwEGNqcM2nAcPtRR9rAX8Hrg4k9kJLCHoTR5chJfz6d",
	}
	logrus.SetLevel(logrus.DebugLevel)

	// compose the MAddrs
	multiAddr := make([]peer.AddrInfo, 0)
	for _, maddrStr := range testMAddrs {
		maddr, err := utils.UnmarshalMaddr(maddrStr)
		require.Equal(t, nil, err)
		peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		require.Equal(t, err, nil)
		multiAddr = append(multiAddr, *peerInfo)
	}

	// generate the discovered peers
	discpeers := NewDiscoveryPeers(context.Background())

	// test that there is no next peer since is empty
	require.Equal(t, false, discpeers.next())

	// add Peer 0 and check if it Next is true and the value is the same one
	discpeers.addPeer(multiAddr[0])
	require.Equal(t, true, discpeers.next())

	np := discpeers.getNextPeer()
	require.Equal(t, np, multiAddr[0])

	require.Equal(t, false, discpeers.next())

	// add Peer 1 and check if it Next is true and the value is the same one
	discpeers.addPeer(multiAddr[1])
	require.Equal(t, true, discpeers.next())

	np = discpeers.getNextPeer()
	require.Equal(t, np, multiAddr[1])

	require.Equal(t, false, discpeers.next())

	// add Peer 2 and check if it Next is true and the value is the same one
	discpeers.addPeer(multiAddr[2])
	require.Equal(t, true, discpeers.next())

	np = discpeers.getNextPeer()
	require.Equal(t, np, multiAddr[2])

	require.Equal(t, false, discpeers.next())

	// we should get an emptpy string
	np = discpeers.getNextPeer()
	require.Equal(t, np, peer.AddrInfo{})

}
