package peer

import (
	"context"
	"fmt"
	"time"
	"math/rand"
	"github.com/protolambda/rumor/metrics"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/protolambda/zrnt/eth2/beacon"
)


type PeerConnectRandomCmd struct {
	*base.Base
	Store      track.ExtendedPeerstore
	GossipMetrics *metrics.GossipMetrics
	Timeout    time.Duration `ask:"--timeout" help:"connection timeout, 0 to disable"`
	Rescan     time.Duration `ask:"--rescan" help:"rescan the peerscore for new peers to connect with this given interval"`
	MaxRetries int       	 `ask:"--max-retries" help:"how many connection attempts until the peer is banned"`

	FilterDigest beacon.ForkDigest `ask:"--filter-digest" help:"Only connect when the peer is known to have the given fork digest in ENR. Or connect to any if not specified."`
	FilterPort   int   `ask:"--filter-port" help:"Only connect to peers that has the given port advertised on the ENR."`
    Filtering    bool  `changed:"filter-digest"`
}

func (c *PeerConnectRandomCmd) Default() {
	c.Timeout = 15 * time.Second
	c.Rescan = 1 * time.Minute
	c.MaxRetries = 5
    c.FilterPort = -1
}

func (c *PeerConnectRandomCmd) Help() string {
	return "Auto-connect to peers in the peerstore."
}

func (c *PeerConnectRandomCmd) Run(ctx context.Context, args ...string) error {
	fmt.Println("Auto connecting")
	h, err := c.Host()
	if err != nil {
		return err
	}

	bgCtx, bgCancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		c.run(bgCtx, h)
		close(done)
	}()
	fmt.Println("Free of lock")

	c.Control.RegisterStop(func(ctx context.Context) error {
		bgCancel()
		c.Log.Infof("Stopped auto-connecting")
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
func (c *PeerConnectRandomCmd) run(ctx context.Context, h host.Host) {
	c.Log.Info("started randomly peering")
	peerCache := make(map[peer.ID]bool,0)
	for { // for loop that will be executing until the ctx is canceled
		// make the first copy of the peerstore
		p := h.Peerstore()
		peerList := p.Peers()
		fmt.Println("the peerstore has been re-scanned")
		peerstoreLen := len(peerList)
		fmt.Println("len peerlist:", peerstoreLen)
		// set up the loop where every given time we will stop it to refresh the peerstore
		reset := false
		go func() {
			// loop to attempt connetions for the given time
			for {
				p := randomPeer(peerList)
				// loop until we arrive to a peer that we didn't connect before
				exists := c.GossipMetrics.ExtraMetrics.AddNewPeer(p)
				for exists == true {
					fmt.Println("Peer", p , "was already contacted")
					p = randomPeer(peerList)
					exists = c.GossipMetrics.ExtraMetrics.AddNewPeer(p)
					// if the peerstore is as big as the generated cache, break the loop and wait
					if peerstoreLen == len(peerCache) {
						fmt.Println("----------------> The Peerstore has been already attempted")
						break
					}
				}
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
				fmt.Println("peer_id", p)
				fmt.Println("addrs", addrInfo.Addrs, "attempting connection to peer")
				// try to connect the peer
				attempts := 1
				for attempts <= c.MaxRetries {
					if err := h.Connect(ctx, addrInfo); err != nil {
						// the connetion failed
						attempts += 1
						c.GossipMetrics.ExtraMetrics.AddNewAttempt(p, false)
						fmt.Println("attempts", attempts, "failed connection attempt", err)
					} else { // connection successfuly made
						fmt.Println("peer_id", p, "successful connection made")
						c.GossipMetrics.ExtraMetrics.AddNewAttempt(p, true)
						// break the loop
						break
					}
					if attempts > c.MaxRetries {
						fmt.Println("attempts", attempts,"failed connection attempt, reached maximum, no retry")
					}
					// if the reset flag is active, kill the go-routine
					if reset == true {
						return
					}
				}
			}
		}()
		time.Sleep(time.Minute * c.Rescan)
		fmt.Println("---------------------- RESET ----------------")
		reset = true
	}
}

