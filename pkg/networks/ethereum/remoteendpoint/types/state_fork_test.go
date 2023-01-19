package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestForkStateDecoder(t *testing.T) {
	var prevVers Version
	err := prevVers.UnmarshalText([]byte("0x00000000"))
	require.Equal(t, nil, err)

	var currentVers Version
	err = currentVers.UnmarshalText([]byte("0x01000000"))
	require.Equal(t, nil, err)

	statefork := StateFork{
		PreviousVersion: prevVers,
		CurrentVersion:  currentVers,
		Epoch:           Epoch(74240),
	}
	bytes, err := statefork.MarshalJSON()
	require.Equal(t, nil, err)
	var stateFork2 StateFork
	err = stateFork2.UnmarshalJSON(bytes)
	require.Equal(t, nil, err)
}
