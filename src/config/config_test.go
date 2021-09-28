package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadJSON(t *testing.T) {
	config_object := NewEmptyConfigData() //NewDefaultConfigData()

	config_object.ReadFromJSON("custom.json")

	require.Equal(t, config_object.GetIP(), "127.0.0.1")
	require.Equal(t, config_object.GetTcpPort(), 100)
	require.Equal(t, config_object.GetUdpPort(), 101)
	require.Equal(t, config_object.GetTopicArray(), []string{"echo", "ohce"})
	require.Equal(t, config_object.GetNetwork(), "testnet")
	require.Equal(t, config_object.GetForkDigest(), "0xdlskgfn")
	require.Equal(t, config_object.GetUserAgent(), "bsc_test")

}
