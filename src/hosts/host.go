package libp2p

import (
	"context"
	"net"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/p2p/protocols/identify"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/src/base"
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
	fullMultiAddr string

	userAgent string
	peerID    peer.ID

	privKey crypto.PrivKey
}

type BasicLibp2pHostOpts struct {
	IP        string
	TCP       string
	UDP       string
	UserAgent string
	PrivKey   crypto.Secp256k1PrivateKey
	LogOpts   base.LogOpts
	// TODO: -Add IdService for the libp2p host
}

// Generate a new Libp2p host from the given context and Options
// TODO: missing argument for app info (givin Privkeys, IPs, ports, userAgents)
func NewBasicLibp2pHost(ctx context.Context, opts BasicLibp2pHostOpts) (*BasicLibp2pHost, error) {
	// Generate Base module struct with basic funtioning
	b := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(base.logOpts{
			ModName:   opts.LogOpts.ModName,
			Output:    opts.LogOpts.Optput,
			Formatter: opts.LogOpts.Formatter,
			Level:     opts.LogOpts.Level,
		}),
	)
	// check the parsed host options
	ip, err := net.ParseIP(opts.IP)
	if err != nil {
		b.Log.Debugf("s% - IP: s%, setting default IP: %s", err, ip.String(), hosts.DefaultIP)
		// If the parsed IP is wrong/empty, simply the default one "0.0.0.0"
		ip, _ := net.ParseIP(hosts.DefaultIP)
	}
	tcp := opts.TCP
	if tcp == "" {
		b.Log.Debugf("empty tcp given, setting tcp port to default to %s", hosts.DefaultTCP)
		tcp = hosts.DefaultTCP
	}
	udp := opts.UDP
	if udp == "" {
		b.Log.Debugf("empty udp given, setting udp port to default to %s", hosts.DefaultUDP)
		udp = hosts.DefaultUDP
	}
	useragent = opts.UserAgent
	if useragent == "" {
		b.Log.Debugf("empty user-agent given, setting user-agent port to default to %s", hosts.DefaultUserAgent)
		useragent = hosts.DefaulUserAgent
	}
	// generate de multiaddress
	multiaddr := sprintf("/ip4/%s/tcp/%s", ip, tcp)
	muladdr := ma.NewMultiaddr(multiaddr)
	b.Log.Debugf("setting multiaddres to %s", muladdr)
	// parse the privKey of the host
	privk := opts.PrivKey
	if len(privk) == "" {
		// TODO: if empty privkey was given, generate new one and export it for later userAgent
		//		Perhaps this key, generation parsing, should be done before
		b.Log.Debugf("empty privkey given, generating new one: %s", privk)
	}
	// Generate the main Libp2p host that will be exposed to the network
	host, err := libp2p.New(
		b.Ctx(),
		libp2p.ListenAddrs(muladdr),
		libp2p.Identify(privk),
		libp2p.UserAgent(useragent),
	)
	if err != nil {
		return nil, err
	}
	peerId := host.ID().String()
	fmaddr := h.Addrs()[0].String() + "/p2p/" + h.ID().String()

	// Gererate the struct that contains all the configuration and structs surrounding the Libp2p Host
	basicHost := &BasicLibp2pHost{
		Base:          b,
		host:          host,
		identify:      host.IDService(),
		ip:            ip,
		tcp:           tcp,
		udp:           udp,
		multiAddr:     muladdr,
		fullMultiAddr: fmaddr,
		peerID:        peerId,
		userAgent:     useragent,
		privKey:       privk,
	}
	b.Log.Debug("setting custom notification functions")
	basicHost.SetCustomNotifications()

	return basicHost, nil
}

// start the libp2pHost module (perhaps NOT NECESARY)
// So far, start listening on the fullMultiAddrs
func (b *BasicLibp2pHost) Start() {
	err := b.host.Network().Listen(b.fullMultiAddr)
	if err != nil {
		b.Log.Errorf("error starting. %s", err)
	} else {
		b.Log.Infof("start listening at %s", b.fullMultiAddr)
	}
}

// cancel the context of the libp2pHost module
func (b *BasicLibp2pHost) Stop() {
	b.Log.Info("stopping Libp2p host")
	return b.Cancel()
}
