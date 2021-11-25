package peering

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_PrunnedPeer(t *testing.T) {

	tNow := time.Now()
	prunnedPeer1 := NewPrunedPeer("Peer1", PositiveDelayType)

	prunnedPeer1.ConnEventHandler("None")
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(6*time.Hour)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(7*time.Hour)))

	require.Equal(t, true, time.Now().Sub(prunnedPeer1.BaseDeprecationTimestamp) < 1*time.Second)

	prunnedPeer1.ConnEventHandler("None")
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(12*time.Hour)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(13*time.Hour)))

	tNow = time.Now()
	prunnedPeer1.ConnEventHandler("") // this should go to NegativeWithHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(2*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(3*time.Minute)))

	prunnedPeer1.ConnEventHandler("Connection reset by peer") // this should maintain in NegativeWithHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(4*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(6*time.Minute)))

	prunnedPeer1.ConnEventHandler("connection refused") // this should maintain in NegativeWithHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(8*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(9*time.Minute)))

	tNow = time.Now()
	prunnedPeer1.ConnEventHandler("no route to host") // this should go to NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(12*time.Hour)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(13*time.Hour)))

	prunnedPeer1.ConnEventHandler("unreachable network") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(24*time.Hour)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(25*time.Hour)))

	prunnedPeer1.ConnEventHandler("i/o timeout") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(48*time.Hour)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(49*time.Hour)))

	prunnedPeer1.ConnEventHandler("peer id mismatch") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(96*time.Hour)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(97*time.Hour)))

	prunnedPeer1.ConnEventHandler("dial to self attempted") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(192*time.Hour)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(193*time.Hour)))

	tNow = time.Now()
	prunnedPeer1.ConnEventHandler("context deadline exceeded") // this should go to NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(2*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(3*time.Minute)))

	prunnedPeer1.ConnEventHandler("dial backoff") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(4*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(5*time.Minute)))

	time.Sleep(5 * time.Second)
	prunnedPeer1.ConnEventHandler("error requesting metadata") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(8*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(9*time.Minute)))

	// check this has not been refreshed
	require.Equal(t, true, time.Now().Sub(prunnedPeer1.BaseDeprecationTimestamp) > 5*time.Second)

	tNow = time.Now()
	prunnedPeer1.ConnEventHandler("None") // this should go to Positive
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(6*time.Hour)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(7*time.Hour)))

}
