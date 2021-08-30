package peer

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/protolambda/zrnt/eth2/beacon"
	log "github.com/sirupsen/logrus"
)

type PeerConnectRandomCmd struct {
	*base.Base
	Store         track.ExtendedPeerstore
	PeerStore *metrics.PeerStore
	Timeout       time.Duration `ask:"--timeout" help:"connection timeout, 0 to disable"`
	Rescan        time.Duration `ask:"--rescan" help:"rescan the peerscore for new peers to connect with this given interval"`
	MaxRetries    int           `ask:"--max-retries" help:"how many connection attempts until the peer is banned"`

	FilterDigest beacon.ForkDigest `ask:"--filter-digest" help:"Only connect when the peer is known to have the given fork digest in ENR. Or connect to any if not specified."`
	FilterPort   int               `ask:"--filter-port" help:"Only connect to peers that has the given port advertised on the ENR."`
	Filtering    bool              `changed:"filter-digest"`
}

func (c *PeerConnectRandomCmd) Default() {
	c.Timeout = 15 * time.Second
	c.Rescan = 10 * time.Minute
	c.MaxRetries = 3
	c.FilterPort = -1
}

func (c *PeerConnectRandomCmd) Help() string {
	return "Auto-connect to peers in the peerstore with a random-peering strategy."
}

func (c *PeerConnectRandomCmd) Run(ctx context.Context, args ...string) error {
	c.Log.Infof("Randomly connecting peers")
	h, err := c.Host()
	if err != nil {
		return err
	}
	fmt.Println("Peerstore Rescan Every:", c.Rescan)
	bgCtx, bgCancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		c.run(bgCtx, h, c.Store)
		close(done)
	}()

	c.Control.RegisterStop(func(ctx context.Context) error {
		bgCancel()
		c.Log.Infof("Stopped auto-connecting")
		fmt.Println("Stop Autoconnected")
		<-done
		return nil
	})
	return nil
}

// Get a random peer from the give peer list/slice
func randomPeer(peerList peer.IDSlice) peer.ID {
	rand.Seed(time.Now().UnixNano())
	return peerList[rand.Intn(len(peerList))]
}

// main loop of the peering strategy.
// every 3-4 minutes generate a local new copy of the peers in the peerstore.
// It randomly selects one of the attempting to connect with it, recording the
// results of the attempts. If the peer was already connected, just dropt it
func (c *PeerConnectRandomCmd) run(ctx context.Context, h host.Host, store track.ExtendedPeerstore) {
	c.Log.Info("started randomly peering")
	quit := make(chan struct{})
	// Set the defer function to cancel the go routine
	defer close(quit)
	// set up the loop where every given time we will stop it to refresh the peerstore
	go func() {
		loopCount := 0
		peerCache := make(map[peer.ID]bool)
		for quit != nil {
			if loopCount >= 10 {
				fmt.Println("Reseting cache")
				// generate a "cache of peers in this raw"
				peerCache = make(map[peer.ID]bool)
				loopCount = 0
			}
			// make the first copy of the peers from the Host
			//p := h.Peerstore()
			//peerList := p.Peers()

			// Make the first copy of the peers from the peerstore
			peerList := store.Peers()
			c.Log.Infof("the peerstore has been re-scanned")
			peerstoreLen := len(peerList)
			c.Log.Infof("len peerlist: %s", peerstoreLen)
			fmt.Println("Peerstore Re-Scanned with", peerstoreLen, "peers")
			t := time.Now()
			// loop to attempt connetions for the given time
			tgap := time.Since(t)
			for tgap < c.Rescan {
				p := randomPeer(peerList)
				// loop until we arrive to a peer that we didn't connect before

				// TODO: Add also IP taken from ENR of discovered peers
				peerMetrics := metrics.NewPeer(p.String())
				c.PeerStore.StoreOrUpdatePeer(peerMetrics)
				_, ok := peerCache[p]
				if ok {
					if len(peerCache) == peerstoreLen {
						loopCount += 1
						break // Temporary commented
					}
					continue
				}
				// add peer to the peerCache for this round
				peerCache[p] = true
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
				ctx, _ := context.WithTimeout(ctx, c.Timeout)
				c.Log.Warnf("addrs %s attempting connection to peer", addrInfo.Addrs)
				// try to connect the peer
				attempts := 0
				for attempts <= c.MaxRetries {
					if err := h.Connect(ctx, addrInfo); err != nil {
						// the connetion failed
						attempts += 1
						err := c.PeerStore.ConnectionAttemptEvent(p.String(), false, err.Error())
						if err != nil {
							log.Error("could not add new connection attemp: ", err)
						}
						c.Log.WithError(err).Warnf("attempts %d failed connection attempt", attempts)
						continue
					} else { // connection successfuly made
						c.Log.Infof("peer_id %s successful connection made", p)
						err := c.PeerStore.ConnectionAttemptEvent(p.String(), true, "None")
						if err != nil {
							log.Error("could not add new connection attemp: ", err)
						}
						break
					}
					if attempts > c.MaxRetries {
						c.Log.Warnf("attempts %d failed connection attempt, reached maximum, no retry", attempts)
						break
					}
				}
				tgap = time.Since(t)

			}

			fmt.Println("Restarting the peering")
			fmt.Println("Peer attempted from the last reset:", len(peerCache))

			// Check if we have received any quit signal
			if quit == nil {
				c.Log.Infof("Channel Quit has been closed")
				fmt.Println("Quit has been closed")
				break
			}
			loopCount += 1
		}
		c.Log.Infof("Go routine to randomly connect has been canceled")
		fmt.Println("Go routine to randomly connect has been canceled")
	}()
}
