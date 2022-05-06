package kdht

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/utils"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"

	kbucket "github.com/libp2p/go-libp2p-kbucket"
)

// fetchNeighbors sends RPC messages to the given peer and asks for its closest peers to an artificial set
// of 15 random peer IDs with increasing common prefix lengths (CPL). The returned peers are streamed
// to the results channel.
func (disc *IPFSDiscService) fetchNeighbors(ctx context.Context, pi peer.AddrInfo) (*RoutingTable, error) {
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
func (c *IPFSDiscService) extractHostInfo(p peer.AddrInfo) models.Peer {
	fpeer := models.NewPeer(p.ID.String())
	// look for the Public single IP
	for _, addr := range p.Addrs {
		// extract ip from mAddrs
		tempIP := utils.ExtractIPFromMAddr(addr)
		// check if IP is public
		if utils.IsIPPublic(tempIP) == true {
			// the IP is public
			fpeer.Ip = tempIP.String()
			fpeer.MAddrs = p.Addrs
		}
	}
	err := ReqIpfsHostInfo(c.h, p.ID, &fpeer)
	if err != nil {
		log.Debugf("unable to fetch peer info. %s", err.Error())
	}
	return fpeer
}

func ReqIpfsHostInfo(h host.Host, peerID peer.ID, p *models.Peer) error {
	// final error
	var finErr error

	// Fulfill the hInfo struct
	ua, err := h.Peerstore().Get(peerID, "AgentVersion")
	if err == nil {
		p.UserAgent = ua.(string)
	} else {
		finErr = errors.Errorf("unable to identify peer. %s", err.Error())
	}

	// Into the new peer to fetch
	pv, err := h.Peerstore().Get(peerID, "ProtocolVersion")
	if err == nil {
		p.ProtocolVersion = pv.(string)
	}
	// Extract protocols
	if protocols, err := h.Peerstore().GetProtocols(peerID); err == nil {
		p.Protocols = protocols
	}
	return finErr
}
