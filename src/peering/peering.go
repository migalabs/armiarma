package peering

/**
This file implements the peering service interface
With this interface, the given host will be able to retreive and connect a set of peers from the peerstore under the chosen strategy.

*/

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/hosts"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

var (
	ModuleName = "PEERING"
	log        = logrus.WithField(
		"module", ModuleName,
	)

	ConnectionRefuseTimeout = 10 * time.Second
	MaxRetries              = 1
	DefaultWorkers          = 10
)

type PeeringOption func(*PeeringService) error

// PeeringService is the main service that will connect peers from the given peerstore and using the given Host.
// It will use the specified peering strategy, which might difer/change from the testing or desired purposes of the run.
type PeeringService struct {
	ctx context.Context

	host      *hosts.BasicLibp2pHost
	PeerStore *db.PeerStore
	strategy  PeeringStrategy
	// Control Flags
	Timeout    time.Duration
	MaxRetries int
}

// Constructor
func NewPeeringService(
	ctx context.Context,
	h *hosts.BasicLibp2pHost,
	peerstore *db.PeerStore,
	opts ...PeeringOption) (PeeringService, error) {

	pServ := PeeringService{
		ctx:        ctx,
		host:       h,
		PeerStore:  peerstore,
		Timeout:    ConnectionRefuseTimeout,
		MaxRetries: MaxRetries, // TODO: Hardcoded to 1 retry, future retries are directly dismissed/dropped by dialing peer
	}
	// iterate through the Options given as args
	for _, opt := range opts {
		err := opt(&pServ)
		if err != nil {
			return pServ, err
		}
	}
	return pServ, nil
}

func WithPeeringStrategy(strategy PeeringStrategy) PeeringOption {
	return func(p *PeeringService) error {
		if strategy == nil {
			return fmt.Errorf("given peering strategy is empty")
		}
		log.Debugf("configuring peering with %s", strategy.Type())
		p.strategy = strategy
		return nil
	}
}

// Run:
// Main peering event selector.
// For every next peer received from the strategy, attempt the connection and record the status of this one.
// Notify the strategy of any conn/disconn recorded.
func (c *PeeringService) Run() {
	log.Debug("starting the peering service")

	// start the peering strategy
	peerStreamChan := c.strategy.Run()

	// set up the routines that will peer and record connections
	for worker := 1; worker <= DefaultWorkers; worker++ {
		workerName := fmt.Sprintf("Peering Worker %d", worker)
		go c.peeringWorker(workerName, peerStreamChan)
	}
	go c.eventRecorderRoutine()
	go c.ServeMetrics()
}

// peeringWorker
// Peering routine that will be launched several times (several workers).
// @param workerID: id of the worker.
// @param peerStreamChan: used to receive the next peer.
func (c *PeeringService) peeringWorker(workerID string, peerStreamChan chan models.Peer) {
	log.Infof("launching %s", workerID)
	peeringCtx := c.ctx
	h := c.host.Host()

	// Request new peer from the peering strategy
	c.strategy.NextPeer()

	// set up the loop that will receive peers to connect
	for {
		select {
		// Next peer arrives
		case nextPeer := <-peerStreamChan:
			log.Debugf("%s -> new peer %s to connect", workerID, nextPeer.PeerId)
			peerID, err := peer.Decode(nextPeer.PeerId)
			if err != nil {
				log.Warnf("%s -> coulnd't extract peer.ID from peer %s", workerID, nextPeer.PeerId)
				// Request the next peer when case is over
				c.strategy.NextPeer()
				continue
			}
			// Check if the peer is already connected by the host
			peerList := h.Network().Peers()
			connected := false
			for _, p := range peerList {
				if p.String() == peerID.String() {
					connected = true
					break
				}
			}
			if connected {
				log.Infof("%s -> Peer %s was already connected", workerID, nextPeer.PeerId)
				c.strategy.NextPeer()
				continue
			}

			// Set the correct address format to connect the peers
			// libp2p complains if we put multi-addresses that include the peer ID into the Addrs list.
			addrs := nextPeer.ExtractPublicAddr()
			transport, _ := peer.SplitAddr(addrs)
			if transport == nil {
				// Request the next peer when case is over
				c.strategy.NextPeer()
				continue
			}
			addrInfo := peer.AddrInfo{
				ID:    peerID,
				Addrs: make([]ma.Multiaddr, 0, 1),
			}
			addrInfo.Addrs = append(addrInfo.Addrs, transport)
			nPeer := models.NewPeer(nextPeer.PeerId)
			connAttStat := ConnectionAttemptStatus{
				Peer: nPeer,
			}
			log.Debugf("%s addrs %s attempting connection to peer", workerID, addrInfo.Addrs)
			// try to connect the peer
			attempts := 0
			timeoutctx, cancel := context.WithTimeout(peeringCtx, c.Timeout)
			for attempts < c.MaxRetries {
				if err := h.Connect(timeoutctx, addrInfo); err != nil {
					log.WithError(err).Debugf("%s attempts %d failed connection attempt", workerID, attempts+1)
					// the connetion failed
					// fill the ConnectionStatus for the given peer connection
					connAttStat.Timestamp = time.Now()
					connAttStat.Successful = false
					connAttStat.RecError = err
					// increment the attempts
					attempts++
					continue
				} else { // connection successfuly made
					log.Debugf("%s peer_id %s successful connection made", workerID, peerID.String())
					// fill the ConnectionStatus for the given peer connection
					connAttStat.Timestamp = time.Now()
					connAttStat.Successful = true
					connAttStat.RecError = nil
					break
				}
			}
			cancel()
			// send it to the strategy
			c.strategy.NewConnectionAttempt(connAttStat)
			// Request the next peer when case is over
			c.strategy.NextPeer()

		// Stoping go routine
		case <-peeringCtx.Done():
			log.Debugf("closing %s", workerID)
			return
		}
	}

}

// eventRecorderRoutine:
// The event selector records the status of any incoming connection and disconnection and
// notifies the strategy of any recorded conn/disconn.
func (c *PeeringService) eventRecorderRoutine() {
	log.Debug("starting the event recorder service")
	// get the connection and disconnection notification channels from the host
	newConnEventChan := c.host.ConnEventNotChannel()
	newIdentPeerChan := c.host.IdentEventNotChannel()
	for {
		select {

		// New connection event
		case newConn := <-newConnEventChan:
			switch newConn.ConnType {
			case int8(1):
				log.Debugf("new conection from %s", newConn.Peer.PeerId)
			case int8(2):
				log.Debugf("new disconnection from %s", newConn.Peer.PeerId)
			default:
				log.Debugf("unrecognized event from peer %s", newConn.Peer.PeerId)
			}
			c.strategy.NewConnectionEvent(newConn)

		// New identification event has been recorded
		case newIdent := <-newIdentPeerChan:
			log.Debugf("new identification %s from peer %s", strconv.FormatBool(newIdent.Peer.IsConnected), newIdent.Peer.PeerId)
			c.strategy.NewIdentificationEvent(newIdent)

			// Stoping go routine
		case <-c.ctx.Done():
			log.Debug("closing peering go routine")
			return
		}
	}
}
