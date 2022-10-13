package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestConnEventInPSQL(t *testing.T) {

	loginStr := "postgresql://test:password@localhost:5432/armiarmadb"
	// generate a new DBclient with the given login string
	dbCli, err := NewDBClient(context.Background(), loginStr, false)
	defer func() {
		dbCli.Close()
	}()
	require.NoError(t, err)
	// initialize only the ConnEvent Table separatelly
	err = dbCli.InitConnEventTable()
	require.NoError(t, err)

	// insert a new row for the
	connEv := genNewTestConnEvent(t, "12D3KooW9pdHR2n4xvYU1RBEgrJMH1kd557QSXYURzEFWeEECjGn")

	err = dbCli.InsertNewConnEvent(connEv)
	require.NoError(t, err)

	// phase 2 -> (Inserting the same peer should result with an error)

	err = dbCli.InsertNewConnEvent(connEv)
	require.NotEqual(t, nil, err)

}

func genNewTestConnEvent(t *testing.T, peerStr string) *models.ConnEvent {
	peer1, err := peer.Decode(peerStr)
	require.NoError(t, err)
	// add ConEvent
	connEv := models.NewConnEvent(peer1)
	lat := time.Duration(int(100000))
	cInfo := models.ConnInfo{
		Direction:  models.InboundConnection,
		ConnTime:   utils.ParseTestTime(t, "2022-10-12T00:00:00.000Z"),
		Latency:    lat,
		Identified: true,
		Att:        make(map[string]interface{}),
		Error:      utils.NoneErr,
	}
	connEv.AddConnInfo(cInfo)

	// add the disconnection
	cDisc := models.EndConnInfo{
		DiscTime: utils.ParseTestTime(t, "2022-10-12T02:00:00.000Z"),
	}
	connEv.AddDisconn(cDisc)

	return connEv
}
