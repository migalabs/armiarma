package ethereum

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/p2p/enode"
	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
)

const ATTNETS_KEY = "attnets"
const ETH2_ENR_KEY = "eth2"

// Expected size for eth2 data (fork version + next fork version + next fork epoch)
const ETH2_DATA_SIZE = 16

var (
	ErrInvalidHex      = errors.New("invalid hex string")
	ErrInvalidDataSize = errors.New("invalid data size")
	ErrNoEth2Data      = errors.New("no eth2 data in ENR")
	ErrDeserialize     = errors.New("failed to deserialize eth2 data")
)

// Attended networks are the networks the node will be participating in
type AttnetsENREntry []byte

func NewAttnetsENREntry(input_bytes string) (AttnetsENREntry, error) {
	// Remove 0x prefix if present
	if len(input_bytes) >= 2 && input_bytes[:2] == "0x" {
		input_bytes = input_bytes[2:]
	}

	result, err := hex.DecodeString(input_bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidHex, err)
	}

	return result, nil
}

func (aee AttnetsENREntry) ENRKey() string {
	return ATTNETS_KEY
}

// With this entry we allow the node to have a registered fork digest
type Eth2ENREntry []byte

func NewEth2DataEntry(input_bytes string) (Eth2ENREntry, error) {
	// Remove 0x prefix if present
	if len(input_bytes) >= 2 && input_bytes[:2] == "0x" {
		input_bytes = input_bytes[2:]
	}

	result, err := hex.DecodeString(input_bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidHex, err)
	}

	// Validate expected size
	if len(result) != ETH2_DATA_SIZE {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidDataSize, ETH2_DATA_SIZE, len(result))
	}

	return result, nil
}

func (eee Eth2ENREntry) ENRKey() string {
	return ETH2_ENR_KEY
}

func (eee Eth2ENREntry) Eth2Data() (*beacon.Eth2Data, error) {
	// Validate data length before deserializing
	if len(eee) == 0 {
		return nil, fmt.Errorf("%w: empty eth2 data", ErrInvalidDataSize)
	}

	if len(eee) != ETH2_DATA_SIZE {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidDataSize, ETH2_DATA_SIZE, len(eee))
	}

	var dat beacon.Eth2Data
	reader := codec.NewDecodingReader(bytes.NewReader(eee), uint64(len(eee)))

	if err := dat.Deserialize(reader); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDeserialize, err)
	}

	// Validate the deserialized data
	if err := validateEth2Data(&dat); err != nil {
		return nil, err
	}

	return &dat, nil
}

// validateEth2Data performs basic validation on the eth2 data
func validateEth2Data(data *beacon.Eth2Data) error {
	// Check if fork digest is empty
	emptyDigest := beacon.ForkDigest{}
	if data.ForkDigest == emptyDigest {
		return fmt.Errorf("fork digest is empty")
	}

	return nil
}

// ParseNodeEth2Data
// * This method will parse the Node and obtain information about it
// @param n: the enode from where to get the information
// @return the Eth2Data object from the beacon package
func ParseNodeEth2Data(n enode.Node) (data *beacon.Eth2Data, exists bool, err error) {
	var eth2 Eth2ENREntry
	if err := n.Load(&eth2); err != nil {
		// Check if it's because the entry doesn't exist
		if err.Error() == "missing ENR key \""+ETH2_ENR_KEY+"\"" {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to load eth2 data from ENR: %v", err)
	}

	// Check if we actually got data
	if len(eth2) == 0 {
		return nil, true, fmt.Errorf("%w: eth2 entry exists but is empty", ErrInvalidDataSize)
	}

	dat, err := eth2.Eth2Data()
	if err != nil {
		return nil, true, fmt.Errorf("failed to parse eth2 data: %v", err)
	}

	return dat, true, nil
}

func GetForkDigestFromEth2Data(b beacon.Eth2Data) string {
	return b.ForkDigest.String()
}

func GetForkDigestFromENode(n enode.Node) (string, error) {
	b_eth2Data, exists, err := ParseNodeEth2Data(n)
	if err != nil {
		return "", fmt.Errorf("failed to parse node eth2 data: %v", err)
	}
	if !exists {
		return "", ErrNoEth2Data
	}
	if b_eth2Data == nil {
		return "", fmt.Errorf("eth2 data is nil despite existing")
	}
	return GetForkDigestFromEth2Data(*b_eth2Data), nil
}

func GetForkDigestFromStatus(b beacon.Status) string {
	return b.ForkDigest.String()
}
