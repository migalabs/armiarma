package ethereum

import (
	"context"
	"sync"

	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/methods"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

// ReqBeaconStatus opens a new Stream from the given host to send a RPC reqresping the BeaconStatus of the given peer.ID.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
func (en *LocalEthereumNode) ReqBeaconStatus(
	ctx context.Context,
	wg *sync.WaitGroup,
	h host.Host,
	peerID peer.ID,
	result *common.Status,
	finErr *error) {

	defer wg.Done()
	// declare the result obj of the RPC call
	var remoteStatus common.Status

	var resCode reqresp.ResponseCode // error by default
	err := methods.StatusRPCv1.RunRequest(ctx, h.NewStream, peerID, new(reqresp.SnappyCompression),
		reqresp.RequestSSZInput{Obj: &en.LocalStatus}, 1,
		func() error {
			return nil
		},
		func(chunk reqresp.ChunkedResponseHandler) error {
			resCode = chunk.ResultCode()
			switch resCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return errors.Wrap(err, msg)
				}
			case reqresp.SuccessCode:
				if err := chunk.ReadObj(&remoteStatus); err != nil {
					return err
				}
			default:
				return errors.New("unexpected result code")
			}
			return nil
		})
	*finErr = err
	*result = remoteStatus
}

func (en *LocalEthereumNode) ServeBeaconStatus(h host.Host) {

	go func() {
		sCtxFn := func() context.Context {
			reqCtx, _ := context.WithTimeout(en.ctx, RPCTimeout)
			return reqCtx
		}
		comp := new(reqresp.SnappyCompression)
		listenReq := func(ctx context.Context, peerId peer.ID, handler reqresp.ChunkedRequestHandler) {
			var reqStatus common.Status
			err := handler.ReadRequest(&reqStatus)
			if err != nil {
				_ = handler.WriteErrorChunk(reqresp.InvalidReqCode, "could not parse status request")
				log.Tracef("failed to read status request: %v from %s", err, peerId.String())
			} else {
				if err := handler.WriteResponseChunk(reqresp.SuccessCode, &en.LocalStatus); err != nil {
					log.Tracef("failed to respond to status request: %v", err)
				} else {
					// update if possible out status
					en.UpdateStatus(reqStatus)
					log.Tracef("handled status request")
				}
			}
		}
		m := methods.StatusRPCv1
		streamHandler := m.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(m.Protocol, streamHandler)
		log.Info("Started serving Beacon Status")
		// wait untill the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving Beacon Status")
	}()
}
