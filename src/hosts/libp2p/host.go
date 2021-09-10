package libp2p

import (
    "context"
    "net"
    "time"
    "error"
    "github.com/multiformats/go-multiaddr"
    "github.com/libp2p/go-libp2p-core/crypto"
    "github.com/libp2p/go-libp2p-core/network"

)

// Struct that defines the Basic Struct asociated to the Libtp2p host
type BasicLibp2pHost struct {
    // Basic host variables
    ctx context.Context
    ctxCancel context.CancelFunc
    
    // loggrus related log printer
    Log *logrus.Logger
    // Basic sevices related with the libp2p host
    host host.Host

    // Basic Host Metadata
    ip net.IP
    tcp string
    udp string
    multiAddr ma.Multiaddr
    
    userAgent string

    privKey  crypto.PrivKey
}

type BasicLibp2pHostOpts struct {
    IP string
    TCP string
    UDP string
    UserAgent string
    PrivKey crypto.Secp256k1PrivateKey
    LogLvl string
    // TODO: -Add IdService for the libp2p host
    //       -Aggregate more data about the log format
    //       -Include info regarding the 
}

// Generate a new Libp2p host from the given context and Options
func NewBasicLibp2pHost(ctx context.Context, opts BasicLibp2pHost) (BasicLibp2pHost, error){
    // Link the host context with the app main context
    hostCtx, hostCancel := context.WithCancel(ctx)
    
    // Generate the logrus logger related to the Libp2p host
    log := logrus.WithField(log.Fields{"module": "libp2pHost"})

    // check the parsed host options
    ip, err := net.ParseIP(opts.IP)
    if err != nil {
        log.Errorf("s% - IP: s%", err, ip.String())
        // If the parsed IP is wrong/empty, simply use 0.0.0.0 as default
        ip, _ := net.ParseIP(hosts.DefaultIP)
    }
    tcp := opts.TCP
    if tcp == "" {
          tcp = hosts.DefaultTCP
    }
    udp := opts.UDP
    if udp == "" {
        udp = hosts.DefaultUDP
    }
    useragent = opts.UserAgent
    if useragent == "" {
        useragent = hosts.UserAgent
    }
    // parse the privKey of the host
    




    // Generate the main Libp2p host that will be exposed to the network
    host, err := libp2p.New(
        hostCtx,
        libp2p.ListenAddrs(),
        
    )
    
    // Gererate the struct that contains all the configuration and structs surrounding the Libp2p Host
    b := BasicLibp2pHost {
        ctx: hostCtx,
        ctxCancel: hostCancel,


    }

}


func (b *BasicLibp2pHost) Start() {
    
}

func (b *BasicLibp2pHost) Stop() {
    
}
