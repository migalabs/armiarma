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
	connAttemtp1 := models.NewConnAttempt(
		host1.ID,
		models.PossitiveAttempt,
		utils.NoError,
		false,
		false,
	)

	// Insert new HostInfo
	q, args := dbCli.UpsertHostInfo(host1)
	_, err = dbCli.SingleQuery(q, args)
	require.NoError(t, err)

	// update hostInfo with peerInfo
	q, args = dbCli.UdatePeerInfo(peer1)
	_, err = dbCli.SingleQuery(q, args)
	require.NoError(t, err)

	// Update hostInfo's with NewConnAttemtp
	q, args = dbCli.UpdateConnAttempt(connAttemtp1)
	_, err = dbCli.SingleQuery(q, args)
	require.NoError(t, err)

	// Read peer info
	ok := dbCli.PeerInfoExists(host1.ID)
	require.Equal(t, true, ok)

	// Check if the FullPeerInfo is matches what we have inserted
	rHostInfo, err := dbCli.GetFullHostInfo(host1.ID)
	require.NoError(t, err)
	require.Equal(t, host1.Newtwork, rHostInfo.Network)
	require.Equal(t, host1.ID.String(), rHostInfo.ID.String())
	require.Equal(t, host1.IP, rHostInfo.IP)
	require.Equal(t, host1.TCP, rHostInfo.TCP)
	require.Equal(t, host1.UDP, rHostInfo.UDP)
	require.Equal(t, host1.MAddrs, rHostInfo.MAddrs)
	require.Equal(t, peer1.RemotePeer.String(), rHostInfo.PeerInfo.RemotePeer.String())
	require.Equal(t, peer1.UserAgent, rHostInfo.PeerInfo.UserAgent)
	require.Equal(t, peer1.ProtocolVersion, rHostInfo.PeerInfo.ProtocolVersion)
	require.Equal(t, peer1.Protocols, rHostInfo.PeerInfo.Protocols)
	require.Equal(t, peer1.Latency, rHostInfo.PeerInfo.Latency)
	require.Equal(t, false, rHostInfo.ControlInfo.Deprecated)
	require.Equal(t, false, rHostInfo.ControlInfo.hasLeftNetwork)
	require.Equal(t, true, rHostInfo.ControlInfo.Attemtped)
	require.Equal(t, time.Time{}.Unix(), rHostInfo.ControlInfo.LastActivity.Unix())
	require.Equal(t, connAttemtp1.LastConnAttempt.Unix(), rHostInfo.ControlInfo.LastConnAttempt.Unix())
	require.Equal(t, connAttemtp1.Error, rHostInfo.ControlInfo.Error)

	// TODO: extend this tests to include edgy cases with Connections and Disconnections
	// TODO: extend this test to include deprecated and left network SQL queries
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

	// 1. Multiaddress
	maddrs, err := ma.NewMultiaddr(fmt.Sprintf("/ipv/%s/tcp/%s/p2p/%s", ip, tcp, peerStr))
	require.NoError(t, err)

	var arrMaddr []ma.Multiaddr
	arrMaddr = append(arrMaddr, maddrs)

	// create a new peer
	hostInfo := models.NewHostInfo(
		pID,
		network,
		models.WithIPAndPorts(ip, tcp, tcp),
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
	protocolVersion := "protocol-version"
	protocols := []string{
		"discv5",
		"gossipsub",
		"rpcs",
	}

	latency := 1 * time.Millisecond

	// create a new peer
	peerInfo := models.NewPeerInfo(pID, userAgent, protocolVersion, protocols, latency)

	return peerInfo
}
