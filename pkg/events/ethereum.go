package events

import (
	"time"

	"github.com/protolambda/zrnt/eth2/beacon/phase0"
)

// EthereumAttestation contains the data for an Ethereum Attestation that was received
type EthereumAttestation struct {
	Attestation *phase0.Attestation `json:"attestation"`
}

// TimedEthereumAttestation contains extra data for an Ethereum Attestation
type TimedEthereumAttestation struct {
	Attestation          *phase0.Attestation   `json:"attestation"`
	AttestationExtraData *AttestationExtraData `json:"attestation_extra_data"`
	PeerInfo             *PeerInfo             `json:"peer_info"`
}

// PeerInfo contains information about a peer
type PeerInfo struct {
	ID              string        `json:"id"`
	IP              string        `json:"ip"`
	Port            int           `json:"port"`
	UserAgent       string        `json:"user_agent"`
	Latency         time.Duration `json:"latency"`
	Protocols       []string      `json:"protocols"`
	ProtocolVersion string        `json:"protocol_version"`
}

// AttestationExtraData contains extra data for an attestation
type AttestationExtraData struct {
	ArrivedAt  time.Time     `json:"arrived_at"`
	P2PMsgID   string        `json:"peer_msg_id"`
	Subnet     int           `json:"subnet"`
	TimeInSlot time.Duration `json:"time_in_slot"`
}
