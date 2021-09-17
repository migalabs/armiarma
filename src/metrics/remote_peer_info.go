package metrics

import (
	"time"

	"github.com/protolambda/zrnt/eth2/beacon"
)

// BASIC HOST INFO

// BasicHostInfo contains the basic Host info that will be requested from the identification of a libp2p peer
type BasicHostInfo struct {
	TimeStamp time.Time
	// Peer Host/Node Info
	PeerID          string
	NodeID          string
	UserAgent       string
	ProtocolVersion string
	Addrs           string
	PubKey          string
	RTT             time.Duration
	Protocols       []string
	// Information regarding the metadata exchange
	Direction string
	// Metadata requested
	MetadataRequest bool
	MetadataSucceed bool
}

// BEACON METADATA

// Basic BeaconMetadata struct that includes the timestamp of the received beacon metadata
type BeaconMetadataStamped struct {
	Timestamp time.Time
	Metadata  beacon.MetaData
}

// Generate a New BeaconMetadata struct and timestamp it
// NOTE: It will append empty metadatas timestamped to track when did we request them
func NewBMetadataStamped(meta beacon.MetaData) BeaconMetadataStamped {
	bmeta := BeaconMetadataStamped{
		Timestamp: time.Now(),
		Metadata:  meta,
	}
	return bmeta
}

// Funciton that returns de timestamp of the BeaconMetadata
func (b *BeaconMetadataStamped) Time() time.Time {
	return b.Timestamp
}

// Funciton that returns de content of the BeaconMetadata
func (b *BeaconMetadataStamped) Content() beacon.MetaData {
	return b.Metadata
}

// BEACON STATUS

//  Basic BeaconMetadata struct that includes The timestamp of the received beacon Status
type BeaconStatusStamped struct {
	Timestamp time.Time
	Status    beacon.Status
}

// Generate a new BeaconStatusStamp struct and timestamp it
// NOTE: It will append empty status timestamped to track when did we request them
func NewBStatusStamped(status beacon.Status) BeaconStatusStamped {
	bstatus := BeaconStatusStamped{
		Timestamp: time.Now(),
		Status:    status,
	}
	return bstatus
}

// Funciton that returns de timestamp of the BeaconMetadata
func (b *BeaconStatusStamped) Time() time.Time {
	return b.Timestamp
}

// Funciton that returns de content of the BeaconMetadata
func (b *BeaconStatusStamped) Content() beacon.Status {
	return b.Status
}
