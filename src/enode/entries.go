package enode

import (
	"encoding/hex"
)

const ATTNETS_KEY = "attnets"
const ETH2_ENR_KEY = "eth2"

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
