package metrics

import (
	//log "github.com/sirupsen/logrus"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Peer(t *testing.T) {
	// TODO
	require.Equal(t, 1, 1)
}

// Testing the HInfo fetch into the peer struct
func Test_FetchHostInfo(t *testing.T) {
	// generate base peer
	peerBase := Peer{}

	// generate the fetching info
	bhost := BasicHostInfo{
		TimeStamp: time.Now(),
		// Peer Host/Node Info
		PeerID:          "Peer1",
		NodeID:          "Node1",
		UserAgent:       "TestPeer",
		ProtocolVersion: "Protocol",
		Addrs:           "/ip4/95.169.232.98/tcp/9000",
		PubKey:          "PubKey",
		RTT:             2 * time.Millisecond,
		Protocols:       make([]string, 0),
		// Information regarding the metadata exchange
		Direction: "inbound",
		// Metadata requested
		MetadataRequest: true,
		MetadataSucceed: false,
	}

	peerBase.FetchHostInfo(bhost)

	// Peer Host/Node Info
	require.Equal(t, peerBase.PeerId, "Peer1")
	require.Equal(t, peerBase.NodeId, "Node1")
	require.Equal(t, peerBase.UserAgent, "TestPeer")
	require.Equal(t, peerBase.Addrs, "/ip4/95.169.232.98/tcp/9000")
	require.Equal(t, peerBase.Ip, "95.169.232.98")
	require.Equal(t, peerBase.Country, "Spain")
	require.Equal(t, peerBase.City, "Barcelona")
	require.Equal(t, peerBase.Pubkey, "PubKey")
	require.Equal(t, float64(2)/1000, peerBase.Latency)
	require.Equal(t, peerBase.MetadataRequest, true)
	require.Equal(t, peerBase.MetadataSucceed, false)

	// Generate second host Info to fetch the info
	// generate the fetching info
	bhost2 := BasicHostInfo{
		TimeStamp: time.Now(),
		// Peer Host/Node Info
		PeerID:          "",
		NodeID:          "",
		UserAgent:       "UpdateUser",
		ProtocolVersion: "",
		Addrs:           "/ip4/212.230.135.2/tcp/9000",
		PubKey:          "PubKey",
		RTT:             3 * time.Millisecond,
		Protocols:       make([]string, 0),
		// Information regarding the metadata exchange
		Direction: "inbound",
		// Metadata requested
		MetadataRequest: false,
		MetadataSucceed: false,
	}

	peerBase.FetchHostInfo(bhost2)

	// Check if fetch was successfull
	require.Equal(t, peerBase.PeerId, "Peer1")
	require.Equal(t, peerBase.NodeId, "Node1")
	require.Equal(t, peerBase.UserAgent, "UpdateUser")
	require.Equal(t, peerBase.Addrs, "/ip4/212.230.135.2/tcp/9000")
	require.Equal(t, peerBase.Ip, "212.230.135.2")
	require.Equal(t, peerBase.Country, "Spain")
	require.Equal(t, peerBase.City, "Alcobendas")
	require.Equal(t, peerBase.Pubkey, "PubKey")
	require.Equal(t, float64(3)/1000, peerBase.Latency)
	require.Equal(t, peerBase.MetadataRequest, true)
	require.Equal(t, peerBase.MetadataSucceed, false)
}
