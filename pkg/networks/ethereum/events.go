package ethereum

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
)

type AttestationReceievedEvent struct {
	Attestation        *phase0.Attestation
	TrackedAttestation *TrackedAttestation
	PeerID             peer.ID
}
