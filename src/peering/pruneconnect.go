package peer

import (
	"context"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/metrics"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/protolambda/rumor/p2p/addrutil"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/protolambda/zrnt/eth2/beacon"
	log "github.com/sirupsen/logrus"
)

var (
	DefaultDelay = 24 // hours of dealy after each negative attempt with delay
)

type PeerPruneConncetCmd struct {
	*base.Base
	Store      track.ExtendedPeerstore
	PeerStore  *metrics.PeerStore
	Timeout    time.Duration `ask:"--timeout" help:"connection timeout, 0 to disable"`
	MaxRetries int           `ask:"--max-retries" help:"how many connection attempts until the peer is banned"`

	FilterDigest beacon.ForkDigest `ask:"--filter-digest" help:"Only connect when the peer is known to have the given fork digest in ENR. Or connect to any if not specified."`
	FilterPort   int               `ask:"--filter-port" help:"Only connect to peers that has the given port advertised on the ENR."`
	Filtering    bool              `changed:"filter-digest"`
}

func (c *PeerPruneConncetCmd) Default() {
	c.Timeout = 10 * time.Second
	c.MaxRetries = 1
	c.FilterPort = -1
}

func (c *PeerPruneConncetCmd) Help() string {
	return "Auto-connect to peers in the peerstore with a random-peering strategy."
}

func (c *PeerPruneConncetCmd) Run(ctx context.Context, args ...string) error {
	c.Log.Infof("Randomly connecting peers")
	h, err := c.Host()
	if err != nil {
		return err
	}
	bgCtx, bgCancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		c.run(bgCtx, h, c.Store)
		close(done)
	}()

	c.Control.RegisterStop(func(ctx context.Context) error {
		bgCancel()
		c.Log.Infof("Stopped auto-connecting")
		<-done
		return nil
	})
	return nil
}

// main loop of the peering strategy.
// every 3-4 minutes generate a local new copy of the peers in the peerstore.
// It randomly selects one of the attempting to connect with it, recording the
// results of the attempts. If the peer was already connected, just dropt it
func (c *PeerPruneConncetCmd) run(ctx context.Context, h host.Host, store track.ExtendedPeerstore) {
	c.Log.Info("started randomly peering")
	quit := make(chan struct{})
	// Set the defer function to cancel the go routine
	defer close(quit)
	// set up the loop where every given time we will stop it to refresh the peerstore
	go func() {
		for quit != nil {
			peercount := 0
			// Make the copy of the peers from the peerstore
			peerList := store.Peers()
			log.Infof("the peerstore has been re-scanned")
			peerstoreLen := len(peerList)
			log.Infof("len peerlist: %d", peerstoreLen)
			t := time.Now()
			for _, p := range peerList {
				// check if the peers is the crawler itself
				if p == h.ID() {
					log.Debug("Calling to self host")
					continue
				}
				// read info about the peer
				pinfo, err := c.PeerStore.GetPeerData(p.String())
				if err != nil {
					log.Warn(err)
					pinfo = metrics.NewPeer(p.String())
				}
				// check if peer has been already deprecated for being many hours without connected
				wtime := pinfo.DaysToWait()
				if wtime != 0 {
					lconn, err := pinfo.LastAttempt()
					if err != nil {
						log.Warnf("ERROR, the peer should have a last connection attempt but list is empty")
					}
					lconnSecs := lconn.Add(time.Duration(wtime*DefaultDelay) * time.Hour).Unix()
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
				addrs := c.Store.Addrs(p)
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
				peerEnr := c.Store.LatestENR(p)
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
				/* Deprecated for now, too many request for the IP-Localization
				country, city, err := utils.GetLocationFromIp(peer.Ip)
				if err != nil {
					log.Warn("could not get location from ip: ", peer.Ip, err)
				} else {
					peer.Country = country
					peer.City = city
				}
				*/
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
			log.Infof("Time to ping the entire peerstore (except deprecated): %s", tIter)
			log.Infof("Peer attempted from the last reset: %d", len(peerList))
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
		log.Infof("Go routine to randomly connect has been canceled")
	}()
}

// Worker Implementations
func peeringWorker(ctx context.Context, ps *metrics.PeerStore, peerChan chan string) {

}

// function that selects actuation method for each of the possible errors while actively dialing peers
//
func (c *PeerPruneConncetCmd) RecErrorHandler(pe peer.ID, rec_err string) {
	var fn func(p *metrics.Peer)
	switch utils.FilterError(rec_err) {
	case "Connection reset by peer":
		fn = func(p *metrics.Peer) {
			p.AddNegConnAtt()
		}
	case "i/o timeout":
		fn = func(p *metrics.Peer) {
			p.AddNegConnAttWithPenalty()
		}
	case "dial to self attempted":
		// we tried to peer ourselfs! deprecate the peer
		fn = func(p *metrics.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	case "dial backoff":
		fn = func(p *metrics.Peer) {
			p.AddNegConnAtt()
		}
	case "connection refused":
		fn = func(p *metrics.Peer) {
			p.AddNegConnAtt()
		}
	case "no route to host":
		fn = func(p *metrics.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	case "unreachable network":
		fn = func(p *metrics.Peer) {
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
			// Info about the peer should be added to the metrics
			// *** Carefull - problems with reading the pubkey ***
			//newP := metrics.NewPeer(newPeerID.String())
			//c.PeerStore.Store(newPeerID.String(), newP)
			_, _ = c.Store.UpdateENRMaybe(newPeerID, enr)
			c.Store.AddAddrs(newPeerID, addrs, time.Duration(48)*time.Hour)
		*/
		fn = func(p *metrics.Peer) {
			p.AddNegConnAtt()
			p.Deprecated = true
		}
	default:
		fn = func(p *metrics.Peer) {
			p.AddNegConnAttWithPenalty()
		}
	}
	c.PeerStore.AddNewNegConnectionAttempt(pe.String(), rec_err, fn)
}
