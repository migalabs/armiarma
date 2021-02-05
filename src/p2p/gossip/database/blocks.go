package database

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/zrnt/eth2/beacon"
)

// BeaconBlock Message Metadata Struct
type ReceivedBeaconBlock struct {
	ReceivedTime    time.Time
	FromPeer        peer.ID
	BlockRoot       beacon.Root
	SignedBeaconBlock     beacon.SignedBeaconBlock
}

func (c *ReceivedBeaconBlock) Time() time.Time {
	return c.ReceivedTime
}

func (c *ReceivedBeaconBlock) From() peer.ID {
	return c.FromPeer
}

func (c *ReceivedBeaconBlock) ID() interface{} {
	return c.BlockRoot
}

func (c *ReceivedBeaconBlock) MessageContent() interface{} {
	return c.SignedBeaconBlock
}

func NewReceivedBeaconBlock(from peer.ID, block beacon.SignedBeaconBlock) *ReceivedBeaconBlock {
	rbb := &ReceivedBeaconBlock{
		ReceivedTime:   GetCurrentTime(),
		FromPeer:       from,
		BlockRoot:      block.Message.StateRoot,
		SignedBeaconBlock:    block,
	}
	return rbb
}
