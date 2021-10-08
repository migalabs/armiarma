package enode

import (
	"encoding/hex"
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
