package ethereum

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/p2p/enode"
	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
)

const ATTNETS_KEY = "attnets"
const ETH2_ENR_KEY = "eth2"

// Attended networks are the networks the node will be participating in
type AttnetsENREntry []byte

func NewAttnetsENREntry(input_bytes string) AttnetsENREntry {

	result, err := hex.DecodeString(input_bytes)

	if err != nil {
		return nil
	}

	return result
}

func (aee AttnetsENREntry) ENRKey() string {
	return ATTNETS_KEY
}

// With this entry we allow the node to have a registered fork digest
type Eth2ENREntry []byte

func NewEth2DataEntry(input_bytes string) Eth2ENREntry {
	result, err := hex.DecodeString(input_bytes)

	if err != nil {
		return nil
	}

	return result
}

func (eee Eth2ENREntry) ENRKey() string {
	return ETH2_ENR_KEY
}

func (eee Eth2ENREntry) Eth2Data() (*beacon.Eth2Data, error) {
	var dat beacon.Eth2Data
	if err := dat.Deserialize(codec.NewDecodingReader(bytes.NewReader(eee), uint64(len(eee)))); err != nil {
		return nil, err
	}
	return &dat, nil
}

// ParseNodeEth2Data
// * This method will parse the Node and obtain information about it
// @param n: the enode from where to get the information
// @return the Eth2Data object from the beacon package
func ParseNodeEth2Data(n enode.Node) (data *beacon.Eth2Data, exists bool, err error) {
	var eth2 Eth2ENREntry
	if err := n.Load(&eth2); err != nil {
		return nil, false, nil
	}
	dat, err := eth2.Eth2Data()
	if err != nil {
		return nil, true, err
	}
	return dat, true, nil
}

func GetForkDigestFromEth2Data(b beacon.Eth2Data) string {
	return b.ForkDigest.String()
}

func GetForkDigestFromENode(n enode.Node) (string, error) {
	b_eth2Data, exists, err := ParseNodeEth2Data(n)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("node does not exist")
	}
	return GetForkDigestFromEth2Data(*b_eth2Data), nil
}

func GetForkDigestFromStatus(b beacon.Status) string {
	return b.ForkDigest.String()
}
