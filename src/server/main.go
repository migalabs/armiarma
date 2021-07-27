
package main

import (
	log "github.com/sirupsen/logrus"
	"net"
	"flag"
	"fmt"
	//"context"
	//"time"

	"google.golang.org/grpc"
  "github.com/protolambda/rumor/server/aggregator"
  pb "github.com/protolambda/rumor/proto"

)

const (
	port = ":50051"
)

var (
	rpcHost = flag.String("rpc-host", "0.0.0.0", "RPC host for the gRPC")
	rpcPort = flag.Int("rpc-port", 50051, "RPC port for the gRPC")
)

func main() {
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *rpcHost, *rpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	
	s := grpc.NewServer()
	pb.RegisterAggregatorServer(s, &aggregator.Server{})
	log.Info("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve: %v", err)
	}
}
