package events

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/migalabs/armiarma/pkg/networks/ethereum"
	"github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
)

type Topic string

type Forwarder struct {
	ctx  context.Context
	ip   string
	port int

	server        *sse.Server
	db            *postgresql.DBClient
	ethMsgHandler *ethereum.EthMessageHandler

	// Store downstream attestation events in a channel so
	// that we don't block the eth2 handler.
	attestationCh chan *ethereum.AttestationReceievedEvent

	once sync.Once
}

func NewForwarder(ip string, port int, db *postgresql.DBClient, ethMsgHandler *ethereum.EthMessageHandler) *Forwarder {
	server := sse.New()

	// Disable auto replay. If a consumer is not connected, it will never receive the event.
	server.AutoReplay = false

	return &Forwarder{
		ip:            ip,
		port:          port,
		server:        server,
		db:            db,
		ethMsgHandler: ethMsgHandler,
		attestationCh: make(chan *ethereum.AttestationReceievedEvent, 10000),
	}
}

func (f *Forwarder) Start(ctx context.Context) error {
	f.ctx = ctx

	// Only start if we have a valid IP and port
	if f.ip == "" || f.port == 0 {
		log.WithField("address", f.ip).WithField("port", f.port).Debug("Not starting SSE server as no IP or port provided")

		return nil
	}

	var err error

	f.once.Do(func() {
		f.subscribeDownstream(ctx)

		err = f.startHTTPServer()
		if err != nil {
			return
		}
	})

	return err
}

func (f *Forwarder) Stop() {
	f.server.Close()
}

func (f *Forwarder) startHTTPServer() error {
	// Create a new Mux and set the handler
	sseMux := http.NewServeMux()
	sseMux.HandleFunc("/events", f.server.ServeHTTP)

	log.WithField("address", f.ip).WithField("port", f.port).Info("Starting SSE server")

	errCh := make(chan error, 1)

	// Start the HTTP server
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", f.ip, f.port), sseMux)
		if err != nil {
			errCh <- err
		}
	}()

	return <-errCh
}

func (f *Forwarder) subscribeDownstream(ctx context.Context) {
	f.ethMsgHandler.OnAttestation(func(event *ethereum.AttestationReceievedEvent) {
		f.attestationCh <- event
	})
}

func (f *Forwarder) startWorkers() {
	// start the event workers that process events out of the channels
	for i := 0; i < 10; i++ {
		go f.eventWorker()
	}
}

func (f *Forwarder) eventWorker() {
	for {
		select {
		case event := <-f.attestationCh:
			f.processAttestationEvent(event)
		case <-f.ctx.Done():
			return
		}
	}
}

func (f *Forwarder) processAttestationEvent(e *ethereum.AttestationReceievedEvent) {
	// Publish the raw attestation straight away
	if err := f.publishEthereumAttestation(&EthereumAttestation{
		Attestation: e.Attestation,
	}); err != nil {
		log.WithError(err).Error("error publishing raw attestation to SSE server")
	}

	// Build the timed attestation event
	info, err := f.db.GetFullHostInfo(e.PeerID)
	if err != nil {
		log.WithError(err).Error("error getting host info from DB when handling a new attestation")

		return
	}

	// Publish the timed event
	if err := f.publishTimedEthereumAttestation(&TimedEthereumAttestation{
		Attestation: e.Attestation,
		AttestationExtraData: &AttestationExtraData{
			ArrivedAt:  e.TrackedAttestation.ArrivalTime,
			P2PMsgID:   e.TrackedAttestation.MsgID,
			Subnet:     e.TrackedAttestation.Subnet,
			TimeInSlot: e.TrackedAttestation.TimeInSlot,
		},
		PeerInfo: &PeerInfo{
			ID:              string(info.ID),
			IP:              info.IP,
			Port:            info.Port,
			UserAgent:       info.PeerInfo.UserAgent,
			Latency:         info.PeerInfo.Latency,
			Protocols:       info.PeerInfo.Protocols,
			ProtocolVersion: info.PeerInfo.ProtocolVersion,
		},
	}); err != nil {
		log.WithError(err).Error("error publishing timed attestation to SSE server")
	}

}
