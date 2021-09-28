package enode

import (
	"encoding/hex"
)

const ATTNETS_KEY = "attnets"
const ETH2_ENR_KEY = "eth2"

type AttnetsENREntry []byte

func NewAttnetsENREntry(input_bytes string) (AttnetsENREntry, error) {

	result, err := hex.DecodeString(input_bytes)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (aee AttnetsENREntry) ENRKey() string {
	return ATTNETS_KEY
}

type Eth2ENREntry []byte

func NewEth2DataEntry(input_bytes string) (Eth2ENREntry, error) {
	result, err := hex.DecodeString(input_bytes)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (eee Eth2ENREntry) ENRKey() string {
	return ETH2_ENR_KEY
}
