package ethereum

import (
	"context"
	"encoding/hex"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/methods"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

// BeaconPing
func (en *LocalEthereumNode) ServeBeaconPing(h host.Host) {
	go func() {
		sCtxFn := func() context.Context {
			reqCtx, _ := context.WithTimeout(en.ctx, RPCTimeout)
			return reqCtx
		}
		comp := new(reqresp.SnappyCompression)
		listenReq := func(ctx context.Context, peerId peer.ID, handler reqresp.ChunkedRequestHandler) {
			var ping common.Ping
			err := handler.ReadRequest(&ping)
			if err != nil {
				_ = handler.WriteErrorChunk(reqresp.InvalidReqCode, "could not parse ping request")
				log.Tracef("failed to read ping request: %v from %s", err, peerId.String())
			} else {
				if err := handler.WriteResponseChunk(reqresp.SuccessCode, &ping); err != nil {
					log.Tracef("failed to respond to ping request: %v", err)
				} else {
					log.Tracef("handled ping request", ping)
				}
			}
		}
		m := methods.PingRPCv1
		streamHandler := m.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(m.Protocol, streamHandler)
		log.Info("Started serving ping")
		// wait until the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving ping")
	}()
}

// BeaconGoodbye
func (en *LocalEthereumNode) ServeBeaconGoodbye(h host.Host) {
	go func() {
		sCtxFn := func() context.Context {
			reqCtx, _ := context.WithTimeout(en.ctx, RPCTimeout)
			return reqCtx
		}
		comp := new(reqresp.SnappyCompression)
		listenReq := func(ctx context.Context, peerId peer.ID, handler reqresp.ChunkedRequestHandler) {
			var goodbye common.Goodbye
			err := handler.ReadRequest(&goodbye)
			if err != nil {
				_ = handler.WriteErrorChunk(reqresp.InvalidReqCode, "could not parse goodbye request")
				log.Tracef("failed to read goodbye request: %v from %s", err, peerId.String())
			} else {
				if err := handler.WriteResponseChunk(reqresp.SuccessCode, &goodbye); err != nil {
					log.Tracef("failed to respond to goodbye request: %v", err)
				} else {
					log.Tracef("handled goodbye request", goodbye)
				}
			}
		}
		m := methods.GoodbyeRPCv1
		streamHandler := m.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(m.Protocol, streamHandler)
		log.Info("Started serving goodbye")
		// wait until the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving goodbye")
	}()
}

// Metadata
// ReqBeaconMetadata opens a new Stream from the given host to send a RPC reqresping the BeaconStatus of the given peer.ID.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
func (en *LocalEthereumNode) ReqBeaconMetadata(
	ctx context.Context,
	wg *sync.WaitGroup,
	h host.Host,
	peerID peer.ID,
	result *common.MetaData,
	finErr *error) {

	defer wg.Done()

	// declare the result output of the RPC call
	var metadata common.MetaData

	attnets := new(common.AttnetBits)
	bytes, err := hex.DecodeString("ffffffffffffffff") //TODO: remove hardcodding!!!!
	if err != nil {
		log.Panic("unable to decode ForkDigest", err.Error())
	}
	attnets.UnmarshalText(bytes)

	// Generate the Server Error Code
	var resCode reqresp.ResponseCode // error by default
	// record the error into the error channel
	err = methods.MetaDataRPCv2NoSnappy.RunRequest(ctx, h.NewStream, peerID, new(reqresp.SnappyCompression),
		reqresp.RequestSSZInput{Obj: nil}, 1,
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
				if err := chunk.ReadObj(&metadata); err != nil {
					return errors.Wrap(err, "from reqresping BeaconMetadata RPC")
				}
			default:
				return errors.New("unexpected result code for BeaconMetadata RPC reqresp")
			}
			return nil
		})
	*finErr = err
	*result = metadata
}

func (en *LocalEthereumNode) ServeBeaconMetadata(h host.Host) {

	go func() {
		sCtxFn := func() context.Context {
			reqCtx, _ := context.WithTimeout(en.ctx, RPCTimeout)
			return reqCtx
		}
		comp := new(reqresp.SnappyCompression)
		listenReq := func(ctx context.Context, peerId peer.ID, handler reqresp.ChunkedRequestHandler) {
			var reqMetadata common.MetaData
			err := handler.ReadRequest(&reqMetadata)
			if err != nil {
				_ = handler.WriteErrorChunk(reqresp.InvalidReqCode, "could not parse status request")
				log.Tracef("failed to read metadata request: %v from %s", err, peerId.String())
			} else {
				if err := handler.WriteResponseChunk(reqresp.SuccessCode, &en.LocalMetadata); err != nil {
					log.Tracef("failed to respond to metadata request: %v", err)
				} else {
					log.Tracef("handled metadata request")
				}
			}
		}
		m := methods.MetaDataRPCv2
		streamHandler := m.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(m.Protocol, streamHandler)
		log.Info("Started serving Beacon Metadata")
		// wait until the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving Beacon Metadata")
	}()
}

