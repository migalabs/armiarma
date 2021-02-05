package gossip

import (
    "time"

    "github.com/protolambda/zrnt/eth2/beacon"
    "github.com/libp2p/go-libp2p-core/peer"
)


// Attestation Message Metadata Struct
type ReceivedAttestation struct {
    ReceivedTime    time.Time
    From            peer.ID
    BeaconAttestation   beacon.Attestation
}

// Proposer Slashing Message Metadata Struct
type ReceivedProposerSlashing struct {
    ReceivedTime    time.Time
    From            peer.ID
    ProposerSlashing    beacon.ProposerSlashing
}

// Attester Slashing Message Metadata Struct
type ReceivedAttesterSlashing struct {
    ReceivedTime    time.Time
    From            peer.ID
    AttesterSlashing    beacon.AttesterSlashing
}

// Voluntary Exit Message Metadata Struct
type ReceivedVoluntaryExit struct {
    ReceivedTime    time.Time
    From            peer.ID
    VoluntaryExit   beacon.VoluntaryExit
}






