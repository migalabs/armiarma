package ethereum

import (
	"context"

	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/methods"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

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
		// wait untill the ctx is down
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
		// wait untill the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving blobs_by_root")
	}()
}
