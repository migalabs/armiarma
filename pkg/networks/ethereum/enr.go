package ethereum

import (
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"math/bits"
	"net"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/pkg/errors"

	gcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

var (
	EnrValidationError   error = errors.New("error validating ENR")
	Eth2DataParsingError error = errors.New("error parsing eth2 data")
)

var (
	EnrHostInfoAttribute string = "enr-info"
)

type EnrNode struct {
	Timestamp time.Time
	ID        enode.ID
	IP        net.IP
	Seq       uint64
	UDP       int
	TCP       int
	Pubkey    *ecdsa.PublicKey
	Eth2Data  *common.Eth2Data
	Attnets   *Attnets
}

func NewEnrNode(nodeID enode.ID) *EnrNode {

	return &EnrNode{
		Timestamp: time.Now(),
		ID:        nodeID,
		Pubkey:    new(ecdsa.PublicKey),
		Eth2Data:  new(common.Eth2Data),
		Attnets:   new(Attnets),
	}
}

// define the Handler for when we discover a new ENR
func ParseEnr(node *enode.Node) (*EnrNode, error) {
	// check if the node is valid
	err := node.ValidateComplete()
	if err != nil {
		return &EnrNode{}, EnrValidationError
	}

	// create a new ENR node
	enrNode := NewEnrNode(node.ID())

	// compose the rest of the info
	enrNode.Seq = node.Seq()
	enrNode.IP = node.IP()
	enrNode.UDP = node.UDP()
	enrNode.TCP = node.TCP()
	enrNode.Pubkey = node.Pubkey()

	// Retrieve the Fork Digest and the attestnets
	eth2Data, ok, err := ParseNodeEth2Data(*node)
	if !ok {
		eth2Data = new(common.Eth2Data)
	} else {
		if err != nil {
			return &EnrNode{}, Eth2DataParsingError
		}
	}
	enrNode.Eth2Data = eth2Data

	// in this case ParseAttnets will always return or a new(attnets) or a filled one
	attnets, _, _ := ParseAttnets(*node)
	enrNode.Attnets = attnets

	return enrNode, nil
}

func (enr *EnrNode) GetPeerID() (peer.ID, error) {
	// Get the public key and the peer.ID of the discovered peer
	pubkey, err := utils.ConvertECDSAPubkeyToSecp2561k(enr.Pubkey)
	if err != nil {
		return *new(peer.ID), errors.Errorf("error converting geth pubkey to libp2p pubkey")
	}

	peerId, err := peer.IDFromPublicKey(pubkey)
	if err != nil {
		return peerId, errors.Errorf("error extracting peer.ID from node %s", enr.ID)
	}
	return peerId, nil
}

func (enr *EnrNode) GetPubkeyString() string {
	pubBytes := gcrypto.FromECDSAPub(enr.Pubkey)
	pubkey := hex.EncodeToString(pubBytes)
	return pubkey
}

func (enr *EnrNode) GetAttnetsString() string {
	return hex.EncodeToString(enr.Attnets.Raw[:])
}

type Attnets struct {
	Raw       AttnetsENREntry
	NetNumber int
}

// ParseAttnets returns always an initialized Attnet object
// If the Ethereum Node doesn't have the Attnets key-value NetNumber will be -1
func ParseAttnets(node enode.Node) (attnets *Attnets, exists bool, err error) {
	attNetEntry := new(AttnetsENREntry)
	att := &Attnets{
		Raw:       *attNetEntry,
		NetNumber: -1,
	}

	err = node.Load(&att.Raw)
	if err != nil {
		return att, false, nil
	}

	// count the number of bits in the Attnets
	att.NetNumber = CountBits(att.Raw[:])
	return att, true, nil
}

func CountBits(byteArr []byte) int {
	rawInt := binary.BigEndian.Uint64(byteArr)
	return bits.OnesCount64(rawInt)
}
