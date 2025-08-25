package ethereum

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"time"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/protolambda/zrnt/eth2/beacon/common"

	"github.com/migalabs/armiarma/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	RPCTimeout time.Duration = 20 * time.Second
	ETH2_DATA_HEX_LENGTH = 32 // 16 bytes = 32 hex chars
	FORK_DIGEST_HEX_LENGTH = 8 // 4 bytes = 8 hex chars
	ATTNETS_HEX_LENGTH = 16 // 8 bytes = 16 hex chars
)

type LocalEthereumNode struct {
	ctx context.Context
	// ethereum node
	ethNode *enode.LocalNode
	// node's metadata in the network
	LocalStatus   common.Status
	LocalMetadata common.MetaData
	// Network Details
	networkGenesis time.Time
}

// NewLocalNode will create a LocalNode object using the given arguments.
func NewLocalEthereumNode(
	ctx context.Context,
	privKey *ecdsa.PrivateKey,
	status common.Status,
	matadata common.MetaData,
	forkDigest string) *LocalEthereumNode {

	// db where to store the ENRs
	ethDB, err := enode.OpenDB("")
	if err != nil {
		log.Panicf("Could not create local DB %s", err)
	}
	log.Infof("Creating Local Node")

	// select network based on the network that we are participating in
	var genesis time.Time
	switch forkDigest {
	// Mainnet
	case ForkDigests[Phase0Key], ForkDigests[AltairKey], ForkDigests[BellatrixKey]:
		genesis = MainnetGenesis
	// Prater
	case ForkDigests[PraterPhase0Key], ForkDigests[PraterBellatrixKey]:
		genesis = GoerliGenesis
	// Gnosis
	case ForkDigests[GnosisPhase0Key], ForkDigests[GnosisBellatrixKey]:
		genesis = GnosisGenesis
	// Mainnet
	default:
		genesis = MainnetGenesis
	}

	return &LocalEthereumNode{
		ctx:            ctx,
		ethNode:        enode.NewLocalNode(ethDB, privKey),
		networkGenesis: genesis,
	}
}

func (en *LocalEthereumNode) GetNetworkGenesis() time.Time {
	return en.networkGenesis
}

func (en *LocalEthereumNode) UpdateStatus(newStatus common.Status) {
	// check if the new one is newer than ours
	if newStatus.HeadSlot > en.LocalStatus.HeadSlot {
		en.LocalStatus = newStatus
	}
}

// Useless, we will never update our metadata
func (en *LocalEthereumNode) UpdateMetadata() {
	// check if the new one is emtpy
	// check if the new one is newer than ours
}

func (en *LocalEthereumNode) Network() utils.NetworkType {
	return utils.EthereumNetwork
}

// SetForkDigest adds any given ForkDigest into the local node's enr
func (en *LocalEthereumNode) SetForkDigest(forkDigest string) error {
	// Normalize the input
	forkDigest = strings.TrimSpace(forkDigest)
	
	// If only the 4-byte fork digest is provided, we need the full eth2 data
	if len(strings.TrimPrefix(forkDigest, "0x")) == FORK_DIGEST_HEX_LENGTH {
		log.Warnf("Fork digest is only 4 bytes: %s", forkDigest)
		return fmt.Errorf("incomplete fork digest: need full 16-byte eth2 data, got only 4-byte fork digest")
	}
	
	// Validate length (should be 32 hex chars for 16 bytes)
	hexStr := strings.TrimPrefix(forkDigest, "0x")
	if len(hexStr) != ETH2_DATA_HEX_LENGTH {
		log.Warnf("Non-standard fork digest length: expected %d hex chars, got %d", ETH2_DATA_HEX_LENGTH, len(hexStr))
		return fmt.Errorf("invalid eth2 data length: expected %d hex chars, got %d", ETH2_DATA_HEX_LENGTH, len(hexStr))
	}
	
	// Validate hex format
	if _, err := hex.DecodeString(hexStr); err != nil {
		log.Warnf("Invalid hex format for fork digest: %v", err)
		return fmt.Errorf("invalid hex format: %v", err)
	}
	
	// Create the entry with validation
	entry, err := NewEth2DataEntry(forkDigest)
	if err != nil {
		return fmt.Errorf("failed to create eth2 data entry: %v", err)
	}
	
	// Validate that the entry can be deserialized
	if _, err := entry.Eth2Data(); err != nil {
		return fmt.Errorf("invalid eth2 data format: %v", err)
	}
	
	// Add to ENR
	en.addEntries(entry)
	log.Debugf("Set fork digest in ENR: %s", forkDigest)
	
	return nil
}

// SetAttNetworks adds any given set of Attnets into the local node's enr
func (en *LocalEthereumNode) SetAttNetworks(networks string) error {
	// Normalize the input
	networks = strings.TrimSpace(networks)
	hexStr := strings.TrimPrefix(networks, "0x")
	
	// Validate length (typically 16 hex chars for 8 bytes)
	if len(hexStr) == 0 {
		return fmt.Errorf("empty attestation networks")
	}
	
	// Allow variable length but warn if not standard
	if len(hexStr) != ATTNETS_HEX_LENGTH {
		log.Warnf("Non-standard attnets length: expected %d hex chars, got %d", 
			ATTNETS_HEX_LENGTH, len(hexStr))
	}
	
	// Validate hex format
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return fmt.Errorf("invalid hex format: %v", err)
	}
	
	// Create the entry with validation
	entry, err := NewAttnetsENREntry(networks)
	if err != nil {
		return fmt.Errorf("failed to create attnets entry: %v", err)
	}
	
	// Add to ENR
	en.addEntries(entry)
	
	// Log the attestation subnet participation
	bitCount := CountBits(decoded)
	log.Debugf("Set attestation networks in ENR: %s (%d subnets)", networks, bitCount)
	
	return nil
}

// AddEntries modifies the local Ethereum Node's ENR adding a new entry to the Key-Value
func (en *LocalEthereumNode) addEntries(entry enr.Entry) {
	en.ethNode.Set(entry)
}

func (en *LocalEthereumNode) EthNode() *enode.LocalNode {
	return en.ethNode
}

func ComposeQuickBeaconStatus(forkDigest string) common.Status {
	frkDgst := new(common.ForkDigest)
	frkDgst.UnmarshalText([]byte(forkDigest))

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
