package ethereum

import (
	"context"

	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/methods"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
)

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
		// wait untill the ctx is down
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
		// wait untill the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving block_by_root")
	}()
}
