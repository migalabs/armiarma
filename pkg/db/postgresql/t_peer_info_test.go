package postgresql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestPeerInfoInPSQL(t *testing.T) {
	network := utils.EthereumNetwork
	loginStr := "postgresql://test:password@localhost:5432/armiarmadb"
	// generate a new DBclient with the given login string
	dbCli, err := NewDBClient(context.Background(), network, loginStr, false)
	defer func() {
		dbCli.Close()
	}()
	require.NoError(t, err)
	// initialize only the ConnEvent Table separatelly
	err = dbCli.InitPeerInfoTable()
	require.NoError(t, err)

	// insert a new row for the
	peer1 := genNewTestPeer(t, network, "12D3KooW9pdHR2n4xvYU1RBEgrJMH1kd557QSXYURzEFWeEECjGn", "migalabs-crawler", "9000", "192.168.1.1")

	err = dbCli.InsertNewPeerInfo(peer1)
	require.NoError(t, err)
}

func genNewTestPeer(
	t *testing.T,
	network utils.NetworkType,
	peerStr string,
	ip string,
	userAgent string,
	port string) *models.PeerInfo {
	pID, err := peer.Decode(peerStr)
	require.NoError(t, err)
	// create a new peer
	peerInfo := models.NewPeerInfo(pID, network)

	// compose the peer info

	// 1. Multiaddress
	maddrs, err := ma.NewMultiaddr(fmt.Sprintf("/ipv/%s/tcp/%s/p2p/%s", ip, port, peerStr))
	require.NoError(t, err)
	var arrMaddr []ma.Multiaddr
	arrMaddr = append(arrMaddr, maddrs)
	peerInfo.MAddrs = arrMaddr

	// 2. Protocols and UserAgent
	peerInfo.UserAgent = userAgent
	peerInfo.ProtocolVersion = "protocol-version"
	peerInfo.Protocols = []string{
		"discv5",
		"gossipsub",
		"rpcs",
	}
	peerInfo.Latency = 1 * time.Millisecond

	// TODO: Add ControlInfo when finished

	return peerInfo
}
