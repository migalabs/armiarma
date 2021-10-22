package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/protocol"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/hosts"
)

var protocolID = protocol.ID("/eth2/beacon_chain/req/metadata/1/ssz_snappy")

const PKG_NAME = "RPC"

type Custom_rpc struct {
	*base.Base
	Client   gorpc.Client
	Server   gorpc.Server
	host     *hosts.BasicLibp2pHost
	protocol protocol.ID
	counter  int
}

func StartClientServer(ctx context.Context, h hosts.BasicLibp2pHost, stdOpts base.LogOpts) Custom_rpc {

	custom_log_opts := rpcLoggerOpts(stdOpts)
	new_base, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(custom_log_opts),
	)

	if err != nil {
		fmt.Errorf("Could not create base object")
	}

	// server := gorpc.NewServer(h.Host(), protocolID)
	// client := gorpc.NewClientWithServer(h.Host(), protocolID, server)

	return Custom_rpc{
		Base: new_base,
		// Client:   *client,
		// Server:   *server,
		host:     &h,
		protocol: protocolID,
		counter:  0,
	}

}

func (r *Custom_rpc) StartServer() {

	rpcServer := gorpc.NewServer(r.host.Host(), protocolID)
	r.Server = *rpcServer

	svc := MetadataService{}
	err := rpcServer.Register(&svc)
	if err != nil {
		r.Log.Panicf(err.Error())
	}

	r.Log.Debugf("RPC Service is Up&Running")

	// for {
	time.Sleep(time.Second * 1)
	// }
}

/*func startClient(h hosts.BasicLibp2pHost, pingCount, randomDataSize int) {

	rpcClient := gorpc.NewClient(h.Host(), protocolID)
	numCalls := 0
	durations := []time.Duration{}
	betweenPingsSleep := time.Second * 1

	for {

		for peerID := range h.Host().Network().Peerstore().Peers() {

			var reply PingReply
			var args PingArgs

			c := randomDataSize
			b := make([]byte, c)
			_, err := rand.Read(b)
			if err != nil {
				panic(err)
			}

			args.Data = b

			time.Sleep(betweenPingsSleep)
			startTime := time.Now()
			err = rpcClient.Call(peer.ID, "PingService", "Ping", args, &reply)
			if err != nil {
				panic(err)
			}
			if !bytes.Equal(reply.Data, b) {
				panic("Received wrong amount of bytes back!")
			}
			endTime := time.Now()
			diff := endTime.Sub(startTime)
			fmt.Printf("%d bytes from %s (%s): seq=%d time=%s\n", c, peerInfo.ID.String(), peerInfo.Addrs[0].String(), numCalls+1, diff)
			numCalls += 1
			durations = append(durations, diff)
		}

	}

	totalDuration := int64(0)
	for _, dur := range durations {
		totalDuration = totalDuration + dur.Nanoseconds()
	}
	averageDuration := totalDuration / int64(len(durations))
	fmt.Printf("Average duration for ping reply: %s\n", time.Duration(averageDuration))

}*/

func rpcLoggerOpts(i_opts base.LogOpts) base.LogOpts {
	i_opts.ModName = PKG_NAME
	return i_opts
}
