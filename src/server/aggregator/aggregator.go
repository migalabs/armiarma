package aggregator

import (
	"context"
	"log"

	pb "github.com/protolambda/rumor/proto"
)

type Server struct {
	pb.UnimplementedAggregatorServer
}

func (s *Server) NewPeerMetadata(ctx context.Context, in *pb.NewPeerMetadataRequest) (*pb.NewPeerMetadataReply, error) {
	log.Printf("Received new Peer information: %v", in)
	return &pb.NewPeerMetadataReply{Dummy: "Ok" }, nil
}
