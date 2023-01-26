package kdht

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"

	kbucket "github.com/libp2p/go-libp2p-kbucket"
)

// fetchNeighbors sends RPC messages to the given peer and asks for its closest peers to an artificial set
// of 15 random peer IDs with increasing common prefix lengths (CPL). The returned peers are streamed
// to the results channel.
func (disc *KadDHTDiscService) fetchNeighbors(ctx context.Context, pi peer.AddrInfo) (*RoutingTable, error) {
	rt, err := kbucket.NewRoutingTable(20, kbucket.ConvertPeerID(pi.ID), time.Hour, nil, time.Hour, nil)
	if err != nil {
		return nil, err
	}

	allNeighborsLk := sync.RWMutex{}
	allNeighbors := map[peer.ID]peer.AddrInfo{}

	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	var errorBits uint32

	errg := errgroup.Group{}
	for i := uint(0); i <= 15; i++ { // 15 is maximum
		count := i // Copy value
		errg.Go(func() error {
			// Generate a peer with the given common prefix length
			rpi, err := rt.GenRandPeerID(count)
			if err != nil {
				atomic.StoreUint32(&errorBits, 1<<count)
				return errors.Wrapf(err, "generating random peer ID with CPL %d", count)
			}

			neighbors, err := disc.pm.GetClosestPeers(ctx, pi.ID, rpi)
			if err != nil {
				atomic.StoreUint32(&errorBits, 1<<count)
				return errors.Wrapf(err, "getting closest peer with CPL %d", count)
			}

			allNeighborsLk.Lock()
			defer allNeighborsLk.Unlock()
			for _, n := range neighbors {
				allNeighbors[n.ID] = *n
			}
			return nil
		})
	}
	err = errg.Wait()

	routingTable := &RoutingTable{
		PeerID:    pi.ID,
		Neighbors: []peer.AddrInfo{},
		ErrorBits: uint16(atomic.LoadUint32(&errorBits)),
		Error:     err,
	}

	for _, n := range allNeighbors {
		routingTable.Neighbors = append(routingTable.Neighbors, n)
	}

	return routingTable, err
}

// RoutingTable captures the routing table information and crawl error of a particular peer
type RoutingTable struct {
	// PeerID is the peer whose neighbors (routing table entries) are in the array below.
	PeerID peer.ID
	// The peers that are in the routing table of the above peer
	Neighbors []peer.AddrInfo
	// First error that has occurred during crawling that peer
	Error error
	// Little Endian representation of at which CPLs errors occurred during neighbors fetches.
	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	ErrorBits uint16
}

// extract info from the peer and compose
func (c *KadDHTDiscService) extractHostInfo(p peer.AddrInfo) *models.HostInfo {

	var mAddrs []ma.Multiaddr
	// look for the Public single IP
	for _, addr := range p.Addrs {
		// extract ip from mAddrs
		tempIP := utils.ExtractIPFromMAddr(addr)
		// check if IP is public
		if utils.IsIPPublic(tempIP) == true {
			// the IP is public
			mAddrs = append(mAddrs, addr)
		}
	}

	hInfo := models.NewHostInfo(
		p.ID,
		c.network,
		models.WithMultiaddress(mAddrs),
	)

	err := ReqIpfsPeerInfo(c.h, p.ID, hInfo)
	if err != nil {
		log.Debugf("unable to fetch peer info. %s", err.Error())
	}
	return hInfo
}

func ReqIpfsPeerInfo(h host.Host, peerID peer.ID, hInfo *models.HostInfo) error {
	// final error
	var finErr error

	var userAgent string
	// Fulfill the hInfo struct
	ua, err := h.Peerstore().Get(peerID, "AgentVersion")
	if err == nil {
		userAgent = ua.(string)
	} else {
		finErr = errors.Errorf("unable to identify peer. %s", err.Error())
	}

	var protocolVersion string
	// Into the new peer to fetch
	pv, err := h.Peerstore().Get(peerID, "ProtocolVersion")
	if err == nil {
		protocolVersion = pv.(string)
	}

	protocols := make([]string, 0)
	// Extract protocols
	if ps, err := h.Peerstore().GetProtocols(peerID); err == nil {
		copy(protocols, ps)
	}

	pInfo := models.NewPeerInfo(
		peerID,
		userAgent,
		protocolVersion,
		protocols,
		time.Duration(0),
	)

	hInfo.IdentifyHost(pInfo)
	return finErr
}
