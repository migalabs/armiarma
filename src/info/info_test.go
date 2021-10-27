package info

import (
	"strings"
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
	require.Equal(t, info_object.GetTopicArray(), []string{"echo", "ohce"})
	require.Equal(t, info_object.GetNetwork(), "testnet")
	require.Equal(t, info_object.GetForkDigest(), "0xdlskgfn")
	require.Equal(t, info_object.GetUserAgent(), "bsc_test")
	require.Equal(t, info_object.GetPrivKeyString(), "026c60367b01fe3d7c7460bce1d585260ce465fa0abcb6e13619f88bf0dad54f")

}

func Test_CustomInfoDataFail(t *testing.T) {
	stdOpts := base.LogOpts{
		Output:    "terminal",
		Formatter: "text",
	}
	info_object := NewCustomInfoData("../config/config_fail.json", stdOpts)

	require.Equal(t, info_object.GetIP().String(), "127.0.0.1")
	require.Equal(t, info_object.GetTcpPort(), 100)
	require.Equal(t, info_object.GetUdpPort(), 101)
	require.Equal(t, len(info_object.GetTopicArray()), len(strings.Split(DefaultTopicArray, ",")))
	require.Equal(t, info_object.GetNetwork(), "testnet")
	require.Equal(t, info_object.GetForkDigest(), DefaultForkDigest)
	require.Equal(t, info_object.GetUserAgent(), "bsc_test")

}
