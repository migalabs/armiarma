package peering

/**
This file implements the peering service interface
With this interface, the given host will be able to retreive and connect a set of peers from the peerstore under the chosen strategy.

*/

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	ConnectionRefuseTimeout = 3 * time.Second
)

type PeeringOption func(*PeeringService) error

type PeeringOpts struct {
	InfoObj *info.InfoData
	LogOpts base.LogOpts
}

// PeeringService is the main service that will connect peers from the given peerstore and using the given Host.
// It will use the specified peering strategy, which might difer/change from the testing or desired purposes of the run.
type PeeringService struct {
	*base.Base
	InfoObj   *info.InfoData
	host      *hosts.BasicLibp2pHost
	PeerStore *db.PeerStore
	strategy  PeeringStrategy
	//
	Timeout    time.Duration
	MaxRetries int
}

func NewPeeringService(ctx context.Context, h *hosts.BasicLibp2pHost, peerstore *db.PeerStore,
	peeringOpts *PeeringOpts, opts ...PeeringOption) (PeeringService, error) {
	// TODO: cancel is still not implemented in the BaseCreation
	peeringCtx, _ := context.WithCancel(ctx)
	logOpts := peeringOpts.LogOpts
	logOpts.ModName = "peering service"
	b, err := base.NewBase(
		base.WithContext(peeringCtx),
		base.WithLogger(logOpts),
	)
	if err != nil {
		return PeeringService{}, err
	}
	pServ := PeeringService{
		Base:       b,
		InfoObj:    peeringOpts.InfoObj,
		host:       h,
		PeerStore:  peerstore,
		Timeout:    15 * time.Second, // TODO: Hardcoded to 15 seconds
		MaxRetries: 1,                // TODO: Hardcoded to 1 retry, future retries are directly dismissed/dropped by dialing peer
	}
	// iterate through the Options given as args
	for _, opt := range opts {
		err := opt(&pServ)
		if err != nil {
			return pServ, err
		}
	}
	/* -- Check Performance of the previous PeeringServ + Pruning
	// check if there is any peering strategy assigned
	if pServ.strategy == nil {
		// Choose the Default strategy (Pruning)
		prOpts := PruningOpts{
			AggregatedDelay: 24 * time.Hour,
			logOpts:         peeringOpts.LogOpts,
		}
		pruning := NewPruningStrategy(pServ.Base.Ctx(), peerstore, prOpts)
		pServ.strategy = pruning
	}
	*/
	return pServ, nil
}

func WithPeeringStrategy(strategy PeeringStrategy) PeeringOption {
	return func(p *PeeringService) error {
		if strategy == nil {
			return fmt.Errorf("given peering strategy is empty")
		}
		p.Log.Debugf("configuring peering with %s", strategy.Type())
		p.strategy = strategy
		return nil
	}
}

// Run
// * Main peering event selector,
// * For every next peer received from the strategy, attempt the connection and record the status of this one
// * Notify the strategy of any conn/disconn recorded
func (c *PeeringService) Run() {
	c.Log.Debug("starting the peering service")
	peeringCtx := c.Ctx()
	h := c.host.Host()
	// get the connection and disconnection notification channels from the host
	newConnChan := c.host.ConnNotChan()
	newDisconnChan := c.host.DisconnNotChan()

	// start the peering strategy
	peerStreamChan := c.strategy.Run()
	// Request new peer from the peering strategy
	c.strategy.NextPeer()

	// set up the loop where every given time we will stop it to refresh the peerstore
	go func() {
		for {
			select {
			// Next peer arrives
			case nextPeer := <-peerStreamChan:
				c.Log.Debugf("new peer %s to connect", nextPeer.PeerId)
				peerID, err := peer.Decode(nextPeer.PeerId)
				if err != nil {
					c.Log.Warnf("coulnd't extract peer.ID from peer %s", nextPeer.PeerId)
					// Request the next peer when case is over
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
				nPeer := db.NewPeer(nextPeer.PeerId)
				connAttStat := ConnectionAttemptStatus{
					Peer: nPeer,
				}
				c.Log.Debugf("addrs %s attempting connection to peer", addrInfo.Addrs)
				// try to connect the peer
				attempts := 1
				timeoutctx, cancel := context.WithTimeout(peeringCtx, c.Timeout)
				defer cancel()
				for attempts <= c.MaxRetries {
					if err := h.Connect(timeoutctx, addrInfo); err != nil {
						c.Log.WithError(err).Debugf("attempts %d failed connection attempt", attempts)
						// the connetion failed
						/*
							// TODO: think about edgy case for when the connection gets refused by peer but connection handler notifies
							timeoutctx, cancel := context.WithTimeout(peeringCtx, ConnectionRefuseTimeout)
							select {
							case newConn := <-newConnChan:
								// edgy case, connection refused
								cancel()
							case <- timeoutctx.Done():
								// There was no not from the peer
							}
						*/
						// fill the ConnectionStatus for the given peer connection
						connAttStat.Timestamp = time.Now()
						connAttStat.Successful = false
						connAttStat.RecError = err
						// increment the attempts
						attempts++
						continue
					} else { // connection successfuly made
						c.Log.Debugf("peer_id %s successful connection made", peerID.String())
						// fill the ConnectionStatus for the given peer connection
						connAttStat.Timestamp = time.Now()
						connAttStat.Successful = true
						connAttStat.RecError = nil
						// break the loop
						break
					}
				}
				// send it to the strategy
				c.strategy.NewConnectionAttempt(connAttStat)
				// Request the next peer when case is over
				c.strategy.NextPeer()

			// New connection
			case newConn := <-newConnChan:
				c.Log.Debugf("new conection from %s", newConn.Peer.PeerId)
				c.strategy.NewConnection(newConn)

			// New disconnection
			case newDisconn := <-newDisconnChan:
				c.Log.Debugf("new disconection from %s", newDisconn.Peer.PeerId)
				c.strategy.NewDisconnection(newDisconn)

			// Stoping go routine
			case <-peeringCtx.Done():
				c.Log.Debug("closing peering go routine")
			}
		}
	}()
}

func (c *PeeringService) peeringRoutine() {

}

func (c *PeeringService) Close() {
	c.Log.Info("stoping the peering service")
	// Stop the strategy
	c.strategy.Close()
	// finish the module context
	c.Cancel()
}
