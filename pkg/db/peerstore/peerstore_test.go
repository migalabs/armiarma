package peerstore

import (
	"testing"

	"github.com/migalabs/armiarma/pkg/utils"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/stretchr/testify/require"
)

var peerIDs []string = []string{
	"16Uiu2HAm3r5n99y8TCyz83szHtpnwhvoc13ZyJjWfGRPeRp89Wa8",
	"16Uiu2HAm661DnLn4Ff11V8iJWk76Ykep2pfSVZzo6H6mV3iVN7g5",
	"16Uiu2HAm5NTiy9xDgumG9pNyPrfzJ66xZYLCx9ZG2eirbRimr3T5",
}

var peerAddrs [][]string = [][]string{
	{"/ip4/127.0.0.1/tcp/9001"},
	{"/ip4/127.0.0.1/tcp/9002"},
	{"/ip4/127.0.0.1/tcp/9003"},
}

func composeTestPeers(t *testing.T) []PersistablePeer {
	persistables := make([]PersistablePeer, 0, 3)

	for idx, pidStr := range peerIDs {
		// Get the peerID
		peerID, err := peer.Decode(pidStr)
		require.NoError(t, err)
		// Get the maddrs
		maddrs, err := ma.NewMultiaddr(peerAddrs[idx][0])
		require.NoError(t, err)

		maddresses := make([]ma.Multiaddr, 0)
		maddresses = append(maddresses, maddrs)
		// Compose the AddrInfo
		persis := NewPersistable(peerID, maddresses, utils.EthereumNetwork)
		persistables = append(persistables, *persis)
	}
	// check that the actual len if 3
	require.Equal(t, 3, len(persistables))
	return persistables
}

func TestAddInfo(t *testing.T) {
	// Order of the test
	// 1. Generate a new db
	// 2. Store all the Peers
	// 3. Check that the len is 3
	// 4. Remove 1 by 1 the peers
	// 5. Check len again
	// 6. Remove the db folder

	persistables := composeTestPeers(t)

	peerstr := NewPeerstore("./")

	// insert into DB the peers
	for idx, persis := range persistables {
		err := peerstr.StorePeer(persis)
		require.NoError(t, err)

		// -- checks --
		// len of peerstore
		pids := peerstr.GetPeerIDs()
		ps := peerstr.GetAllPeers()
		require.Equal(t, idx+1, len(pids))
		require.Equal(t, idx+1, len(ps))

		// load peer
		newP, ok := peerstr.LoadPeer(persis.ID.String())
		require.Equal(t, true, ok)

		// content of the stored peer
		require.Equal(t, newP.ID, persis.ID)
		require.Equal(t, newP.Addrs, persis.Addrs)
		require.Equal(t, newP.Network, persis.Network)
	}

	// remove from the peerstore the peers
	for idx, persis := range persistables {
		peerstr.DeletePeer(persis.ID.String())

		// -- checks --
		// len of peerstore
		pids := peerstr.GetPeerIDs()
		ps := peerstr.GetAllPeers()
		require.Equal(t, len(persistables)-(idx+1), len(pids))
		require.Equal(t, len(persistables)-(idx+1), len(ps))

		// load peer
		_, ok := peerstr.LoadPeer(persis.ID.String())
		require.Equal(t, false, ok)
	}

	// remove test peerstore
	err := utils.RemoveFolderOrFile(peerstoreFile)
	require.NoError(t, err)
}
