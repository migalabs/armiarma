package endpoint

import (
	"context"

	"github.com/migalabs/armiarma/src/onchaindata/eth2/endpoint/types"
	"github.com/pkg/errors"
)

// receives: state as string
// State identifier. Can be one of: "head" (canonical head in node's view), "genesis", "finalized", "justified", <slot>, <hex encoded stateRoot with 0x prefix>.
func (c *InfuraClient) ReqStateFork(ctx context.Context, state string) (statefork types.StateFork, err error) {
	if !c.IsInitialized() {
		return statefork, errors.New("infura client is not initialized")
	}
	// Compose the request for the StateFork request
	req := ReplaceEndpointWithRequest(BEACON_STATE_FORK, "state", state)
	err = c.NewHttpsRequest(ctx, req, &statefork)
	return statefork, err
}
