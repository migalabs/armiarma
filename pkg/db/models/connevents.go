package models

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

type ConnDirection int8

const (
	UnsetConnection ConnDirection = iota
	InboundConnection
	OutboundConnection
)

func DirectionIndexToString(connDir ConnDirection) string {
	var str string
	switch connDir {
	case InboundConnection:
		str = "inbound"
	case OutboundConnection:
		str = "outbound"
	default:
		str = "unset"
	}
	return str
}

// Based on the current logic of the crawler
// 1. Receive the Connection -> gen and save the ConnEvent with the Remore.Peer.ID and the direction
// 2. while we identify the peer or the metadata, the disconnection might come
// 3. persist the Metadata or the disconn to the prev added

//
type EventTrace struct {
	PeerID peer.ID
	Event  interface{}
}

// the struct of a connection and its info to a given peer
type ConnEvent struct {
	PeerID peer.ID

	ConnInfo
	EndConnInfo
}

type ConnInfo struct {
	Direction  ConnDirection // "inbound"/"outbound"
	ConnTime   time.Time
	Latency    time.Duration
	Identified bool
	Att        map[string]interface{}
	Error      string
}

type EndConnInfo struct {
	DiscTime     time.Time
	ConnDuration time.Duration
}

// Create a new connection event that will summarize the interaction with a given peer
func NewConnEvent(pID peer.ID) *ConnEvent {
	return &ConnEvent{
		PeerID: pID,
		ConnInfo: ConnInfo{
			Att: make(map[string]interface{}, 0),
		},
	}
}

// AddConnInfo includes the connection control info to the ConnEvent struct
func (c *ConnEvent) AddConnInfo(connInfo ConnInfo) {
	// update the missing values to the ConnEvent
	c.Direction = connInfo.Direction
	c.ConnTime = connInfo.ConnTime
	c.Latency = connInfo.Latency
	c.Identified = connInfo.Identified

	// filter in the Error to avoid overwriting important info
	// only write the error if it's none or err_requesting_metadata

	// ---- removed so far since our only edge case is connection triggered and followed by ContextDeadlineExceeded
	// if connInfo.Error == utils.NoneErr || connInfo.Error == utils.MetadataReqErr {
	// 	c.Error = connInfo.Error
	// }
	c.Error = connInfo.Error

	for k, v := range connInfo.Att {
		c.Att[k] = v
	}
	// check if there was already a DisconnectionEvent to calculate the Duration
	if c.DiscTime != (time.Time{}) {
		c.ConnDuration = c.DiscTime.Sub(c.ConnTime)
	}
}

// AddDisconn aggregates the disconnection time and precalculates the total duration time
func (c *ConnEvent) AddDisconn(discEv EndConnInfo) {
	// check if the ConnectionEvent has alredy a a connection
	if c.ConnTime != (time.Time{}) {
		// only calculate the duration if we have the connection time and the disconnection time (same for the connections)
		c.ConnDuration = discEv.DiscTime.Sub(c.ConnTime)
	}
	c.DiscTime = discEv.DiscTime
}

func (c *ConnEvent) IsReadyToPersist() bool {
	return (c.ConnTime != (time.Time{}) &&
		c.DiscTime != (time.Time{}) &&
		c.ConnDuration != time.Duration(uint64(0)))
}
