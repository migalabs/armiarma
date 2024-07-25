package peering

/**
This file implements the peering service interface
With this interface, the given host will be able to retreive and connect a set of peers from the peerstore under the chosen strategy.

*/

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
//	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	psql "github.com/hdser/armiarma/pkg/db/redshift"
	"github.com/migalabs/armiarma/pkg/hosts"
	log "github.com/sirupsen/logrus"
)

var (
	ConnectionRefuseTimeout = 20 * time.Second
	MaxRetries              = 1
	DefaultWorkers          = 500
)

type PeeringOption func(*PeeringService) error

// PeeringService is the main service that will connect peers from the given peerstore and using the given Host.
// It will use the specified peering strategy, which might difer/change from the testing or desired purposes of the run.
type PeeringService struct {
	ctx context.Context

	host     *hosts.BasicLibp2pHost
	DBClient *psql.DBClient
	strategy PeeringStrategy
	// Control Flags
	Timeout    time.Duration
	MaxRetries int

	// metrics
	m                 sync.RWMutex
	errorDistribution map[string]int
}

// Constructor
func NewPeeringService(
	ctx context.Context,
	h *hosts.BasicLibp2pHost,
	dbClient *psql.DBClient,
	opts ...PeeringOption) (PeeringService, error) {

	pServ := PeeringService{
		ctx:               ctx,
		host:              h,
		DBClient:          dbClient,
		Timeout:           ConnectionRefuseTimeout,
		MaxRetries:        MaxRetries,
		errorDistribution: make(map[string]int, 0),
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
		log.Info("configuring crawler with peering strategy: %s", strategy.Type())
		p.strategy = strategy
		return nil
	}
}

// Run:
// Main peering event selector.
// For every next peer received from the strategy, attempt the connection and record the status of this one.
// Notify the strategy of any conn/disconn recorded.
func (c *PeeringService) Run() {
	log.Info("starting the peering service")

	// start the peering strategy
	peerStreamChan := c.strategy.Run()

	// set up the routines that will peer and record connections
	for worker := 1; worker <= DefaultWorkers; worker++ {
		workerName := fmt.Sprintf("Peering Worker %d", worker)
		go c.peeringWorker(workerName, peerStreamChan)
	}
	go c.eventRecorderRoutine()
}

// peeringWorker
// Peering routine that will be launched several times (several workers).
// @param workerID: id of the worker.
// @param peerStreamChan: used to receive the next peer.
func (c *PeeringService) peeringWorker(workerID string, peerStreamChan chan *models.HostInfo) {
	logEntry := log.WithFields(log.Fields{
		"peering-worker": workerID,
	})
	logEntry.Debug("launching worker")

	h := c.host.Host()

	// Request new peer from the peering strategy
	c.strategy.NextPeer()

	// set up the loop that will receive peers to connect
	for {
		select {
		// Next peer arrives
		case nextPeer := <-peerStreamChan:
			logEntry.Tracef("%s -> new peer %+v to connect", workerID, nextPeer)

			// Check if the peer is already connected by the host
			peerList := h.Network().Peers()
			connected := false
			for _, p := range peerList {
				if p.String() == nextPeer.ID.String() {
					connected = true
					break
				}
			}
			if connected {
				logEntry.Tracef("%s -> Peer %s was already connected", workerID, nextPeer.ID.String())
				c.strategy.NextPeer()
				continue
			}

			addrInfo := nextPeer.ComposeAddrsInfo()

			// control info for the attempt
			var attStatus models.AttemptStatus = models.NegativeAttempt
			var attError string = ""
			var deprecable bool = false
			var leftNet bool = false

			// try to connect the peer
			logEntry.Debugf("%s addrs %s attempting connection to peer", workerID, addrInfo.Addrs)
			attempts := 0
			timeoutctx, cancel := context.WithTimeout(c.ctx, c.Timeout)
			for attempts < c.MaxRetries {
				if err := h.Connect(timeoutctx, addrInfo); err != nil { // there was an error
					logEntry.WithError(err).Debugf("%s attempts %d failed connection attempt to %+v",
						workerID, attempts+1, addrInfo)
					attError = hosts.ParseConError(err)
					attempts++
					continue
				} else { // connection successfuly made
					logEntry.Debugf("successful connection to %s", nextPeer.ID.String())
					attStatus = models.PossitiveAttempt
					attError = hosts.NoConnError
					break
				}
			}
			cancel()

			// generate the connectionAttempt
			connAttempt := models.NewConnAttempt(
				nextPeer.ID,
				attStatus,
				attError,
				deprecable,
				leftNet,
			)

			// send it to the strategy
			c.strategy.NewConnectionAttempt(connAttempt)
			// Request the next peer when case is over
			c.strategy.NextPeer()

		// Stoping go routine
		case <-c.ctx.Done():
			logEntry.Infof("closing")
			return
		}
	}

}

// eventRecorderRoutine:
// The event selector records the status of any incoming connection and disconnection and
// notifies the strategy of any recorded conn/disconn.
func (c *PeeringService) eventRecorderRoutine() {
	log.Trace("starting the event recorder service")
	// get the connection and disconnection notification channels from the host
	newConnEventChan := c.host.ConnEventNotChannel()
	newIdentPeerChan := c.host.IdentEventNotChannel()
	for {
		select {

		// New connection event
		case newConn := <-newConnEventChan:
			c.strategy.NewConnectionEvent(newConn)

		// New identification event has been recorded
		case newIdent := <-newIdentPeerChan:
			c.strategy.NewIdentificationEvent(newIdent)

		// Stoping go routine
		case <-c.ctx.Done():
			log.Debug("closing peering go routine")
			return
		}
	}
}
