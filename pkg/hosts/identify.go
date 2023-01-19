package hosts

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/migalabs/armiarma/pkg/utils/apis"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
)

// Identify the peer from the Libp2p Identify Service
type HostWithIDService interface {
	IDService() identify.IDService
}

// ReqBeaconMetadata:
// ReqHostInfo returns the basic host information regarding a given peer, from the libp2p perspective
// it aggregates the info from the libp2p Identify protocol adding some extra info such as RTT between local host and remote peer
// return empty struct and error if failure on the identify process.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
func ReqHostInfo(ctx context.Context, wg *sync.WaitGroup, h host.Host, ipLoc *apis.PeerLocalizer, conn network.Conn, peer *models.Peer, errIdent *error) {
	defer wg.Done()

	peerID := conn.RemotePeer()

	var finErr error
	// Identify Peer to access main data
	// convert host to IDService
	withIdentify, ok := h.(HostWithIDService)
	if !ok {
		finErr = errors.Errorf("host does not support libp2p identify protocol")
		*errIdent = finErr
		return
	}
	t := time.Now()
	idService := withIdentify.IDService()
	if idService == nil {
		finErr = errors.Errorf("libp2p identify not enabled on this host")
		*errIdent = finErr
		return
	}
	peer.MetadataRequest = true
	var rtt time.Duration
	select {
	case <-idService.IdentifyWait(conn):
		peer.MetadataSucceed = true
		peer.LastIdentifyTimestamp = time.Now()
		rtt = time.Since(t)
	case <-ctx.Done():
		finErr = errors.Errorf("identification error caused by timed out")
		*errIdent = finErr
		return
	}
	// Fill the the metrics
	// Into the new peer to fetch
	pv, err := h.Peerstore().Get(peerID, "ProtocolVersion")
	if err == nil {
		peer.ProtocolVersion = pv.(string)
	}

	prot, err := h.Peerstore().GetProtocols(peerID)
	if err == nil {
		peer.Protocols = prot
	}
	// Update the values of the
	peer.Latency = float64(rtt/time.Millisecond) / 1000
	peer.PeerId = peerID.String()
	multiAddrStr := conn.RemoteMultiaddr().String() + "/p2p/" + peerID.String()
	multiAddr, err := ma.NewMultiaddr(multiAddrStr)
	if err != nil {
		finErr = errors.Wrap(err, "unable to compose the maddrs")
		*errIdent = finErr
		fmt.Println("unable to extract multiaddrs")
		return
	}
	// generate array of MAddr to fit the models.Peer struct
	mAddrs := make([]ma.Multiaddr, 0)
	mAddrs = append(mAddrs, multiAddr)
	peer.MAddrs = mAddrs

	// Update, reqresp location from a peer only when the connection is inbound
	// if the connection is outbound, we already had the IP located from the ENR

	if conn.Stat().Direction.String() == "Inbound" {
		peer.Ip = utils.ExtractIPFromMAddr(multiAddr).String()
		locResp, err := ipLoc.LocateIP(peer.Ip)
		if err != nil {
			// TODO: think about a better idea to integrate a logger into this functions
			log.Warnf("error when fetching country/city from ip %s. %s", peer.Ip, err.Error())
		} else {
			peer.Country = locResp.Country
			peer.City = locResp.City
			peer.CountryCode = locResp.CountryCode
		}
	}

	peer.Ip = utils.ExtractIPFromMAddr(multiAddr).String()

	// locResp, err := ipLoc.LocateIP(peer.Ip)
	// if err != nil {
	// 	// TODO: think about a better idea to integrate a logger into this functions
	// 	log.Warnf("error when fetching country/city from ip %s. %s", peer.Ip, err.Error())
	// } else {
	// 	peer.Country = locResp.Country
	// 	peer.City = locResp.City
	// 	peer.CountryCode = locResp.CountryCode
	// }

	// Fulfill the hInfo struct
	ua, err := h.Peerstore().Get(peerID, "AgentVersion")
	if err == nil {
		peer.UserAgent = ua.(string)
		// Extract Client type and version
		peer.ClientName, peer.ClientVersion = utils.FilterClientType(peer.UserAgent)
		peer.ClientOS = "TODO"
	} else {
		// EDGY CASE: when peers refuse the connection, the callback gets called and the identify protocol
		// returns an empty struct (we are unable to identify them)
		finErr = errors.Errorf("unable to identify peer")
		peer.MetadataSucceed = false
	}
	pubk, err := conn.RemotePublicKey().Raw()
	if err == nil {
		peer.SetAtt("pubkey", hex.EncodeToString(pubk))
	}
	// return the erro defined in the top
	// nil if we could identify it, ident error if we couldnt line 181
	*errIdent = finErr
}
