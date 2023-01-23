package models

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/utils"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

type RemoteHostOptions func(*HostInfo) error

const (
	DeprecableTime = 24 * time.Hour
	// if in 2 months we didn't connect the peer,
	// say that it left the network
	// unless discv5 says the opposite
	LeftNetworkTime = 24 * time.Hour * 60
)

// HostInfo is the basic struct that contains all the information needed to connect, identify and monitor a Peer
type HostInfo struct {
	sync.RWMutex

	// AddrInfo
	ID     peer.ID
	IP     string
	UDP    int
	TCP    int
	MAddrs []ma.Multiaddr

	// network
	Network utils.NetworkType

	// Indetification
	PeerInfo PeerInfo

	// Control Info
	ControlInfo ControlInfo

	Attr map[string]interface{}
}

// NewHostInfo returns a new structure of the PeerInfo field for the specific network passed as argk
func NewHostInfo(peerID peer.ID, network utils.NetworkType, opts ...RemoteHostOptions) *HostInfo {
	hInfo := &HostInfo{
		ID:      peerID,
		MAddrs:  make([]ma.Multiaddr, 0),
		Network: network,
		Attr:    make(map[string]interface{}),
	}

	// apply all the Options
	for _, opt := range opts {
		err := opt(hInfo)
		if err != nil {
			log.Error("unable to init HostInfo with folling Option", opt)
		}
	}

	return hInfo
}

// HostInfo Options

func WithIPAndPorts(ip string, tcp, udp int) RemoteHostOptions {
	return func(h *HostInfo) error {
		h.Lock()
		defer h.Unlock()

		h.IP = ip
		h.UDP = udp
		h.TCP = tcp

		// Compose Muliaddress from data
		mAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, tcp))
		if err != nil {
			return err
		}
		// add single address to the HostInfo
		h.MAddrs = append(h.MAddrs, mAddr)
		return nil
	}
}

func WithMultiaddress(mAddrs []ma.Multiaddr) RemoteHostOptions {
	return func(h *HostInfo) error {
		h.Lock()
		defer h.Unlock()

		for _, a := range mAddrs {
			h.MAddrs = append(h.MAddrs, a)
		}
		// TODO: Extract public IP, tcp and udp from ma
		return nil
	}
}

// ComposeAddrsInfo returns the PeerId and Multiaddres in the peer.AddrsInfo format
// Essential for libp2p.Connect() operation
func (h HostInfo) ComposeAddrsInfo() peer.AddrInfo {
	h.RLock()
	defer h.RUnlock()

	// generate new AddrInfo struct
	addrInfo := peer.AddrInfo{
		ID:    h.ID,
		Addrs: make([]ma.Multiaddr, 0),
	}
	// append the MAddrs
	addrInfo.Addrs = h.MAddrs

	return addrInfo
}

func (h *HostInfo) AddAtt(key string, attr interface{}) {
	h.Lock()
	defer h.Unlock()

	h.Attr[key] = attr
}

func (h *HostInfo) IdentifyHost(pInfo *PeerInfo) {
	h.Lock()
	defer h.Unlock()
	h.PeerInfo = *pInfo
}

func (h *HostInfo) IsHostIdentified() bool {
	return h.PeerInfo.IsPeerIdentified()
}

// PeerInfo contains all the info that can be extracted from the Libp2p.IDService
type PeerInfo struct {
	// Indetification
	RemotePeer      peer.ID
	UserAgent       string
	ProtocolVersion string
	Protocols       []string
	Latency         time.Duration
}

func NewEmptyPeerInfo() *PeerInfo {
	return &PeerInfo{
		Protocols: make([]string, 0),
	}
}

// IdentifyHost updates if the fileds are not empty the fields that identify the peer in the network
func NewPeerInfo(remotePeer peer.ID, userAgent, protocolVersion string, protocols []string, latency time.Duration) *PeerInfo {
	pInfo := &PeerInfo{
		RemotePeer:      remotePeer,
		UserAgent:       userAgent,
		ProtocolVersion: protocolVersion,
		Protocols:       make([]string, 0),
		Latency:         latency,
	}

	for _, protocol := range protocols {
		pInfo.Protocols = append(pInfo.Protocols, protocol)
	}

	return pInfo
}

// IsHostIdentified checks if the Peer was already identified before
func (p *PeerInfo) IsPeerIdentified() bool {
	return p.UserAgent != "" || p.ProtocolVersion != "" || len(p.Protocols) > 0
}

type ControlInfo struct {
	RemotePeer peer.ID

	// major variables
	Deprecated  bool
	LeftNetwork bool

	// control timestamps
	Attempted       bool
	LastActivity    time.Time
	LastConnAttempt time.Time
	LastError       string
}

func NewControlInfo() *ControlInfo {
	return &ControlInfo{
		LastError: "",
	}
}
