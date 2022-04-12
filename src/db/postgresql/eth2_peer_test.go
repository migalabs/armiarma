package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/utils"
	//"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestPeerLoadAndStore(t *testing.T) {
	//logrus.SetLevel(logrus.DebugLevel)
	url := "postgres://armiarmacrawler:ar_Mi_arm4@localhost:5432/armiarmadb"
	psqlDB, err := ConnectToDB(context.Background(), url)
	require.Equal(t, nil, err)

	msgMet := models.MessageMetric{
		Count:            12,
		FirstMessageTime: parseTime("2021-08-23T01:00:00.000Z", t),
		LastMessageTime:  parseTime("2021-08-23T01:00:00.000Z", t),
	}

	bStatus, err := models.ParseBeaconStatusFromBasicTypes(
		parseTime("2021-08-23T01:00:00.000Z", t),
		"0xafcaaba0",
		"0x7ed9458a2c9d6af4741e26df3410e87de37dfbdc5ac68aa810be7b94ff13328c",
		int64(123),
		"0x7ed9458a2c9d6af4741e26df3410e87de37dfbdc5ac68aa810be7b94ff13328c",
		int64(12345),
	)
	require.Equal(t, nil, err)
	// generate first peer
	peer1 := models.Peer{
		PeerId:                "Peer1",
		Pubkey:                "ASWDSFAWSF",
		NodeId:                "Node1",
		UserAgent:             "TestPeer",
		ClientName:            "TestClient",
		ClientOS:              "linux",
		ClientVersion:         "v1.0.0",
		BlockchainNodeENR:     "AWAW123111231J231K23JH123K12",
		Ip:                    "123.12.12.12",
		Country:               "Spain",
		CountryCode:           "ES",
		City:                  "Barcelona",
		Latency:               float64(0.12),
		Protocols:             []string{"pubsub", "gossipsub"},
		ProtocolVersion:       "pubsub:v1",
		ConnectedDirection:    []string{"inbound"},
		IsConnected:           true,
		Attempted:             true,
		Succeed:               true,
		Attempts:              1,
		Error:                 []string{"None"},
		LastErrorTimestamp:    parseTime("0001-01-01T00:00:00.000Z", t),
		Deprecated:            false,
		LastIdentifyTimestamp: parseTime("2021-08-23T01:00:00.000Z", t),
		NegativeConnAttempts:  []time.Time{parseTime("2021-08-23T01:00:00.000Z", t)},
		ConnectionTimes:       []time.Time{parseTime("2021-08-23T01:00:00.000Z", t)},
		DisconnectionTimes:    []time.Time{parseTime("2021-08-23T01:00:00.000Z", t)},
		MetadataRequest:       true,
		MetadataSucceed:       true,
		LastExport:            123123123,
		MessageMetrics:        make(map[string]models.MessageMetric),
		BeaconStatus:          bStatus,
	}
	// generate multiaddres
	addreses := []string{"/ip4/51.89.42.176/tcp/9000", "/ip4/123.123.123.123/tcp/9000"}
	for _, ma := range addreses {
		maddres, err := utils.UnmarshalMaddr(ma)
		if err != nil {
			log.Warnf("unable to generate multiaddres from %s. %s", ma, err.Error())
			continue
		}
		peer1.MAddrs = append(peer1.MAddrs, maddres)
	}

	// Generate message metrics
	peer1.MessageMetrics["testTopic"] = msgMet
	peer1.MessageMetrics["testTopic2"] = msgMet

	psqlDB.StorePeer(peer1.PeerId, peer1)

	readPeer, ok := psqlDB.LoadPeer(peer1.PeerId)
	require.Equal(t, true, ok)

	require.Equal(t, readPeer.PeerId, peer1.PeerId)
	require.Equal(t, readPeer.Pubkey, peer1.Pubkey)
	require.Equal(t, readPeer.NodeId, peer1.NodeId)
	require.Equal(t, readPeer.UserAgent, peer1.UserAgent)
	require.Equal(t, readPeer.ClientName, peer1.ClientName)
	require.Equal(t, readPeer.ClientOS, peer1.ClientOS)
	require.Equal(t, readPeer.ClientVersion, peer1.ClientVersion)
	require.Equal(t, readPeer.BlockchainNodeENR, peer1.BlockchainNodeENR)

	require.Equal(t, readPeer.Ip, peer1.Ip)
	require.Equal(t, readPeer.Country, peer1.Country)
	require.Equal(t, readPeer.CountryCode, peer1.CountryCode)
	require.Equal(t, readPeer.City, peer1.City)
	require.Equal(t, readPeer.Latency, peer1.Latency)
	require.Equal(t, readPeer.Protocols, peer1.Protocols)
	require.Equal(t, readPeer.ProtocolVersion, peer1.ProtocolVersion)

	require.Equal(t, readPeer.ConnectedDirection, peer1.ConnectedDirection)
	require.Equal(t, readPeer.IsConnected, peer1.IsConnected)
	require.Equal(t, readPeer.Attempted, peer1.Attempted)
	require.Equal(t, readPeer.Succeed, peer1.Succeed)
	require.Equal(t, readPeer.Attempts, peer1.Attempts)
	require.Equal(t, readPeer.Error, peer1.Error)
	require.Equal(t, readPeer.LastErrorTimestamp, peer1.LastErrorTimestamp)
	require.Equal(t, readPeer.Deprecated, peer1.Deprecated)
	require.Equal(t, readPeer.LastIdentifyTimestamp, peer1.LastIdentifyTimestamp)

	require.Equal(t, readPeer.NegativeConnAttempts, peer1.NegativeConnAttempts)
	require.Equal(t, readPeer.ConnectionTimes, peer1.ConnectionTimes)
	require.Equal(t, readPeer.DisconnectionTimes, peer1.DisconnectionTimes)
	require.Equal(t, readPeer.MetadataRequest, peer1.MetadataRequest)
	require.Equal(t, readPeer.MetadataSucceed, peer1.MetadataSucceed)

	require.Equal(t, readPeer.LastExport, peer1.LastExport)

	require.Equal(t, bStatus, readPeer.BeaconStatus)

	require.Equal(t, len(readPeer.MessageMetrics), 2)
	require.Equal(t, readPeer.MessageMetrics["testTopic"], msgMet)
	require.Equal(t, readPeer.MessageMetrics["testTopic2"], msgMet)

	// Update the peerInfo
	peer1.UserAgent = "TestPeer"
	psqlDB.StorePeer(peer1.PeerId, peer1)

	readPeer, ok = psqlDB.LoadPeer(peer1.PeerId)
	require.Equal(t, true, ok)

	require.Equal(t, readPeer.UserAgent, peer1.UserAgent)

	// Delete peers from the test db
	psqlDB.DeletePeer(peer1.PeerId)
	peers := psqlDB.GetPeers()
	require.Equal(t, 0, len(peers))
}

