/*
	Copyright © 2021 Miga Labs
*/
package kdht

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/discovery"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/migalabs/armiarma/pkg/utils/apis"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	kdht "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"

	"github.com/sirupsen/logrus"
)

var (
	graceTime = 3 * time.Second
	timeout   = 10 * time.Second

	workers    = 25
	ModuleName = "KDHT-DISC"
	log        = logrus.WithField(
		"module", ModuleName,
	)
)

// IPFS discovery service with Kademlia DHT https://github.com/libp2p/go-libp2p-kad-dht
// Fulfilling the basic PeerDiscovery interfce for the Armiarma Crawler
type IPFSDiscService struct {
	ctx context.Context

	h     host.Host
	ipLoc *apis.PeerLocalizer

	timeout time.Duration
	pm      *pb.ProtocolMessenger
	ipfsDHT *kdht.IpfsDHT

	discPeers *discoveredPeers

	bootnodes []peer.AddrInfo
}

func NewIPFSDiscService(ctx context.Context, h host.Host, protocols []string, bootstrapnodes []peer.AddrInfo, timeout time.Duration) IPFSDiscService {

	ms := &msgSender{
		h:         h,
		protocols: protocol.ConvertFromStrings(protocols),
		timeout:   timeout,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		log.Panicf("unable to generate protocol messenger for kdht %s", err)
	}

	// Generate the new Kademlia DHT
	peerkdht, err := kdht.New(ctx, h)
	if err != nil {
		log.Panicf("unable to generate the kdht %s", err)
	}

	// bootstrap
	log.Info("setting the bootstrap to dht")
	err = peerkdht.Bootstrap(ctx)
	if err != nil {
		log.Panicf("unable to generate the bootstrap %s", err)
	}

	// Peer Discovery
	connectablePeers := NewDiscoveryPeers(ctx)

	// get official bootstrap peers reading from the bootstrapnodes file
	// Alternatively, user 'bootstrapNodes := kdht.GetDefaultBootstrapPeerAddrInfos()' for IPFS ones
	log.Infof("Bootnodes: %s", bootstrapnodes)
	ipfsDisc := IPFSDiscService{
		ctx:       ctx,
		h:         h,
		timeout:   timeout,
		pm:        pm,
		ipfsDHT:   peerkdht,
		discPeers: &connectablePeers,
		bootnodes: bootstrapnodes,
	}
	return ipfsDisc
}

