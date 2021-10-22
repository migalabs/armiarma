package peering

/**
This file implements the peering service interface
With this interface, the given host will be able to retreive and connect a set of peers from the peerstore under the chosen strategy.

*/

import (
	"context"
	"fmt"
	"time"

	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	"github.com/migalabs/armiarma/src/metrics"
)

type PeeringOption func(*PeeringService) error

type PeeringOpts struct {
	InfoObj info.InfoData
	LogOpts base.LogOpts
}

// PeeringService is the main service that will connect peers from the given peerstore and using the given Host.
// It will use the specified peering strategy, which might difer/change from the testing or desired purposes of the run.
type PeeringService struct {
	*base.Base
	InfoObj   *info.InfoData
	host      *hosts.BasicLibp2pHost
	peerstore *metrics.PeerStore
	strategy  *PeeringStrategy
}

func NewPeeringService(ctx context.Context, h *hosts.BasicLibp2pHost, peerstore *metrics.PeerStore, peeringOpts PeeringOpts, opts ...PeeringOption) (*PeeringService, error) {
	// TODO: cancel is still not implemented in the BaseCreation
	peeringCtx, _ := context.WithCancel(ctx)
	logOpts := peeringOpts.LogOpts
	logOpts.ModName = "peering service"
	b, err := base.NewBase(
		base.WithContext(peeringCtx),
		base.WithLogger(logOpts),
	)
	if err != nil {
		return nil, err
	}

	pServ := &PeeringService{
		Base:      b,
		InfoObj:   peeringOpts.InfoObj,
		host:      h,
		peerstore: peerstore,
	}
	// iterate through the Options given as args
	for _, opt := range opts {
		err := opt(pServ)
		if err != nil {
			return nil, err
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

func WithPeeringStrategy(strategy *PeeringStrategy) PeeringOption {
	return func(p *PeeringService) error {
		if strategy == nil {
			return fmt.Errorf("given peering strategy is empty")
		}
		c.Log.Debugf("configuring peering with %s", strategy.Type())
		p.strategy = p.strategy
		return nil
	}
}

func (c *PeeringService) Start() {
	c.Log.Debug("starting the peering service")
	quit := make(chan struct{})
	// Set the defer function to cancel the go routine
	defer close(quit)
	h := c.host.Host()
	// set up the loop where every given time we will stop it to refresh the peerstore
	go func() {
		for quit != nil {
			peercount := 0
			// Make the copy of the peers from the peerstore
			peerList := c.peerstore.Peers()
			c.Log.Infof("the peerstore has been re-scanned")
			peerstoreLen := len(peerList)
			c.Log.Infof("len peerlist: %d", peerstoreLen)
			t := time.Now()
			for _, p := range peerList {
				// check if the peers is the crawler itself
				if p == h.ID() {
					c.Log.Debug("Calling to self host")
					continue
				}
				// read info about the peer
				pinfo, err := c.peerstore.GetPeerData(p.String())
				if err != nil {
					c.Log.Warn(err)
					pinfo = metrics.NewPeer(p.String())
				}
				// check if peer has been already deprecated for being many hours without connected
				wtime := pinfo.DaysToWait()
				if wtime != 0 {
					lconn, err := pinfo.LastAttempt()
					if err != nil {
						log.Warnf("ERROR, the peer should have a last connection attempt but list is empty")
					}
					lconnSecs := lconn.Add(time.Duration(wtime*12) * time.Hour).Unix()
					tnow := time.Now().Unix()
					// Compare time now with last connection plus waiting list
					if (tnow - lconnSecs) <= 0 {
						// If result is lower than 0, still have time to wait
						// continue to next peer
						continue
					}
				}
				peercount++
				// Set the correct address format to connect the peers
				// libp2p complains if we put multi-addresses that include the peer ID into the Addrs list.
				addrs := p.Addrs(p)
				addrInfo := peer.AddrInfo{
					ID:    p,
					Addrs: make([]ma.Multiaddr, 0, len(addrs)),
				}
				for _, m := range addrs {
					transport, _ := peer.SplitAddr(m)
					if transport == nil {
						continue
					}
					addrInfo.Addrs = append(addrInfo.Addrs, transport)
				}

				// compose all the detailed info for the given peer
				peerEnr := pinfo.ENR(p)
				// ensure the enr is not nil
				if peerEnr == nil {
					continue
				}
				addr, err := addrutil.EnodeToMultiAddr(peerEnr)
				if err != nil {
					log.Error("failed to parse ENR address into multi-addr for libp2p: %s", err)
				}

				pinfo.Pubkey = p.String()
				pinfo.NodeId = peerEnr.ID().String()
				pinfo.Ip = peerEnr.IP().String()
				pinfo.Addrs = addr.String()

				c.PeerStore.StoreOrUpdatePeer(pinfo)

				ctx, cancel := context.WithTimeout(ctx, c.Timeout)
				defer cancel()
				c.Log.Warnf("addrs %s attempting connection to peer", addrInfo.Addrs)
				// try to connect the peer
				attempts := 0
				for attempts < c.MaxRetries {
					if err := h.Connect(ctx, addrInfo); err != nil {
						c.Log.WithError(err).Warnf("attempts %d failed connection attempt", attempts)
						// the connetion failed
						c.RecErrorHandler(p, err.Error())
						attempts++
						continue
					} else { // connection successfuly made
						c.Log.Infof("peer_id %s successful connection made", p)
						c.PeerStore.AddNewPosConnectionAttempt(p.String())
						// break the loop
						break
					}
					if attempts >= c.MaxRetries {
						c.Log.Warnf("attempts %d failed connection attempt, reached maximum, no retry", attempts)
						break
					}
				}
			}
			tIter := time.Since(t)
			// Measure all the PeerStore iteration times
			c.PeerStore.NewPeerstoreIteration(tIter)
			// Measure the time of the entire PeerStore loop
			c.log.Infof("Time to ping the entire peerstore (except deprecated): %s", tIter)
			c.log.Infof("Peer attempted from the last reset: %d", len(peerList))
			/*
				// Force Garbage collector
				runtime.GC()
				runtime.FreeOSMemory()
			*/

			// Check if we have received any quit signal
			if quit == nil {
				log.Infof("Quit has been closed")
				break
			}
		}
		c.log.Infof("Go routine to randomly connect has been canceled")
	}()
}

func (p *PeeringService) Stop() {
	c.Log.Debug("stopìng the peering service")
}
