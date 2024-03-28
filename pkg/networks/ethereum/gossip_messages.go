package ethereum

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
)

var (
	ErrorNoSubnet          = errors.New("no subnet found in topic")
	ErrorNotParsableSubnet = errors.New("not parseable subnet int")
)

// Tracked Message basis
type GossipSubTrackedMessage interface {
	ReceivedFrom() peer.ID
	MessageID() string
	ArrivalTime() time.Time
	Message() any // Not safe, but will work (should be marshaleable anyways)
}

type TrackedMessage struct {
	MsgID  string
	Sender peer.ID
	Time   time.Time
	Msg    any // Not safe, but will work (same as before)
}

func (m *TrackedMessage) ReceivedFrom() peer.ID {
	return m.Sender
}

func (m *TrackedMessage) MessageID() string {
	return m.MsgID
}

func (m *TrackedMessage) ArrivalTime() peer.ID {
	return m.Sender
}

func (m *TrackedMessage) Message() any { // any = Marshaleable
	return m.Msg
}


// Dummy message (For not yet ready topics)
type DummyMessage struct {
	TrackedMessage
}

func (m *DummyMessage) IsZero() bool {
	return true
}

// Ethereum Message-Specifics
// Beacon Block
type TrackedBeaconBlock struct {
	TrackedMessage
	TimeInSlot time.Duration // exact time inside the slot (range between 0secs and 12s*32slots)
	ValIndex   uint64
	Slot       uint64
}

func (a *TrackedBeaconBlock) IsZero() bool {
	return a.Slot == 0
}

// Attestations
type TrackedAttestation struct {
	TrackedMessage
	TimeInSlot time.Duration // exact time inside the slot (range between 0secs and 12s*32slots)
	Subnet     int
	ValPubkey  string
	Slot       uint64
}

func (a *TrackedAttestation) IsZero() bool {
	return a.Slot == 0
}

// Aggregations and Proofs
type TrackedAggregateAndProof struct {
	TrackedMessage
	TimeInSlot time.Duration // exact time inside the slot (range between 0secs and 12s*32slots)
	Slot       uint64
}

func (a *TrackedAggregateAndProof) IsZero() bool {
	return a.Slot == 0
}

// Voluntarý Exits
type TrackedVoluntaryExit struct {
	TrackedMessage
	Epoch    uint64
	ValIndex uint64
}

func (a *TrackedVoluntaryExit) IsZero() bool {
	return a.Epoch == 0
}

// Propose Slashing
type TrackedProposerSlashing struct {
	TrackedMessage
	Slot          uint64
	ProposerIndex uint64
}

func (a *TrackedProposerSlashing) IsZero() bool {
	return a.Slot == 0
}

// Attester Slashing
type TrackedAttesterSlashing struct {
	TrackedMessage
	Epoch    uint64
	ValIndex uint64
}

func (a *TrackedAttesterSlashing) IsZero() bool {
	return a.Epoch == 0
}

// SyncAggregations: https://github.com/protolambda/zrnt/blob/6bc42739f502a06171cc6f2378ec7aa556e41182/eth2/beacon/altair/sync_contribution.go#L14
type TrackedSyncAggregate struct {
	TrackedMessage
	AggragatorIndex uint64
	TimeInSlot      time.Duration
	Slot            uint64
}

func (a *TrackedSyncAggregate) IsZero() bool {
	return a.Slot == 0
}

// SyncVote
type TrackedSyncMessage struct {
	TrackedMessage
	ValIndex   uint64
	TimeInSlot time.Duration
	Slot       uint64
}

func (a *TrackedSyncMessage) IsZero() bool {
	return a.Slot == 0
}

// BLS_Changes (TODO)

// blobs (TODO: - zrnt doesn't include the blob struct, still looking for the time to implement the entire structure, the SSZ serialization, the view, the tree hashing, etc)
// Experimental using the Eth2Clients library from attestant -> https://github.com/attestantio/go-eth2-client/blob/2d68bcd60d23ca11bbf073332f86a15b83b7a265/spec/deneb/blobsidecar.go#L24
type TrackedBlobSidecards struct {
	TrackedMessage
	BlobIndex uint64
	BeaconBlockRoot string
}

func (a *TrackedBlobSidecards) IsZero() bool {
	return a.BeaconBlockRoot != ""
}