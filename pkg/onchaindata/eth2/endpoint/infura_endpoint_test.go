package endpoint

import (
	"context"

	"testing"

	"github.com/stretchr/testify/require"

	"github.com/migalabs/armiarma/pkg/onchaindata/eth2/endpoint/types"
)

var infuraTestEndpoint = "https://20PdJoS82pnejJJ9joDMnbjsQ32:0c9b868d8621332ea91c7fc24c5fc34f@eth2-beacon-mainnet.infura.io"

func TestInfuraNewHttpsRequest(t *testing.T) {
	mainCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// generate infura cli
	infuraCli, err := NewInfuraClient(infuraTestEndpoint)
	require.Equal(t, err, nil)
	// make genesis request function
	genesis := types.Genesis{}
	err = infuraCli.NewHttpsRequest(mainCtx, GENESIS_ENPOINT, &genesis)
	require.Equal(t, err, nil)
}

func TestGenesisReq(t *testing.T) {
	mainCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// generate infura cli
	infuraCli, err := NewInfuraClient(infuraTestEndpoint)
	require.Equal(t, err, nil)
	// make genesis request function
	_, err = infuraCli.ReqGenesis(mainCtx)
	require.Equal(t, err, nil)

}

func TestStateForkReq(t *testing.T) {
	mainCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// generate infura cli
	infuraCli, err := NewInfuraClient(infuraTestEndpoint)
	require.Equal(t, err, nil)
	// make genesis request function
	_, err = infuraCli.ReqStateFork(mainCtx, "head")
	require.Equal(t, err, nil)

}
