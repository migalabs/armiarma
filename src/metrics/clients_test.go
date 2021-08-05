package metrics

import (
	"testing"
	"github.com/stretchr/testify/require"
	"fmt"
)

func Test_Clients(t *testing.T) {
	clients := NewClients()

	clients.AddClientVersion("prysm", "v1.0.0")
	clients.AddClientVersion("prysm", "v1.0.0")
	clients.AddClientVersion("prysm", "v1.0.0")
	clients.AddClientVersion("prysm", "v1.0.0")
	clients.AddClientVersion("prysm", "v1.0.1")
	clients.AddClientVersion("prysm", "v1.0.1")
	clients.AddClientVersion("prysm", "v1.0.0")
	clients.AddClientVersion("lighthouse", "v1.0.0")
	clients.AddClientVersion("lighthouse", "v1.0.0")
	clients.AddClientVersion("lighthouse", "v1.0.0")
	clients.AddClientVersion("lighthouse", "v2.0.0")
	clients.AddClientVersion("lighthouse", "v2.0.0")
	clients.AddClientVersion("lighthouse", "v1.0.0")
	clients.AddClientVersion("lighthouse", "v2.0.0")
	clients.AddClientVersion("lighthouse", "v3.0.0")

  require.Equal(t, clients.GetPeersOfClient("prysm"), 7)
	require.Equal(t, clients.GetPeersOfClient("lighthouse"), 8)
	require.Equal(t, len(clients.GetClientNames()), 2)

	require.Equal(t, len(clients.Clients["prysm"]), 2)
	require.Equal(t, len(clients.Clients["lighthouse"]), 3)
}
