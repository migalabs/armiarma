package config

import (
	"testing"

	"github.com/migalabs/armiarma/src/base"
	"github.com/stretchr/testify/require"
)

func Test_ReadJSON_Success(t *testing.T) {
	baseConfigOpts := base.LogOpts{
		ModName:   "Config",
		Output:    "terminal",
		Formatter: "text",
		Level:     "debug",
	}
	config_object := NewEmptyConfigData(baseConfigOpts)

	config_object.ReadFromJSON("config_success.json")
	require.Equal(t, config_object.GetIP(), "127.0.0.1")
	require.Equal(t, config_object.GetTcpPort(), 100)
	require.Equal(t, config_object.GetUdpPort(), 101)
	require.Equal(t, config_object.GetTopicArray(), []string{"BeaconBlock"})
	require.Equal(t, config_object.GetNetwork(), "testnet")
	require.Equal(t, config_object.GetEth2Endpoint(), "https://infura.test.endpoint")
	require.Equal(t, config_object.GetForkDigest(), "Altair")
	require.Equal(t, config_object.GetUserAgent(), "bsc_test")
	require.Equal(t, config_object.GetPrivKey(), "026c60367b01fe3d7c7460bce1d585260ce465fa0abcb6e13619f88bf0dad54f")
}

func Test_ReadJSON_Fail(t *testing.T) {
	baseConfigOpts := base.LogOpts{
		ModName:   "Config",
		Output:    "terminal",
		Formatter: "text",
		Level:     "debug",
	}
	config_object := NewEmptyConfigData(baseConfigOpts)

	config_object.ReadFromJSON("config_fail.json")

	require.Equal(t, config_object.GetIP(), "127.0.0.1")
	require.Equal(t, config_object.GetTcpPort(), 100)
	require.Equal(t, config_object.GetUdpPort(), 101)
	require.Equal(t, len(config_object.GetTopicArray()), 0)
	require.Equal(t, config_object.GetNetwork(), "testnet")
	require.Equal(t, config_object.GetEth2Endpoint(), "https://infura.test.endpoint")
	require.Equal(t, config_object.GetForkDigest(), "")
	require.Equal(t, config_object.GetUserAgent(), "bsc_test")
}
