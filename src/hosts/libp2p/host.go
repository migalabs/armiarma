package libp2p

import (
	"context"
	"net"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
)

// Struct that defines the Basic Struct asociated to the Libtp2p host
type BasicLibp2pHost struct {
	// Basic host variables
	ctx       context.Context
	ctxCancel context.CancelFunc

	// loggrus related log printer
	Log *logrus.Logger
	// Basic sevices related with the libp2p host
	host host.Host

	userAgent string

	privKey crypto.Secp256k1PrivateKey
}

type BasicLibp2pHostOpts struct {
	IP        string
	TCP       string
	UDP       string
	UserAgent string
	PrivKey   string
	LogLvl    string
	// TODO: -Add IdService for the libp2p host
	//       -Aggregate more data about the log format
	//       -Include info regarding the
}

// Generate a new Libp2p host from the given context and Options
func NewBasicLibp2pHost(ctx context.Context, opts BasicLibp2pHost) (BasicLibp2pHost, error) {
	// Link the host context with the app main context
	hostCtx, hostCancel := context.WithCancel(ctx)

	// Generate the logrus logger related to the Libp2p host
	log := logrus.WithField(log.Fields{"module": "libp2pHost"})

	// check the parsed host options
	ip, err := net.ParseIP(opts.IP)
	if err != nil {
		log.Debugf("s% - IP: s%, setting default IP: %s", err, ip.String(), hosts.DefaultIP)
		// If the parsed IP is wrong/empty, simply the default one "0.0.0.0"
		ip, _ := net.ParseIP(hosts.DefaultIP)
	}
	tcp := opts.TCP
	if tcp == "" {
		log.Debugf("")
		tcp = hosts.DefaultTCP
	}
	udp := opts.UDP
	if udp == "" {
		udp = hosts.DefaultUDP
	}
	useragent = opts.UserAgent
	if useragent == "" {
		useragent = hosts.DefaulUserAgent
	}
	// parse the privKey of the host
	// check if the received string contains anything, and check if

	// Generate the main Libp2p host that will be exposed to the network
	host, err := libp2p.New(
		hostCtx,
		libp2p.ListenAddrs(),
	)

	// Gererate the struct that contains all the configuration and structs surrounding the Libp2p Host
	b := BasicLibp2pHost{
		ctx:       hostCtx,
		ctxCancel: hostCancel,
	}

}

func (b *BasicLibp2pHost) Start() {

}

func (b *BasicLibp2pHost) Stop() {

}
