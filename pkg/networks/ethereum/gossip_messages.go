package ethereum

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

type TrackedAttestation struct {
	ArrivalTime time.Time
	Sender      peer.ID
	Pubkey      string
	Slot        int64
}

func (a *TrackedAttestation) IsZero() bool {
	return a.Slot == 0
}