// Status
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
	err := methods.StatusRPCv1NoSnappy.RunRequest(ctx, h.NewStream, peerID, new(reqresp.SnappyCompression),
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
	// update the local status with the remote one if the local on is older
	en.UpdateStatus(remoteStatus)
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
		// wait until the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving Beacon Status")
	}()
}

// Blocks
func (en *LocalEthereumNode) ServeBeaconBlocksByRangeV2(h host.Host) {
	go func() {
		sCtxFn := func() context.Context {
			reqCtx, _ := context.WithTimeout(en.ctx, RPCTimeout)
			return reqCtx
		}
		comp := new(reqresp.SnappyCompression)
		listenReq := func(ctx context.Context, peerId peer.ID, handler reqresp.ChunkedRequestHandler) {
			blockRange := new(methods.BlocksByRootReq)
			err := handler.ReadRequest(blockRange)
			if err != nil {
				_ = handler.WriteErrorChunk(reqresp.InvalidReqCode, "could not parse block_by_range request")
				log.Errorf("failed to read block_by_range request: %v from %s", err, peerId.String())
			} else {
				log.Infof("dropped block_by_range request %v", *blockRange)
			}
		}
		m := methods.BlocksByRangeRPCv2
		streamHandler := m.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(m.Protocol, streamHandler)
		log.Info("Started serving block_by_range")
		// wait until the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving block_by_range")
	}()
}

func (en *LocalEthereumNode) ServeBeaconBlocksByRootV2(h host.Host) {
	go func() {
		sCtxFn := func() context.Context {
			reqCtx, _ := context.WithTimeout(en.ctx, RPCTimeout)
			return reqCtx
		}
		comp := new(reqresp.SnappyCompression)
		listenReq := func(ctx context.Context, peerId peer.ID, handler reqresp.ChunkedRequestHandler) {
			blockRoot := new(methods.BlocksByRootReq)
			err := handler.ReadRequest(blockRoot)
			if err != nil {
				_ = handler.WriteErrorChunk(reqresp.InvalidReqCode, "could not parse block_by_root request")
				log.Error("failed to read block_by_root request: %v from %s", err, peerId.String())
			} else {
				log.Infof("dropped block_by_root request %v", *blockRoot)
			}
		}
		m := methods.BlocksByRootRPCv2
		streamHandler := m.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(m.Protocol, streamHandler)
		log.Info("Started serving block_by_root")
		// wait until the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving block_by_root")
	}()
}

// Blobs
func (en *LocalEthereumNode) ServeBeaconBlobsByRangeV1(h host.Host) {
	go func() {
		sCtxFn := func() context.Context {
			reqCtx, _ := context.WithTimeout(en.ctx, RPCTimeout)
			return reqCtx
		}
		comp := new(reqresp.SnappyCompression)
		listenReq := func(ctx context.Context, peerId peer.ID, handler reqresp.ChunkedRequestHandler) {
			blobsRange := new(methods.BlobsByRangeReqV1)
			err := handler.ReadRequest(blobsRange)
			if err != nil {
				_ = handler.WriteErrorChunk(reqresp.InvalidReqCode, "could not parse blobs_by_range request")
				log.Errorf("failed to read blobs_by_range request: %v from %s", err, peerId.String())
			} else {
				log.Info("dropped blobs_by_range request", *blobsRange)
			}
		}
		b := methods.BlobsByRangeRPCv1
		streamHandler := b.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(b.Protocol, streamHandler)
		log.Info("Started serving blobs_by_range")
		// wait until the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving blobs_by_range")
	}()
}

func (en *LocalEthereumNode) ServeBeaconBlobsByRootV1(h host.Host) {
	go func() {
		sCtxFn := func() context.Context {
			reqCtx, _ := context.WithTimeout(en.ctx, RPCTimeout)
			return reqCtx
		}
		comp := new(reqresp.SnappyCompression)
		listenReq := func(ctx context.Context, peerId peer.ID, handler reqresp.ChunkedRequestHandler) {
			blobRoots := new(methods.BlobByRootV1)
			err := handler.ReadRequest(blobRoots)
			if err != nil {
				_ = handler.WriteErrorChunk(reqresp.InvalidReqCode, "could not parse blobs_by_root request")
				log.Errorf("failed to read blobs_by_root request: %v from %s", err, peerId.String())
			} else {
				log.Info("dropped blobs_by_root request", *blobRoots)
			}
		}
		b := methods.BlobsByRootRPCv1
		streamHandler := b.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(b.Protocol, streamHandler)
		log.Info("Started serving blobs_by_root")
		// wait until the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving blobs_by_root")
	}()
}
