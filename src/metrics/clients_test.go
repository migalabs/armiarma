package metrics

import (
	"testing"
	"github.com/stretchr/testify/require"
	"fmt"
)


func Test_Clients(t *testing.T) {
	fmt.Println("hj there")
	require.Equal(t, 1, 1)
	fmt.Println("hj there")

	clients := NewClients()

	clients.AddClientVersion("cl1", "ver1")
	clients.AddClientVersion("cl1", "ver2")
	clients.AddClientVersion("cl1", "ver2")

	fmt.Println("asd", clients.Clients)
}
