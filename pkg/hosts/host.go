package hosts

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	mplex "github.com/libp2p/go-libp2p-mplex"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"

	log "github.com/sirupsen/logrus"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/migalabs/armiarma/pkg/utils/apis"

	"github.com/pkg/errors"
)

var (
	ConnNotChannSize = 256
)

type P2pNetwork interface {
	Network() utils.NetworkType
}

// Struct that defines the Basic Struct asociated to the Libtp2p host
type BasicLibp2pHost struct {
	ctx context.Context

	// Basic sevices related with the libp2p host
	host        host.Host
	identify    identify.IDService
	IpLocator   *apis.IpLocator
	NetworkNode P2pNetwork

	// Basic Host Metadata
	multiAddr ma.Multiaddr

	connEventNotChannel chan *models.EventTrace
	identNotChannel     chan IdentificationEvent
	peerID              peer.ID
}

// NewBasicLibp2pEth2Host generate a new Libp2p host from the given context and Options, for Eth2 network (or similar).
func NewBasicLibp2pEth2Host(
	ctx context.Context,
	ip string,
	port int,
	privKey crypto.PrivKey,
	userAgent string,
	netNode P2pNetwork,
	ipLocator *apis.IpLocator) (*BasicLibp2pHost, error) {

	// generate de multiaddress
	multiaddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, port))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("couldn't generate multiaddress from ip %s and tcp %s", ip, port))
	}

	// resource manager we don't want the host to be limited by anything
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

	// Generate the main Libp2p host that will be exposed to the network
	host, err := libp2p.New(
		libp2p.ListenAddrs(multiaddr),
		libp2p.Identity(privKey),
		libp2p.UserAgent(userAgent),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Muxer(mplex.ID, mplex.DefaultTransport),
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
		libp2p.NATPortMap(),
		libp2p.ResourceManager(rm),
		libp2p.ConnectionManager(connmgr.NullConnMgr{}),
	)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"maddrs": multiaddr.String(),
		"peerID": host.ID().String(),
	}).Info("libp2p successfully generated")

	// generate the identify service
	ids, err := identify.NewIDService(
		host,
		identify.UserAgent(userAgent),
		identify.DisableSignedPeerRecord(),
	)
	if err != nil {
		return nil, err
	}

	// Gererate the struct that contains all the configuration and structs surrounding the Libp2p Host
	basicHost := &BasicLibp2pHost{
		ctx:                 ctx,
		NetworkNode:         netNode,
		host:                host,
		identify:            ids,
		IpLocator:           ipLocator,
		multiAddr:           multiaddr,
		peerID:              host.ID(),
		connEventNotChannel: make(chan *models.EventTrace, ConnNotChannSize),
		identNotChannel:     make(chan IdentificationEvent, ConnNotChannSize),
	}
	log.Debug("setting custom notification functions")
	basicHost.SetCustomNotifications()

	return basicHost, nil
}

// This should be move to to a range of options inside the NewBasicHost

// NewBasicLibp2pFilecoinHost:
// Generate a new Libp2p host from the given context and Options, for Filecoin network.
// func NewBasicLibp2pIpfsHost(ctx context.Context, infObj info.IpfsInfoData, ipLocator *apis., ps *db.PeerStore) (*BasicLibp2pHost, error) {
// 	ip := infObj.IP.String()
// 	tcp := fmt.Sprintf("%d", infObj.TcpPort)
// 	privkey := infObj.PrivateKey
// 	userAgent := infObj.UserAgent

// 	// generate de multiaddress
// 	multiaddr := fmt.Sprintf("/ip4/%s/tcp/%s", ip, tcp)
// 	muladdr, err := ma.NewMultiaddr(multiaddr)
// 	if err != nil {
// 		log.Debugf("couldn't generate multiaddress from ip %s and tcp %s", ip, tcp)
// 		multiaddr = fmt.Sprintf("/ip4/%s/tcp/%s", DefaultIP, DefaultTCP)
// 		muladdr, _ = ma.NewMultiaddr(multiaddr)
// 	}
// 	log.Debugf("setting multiaddress to %s", muladdr)

// 	// Generate the main Libp2p host that will be exposed to the network
// 	host, err := libp2p.New(
// 		libp2p.Identity(privkey),
// 		libp2p.ListenAddrs(muladdr),
// 		libp2p.Ping(true),
// 		libp2p.UserAgent(userAgent),
// 		libp2p.NATPortMap(),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	peerId := host.ID().String()
// 	fmaddr := host.Addrs()[0].String() + "/p2p/" + host.ID().String()
// 	localMultiaddr, _ := ma.NewMultiaddr(fmaddr)
// 	log.Debugf("full multiaddress %s", localMultiaddr)
// 	// generate the identify service
// 	ids, err := identify.NewIDService(host, identify.UserAgent(userAgent), identify.DisableSignedPeerRecord())
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Gererate the struct that contains all the configuration and structs surrounding the Libp2p Host
// 	basicHost := &BasicLibp2pHost{
// 		ctx:                 ctx,
// 		Network:             "ipfs",
// 		host:                host,
// 		identify:            ids,
// 		PeerStore:           ps,
// 		IpLocator:           ipLocator,
// 		multiAddr:           muladdr,
// 		fullMultiAddr:       localMultiaddr,
// 		peerID:              peer.ID(peerId),
// 		connEventNotChannel: make(chan ConnectionEvent, ConnNotChannSize),
// 		identNotChannel:     make(chan IdentificationEvent, ConnNotChannSize),
// 	}
// 	log.Debug("setting custom notification functions")
// 	basicHost.SetCustomNotifications()

// 	return basicHost, nil
// }

func (b *BasicLibp2pHost) Host() host.Host {
	return b.host
}

// Start spawns the libp2pHost module
// So far, start listening on the multiAddrs.
func (b *BasicLibp2pHost) Start() error {
	return b.host.Network().Listen()
}

func (b *BasicLibp2pHost) Ctx() context.Context {
	return b.ctx
}

// RecConnEvent
// Record Connection Event
// @param connEvent: the event to insert in the notification channel
func (b *BasicLibp2pHost) RecConnEvent(eventTrace *models.EventTrace) {
	b.connEventNotChannel <- eventTrace
}

func (b *BasicLibp2pHost) ConnEventNotChannel() chan *models.EventTrace {
	return b.connEventNotChannel
}

// RecIdentEvent
// Record Identification Event
// @param identEvent: the event to insert in the notification channel
func (b *BasicLibp2pHost) RecIdentEvent(identEvent IdentificationEvent) {
	b.identNotChannel <- identEvent
}

func (b *BasicLibp2pHost) IdentEventNotChannel() chan IdentificationEvent {
	return b.identNotChannel
}
