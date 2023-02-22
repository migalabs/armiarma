package ethereum

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/methods"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	log "github.com/sirupsen/logrus"
)

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
		m.Protocol = m.Protocol + "_snappy" // TODO: add snappy support for RPC calls
		streamHandler := m.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(m.Protocol, streamHandler)
		log.Info("Started serving ping")
		// wait untill the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving ping")
	}()
}

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
		m.Protocol = m.Protocol + "_snappy" // TODO: add snappy support for RPC calls
		streamHandler := m.MakeStreamHandler(sCtxFn, comp, listenReq)
		h.SetStreamHandler(m.Protocol, streamHandler)
		log.Info("Started serving goodbye")
		// wait untill the ctx is down
		<-en.ctx.Done() // TODO: do it better
		log.Info("Stopped serving goodbye")
	}()
}