func (disc *IPFSDiscService) Start() {
	// get official bootstrap peers
	//bootstrapNodes := kdht.GetDefaultBootstrapPeerAddrInfos()
	// add the bootnodes to the list of known peers
	bnCnt := 0
	for _, bootnode := range disc.bootnodes {
		disc.discPeers.addPeer(bootnode)
		bnCnt++
	}
	log.Infof("Adding %d bootstrap nodes", bnCnt)

	workerTask := func(workerlog logrus.FieldLogger, nextp peer.AddrInfo) {
		ctx := network.WithDialPeerTimeout(disc.ctx, timeout)
		// Force direct dials will prevent swarm to run into dial backoff errors. It also prevents proxied connections.
		ctx = network.WithForceDirectDial(ctx, "prevent backoff")

		workerlog.Debugf("connecting %s", nextp.ID.String())
		if err := disc.h.Connect(ctx, nextp); err != nil {
			workerlog.Debugf("unable to connect peer %s - %s", nextp.ID.String(), err.Error())
		} else {
			workerlog.Debugf("Connection established with bootstrap node: %s", nextp.ID.String())
			// If peer was connectable, req all the possible info
			// try to request neighbors to connected peer
			//addinfo, _ := utils.CompAddrInfo(nextp.ID.String(), nextp.MAddrs)
			neighborsRt, err := disc.fetchNeighbors(disc.ctx, nextp)
			if err != nil {
				workerlog.Debugf("unable to request neibors to peer. %s", err.Error())
			}
			// add peer to connectable list
			log.Debugf("%d neighbours for peer", len(neighborsRt.Neighbors))
			for _, newPeer := range neighborsRt.Neighbors {
				// add neihbors to the peer list
				disc.discPeers.addPeer(newPeer)
			}
			// Free connection resources
			if err := disc.h.Network().ClosePeer(nextp.ID); err != nil {
				log.Warnf("Could not close connection to peer %s", err)
			}
		}
	}

	for i := 0; i < workers; i++ {
		log.Debugf("launching worker %d", i)
		workerid := i
		workerlog := log.WithField(
			"worker", workerid,
		)
		go func() {
			for {
				// get a new random peer to connect and request new peers
				nextp, ok := disc.discPeers.getBootstrapPeer() //getRandomPeer()
				if !ok {
					workerlog.Debugf("there is no new peer to dial")
					time.Sleep(graceTime)
					continue
				}

				workerTask(workerlog, nextp)

				// check if the context has died to cancel the routine
				if disc.ctx.Err() != nil {
					workerlog.Info("closing discover peer worker")
					return
				}
			}
		}()
	}
	// Random Peers
	for i := 0; i < workers; i++ {
		log.Debugf("launching worker %d", i)
		workerid := workers + i
		workerlog := log.WithField(
			"worker", workerid,
		)
		go func() {
			for {
				// get a new random peer to connect and request new peers
				nextp, ok := disc.discPeers.getRandomPeer()
				if !ok {
					workerlog.Debugf("there is no new peer to dial")
					time.Sleep(graceTime)
					continue
				}

				workerTask(workerlog, nextp)

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
	pubAddrs := utils.GetPublicAddrsFromAddrArray(addrinfo.Addrs)
	p.MAddrs = append(p.MAddrs, pubAddrs)
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
	bp     *uint64
}

func NewDiscoveryPeers(ctx context.Context) discoveredPeers {
	var rp uint64 = 0
	var wp uint64 = 0
	var bp uint64 = 0
	log.Trace("generating new DiscoveryPeers")
	dp := discoveredPeers{
		ctx:    ctx,
		pArray: make([]*peer.AddrInfo, 0),
		rp:     &rp,
		wp:     &wp,
		bp:     &bp,
	}
	return dp
}

// adds a new peer to the sync.Map and the array
func (d *discoveredPeers) addPeer(p peer.AddrInfo) {
	log.Debugf("Adding new Peer %s", p.ID.String())
	// check if the peer is already in the list
	_, ok := d.pMap.Load(p.ID.String())
	if ok {
		log.Debugf("peer %s already in peer list", p.ID.String())
		return
	}
	// Add peer to the sync Map
	d.pMap.Store(p.ID.String(), &p)
	// mutex and add it to the array
	log.Debugf("Lock adding peer to array. Peer: %s", p.ID.String())
	d.m.Lock()
	d.pArray = append(d.pArray, &p)
	log.Debugf("len array: %d", len(d.pArray))
	d.m.Unlock()
	log.Debugf("Unlock adding peer to array. Peer: %s", p.ID.String())
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
	log.Debugf("getting next peer")
	// check if there is actually a new peer to read
	if !d.next() {
		log.Debugf("no next peer, returning empty peer")
		return peer.AddrInfo{}
	}
	// get the next peer found from the array
	log.Debugf("Lock reading peer from array")
	d.m.Lock()
	addinfo := d.pArray[*d.rp]
	d.m.Unlock()
	log.Debugf("Unlock reading peer from array")
	// Increase the pointer and check
	atomic.AddUint64(d.rp, 1)

	return *addinfo
}

//
func (d *discoveredPeers) getLen() int {
	d.m.Lock()
	l := len(d.pArray)
	log.Debugf("len array: %d", len(d.pArray))
	d.m.Unlock()
	return l
}

func (d *discoveredPeers) isEmpty() bool {
	return d.getLen() == 0
}

// returns the addrinfo of the next discovered peer, only if there a new one to read
// empty addrinfo otherwise
func (d *discoveredPeers) getBootstrapPeer() (peer.AddrInfo, bool) {
	log.Debugf("getting next peer to request")
	// check if there list of peers is empty
	if d.isEmpty() {
		log.Debugf("empty list of peers to bootstrap")
		return peer.AddrInfo{}, false
	}

	log.Debugf("len of pArray: %d", d.getLen())
	// cehck if bootstrap pointer is bigger than d.getLen()
	if atomic.LoadUint64(d.bp) >= uint64(d.getLen()) {
		atomic.StoreUint64(d.bp, 0)
	}

	// get the next peer found from the array

	d.m.Lock()
	log.Debugf("Lock reading peer from array")
	addinfo := d.pArray[*d.bp]
	log.Debugf("next pointer = %d", *d.bp)
	*d.bp++
	d.m.Unlock()
	log.Debugf("Unlock reading peer from array")
	return *addinfo, true
}

// returns the addrinfo of the next discovered peer, only if there a new one to read
// empty addrinfo otherwise
func (d *discoveredPeers) getRandomPeer() (peer.AddrInfo, bool) {
	log.Debugf("getting random peer to request")
	// check if there list of peers is empty
	if d.isEmpty() {
		log.Debugf("empty list of peers to bootstrap")
		return peer.AddrInfo{}, false
	}
	// get random number
	rand.Seed(time.Now().Unix())
	randpointer := rand.Intn(d.getLen())
	log.Debugf("len of pArray: %d", d.getLen())
	log.Debugf("random pointer = %d", randpointer)

	// get the next peer found from the array
	log.Debugf("Lock reading peer from array")
	d.m.Lock()
	addinfo := d.pArray[randpointer]
	d.m.Unlock()
	log.Debugf("Unlock reading peer from array")
	return *addinfo, true
}

// ImportBootNodeList
// This method will read the bootnodes list in string format and create an
// enode array with the parsed ENRs of the bootnodes.
// @param import_json_file represents the file where to read the bootnodes from.
// This file is configured in the config file.
func ReadIpfsBootnodeFile(jfile string) ([]peer.AddrInfo, error) {

	// where we will store the result
	bootNodeList := make([]peer.AddrInfo, 0)

	// where we will unmarshal from file
	bootNodeListString := discovery.BootNodeListString{}

	// check if file exists
	if _, err := os.Stat(jfile); os.IsNotExist(err) {
		return bootNodeList, errors.New("Bootnodes file does not exist")
	} else {
		// exists
		file, err := ioutil.ReadFile(jfile)
		if err == nil {
			err := json.Unmarshal([]byte(file), &bootNodeListString)
			if err != nil {
				return bootNodeList, errors.Wrap(err, "Could not Unmarshal BootNodes file: "+jfile)
			}
		} else {
			return bootNodeList, errors.Wrap(err, "Could not read BootNodes file: %s"+jfile)
		}
	}

	// parse bootnode strings into enodes
	for _, element := range bootNodeListString.BootNodes {
		// unmarshall MAddrs
		maddr, err := utils.UnmarshalMaddr(element)
		if err != nil {
			return bootNodeList, err
		}
		// comopose AddrInfo
		addinfo, err := peer.AddrInfosFromP2pAddrs(maddr)
		if err != nil {
			return bootNodeList, err
		}
		if len(addinfo) <= 0 {
			return bootNodeList, errors.New("error empty generating AddrInfo from bootnode MAddrs")
		}
		bootNodeList = append(bootNodeList, addinfo[0])
	}

	log.Debugf("bootnodes in File: %d", len(bootNodeListString.BootNodes))
	log.Debugf("imported: %d", len(bootNodeList))
	return bootNodeList, nil

}
