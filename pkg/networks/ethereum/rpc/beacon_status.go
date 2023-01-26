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
var ReqBeaconStatus com.RPCRequest = func(ctx context.Context, wg *sync.WaitGroup, h host.Host, peerID peer.ID, result com.RPCResult, finErr *error) {
	defer wg.Done()

	// declare the result obj of the RPC call
	var status common.Status

	frkDgst := new(common.ForkDigest)
	b, err := hex.DecodeString("4a26c58b") //TODO: remove hardcodding!!!!
	if err != nil {
		log.Panic("unable to decode ForkDigest", err.Error())
	}
	frkDgst.UnmarshalText(b)

	ourStatus := &common.Status{
		ForkDigest:     *frkDgst,
		FinalizedRoot:  common.Root{},
		FinalizedEpoch: 0,
		HeadRoot:       common.Root{},
		HeadSlot:       0,
	}
	var resCode reqresp.ResponseCode // error by default
	err = methods.StatusRPCv1.RunRequest(ctx, h.NewStream, peerID, new(reqresp.SnappyCompression),
		reqresp.RequestSSZInput{Obj: ourStatus}, 1,
		func() error {
			return nil
		},
		func(chunk reqresp.ChunkedResponseHandler) error {
			resCode = chunk.ResultCode()
			switch resCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return errors.New(fmt.Sprintf("%s: %s", msg, err))
				}
			case reqresp.SuccessCode:
				if err := chunk.ReadObj(&status); err != nil {
					return err
				}
			default:
				return errors.New("unexpected result code")
			}
			return nil
		})
	*finErr = err
	result = status
}
