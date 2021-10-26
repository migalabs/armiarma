package peering

/**
This file implements the peering service interface
With this interface, the given host will be able to retreive and connect a set of peers from the peerstore under the chosen strategy.

*/

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/db/utils"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
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
	strategy  *PeeringStrategy
	//
	Timeout    time.Duration
	MaxRetries int
}

func NewPeeringService(ctx context.Context, h *hosts.BasicLibp2pHost, peerstore *db.PeerStore, peeringOpts *PeeringOpts, opts ...PeeringOption) (*PeeringService, error) {
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
		Base:       b,
		InfoObj:    peeringOpts.InfoObj,
		host:       h,
		PeerStore:  peerstore,
		Timeout:    15 * time.Second, // TODO: Hardcoded to 15 seconds
		MaxRetries: 1,                // TODO: Hardcoded to 1 retry, future retries are directly dismissed/dropped by dialing peer
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
		p.Log.Debugf("configuring peering with %s", (*strategy).Type())
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
			peerList := c.PeerStore.GetPeerList()
			c.Log.Infof("the peerstore has been re-scanned with %d peer", len(peerList))

			t := time.Now()
			for _, p := range peerList {
				// check if the peers is the crawler itself
				if p == h.ID() {
					c.Log.Debug("Calling to self host")
					continue
				}
				// read info about the peer
				pinfo, err := c.PeerStore.GetPeerData(p.String())
				if err != nil {
					c.Log.Debug(err)
					pinfo = db.NewPeer(p.String())
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
				addrs := pinfo.ExtractPublicAddr()
				addrInfo := peer.AddrInfo{
					ID:    p,
					Addrs: make([]ma.Multiaddr, 0, 1),
				}

				transport, _ := peer.SplitAddr(addrs)
				if transport == nil {
					continue
				}
				addrInfo.Addrs = append(addrInfo.Addrs, transport)

				// compose all the detailed info for the given peer
				peerEnr := pinfo.GetBlockchainNode()

				pinfo.Pubkey = p.String()
				pinfo.NodeId = peerEnr.ID().String()
				pinfo.Ip = peerEnr.IP().String()

				c.PeerStore.StoreOrUpdatePeer(pinfo)

				ctx, cancel := context.WithTimeout(c.Ctx(), c.Timeout)
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
			c.Log.Infof("Time to ping the entire peerstore (except deprecated): %s", tIter)
			c.Log.Infof("Peer attempted from the last reset: %d", len(peerList))
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
		c.Log.Infof("Go routine to randomly connect has been canceled")
	}()
}

func (c *PeeringService) Stop() {
	c.Log.Debug("stopìng the peering service")
}

// function that selects actuation method for each of the possible errors while actively dialing peers
func (c *PeeringService) RecErrorHandler(pe peer.ID, rec_err string) {
	var fn func(p *db.Peer)
	switch utils.FilterError(rec_err) {
	case "Connection reset by peer":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
		}
	case "i/o timeout":
		fn = func(p *db.Peer) {
			p.AddNegConnAttWithPenalty()
		}
	case "dial to self attempted":
		// we tried to peer ourselfs! deprecate the peer
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	case "dial backoff":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
		}
	case "connection refused":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
		}
	case "no route to host":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	case "unreachable network":
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	case "peer id mismatch, peer dissmissed":
		// deprecate old peer and generate a new one
		// trim new peerID from error message
		np := strings.Split(rec_err, "key matches ")[1]
		np = strings.Replace(np, ")", "", -1)
		//newPeerID := peer.ID(np)
		//f.WriteString(fmt.Sprintf("%s shifted to %s\n", pe.String(), newPeerID))
		// Generate a new Peer with the addrs of the previous one and the ID suggested at the
		log.Infof("deprecating peer %s, but adding possible new peer %s", pe.String(), np)
		/*
			_, err := newPeerID.ExtractPublicKey()
			if err != nil {
				fmt.Println("error obtainign pubkey from peerid", err)
			} else {
				fmt.Println("pubkey success, obtained")
			}
			TODO: -Fix empty pubkey originated from adding the new PeerID to the Peerstore
					-Apparently the len of the new peer doesn't have the same one as the previous one
			// Generate new Addrs for the possible new discovered peer
			addrs := c.Store.Addrs(pe)
			enr := c.Store.LatestENR(pe)
			fmt.Println("len old", len(pe.String()), "len new", len(newPeerID.String()))
			fmt.Println(rec_err)
			// Info about the peer should be added to the db
			// *** Carefull - problems with reading the pubkey ***
			//newP := db.NewPeer(newPeerID.String())
			//c.PeerStore.Store(newPeerID.String(), newP)
			_, _ = c.Store.UpdateENRMaybe(newPeerID, enr)
			c.Store.AddAddrs(newPeerID, addrs, time.Duration(48)*time.Hour)
		*/
		fn = func(p *db.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	default:
		fn = func(p *db.Peer) {
			p.AddNegConnAttWithPenalty()
		}
	}
	c.PeerStore.AddNewNegConnectionAttempt(pe.String(), rec_err, fn)
}
