package hosts

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/info"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"

	ma "github.com/multiformats/go-multiaddr"
)

// Struct that defines the Basic Struct asociated to the Libtp2p host
type BasicLibp2pHost struct {
	*base.Base
	// Basic sevices related with the libp2p host
	host     host.Host
	identify *identify.IDService

	// Basic Host Metadata
	ip            net.IP
	tcp           string
	udp           string
	multiAddr     ma.Multiaddr
	fullMultiAddr ma.Multiaddr

	userAgent string
	peerID    peer.ID

	privKey crypto.PrivKey
}

type BasicLibp2pHostOpts struct {
	Info_obj info.InfoData
	LogOpts  base.LogOpts
	// TODO: -Add IdService for the libp2p host
}

// Generate a new Libp2p host from the given context and Options
// TODO: missing argument for app info (givin Privkeys, IPs, ports, userAgents)
func NewBasicLibp2pHost(ctx context.Context, opts BasicLibp2pHostOpts) (*BasicLibp2pHost, error) {
	// Generate Base module struct with basic funtioning
	b, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(base.LogOpts{
			ModName:   opts.LogOpts.ModName,
			Output:    opts.LogOpts.Output,
			Formatter: opts.LogOpts.Formatter,
			Level:     opts.LogOpts.Level,
		}),
	)
	if err != nil {
		log.Panicf("couldn't create base for host module. %s", err)
	}
	// check the parsed host options

	ip := opts.Info_obj.GetIPToString()
	tcp := opts.Info_obj.GetIPToString()

	// generate de multiaddress
	multiaddr := fmt.Sprintf("/ip4/%s/tcp/%s", ip, tcp)
	muladdr, err := ma.NewMultiaddr(multiaddr)
	if err != nil {
		b.Log.Debugf("couldn't generate multiaddress from ip %s and tcp %s", ip, tcp)
		multiaddr = fmt.Sprintf("/ip4/%s/tcp/%s", DefaultIP, DefaultTCP)
		muladdr, _ = ma.NewMultiaddr(multiaddr)
	}
	b.Log.Debugf("setting multiaddres to %s", muladdr)

	privkey := opts.Info_obj.GetPrivKey()

	// Generate the main Libp2p host that will be exposed to the network
	host, err := libp2p.New(
		b.Ctx(),
		libp2p.ListenAddrs(muladdr),
		libp2p.Identity(privkey),
		libp2p.UserAgent(opts.Info_obj.GetUserAgent()),
	)
	if err != nil {
		return nil, err
	}
	peerId := host.ID().String()
	fmaddr := host.Addrs()[0].String() + "/p2p/" + host.ID().String()
	localMultiaddr, _ := ma.NewMultiaddr(fmaddr)
	b.Log.Debugf("full multiaddress %s", localMultiaddr)
	// generate the identify service
	ids, err := identify.NewIDService(host, identify.UserAgent(opts.Info_obj.GetUserAgent()), identify.DisableSignedPeerRecord())
	if err != nil {
		b.Log.Error(err)
	}
	// Gererate the struct that contains all the configuration and structs surrounding the Libp2p Host
	basicHost := &BasicLibp2pHost{
		Base:          b,
		host:          host,
		identify:      ids,
		ip:            opts.Info_obj.GetIP(),
		tcp:           opts.Info_obj.GetTcpPortString(),
		udp:           opts.Info_obj.GetUdpPortString(),
		multiAddr:     muladdr,
		fullMultiAddr: localMultiaddr,
		peerID:        peer.ID(peerId),
		userAgent:     opts.Info_obj.GetUserAgent(),
		privKey:       opts.Info_obj.GetPrivKey(),
	}
	b.Log.Debug("setting custom notification functions")
	basicHost.SetCustomNotifications()

	return basicHost, nil
}

// return the libp2p host from the host Module
func (b *BasicLibp2pHost) Host() host.Host {
	return b.host
}

// start the libp2pHost module (perhaps NOT NECESARY)
// So far, start listening on the fullMultiAddrs
func (b *BasicLibp2pHost) Start() {
	err := b.host.Network().Listen()
	if err != nil {
		b.Log.Errorf("error starting. %s", err)
	} else {
		b.Log.Infof("libp2p host module started")
	}
}

// cancel the context of the libp2pHost module
func (b *BasicLibp2pHost) Stop() {
	b.Log.Info("stopping Libp2p host")
	b.Cancel()
}
