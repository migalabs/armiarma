package models

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/utils"
	ma "github.com/multiformats/go-multiaddr"
)

// PeerInfo is the basic struct that contains all the information needed to connect, identify and monitor a Peer
type PeerInfo struct {
	sync.RWMutex

	// AddrInfo
	ID     peer.ID
	MAddrs []ma.Multiaddr

	// Indetification
	UserAgent       string
	ProtocolVersion string
	Protocols       []string
	Latency         time.Duration

	// network
	network utils.P2pNetwork

	// IP related info
	IpInfo IpInfo

	// Networking
	ControlInfo *ControlInfo

	// DB Info
	lastExport int64 // Timestamp in seconds of the last exported time (backup for when we are loading the Peer).

	// Node NetworkNode // specific data for the peer in each of the networks
	// Attributes
	// Application layer / extra attributes
	Att map[string]interface{}
}

// NewPeerInfo returns a new structure of the PeerInfo field for the specific network passed as argk
func NewPeerInfo(peerID peer.ID, network utils.P2pNetwork) *PeerInfo {
	return &PeerInfo{
		network:     network,
		ID:          peerID,
		MAddrs:      make([]ma.Multiaddr, 0),
		ControlInfo: NewControlInfo(),
		Att:         make(map[string]interface{}),
	}
}

// ComposeAddrsInfo returns the PeerId and Multiaddres in the peer.AddrsInfo format
// Essential for libp2p.Connect() operation
func (p PeerInfo) ComposeAddrsInfo() peer.AddrInfo {
	p.RLock()
	defer p.RUnlock()

	// generate new AddrInfo struct
	addrInfo := peer.AddrInfo{
		ID:    p.ID,
		Addrs: make([]ma.Multiaddr, 0),
	}
	// append the MAddrs
	addrInfo.Addrs = p.MAddrs

	return addrInfo
}

// return network of a peer (unable to change it)
func (p PeerInfo) Network() utils.P2pNetwork {
	p.RLock()
	defer p.RUnlock()

	return p.network
}

// IdentifyHost updates if the fileds are not empty the fields that identify the peer in the network
func (p *PeerInfo) IndentifyHost(userAgent, protocolVersion string, protocols []string, latency time.Duration) {
	p.Lock()
	defer p.Unlock()

	// update the host indentification if the host was already identified
	// or if the new data is not empty (asuming that we can update the data)
	if !p.IsHostIdentified() || (userAgent != "" && protocolVersion != "" && len(protocols) > 0) {
		p.UserAgent = userAgent
		p.ProtocolVersion = protocolVersion
		p.Protocols = protocols
		p.Latency = latency
	}
}

// IsHostIdentified checks if the Peer was already identified before
func (p *PeerInfo) IsHostIdentified() bool {
	p.RLock()
	defer p.RUnlock()

	return p.UserAgent != "" || p.ProtocolVersion != "" || len(p.Protocols) > 0
}

// UpdateExportTime updates the time when we updated the time
func (p *PeerInfo) UpdateExportTime(t time.Time) {
	p.Lock()
	defer p.Unlock()

	if t == (time.Time{}) {
		t = time.Now()
	}
	p.lastExport = t.Unix()
}

// LastExportTime returns the last time we exported the Peer to the DB
func (p PeerInfo) LastExportTime() int64 {
	p.RLock()
	defer p.RUnlock()

	return p.lastExport
}

// AddAtt adds a new attribute to the peer
func (p *PeerInfo) AddAttr(key string, value interface{}) {
	p.Lock()
	defer p.Unlock()

	p.Att[key] = value
}

// ReadAtt reads (if it exists) the attribute for the given key
func (p *PeerInfo) ReadAttr(key string) (interface{}, bool) {
	p.RLock()
	defer p.RUnlock()

	value, ok := p.Att[key]
	return value, ok
}

// TODO:
// 		- missing Network related stuff
// 		- ControlInfo related stuff
// 		- Attributes
