package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/control/actor/flags"
	"github.com/protolambda/rumor/control/actor/peer/metadata"
	"github.com/protolambda/rumor/p2p/rpc/reqresp"
	"github.com/protolambda/rumor/p2p/track"
)

var timeout time.Duration = 5 * time.Second

func PollPeerMetadata(p peer.ID, base *base.Base, peerMetadataState *metadata.PeerMetadataState, store track.ExtendedPeerstore, gm *GossipMetrics) {
	start := time.Now()
	fmt.Println(start, "- Requesting Metadata to peer:", p)

	// apply timeout to each poll target in this round
	reqCtx, _ := context.WithTimeout(context.Background(), timeout)

	go func(peerID peer.ID) {
		pingCmd := &metadata.PeerMetadataPingCmd{
			Base:              base,
			PeerMetadataState: peerMetadataState,
			Store:             store,
			Timeout:           timeout,
			Compression:       flags.CompressionFlag{Compression: reqresp.SnappyCompression{}},
			Update:            true,
			ForceUpdate:       true,
			UpdateTimeout:     timeout,
			MaxTries:          uint64(2),
			PeerID:            flags.PeerIDFlag{PeerID: peerID},
		}
		if err := pingCmd.Run(reqCtx); err != nil {
			gm.AddMetadataEvent(p, false)
			fmt.Println("failed to poll peer", p)
		} else {
			gm.AddMetadataEvent(p, true)
			fmt.Println("successfully poll peer", p)
		}
	}(p)
}
