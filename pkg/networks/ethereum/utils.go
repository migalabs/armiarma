package ethereum

import (
	"context"

	rendp "github.com/migalabs/armiarma/pkg/networks/ethereum/remoteendpoint"
	"github.com/pkg/errors"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func GetForkDigetsOfEth2Head(ctx context.Context, infCli *rendp.InfuraClient) (common.ForkDigest, error) {
	if !infCli.IsInitialized() {
		return common.ForkDigest{}, errors.New("infura client is not initialized")
	}
	// request the genesis
	genesis, err := infCli.ReqGenesis(ctx)
	if err != nil {
		return common.ForkDigest{}, errors.Wrap(err, "unable to request genesis to compose forkdigest")
	}
	// request state fork of the Eth2 HEAD
	statefork, err := infCli.ReqStateFork(ctx, "head")
	if err != nil {
		return common.ForkDigest{}, errors.Wrap(err, "unable to request genesis to compose forkdigest")
	}
	forkdigest := common.ComputeForkDigest(statefork.CurrentVersion, genesis.GenesisValidatorsRoot)
	return forkdigest, nil
}
