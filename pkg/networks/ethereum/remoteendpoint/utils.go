package endpoint

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func ReplaceEndpointWithRequest(endpoint string, toBeReplaced string, item string) string {
	replaceable := fmt.Sprintf("{%s}", toBeReplaced)
	return strings.Replace(endpoint, replaceable, item, -1)
}

func GetForkDigetsOfEth2Head(ctx context.Context, infCli *InfuraClient) (common.ForkDigest, error) {
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
