package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/stretchr/testify/require"
)

const (
	loginStr = "postgresql://test:password@localhost:5432/armiarmadb"
)

func TestNewService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// create the DB client
	dbClient, err := NewDBClient(ctx, utils.EthereumNetwork, loginStr, true)
	require.NoError(t, err)

	hInfo := genNewTestHostInfo(
		t,
		utils.EthereumNetwork,
		"12D3KooW9pdHR2n4xvYU1RBEgrJMH1kd557QSXYURzEFWeEECjGn",
		"192.168.1.1",
		9000,
	)
	dbClient.PersistToDB(hInfo)
	pInfo := genNewTestPeerInfo(
		t,
		"12D3KooW9pdHR2n4xvYU1RBEgrJMH1kd557QSXYURzEFWeEECjGn",
		"migalabs-crawler",
	)
	dbClient.PersistToDB(pInfo)
	connAttemtp := models.NewConnAttempt(
		hInfo.ID,
		models.PossitiveAttempt,
		"None",
		false,
		false,
	)
	dbClient.PersistToDB(connAttemtp)

	// Wait at least 3 secs to persist
	time.Sleep(5 * time.Second)

	persistable, err := dbClient.GetPersistable(hInfo.ID.String())
	require.NoError(t, err)
	require.Equal(t, persistable.ID, hInfo.ID)
	require.Equal(t, persistable.Network, hInfo.Network)
	require.Equal(t, len(persistable.Addrs), len(hInfo.MAddrs))

	dbClient.Close()
	cancel()
}
