package postgresql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	loginStr = "postgresql://test:password@localhost:5432/armiarmadb"
)

func TestNewService(t *testing.T) {

	// create the DB client
	dbClient, err := NewDBClient(context.Background(), loginStr, true)
	require.NoError(t, err)

	dbClient.Close()
}
