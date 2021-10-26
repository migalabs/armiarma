package peering

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/db/utils"
	ma "github.com/multiformats/go-multiaddr"

	log "github.com/sirupsen/logrus"
)

var (
	DefaultDelay = 24 * time.Hour // hours of dealy after each negative attempt with delay
)

type PruningOpts struct {
	AggregatedDelay time.Duration
	LogOpts         base.LogOpts
}

// Pruning Strategy is a Peering Strategy that applies penalties to peers that haven't show activity when attempting to connect them.
// Combined with the Deprecated flag in the db.Peer struct, it produces more accourated metrics when exporting, pruning peers that are no longer active.
type PruningStrategy struct {
	*base.Base
	strategyType string
	PeerStore    *db.PeerStore
	// Delay unit time that gets applied to those slashed peers when reporting inactivity errors when activly connecting
	AggregatedDelay time.Duration

	/*
		// TODO: Choose the necessary parameters for the pruning
		FilterDigest beacon.ForkDigest `ask:"--filter-digest" help:"Only connect when the peer is known to have the given fork digest in ENR. Or connect to any if not specified."`
		FilterPort   int               `ask:"--filter-port" help:"Only connect to peers that has the given port advertised on the ENR."`
		Filtering    bool              `changed:"filter-digest"`
	*/
}

func NewPruningStrategy(ctx context.Context, peerstore *db.PeerStore, opts PruningOpts) (*PruningStrategy, error) {
	// TODO: cancel is still not implemented in the BaseCreation
	pruningCtx, _ := context.WithCancel(ctx)
	opts.LogOpts.ModName = "pruning strategy"
	b, err := base.NewBase(
		base.WithContext(pruningCtx),
		base.WithLogger(opts.LogOpts),
	)
	if err != nil {
		return nil, err
	}
	pr := &PruningStrategy{
		Base:         b,
		strategyType: "prunning",
		PeerStore:    peerstore,
	}
	return pr, nil
}

func (c PruningStrategy) Type() string {
	return c.strategyType
}

func (c *PruningStrategy) Start(ctx context.Context, args ...string) {
	c.Log.Info("Setting up pruning peering strategy")
	// Generate subcontext for the go-routine
	bgCtx, bgCancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	// Launch the pruning strategy go-routine
	go func() {
		c.run(bgCtx, c.PeerStore)
		close(done)
	}()

	cntC := make(chan os.Signal)
	signal.Notify(cntC, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cntC     // wait untill syscall.SIGTERM
		bgCancel() // shut down the sub-context
		<-done     // wait untill we track that the go routine was successfully closed
		c.Log.Infof("Stopped auto-connecting")
	}()
}

// main loop of the peering strategy.
// every 3-4 minutes generate a local new copy of the peers in the peerstore.
// It randomly selects one of the attempting to connect with it, recording the
// results of the attempts. If the peer was already connected, just dropt it
func (c *PruningStrategy) run(bgCtx context.Context, store *db.PeerStore) {
	quit := make(chan struct{})
	// Set the defer function to cancel the go routine
	defer close(quit)
	// set up the loop where every given time we will stop it to refresh the peerstore
	go func() {
		for quit != nil {
			peercount := 0
			// Make the copy of the peers from the peerstore
			peerList := store.GetPeerList()
			log.Infof("the peerstore has been re-scanned")
			peerstoreLen := len(peerList)
			log.Infof("len peerlist: %d", peerstoreLen)
			for _, p := range peerList {
				// check if the peers is the crawler itself
				/* TODO: Check from other source the ID of our host
				if p == h.ID() {
					log.Debug("Calling to self host")
					continue
				}
				*/
				// read info about the peer
				pinfo, err := c.PeerStore.GetPeerData(p.String())
				if err != nil {
					log.Warn(err)
					pinfo = db.NewPeer(p.String())
				}
				// check if peer has been already deprecated for being many hours without connected
				wtime := pinfo.DaysToWait()
				if wtime != 0 {
					lconn, err := pinfo.LastAttempt()
					if err != nil {
						log.Warnf("ERROR, the peer should have a last connection attempt but list is empty")
					}
					lconnSecs := lconn.Add(time.Duration(wtime) * DefaultDelay).Unix()
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
				// compose all the detailed info for the given peer
				peerEnr := pinfo.GetBlockchainNode()

				pinfo.Pubkey = p.String()
				pinfo.NodeId = peerEnr.ID().String()
				pinfo.Ip = peerEnr.IP().String()

				store.StoreOrUpdatePeer(pinfo)

				// Check if we have received any quit signal
				if quit == nil {
					log.Infof("Quit has been closed")
					break
				}
			}
		}
	}()
}

// Worker Implementations
func peeringWorker(ctx context.Context, ps *db.PeerStore, peerChan chan string) {

}

// function that selects actuation method for each of the possible errors while actively dialing peers
func (c *PruningStrategy) RecErrorHandler(pe peer.ID, rec_err string) {
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
			// Info about the peer should be added to the metrics
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
