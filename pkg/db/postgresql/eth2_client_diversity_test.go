package postgresql

import (
	"context"
	"testing"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestClientDiversityInsert(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	url := "postgres://armiarmacrawler:ar_Mi_arm4@localhost:5432/armiarmadb"
	ethmodel := NewEth2Model("eth2")
	psqlDB, err := ConnectToDB(context.Background(), url, &ethmodel)
	require.Equal(t, nil, err)

	// create the ClientDiversity obj for adding it to the SQL
	date := parseTime("2021-08-23T01:00:00.000Z", t)
	diversity := models.NewClientDiversity()
	diversity.Timestamp = date
	diversity.Prysm = 1
	diversity.Lighthouse = 2
	diversity.Teku = 3
	diversity.Nimbus = 4
	diversity.Lodestar = 5
	diversity.Grandine = 6
	diversity.Others = 7

	psqlDB.StoreClientDiversitySnapshot(diversity)

	diversity2, err := psqlDB.LoadClientDiversitySnapshot(date)
	require.Equal(t, nil, err)

	require.Equal(t, diversity, diversity2)
}
