package host

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/control/actor/flags"
	"github.com/protolambda/rumor/control/actor/peer/metadata"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/rpc/reqresp"
	"github.com/protolambda/rumor/p2p/track"
	log "github.com/sirupsen/logrus"
	"github.com/pkg/errors"
)

var timeout time.Duration = 5 * time.Second

// Polls metadata for a given peer and return its metadata if ok
func PollPeerMetadata(p peer.ID, base *base.Base, peerMetadataState *metadata.PeerMetadataState, store track.ExtendedPeerstore, gm *metrics.PeerStore) (*track.PeerAllData, error) {
	// apply timeout to each poll target in this round
	reqCtx, _ := context.WithTimeout(context.Background(), timeout)

	log.Info("Requesting metadata for peer: ", p.String())

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
		PeerID:            flags.PeerIDFlag{PeerID: p},
	}
	// TODO: Rethink this
	err := pingCmd.Run(reqCtx)
	if err != nil {
		gm.AddMetadataEvent(p.String(), false)
	} else {
		gm.AddMetadataEvent(p.String(), true)
	}

	// TODO Naive solution. Iterates the store looking if we got the metadata
	// Note that we can be getting an old one?
	for _, peerId := range pingCmd.Store.Peers() {
		if peerId.String() == p.String() {
			peerData := pingCmd.Store.GetAllData(peerId)
			// TODO: Or another criteria for non empty metadata
			if peerData.UserAgent != "" {
				log.Info("Metadata not empty: ", p.String(), "UserAgent: ", peerData.UserAgent)
				return peerData, nil
			}
		}
	}

	return nil, errors.Wrap(err, "could not get metadata")
}
