package ethereum

import (
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"
	"net"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
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

// ParseError contains detailed information about parsing failures
type ParseError struct {
	NodeID    string
	Field     string
	Err       error
	Timestamp time.Time
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse %s for node %s: %v", e.Field, e.NodeID, e.Err)
}

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

	Eth2DataExists  bool
	Eth2ParseError  error
	AttnetsExists   bool
	AttnetsParseErr error
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
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}

	// check if the node is valid
	err := node.ValidateComplete()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", EnrValidationError, err)
	}

	// create a new ENR node
	enrNode := NewEnrNode(node.ID())

	// compose the rest of the info
	enrNode.Seq = node.Seq()
	enrNode.IP = node.IP()
	enrNode.UDP = node.UDP()
	enrNode.TCP = node.TCP()
	enrNode.Pubkey = node.Pubkey()

	// Retrieve the Fork Digest with improved error handling
	eth2Data, exists, err := ParseNodeEth2Data(*node)
	enrNode.Eth2DataExists = exists
	
	if err != nil {
		// Store the error but don't fail the entire parsing
		enrNode.Eth2ParseError = err
		enrNode.Eth2Data = new(common.Eth2Data)
		
		// Log the error for debugging
		fmt.Printf("Warning: Failed to parse eth2 data for node %s: %v\n", node.ID(), err)
	} else if !exists {
		// No eth2 data is not necessarily an error
		enrNode.Eth2Data = new(common.Eth2Data)
	} else {
		enrNode.Eth2Data = eth2Data
	}

	// Parse attnets with improved error handling
	attnets, exists, err := ParseAttnets(*node)
	enrNode.AttnetsExists = exists
	
	if err != nil {
		enrNode.AttnetsParseErr = err
		enrNode.Attnets = new(Attnets)
		fmt.Printf("Warning: Failed to parse attnets for node %s: %v\n", node.ID(), err)
	} else {
		enrNode.Attnets = attnets
	}

	return enrNode, nil
}

// HasValidEth2Data checks if the node has valid eth2 data
func (enr *EnrNode) HasValidEth2Data() bool {
	return enr.Eth2DataExists && enr.Eth2ParseError == nil && enr.Eth2Data != nil
}

// GetForkDigest safely retrieves the fork digest
func (enr *EnrNode) GetForkDigest() (string, error) {
	if !enr.HasValidEth2Data() {
		if enr.Eth2ParseError != nil {
			return "", fmt.Errorf("eth2 data parsing failed: %v", enr.Eth2ParseError)
		}
		return "", fmt.Errorf("no valid eth2 data available")
	}
	
	// Check for empty fork digest
	emptyDigest := common.ForkDigest{}
	if enr.Eth2Data.ForkDigest == emptyDigest {
		return "", fmt.Errorf("fork digest is empty")
	}
	
	return enr.Eth2Data.ForkDigest.String(), nil
}

func (enr *EnrNode) GetPeerID() (peer.ID, error) {
	if enr.Pubkey == nil {
		return peer.ID(""), fmt.Errorf("public key is nil")
	}

	// Get the public key and the peer.ID of the discovered peer
	pubkey, err := utils.ConvertECDSAPubkeyToSecp2561k(enr.Pubkey)
	if err != nil {
		return peer.ID(""), errors.Errorf("error converting geth pubkey to libp2p pubkey: %v", err)
	}

	peerId, err := peer.IDFromPublicKey(pubkey)
	if err != nil {
		return peerId, errors.Errorf("error extracting peer.ID from node %s: %v", enr.ID, err)
	}
	return peerId, nil
}

func (enr *EnrNode) GetPubkeyString() string {
	if enr.Pubkey == nil {
		return ""
	}
	pubBytes := gcrypto.FromECDSAPub(enr.Pubkey)
	pubkey := hex.EncodeToString(pubBytes)
	return pubkey
}

func (enr *EnrNode) GetAttnetsString() string {
	if enr.Attnets == nil || len(enr.Attnets.Raw) == 0 {
		return ""
	}
	return hex.EncodeToString(enr.Attnets.Raw[:])
}

type Attnets struct {
	Raw       AttnetsENREntry
	NetNumber int
}

// ParseAttnets returns always an initialized Attnet object
// If the Ethereum Node doesn't have the Attnets key-value NetNumber will be -1
func ParseAttnets(node enode.Node) (attnets *Attnets, exists bool, err error) {
	att := &Attnets{
		Raw:       make(AttnetsENREntry, 0),
		NetNumber: -1,
	}

	var raw AttnetsENREntry
	err = node.Load(&raw)
	if err != nil {
		// Check if it's because the entry doesn't exist
		if err.Error() == "missing ENR key \""+ATTNETS_KEY+"\"" {
			return att, false, nil
		}
		return att, false, fmt.Errorf("failed to load attnets: %v", err)
	}

	// Validate the data
	if len(raw) == 0 {
		return att, true, fmt.Errorf("attnets data is empty")
	}

	att.Raw = raw
	// count the number of bits in the Attnets
	att.NetNumber = CountBits(att.Raw[:])
	return att, true, nil
}

func CountBits(byteArr []byte) int {
	if len(byteArr) == 0 {
		return 0
	}

	// Handle byte arrays of different lengths
	count := 0
	for _, b := range byteArr {
		count += bits.OnesCount8(b)
	}
	return count
}
