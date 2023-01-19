package rpc

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	com "github.com/migalabs/armiarma/pkg/networks/common"
	"github.com/migalabs/armiarma/pkg/networks/rpc/methods"
	"github.com/migalabs/armiarma/pkg/networks/rpc/reqresp"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

// ReqBeaconMetadata opens a new Stream from the given host to send a RPC reqresping the BeaconStatus of the given peer.ID.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
func ReqBeaconMetadata(ctx context.Context, wg *sync.WaitGroup, h host.Host, peerID peer.ID, result *com.RPCResult, finErr *error) {
	defer wg.Done()

	// Cast RPCResults to BeaconMetadata
	result = result(*common.MetaData)

	// Generate the Server Error Code
	var resCode reqresp.ResponseCode // error by default
	// record the error into the error channel
	err := methods.MetaDataRPCv1.RunRequest(ctx, h.NewStream, peerID, reqresp.SnappyCompression{}, reqresp.RequestSSZInput{Obj: nil}, 1,
		func() error {
			return nil
		},
		func(chunk reqresp.ChunkedResponseHandler) error {
			resCode = chunk.ResultCode()
			switch resCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return errors.Errorf("error reqresping BeaconMetadata RPC: %s", msg)
				}
			case reqresp.SuccessCode:
				if err := chunk.ReadObj(result); err != nil {
					return errors.Wrap(err, "from reqresping BeaconMetadata RPC")
				}
			default:
				return errors.New("unexpected result code for BeaconMetadata RPC reqresp")
			}
			return nil
		})
	*finErr = err
}
