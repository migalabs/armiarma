package ethereum

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	comm "github.com/migalabs/armiarma/pkg/networks/common"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	log "github.com/sirupsen/logrus"
)

const (
	RPCTimeout        time.Duration        = 20 * time.Second
	BeaconStatusRPC   comm.RPCRequestsName = "beacon-status"
	BeaconMetadataRPC comm.RPCRequestsName = "beacon-metadata"
)

type EthereumNetwork struct {
	ctx           context.Context
	LocalStatus   common.Status
	LocalMetadata common.MetaData

	// List of requestable functions that can be asked when
	// we stablish a connection to a node
	// RPCRequestables map[string]comm.RPCRequest
}

func NewEthereumNetwork(ctx context.Context, status common.Status, metadta common.MetaData) *EthereumNetwork {

	ethNet := &EthereumNetwork{
		ctx:           ctx,
		LocalStatus:   status,
		LocalMetadata: metadta,
		// RPCRequestables: make(map[string]comm.RPCRequest),
	}

	// ethNet.RPCRequestables["status_rpc"] = ethNet.ReqBeaconStatus
	// ethNet.RPCRequestables["metadata_rpc"] = ethNet.ReqBeaconMetadata

	return ethNet
}

func (en *EthereumNetwork) UpdateStatus(newStatus common.Status) {
	// check if the new one is newer than ours
	if newStatus.HeadSlot > en.LocalStatus.HeadSlot {
		en.LocalStatus = newStatus
	}
}

// Useless, we will never update our metadata
func (en *EthereumNetwork) UpdateMetadata() {
	// check if the new one is emtpy
	// check if the new one is newer than ours
	// Update ours
}

func (en *EthereumNetwork) NetworkType() utils.NetworkType {
	return utils.EthereumNetwork
}

func ComposeQuickBeaconStatus(forkDigest string) common.Status {
	frkDgst := new(common.ForkDigest)
	b, err := hex.DecodeString(strings.Trim(forkDigest, "0x"))
	if err != nil {
		log.Panic("unable to decode ForkDigest", err.Error())
	}
	frkDgst.UnmarshalText(b)

	return common.Status{
		ForkDigest:     *frkDgst,
		FinalizedRoot:  common.Root{},
		FinalizedEpoch: 0,
		HeadRoot:       common.Root{},
		HeadSlot:       0,
	}

}

func ComposeQuickBeaconMetaData() common.MetaData {
	attnets := new(common.AttnetBits)
	b, err := hex.DecodeString("ffffffffffffffff")
	if err != nil {
		log.Panic("unable to decode Attnets", err.Error())
	}
	attnets.UnmarshalText(b)

	return common.MetaData{
		SeqNumber: common.SeqNr(1),
		Attnets:   *attnets,
	}
}
