/*
	Copyright © 2021 Miga Labs
*/
package kdht

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/utils/apis"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	kdht "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"

	"github.com/sirupsen/logrus"
)

var (
	graceTime = 2 * time.Second

	workers = 1

	ModuleName = "KDHT-DISC"
	log        = logrus.WithField(
		"module", ModuleName,
	)
)

// IPFS discovery service with Kademlia DHT https://github.com/libp2p/go-libp2p-kad-dht
// Fulfilling the basic PeerDiscovery interfce for the Armiarma Crawler
type IPFSDiscService struct {
	ctx    context.Context
	cancel context.CancelFunc

	h     host.Host
	ipLoc *apis.PeerLocalizer

	timeout time.Duration
	pm      *pb.ProtocolMessenger
	ipfsDHT *kdht.IpfsDHT

	discPeers *discoveredPeers

	bootnodes []peer.AddrInfo
}

func NewIPFSDiscService(ctx context.Context, h host.Host, protocols []string, timeout time.Duration) IPFSDiscService {
	mainctx, cancel := context.WithCancel(ctx)

	// Generate necessary messenger for requesting near peers
	pm, err := pb.NewProtocolMessenger(&msgSender{
		h:         h,
		protocols: protocol.ConvertFromStrings(protocols),
		timeout:   10 * time.Second,
	})
	if err != nil {
		return IPFSDiscService{}
	}

	// Generate the new Kademlia DHT
	peerkdht, err := kdht.New(mainctx, h)
	if err != nil {
		log.Error(err)
	}

	// bootstrap
	log.Info("setting the bootstrap to dht")
	err = peerkdht.Bootstrap(mainctx)
	if err != nil {
		log.Error(err)
	}

	// Peer Discovery
	connectablePeers := NewDiscoveryPeers(mainctx)

	// get official bootstrap peers
	bootstrapNodes := kdht.GetDefaultBootstrapPeerAddrInfos()

	ipfsDisc := IPFSDiscService{
		ctx:       mainctx,
		cancel:    cancel,
		h:         h,
		timeout:   timeout,
		pm:        pm,
		ipfsDHT:   peerkdht,
		discPeers: &connectablePeers,
		bootnodes: bootstrapNodes,
	}
	return ipfsDisc
}

func (disc *IPFSDiscService) Start() {
	// get official bootstrap peers
	bootstrapNodes := kdht.GetDefaultBootstrapPeerAddrInfos()
	// add the bootnodes to the list of known peers
	bnCnt := 0
	for _, bootnode := range bootstrapNodes {
		disc.discPeers.addPeer(bootnode)
		bnCnt++
	}
	log.Infof("Adding %d bootstrap nodes", bnCnt)

	for i := 0; i < workers; i++ {
		log.Debugf("launching worker %d", i)
		workerid := i
		workerlog := log.WithField(
			"worker", workerid,
		)
		go func() {
			for {
				// get a new random peer to connect and request new peers
				nextp, ok := disc.discPeers.getRandomPeer()
				if !ok {
					workerlog.Tracef("there is no new peer to dial")
					time.Sleep(graceTime)
					continue
				}

				workerlog.Debugf(" connecting", nextp.ID.String())
				if err := disc.h.Connect(disc.ctx, nextp); err != nil {
					workerlog.Error(err.Error())
				} else {
					workerlog.Debugf("Connection established with bootstrap node:", nextp.ID.String())
					// If peer was connectable, req all the possible info
					// try to request neighbors to connected peer
					//addinfo, _ := utils.CompAddrInfo(nextp.ID.String(), nextp.MAddrs)
					neighborsRt, err := disc.fetchNeighbors(disc.ctx, nextp)
					if err != nil {
						workerlog.Debugf("unable to request neibors to peer. %s", err.Error())
					}
					// add peer to connectable list
					for _, newPeer := range neighborsRt.Neighbors {
						// add neihbors to the peer list
						disc.discPeers.addPeer(newPeer)
					}
				}
				// check if the context has died to cancel the routine
				if disc.ctx.Err() != nil {
					workerlog.Info("closing discover peer worker")
					return
				}
			}
		}()
	}
}

