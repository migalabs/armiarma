package models

import (
	"time"
)

type OldControlInfo struct {
	ConnectedDirection    []string  // The directions of each connection event.
	IsConnected           bool      // If the peer is connected (CheckIfRealConnect).
	Attempted             bool      // If the peer has been attempted to stablish a connection.
	Succeed               bool      // If the connection attempt (outbound) has been successful.
	Attempts              uint64    // Number of attempts done.
	Error                 []string  // List of errors (also adding "None" errors), aligned with connection events.
	LastErrorTimestamp    time.Time // Timestamp of the last error reported for this peer.
	Deprecated            bool      // Flag to rummarize whether the peer is still valid for statistics or not. If true, the peer is not exported CSV / metrics.
	LastIdentifyTimestamp time.Time // Timestamp of the last time the peer was identified (get user agent...).

	NegativeConnAttempts []time.Time // List of dates when the peer retrieved a negative connection attempt (outbound) (if there is a possitive one, clean the array).
	ConnectionTimes      []time.Time // List of registered connections events.
	DisconnectionTimes   []time.Time // List of Disconnection Events.
	MetadataRequest      bool        // If the peer has been attempted to request its metadata.
	MetadataSucceed      bool        // If the peer has been successfully requested its metadata.
}

func NewOldControlInfo() *OldControlInfo {
	return &OldControlInfo{
		ConnectedDirection:   make([]string, 0),
		Error:                make([]string, 0),
		NegativeConnAttempts: make([]time.Time, 0),
		ConnectionTimes:      make([]time.Time, 0),
		DisconnectionTimes:   make([]time.Time, 0),
	}
}
