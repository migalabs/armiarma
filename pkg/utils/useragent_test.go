package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type clientInfoTest struct {
	userAgent     string
	clientName    string
	clientVersion string
	clientOS      string
	clientArch    string
}

var Eth2TestClients []clientInfoTest = []clientInfoTest{
	{
		userAgent:     "teku/teku/v21.8.2/linux-x86_64/corretto-java-16",
		clientName:    "teku",
		clientVersion: "v21.8.2",
		clientOS:      "linux",
		clientArch:    "x86_64",
	},
	{
		userAgent:     "teku/teku/v21.7.0+9-g77b4b9e/linux-x86_64/-ubuntu-openjdk64bitservervm-java-11",
		clientName:    "teku",
		clientVersion: "v21.7.0",
		clientOS:      "linux",
		clientArch:    "x86_64",
	},
	{
		userAgent:     "Prysm/v1.4.3/8bca66ac6408a03af52d65541f58384007ed50ef",
		clientName:    "prysm",
		clientVersion: "v1.4.3",
		clientOS:      "unknown",
		clientArch:    "unknown",
	},
	{
		userAgent:     "Prysm/v1.3.8-hotfix+6c0942/6c09424feb3141b96016bed817d7ade1cd75deb7",
		clientName:    "prysm",
		clientVersion: "v1.3.8",
		clientOS:      "unknown",
		clientArch:    "unknown",
	},
	{
		userAgent:     "Lighthouse/v1.5.1-b0ac346/x86_64-linux",
		clientName:    "lighthouse",
		clientVersion: "v1.5.1",
		clientOS:      "linux",
		clientArch:    "x86_64",
	},
	{
		userAgent:     "Lighthouse/v3.1.2/aarch64-macos",
		clientName:    "lighthouse",
		clientVersion: "v3.1.2",
		clientOS:      "mac",
		clientArch:    "arm",
	},
	{
		userAgent:     "Lighthouse/v2.5.1-df51a73/aarch64-linux",
		clientName:    "lighthouse",
		clientVersion: "v2.5.1",
		clientOS:      "linux",
		clientArch:    "arm",
	},
	{
		userAgent:     "nimbus",
		clientName:    "nimbus",
		clientVersion: "unknown",
		clientOS:      "unknown",
		clientArch:    "unknown",
	},
	{
		userAgent:     "rust-libp2p/0.36.1",
		clientName:    "grandine",
		clientVersion: "0.36.1",
		clientOS:      "unknown",
		clientArch:    "unknown",
	},
	{
		userAgent:     "js-libp2p/0.36.2",
		clientName:    "lodestar",
		clientVersion: "0.36.2",
		clientOS:      "unknown",
		clientArch:    "unknown",
	},
	{
		userAgent:     "lodestar/v1.2.0",
		clientName:    "lodestar",
		clientVersion: "v1.2.0",
		clientOS:      "unknown",
		clientArch:    "unknown",
	},
	{
		userAgent:     "nim-libp2p/0.0.1",
		clientName:    "nimbus",
		clientVersion: "0.0.1",
		clientOS:      "unknown",
		clientArch:    "unknown",
	},
	{
		userAgent:     "erigon/lightclient",
		clientName:    "erigon",
		clientVersion: "unknown",
		clientOS:      "unknown",
		clientArch:    "unknown",
	},
	{
		userAgent:     "erigon",
		clientName:    "erigon",
		clientVersion: "unknown",
		clientOS:      "unknown",
		clientArch:    "unknown",
	},
}

var IPFSTestClients []string = []string{
	"go-ipfs/0.8.0/48f94e2",
	"hydra-booster/0.7.4",
	"storm",
	"kubo/0.15.0-dev/",
	"ioi",
	"punchr/honeypot/dev+",
}

var FilecoinTestClients []string = []string{
	"lotus-1.13.0+mainnet+git.7a55e8e8",
}

func Test_FilterClientType(t *testing.T) {
	for _, cliInf := range Eth2TestClients {
		fmt.Println(cliInf)
		client, version, os, arch := ParseClientType(EthereumNetwork, cliInf.userAgent)
		require.Equal(t, client, cliInf.clientName)
		require.Equal(t, version, cliInf.clientVersion)
		require.Equal(t, os, cliInf.clientOS)
		require.Equal(t, arch, cliInf.clientArch)
	}
}
