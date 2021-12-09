package peering

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Deprecation(t *testing.T) {
	require.Equal(t, 1, 1)

	testPeer := NewPrunedPeer("test", PositiveDelayType)
	testPeer.BaseDeprecationTimestamp = testPeer.BaseDeprecationTimestamp.Add(-DeprecationTime)

	require.Equal(t, true, testPeer.Deprecable())

	testPeer.ConnEventHandler("None")

	require.Equal(t, false, testPeer.Deprecable())

	testPeer.BaseDeprecationTimestamp = testPeer.BaseDeprecationTimestamp.Add(-DeprecationTime)

	testPeer.ConnEventHandler("error requesting metadata")
	require.Equal(t, true, testPeer.Deprecable())

	testPeer.ConnEventHandler("i/o timeout")
	require.Equal(t, true, testPeer.Deprecable())

	testPeer.ConnEventHandler("None")

	require.Equal(t, false, testPeer.Deprecable())

	testPeer.ConnEventHandler("connection reset by peer")
	require.Equal(t, false, testPeer.Deprecable())

	testPeer.ConnEventHandler("dial to self attempted")
	require.Equal(t, false, testPeer.Deprecable())

	testPeer.BaseDeprecationTimestamp = testPeer.BaseDeprecationTimestamp.Add(-DeprecationTime)

	testPeer.ConnEventHandler("connection refused")
	require.Equal(t, true, testPeer.Deprecable())

	testPeer.ConnEventHandler("context deadline exceeded")
	require.Equal(t, true, testPeer.Deprecable())

	testPeer.ConnEventHandler("no route to host")
	require.Equal(t, true, testPeer.Deprecable())

	testPeer.ConnEventHandler("peer id mismatch")
	require.Equal(t, true, testPeer.Deprecable())

	testPeer.ConnEventHandler("rfgdsfghsdfh")
	require.Equal(t, true, testPeer.Deprecable())

	testPeer.ConnEventHandler("None")
	require.Equal(t, false, testPeer.Deprecable())

}
