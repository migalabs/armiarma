package info

import (
	"testing"

	"github.com/migalabs/armiarma/src/base"
	"github.com/stretchr/testify/require"
)

func Test_CustomInfoDataSuccess(t *testing.T) {
	stdOpts := base.LogOpts{
		Output:    "terminal",
		Formatter: "text",
	}
	info_object := NewCustomInfoData("../config/config_success.json", stdOpts)

	require.Equal(t, info_object.GetIP().String(), "127.0.0.1")
	require.Equal(t, info_object.GetTcpPort(), 100)
	require.Equal(t, info_object.GetUdpPort(), 101)
	require.Equal(t, info_object.GetTopicArray(), []string{"/eth2/b5303f2a/beacon_block/ssz_snappy"})
	require.Equal(t, info_object.GetNetwork(), "testnet")
	require.Equal(t, info_object.GetForkDigest(), "Mainnet")
	require.Equal(t, info_object.GetUserAgent(), "bsc_test")
	require.Equal(t, info_object.GetPrivKeyString(), "026c60367b01fe3d7c7460bce1d585260ce465fa0abcb6e13619f88bf0dad54f")
	require.Equal(t, info_object.GetLogLevel(), "debug")
	require.Equal(t, info_object.GetDBPath(), "/etc")
	require.Equal(t, info_object.GetDBType(), "memory")
}

// Test_CustomInfoDataFail
// * This method tests the InfoData creation using a partly failing config file
func Test_CustomInfoDataFail(t *testing.T) {
	stdOpts := base.LogOpts{
		Output:    "terminal",
		Formatter: "text",
	}
	info_object := NewCustomInfoData("../config/config_fail.json", stdOpts)

	require.Equal(t, 5, len(info_object.GetTopicArray())) // at the moment there are five possible topics for one fork digest
	require.Equal(t, info_object.GetForkDigest(), "Mainnet")

}
