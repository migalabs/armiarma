package rpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"

	com "github.com/migalabs/armiarma/pkg/networks/common"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/methods"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

// ReqBeaconStatus opens a new Stream from the given host to send a RPC reqresping the BeaconStatus of the given peer.ID.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
func ReqBeaconStatus(ctx context.Context, wg *sync.WaitGroup, h host.Host, peerID peer.ID, result *com.RPCResult, finErr *error) {
	defer wg.Done()

	// Cast RPCResult into common.Status
	result = result.(*common.Status)

	frkDgst := new(common.ForkDigest)
	b, err := hex.DecodeString("4a26c58b") //TODO: remove hardcodding!!!!
	if err != nil {
		log.Panic("unable to decode ForkDigest", err.Error())
	}
	frkDgst.UnmarshalText(b)

	status := &common.Status{
		ForkDigest:     *frkDgst,
		FinalizedRoot:  common.Root{},
		FinalizedEpoch: 0,
		HeadRoot:       common.Root{},
		HeadSlot:       0,
	}
	var resCode reqresp.ResponseCode // error by default
	err = methods.StatusRPCv1.RunRequest(ctx, h.NewStream, peerID, new(reqresp.SnappyCompression),
		reqresp.RequestSSZInput{Obj: status}, 1,
		func() error {
			return nil
		},
		func(chunk reqresp.ChunkedResponseHandler) error {
			resCode = chunk.ResultCode()
			switch resCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return fmt.Errorf("%s: %w", msg, err)
				}
			case reqresp.SuccessCode:
				var stat common.Status
				if err := chunk.ReadObj(&stat); err != nil {
					return err
				}
				fmt.Println(stat)
				*resultStatus = stat
			default:
				return errors.New("unexpected result code")
			}
			return nil
		})
	*finErr = err
}
