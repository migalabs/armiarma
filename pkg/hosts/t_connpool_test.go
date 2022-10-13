package hosts

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/stretchr/testify/require"
)

var testDBstr = "postgresql://test:password@localhost:5432/armiarmadb"

func TestConnPool(t *testing.T) {

	hostID := new(peer.ID)
	dbCli, err := postgresql.NewDBClient(context.Background(), testDBstr, false)
	require.NoError(t, err)

	// Gen new ConnPool
	connPool := NewConnectionPool(*hostID, dbCli)

	// ---> Case 1 (standard): Connection -> Metadata Succeeded -> Disconnect
	peer1, err := peer.Decode("12D3KooW9pdHR2n4xvYU1RBEgrJMH1kd557QSXYURzEFWeEECjGn")
	require.NoError(t, err)
	// add ConEvent
	connEv1 := models.NewConnEvent(peer1)
	err = connPool.AddNewEvent(connEv1)
	require.NoError(t, err)
	// add ConnInfo
	lat := time.Duration(int(100000))
	cInfo := models.ConnInfo{
		Direction:  models.InboundConnection,
		ConnTime:   ParseTestTime(t, "2022-10-12T00:00:00.000Z"),
		Latency:    lat,
		Identified: true,
		Att:        make(map[string]interface{}),
		Error:      utils.NoneErr,
	}
	connPool.AddConnInfo(peer1, cInfo)
	// Check len of the ConnectionPool
	l := len(connPool.connPool)
	require.Equal(t, 1, l)
	// if is persistable
	persistable := connPool.connPool[peer1.String()].IsReadyToPersist()
	require.Equal(t, false, persistable)

	// add the disconnection
	cDisc := models.EndConnInfo{
		DiscTime: ParseTestTime(t, "2022-10-12T02:00:00.000Z"),
	}
	connPool.AddDisconn(peer1, cDisc)
	// Check len of the ConnectionPool // it should have been persisted already and deleted
	l = len(connPool.connPool)
	require.Equal(t, 0, l)

	// ----> Case 2 (edgy case): Connection -> Disconnect -> Metadata Succeeded
	peer1, err = peer.Decode("12D3KooW9pdHR2n4xvYU1RBEgrJMH1kd557QSXYURzEFWeEECjGn")
	require.NoError(t, err)
	// add ConEvent
	connEv2 := models.NewConnEvent(peer1)
	err = connPool.AddNewEvent(connEv2)
	require.NoError(t, err)

	// // add the disconnection
	cDisc = models.EndConnInfo{
		DiscTime: ParseTestTime(t, "2022-10-12T02:00:00.000Z"),
	}
	connPool.AddDisconn(peer1, cDisc)
	// Check len of the ConnectionPool // it should have been persisted already and deleted
	l = len(connPool.connPool)
	require.Equal(t, 1, l)

	// if is persistable
	persistable = connPool.connPool[peer1.String()].IsReadyToPersist()
	require.Equal(t, false, persistable)
	// add ConnInfo
	lat = time.Duration(int(100000))
	cInfo = models.ConnInfo{
		Direction:  models.OutboundConnection,
		ConnTime:   ParseTestTime(t, "2022-10-12T00:00:00.000Z"),
		Latency:    lat,
		Identified: true,
		Att:        make(map[string]interface{}),
		Error:      utils.NoneErr,
	}
	connPool.AddConnInfo(peer1, cInfo)
	// Check len of the ConnectionPool
	l = len(connPool.connPool)
	require.Equal(t, 0, l)

	dbCli.Close()
}
