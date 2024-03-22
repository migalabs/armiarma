package hosts

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils/apis"
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
func ReqHostInfo(ctx context.Context, wg *sync.WaitGroup, h host.Host, ipLoc *apis.IpLocator, conn network.Conn, hInfo *models.HostInfo, errIdent *error) {
	defer wg.Done()

	peerID := conn.RemotePeer()

	var finErr error = errors.New("None")
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
	var rtt time.Duration
	select {
	case <-idService.IdentifyWait(conn):
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
		hInfo.PeerInfo.ProtocolVersion = pv.(string)
	}

	prots, err := h.Peerstore().GetProtocols(peerID)
	if err == nil {
		supportedProtocols := make([]string, len(prots))
		for i, prot := range prots {
			supportedProtocols[i] = string(prot)
		}
		hInfo.PeerInfo.Protocols = supportedProtocols
	}
	// Update the values of the
	hInfo.PeerInfo.Latency = rtt
	hInfo.PeerInfo.RemotePeer = peerID

	// Fulfill the hInfo struct
	ua, err := h.Peerstore().Get(peerID, "AgentVersion")
	if err == nil {
		hInfo.PeerInfo.UserAgent = ua.(string)
	} else {
		// EDGY CASE: when peers refuse the connection, the callback gets called and the identify protocol
		// returns an empty struct (we are unable to identify them)
		finErr = errors.Errorf("unable to identify peer")
		// peer.MetadataSucceed = false
	}
	// pubk, err := conn.RemotePublicKey().Raw()
	// if err == nil {
	// 	peer.SetAtt("pubkey", hex.EncodeToString(pubk))
	// }
	// return the erro defined in the top
	// nil if we could identify it, ident error if we couldnt line 181
	*errIdent = finErr
}
