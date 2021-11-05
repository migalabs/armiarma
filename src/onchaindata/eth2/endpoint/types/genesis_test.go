package types

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGenesisDecoder(t *testing.T) {
	var genVR Root
	err := genVR.UnmarshalText([]byte("0x4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95"))
	require.Equal(t, nil, err)

	var genFV Version
	err = genFV.UnmarshalText([]byte("0x00000000"))
	require.Equal(t, nil, err)

	genesis := Genesis{
		GenesisTime:           time.Unix(1606824023, 0),
		GenesisValidatorsRoot: genVR,
		GenesisForkVersion:    genFV,
	}
	bytes, err := genesis.MarshalJSON()
	require.Equal(t, nil, err)
	var gen2 Genesis
	err = gen2.UnmarshalJSON(bytes)
	require.Equal(t, nil, err)
}
