package ethereum

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/protolambda/zrnt/eth2/beacon/common"

	"github.com/migalabs/armiarma/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	RPCTimeout time.Duration = 20 * time.Second
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
	case ForkDigests[GnosisPhase0Key], ForkDigests[GnosisAltairKey], ForkDigests[GnosisBellatrixKey]:
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
func (en *LocalEthereumNode) SetForkDigest(forkDigest string) {
	// TODO: parse to see if it's a valid ForkDigets (len, blabla)
	en.addEntries(NewEth2DataEntry(forkDigest))
}

// SetAttNetworks adds any given set of Attnets into the local node's enr
func (en *LocalEthereumNode) SetAttNetworks(networks string) {
	// TODO: parse to see if it's a valid ForkDigets (len, blabla)
	en.addEntries(NewAttnetsENREntry(networks))
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
