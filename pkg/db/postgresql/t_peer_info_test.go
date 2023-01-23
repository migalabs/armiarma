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

	// insert a new row for HostInfo and PeerInfo
	host1 := genNewTestHostInfo(
		t,
		network,
		"12D3KooW9pdHR2n4xvYU1RBEgrJMH1kd557QSXYURzEFWeEECjGn",
		"192.168.1.1",
		"9000",
	)
	peer1 := genNewTestPeerInfo(
		t,
		"12D3KooW9pdHR2n4xvYU1RBEgrJMH1kd557QSXYURzEFWeEECjGn",
		"migalabs-crawler",
	)

	q, args := dbCli.UpsertHostInfo(host1)
	_, err = dbCli.SingleQuery(q, args)
	require.NoError(t, err)

	q, args := dbCli.UpsertPeerInfo(peer1)
	_, err = dbCli.SingleQuery(q, args)
	require.NoError(t, err)

	// Read peer info
	ok := dbCli.PeerInfoExists(host1.ID)
	require.Equal(t, true, ok)
}

func genNewTestHostInfo(
	t *testing.T,
	network utils.NetworkType,
	peerStr string,
	ip string,
	tcp string) *models.PeerInfo {

	// Decode PeerId
	pID, err := peer.Decode(peerStr)
	require.NoError(t, err)

	// compose the peer info

	// 1. Multiaddress
	maddrs, err := ma.NewMultiaddr(fmt.Sprintf("/ipv/%s/tcp/%s/p2p/%s", ip, tcp, peerStr))
	require.NoError(t, err)

	var arrMaddr []ma.Multiaddr
	arrMaddr = append(arrMaddr, maddrs)

	// create a new peer
	hostInfo := models.NewHostInfo(
		pID,
		network,
		WithIPAndPorts(ip, tcp, tcp),
	)

	return hostInfo
}

func genNewTestPeerInfo(
	t *testing.T,
	peerStr string,
	userAgent string) *models.PeerInfo {

	// Decode the peer info
	pID, err := peer.Decode(peerStr)
	require.NoError(t, err)

	// Protocols and UserAgent
	protocolVersion = "protocol-version"
	protocols = []string{
		"discv5",
		"gossipsub",
		"rpcs",
	}

	peerInfo.Latency = 1 * time.Millisecond

	// create a new peer
	peerInfo := models.NewPeerInfo(pID, userAgent, protocolVersion, protocol, latency)

	return peerInfo
}