// return the the true if there is any new discovered peer, false if not
func (disc *IPFSDiscService) Next() bool {
	return disc.discPeers.next()
}

// fills the given peer struct that satifies the core.Basic standars empty
func (disc *IPFSDiscService) Peer() (models.Peer, bool) {
	// get the next peer from the discovered peers and filter it
	addrinfo := disc.discPeers.getNextPeer()
	if len(addrinfo.Addrs) == 0 {
		return models.Peer{}, false
	}
	// build the DiscoveredPeer from the PeerInfo
	p := models.NewPeer("")
	p.PeerId = addrinfo.ID.String()
	p.MAddrs = addrinfo.Addrs[:]

	// TODO: Not sure if there is actually an iterest to return IP / UserAgent / Protocols... /
	return p, true
}

type discoveredPeers struct {
	ctx    context.Context
	m      sync.Mutex
	pMap   sync.Map
	pArray []*peer.AddrInfo
	rp     *uint64
	wp     *uint64
}

func NewDiscoveryPeers(ctx context.Context) discoveredPeers {
	var rp uint64 = 0
	var wp uint64 = 0
	log.Trace("generating new DiscoveryPeers")
	dp := discoveredPeers{
		ctx:    ctx,
		pArray: make([]*peer.AddrInfo, 0),
		rp:     &rp,
		wp:     &wp,
	}
	return dp
}

// adds a new peer to the sync.Map and the array
func (d *discoveredPeers) addPeer(p peer.AddrInfo) {
	log.Tracef("Adding new Peer %s", p.ID.String())
	// check if the peer is already in the list
	_, ok := d.pMap.Load(p.ID.String())
	if ok {
		log.Tracef("peer %s already in peer list", p.ID.String())
		return
	}
	// Add peer to the sync Map
	d.pMap.Store(p.ID.String(), &p)
	// mutex and add it to the array
	log.Tracef("Lock adding peer to array. Peer: %s", p.ID.String())
	d.m.Lock()
	d.pArray = append(d.pArray, &p)
	d.m.Unlock()
	log.Tracef("Unlock adding peer to array. Peer: %s", p.ID.String())
	// increase the writer pointer
	atomic.AddUint64(d.wp, 1)
}

// returns weather there is a new peer to discover or not
func (d *discoveredPeers) next() bool {
	// check if the writing pointer is bigger than the reading one
	// meanning that there is a new item to read
	return atomic.LoadUint64(d.wp) > atomic.LoadUint64(d.rp)
}

// returns the addrinfo of the next discovered peer, only if there a new one to read
// empty addrinfo otherwise
func (d *discoveredPeers) getNextPeer() peer.AddrInfo {
	log.Tracef("getting next peer")
	// check if there is actually a new peer to read
	if !d.next() {
		log.Tracef("no next peer, returning empty peer")
		return peer.AddrInfo{}
	}
	// get the next peer found from the array
	log.Tracef("Lock reading peer from array")
	d.m.Lock()
	addinfo := d.pArray[*d.rp]
	d.m.Unlock()
	log.Tracef("Unlock reading peer from array")
	// Increase the pointer and check
	atomic.AddUint64(d.rp, 1)

	return *addinfo
}

//
func (d *discoveredPeers) getLen() int {
	d.m.Lock()
	l := len(d.pArray)
	d.m.Unlock()
	return l
}

func (d *discoveredPeers) isEmpty() bool {
	return d.getLen() == 0
}

// returns the addrinfo of the next discovered peer, only if there a new one to read
// empty addrinfo otherwise
func (d *discoveredPeers) getRandomPeer() (peer.AddrInfo, bool) {
	log.Tracef("getting random peer to request")
	// check if there list of peers is empty
	if d.isEmpty() {
		log.Tracef("empty list of peers to bootstrap")
		return peer.AddrInfo{}, false
	}
	// get random number
	rand.Seed(time.Now().Unix())
	randpointer := rand.Intn(d.getLen())
	log.Debugf("random pointer = %d", randpointer)

	// get the next peer found from the array
	log.Tracef("Lock reading peer from array")
	d.m.Lock()
	addinfo := d.pArray[randpointer]
	d.m.Unlock()
	log.Tracef("Unlock reading peer from array")
	return *addinfo, true
}