func parseTime(strTime string, t *testing.T) time.Time {
	parsedTime, err := time.Parse(time.RFC3339, strTime)
	require.NoError(t, err)
	return parsedTime
}

func TestLastToolActivity(t *testing.T) {
	//logrus.SetLevel(logrus.DebugLevel)
	url := "postgres://armiarmacrawler:ar_Mi_arm4@localhost:5432/armiarmadb"
	psqlDB, err := ConnectToDB(context.Background(), url)
	require.Equal(t, nil, err)

	// generate first peer
	peer1 := models.Peer{
		PeerId:               "Peer1",
		NegativeConnAttempts: []time.Time{parseTime("2022-01-21T01:00:01.000Z", t), parseTime("2022-01-21T01:00:04.000Z", t)},
		ConnectionTimes:      []time.Time{parseTime("2022-01-21T01:00:02.000Z", t), parseTime("2022-01-21T01:00:05.000Z", t)},
		DisconnectionTimes:   []time.Time{parseTime("2022-01-21T01:00:03.000Z", t), parseTime("2022-01-21T01:00:06.000Z", t)},
	}
	psqlDB.StorePeer(peer1.PeerId, peer1)

	// generate first peer
	peer2 := models.Peer{
		PeerId:               "Peer2",
		NegativeConnAttempts: []time.Time{parseTime("2022-01-22T01:00:01.000Z", t), parseTime("2022-03-22T01:00:04.000Z", t)},
		ConnectionTimes:      []time.Time{parseTime("2022-01-22T01:00:02.000Z", t), parseTime("2022-01-22T01:00:05.000Z", t)},
		DisconnectionTimes:   []time.Time{parseTime("2022-01-22T01:00:03.000Z", t), parseTime("2022-01-22T01:00:06.000Z", t)},
	}
	psqlDB.StorePeer(peer2.PeerId, peer2)

	//peers := psqlDB.GetPeers()
	//require.Equal(t, 2, len(peers))
	lastActivity, err := psqlDB.GetLastActivityTime()
	require.Equal(t, nil, err)
	require.Equal(t, parseTime("2022-03-22T01:00:04.000Z", t), lastActivity)

	// Delete peers from the test db
	psqlDB.DeletePeer(peer1.PeerId)
	psqlDB.DeletePeer(peer2.PeerId)

	peers := psqlDB.GetPeers()
	require.Equal(t, 0, len(peers))

}

func TestPeerConnectedCheck(t *testing.T) {
	//logrus.SetLevel(logrus.DebugLevel)
	url := "postgres://armiarmacrawler:ar_Mi_arm4@localhost:5432/armiarmadb"
	psqlDB, err := ConnectToDB(context.Background(), url)
	require.Equal(t, nil, err)

	// generate first peer
	peer1 := models.Peer{
		PeerId:          "Peer1",
		IsConnected:     true,
		ConnectionTimes: []time.Time{parseTime("2022-01-21T01:00:01.000Z", t)},
	}
	psqlDB.StorePeer(peer1.PeerId, peer1)

	// generate first peer
	peer2 := models.Peer{
		PeerId:               "Peer2",
		IsConnected:          false,
		NegativeConnAttempts: []time.Time{parseTime("2022-01-22T01:00:01.000Z", t), parseTime("2022-03-22T01:00:04.000Z", t)},
	}
	psqlDB.StorePeer(peer2.PeerId, peer2)

	ok := psqlDB.CheckPeersSummaryTableStatus()
	require.Equal(t, true, ok)

	readPeer, ok := psqlDB.LoadPeer(peer1.PeerId)
	require.Equal(t, true, ok)
	require.Equal(t, 1, len(readPeer.DisconnectionTimes))
	require.Equal(t, false, readPeer.IsConnected)

	// Delete peers from the test db
	psqlDB.DeletePeer(peer1.PeerId)
	psqlDB.DeletePeer(peer2.PeerId)

	peers := psqlDB.GetPeers()
	require.Equal(t, 0, len(peers))
}
