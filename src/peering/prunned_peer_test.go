package peering

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_PrunnedPeerDelays(t *testing.T) {

	tNow := time.Now()

	prunnedPeer1 := NewPrunedPeer("Peer1", PositiveDelayType)
	time.Sleep(5 * time.Second)

	prunnedPeer1.ConnEventHandler("None")
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(128*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(129*time.Minute)))

	// test that the BaseDeprecationTime has been updated
	require.Equal(t, true, time.Now().Sub(prunnedPeer1.BaseDeprecationTimestamp) < 1*time.Second)

	prunnedPeer1.ConnEventHandler("None")
	fmt.Println(prunnedPeer1.DelayObj.GetDegree())
	fmt.Println(prunnedPeer1.NextConnection())
	fmt.Println(tNow.Add(256 * time.Minute))
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(256*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(257*time.Minute)))

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
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(256*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(257*time.Minute)))

	prunnedPeer1.ConnEventHandler("unreachable network") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(512*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(513*time.Minute)))

	prunnedPeer1.ConnEventHandler("unreachable network") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(1024*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(1025*time.Minute)))

	prunnedPeer1.ConnEventHandler("unreachable network") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(2048*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(2049*time.Minute)))

	// check that it maintains the max delay
	prunnedPeer1.ConnEventHandler("unreachable network") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(2048*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(2049*time.Minute)))

	prunnedPeer1.ConnEventHandler("peer id mismatch") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(2048*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(2049*time.Minute)))

	prunnedPeer1.ConnEventHandler("dial to self attempted") // this should maintain in NegativeWithNoHope
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(2048*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(2049*time.Minute)))

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

	prunnedPeer1.ConnEventHandler("i/o timeout") // this should maintain in NegativeWithNoHope
	fmt.Println(prunnedPeer1.NextConnection())
	fmt.Println(tNow)
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(16*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(17*time.Minute)))

	// check this has not been refreshed
	require.Equal(t, true, time.Now().Sub(prunnedPeer1.BaseDeprecationTimestamp) > 5*time.Second)

	tNow = time.Now()
	prunnedPeer1.ConnEventHandler("None") // this should go to Positive
	require.Equal(t, true, prunnedPeer1.NextConnection().After(tNow.Add(128*time.Minute)))
	require.Equal(t, false, prunnedPeer1.NextConnection().After(tNow.Add(129*time.Minute)))

}
