package serverexporter

import (
	"context"
	"fmt"
	"time"

	pb "github.com/protolambda/rumor/proto"
	"google.golang.org/grpc"
)

// TODO:
//      - Aggregate a Context framework to shut down the endless loop
//      - Choose the initialization parameteres / place in the code
//      - Wait untill the gRPC format is defined

type ServerEndpoint struct {
	IP   string
	Port string
	// Notification channels
	NewPeerDiscov   chan *pb.NewPeerMetadataRequest
	NewPeerConn     chan *pb.NewPeerMetadataRequest
	NewPeerDisc     chan *pb.NewPeerMetadataRequest
	NewPeerMetadata chan *pb.NewPeerMetadataRequest
	close           chan struct{}
}

// Initialize the Peer Metrics Export for towards the Armiarma Server Endpoint
func NewServerExport(ip string, port string) *ServerEndpoint {
	se := &ServerEndpoint{
		IP:              ip,
		Port:            port,
		NewPeerDiscov:   make(chan *pb.NewPeerMetadataRequest, 200),
		NewPeerConn:     make(chan *pb.NewPeerMetadataRequest, 200),
		NewPeerDisc:     make(chan *pb.NewPeerMetadataRequest, 200),
		NewPeerMetadata: make(chan *pb.NewPeerMetadataRequest, 200),
		close:           make(chan struct{}),
	}
	return se
}

// Function that runs the routiene to export the data to the Armiarma server
func (c *ServerEndpoint) Start() {
	fmt.Println("Starting Server Export towards Armiarma Server: ", c.IP, ":", c.Port)
	// (generate the gRPC host and configure it)
	serverAddr := c.IP + ":" + c.Port
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Println("connection with the Armiarma Server couldn't be made")
	}
	// keep the gRPC method
	client := pb.NewAggregatorClient(conn)
	// Call the Export Infinite Loop
	go c.exportLoop(conn, &client)
}

// Close/Stop fucntion to finish the end loop to export the metadata
func (c *ServerEndpoint) Stop() {
	fmt.Println("sending closing signal to end the server exporter loop")
	<-c.close
}

// Main loop where the connection with the Armiarma Server will be orchestrated
func (c *ServerEndpoint) exportLoop(conn *grpc.ClientConn, client *pb.AggregatorClient) {
	// Initialize and blablabla
	for {
		// Infinite loop that will receive each of the events and act accordingly
		select {
		case msg := <-c.NewPeerDiscov:
			NotifyNewPeerDiscov(*client, msg)
		case msg := <-c.NewPeerConn:
			NotifyNewPeerConn(*client, msg)
		case msg := <-c.NewPeerDisc:
			NotifyNewPeerDisconn(*client, msg)
		case msg := <-c.NewPeerMetadata:
			NotifyNewPeerMetadata(*client, msg)

		// Close loop channel
		case <-c.close:
			conn.Close()
			fmt.Println("Closing Server Metrics Exporting! Ciao Armiarma service ")
			return
		}
	}
}

// Send gRPC NewPeerDiscov message to the provided endpoint of the Armiarma Server
func NotifyNewPeerDiscov(client pb.AggregatorClient, msg *pb.NewPeerMetadataRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := client.NewPeerMetadata(ctx, msg)
	if err != nil {
		fmt.Println("error reporting peer discovery to server: ", err)
	}
	fmt.Printf("Response:", r)
}

// Send gRPC NewPeerMetadata message to the provided endpoint of the Armiarma Server
func NotifyNewPeerMetadata(client pb.AggregatorClient, msg *pb.NewPeerMetadataRequest) {
	fmt.Println("Sending metadata message to Armiarma Server")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := client.NewPeerMetadata(ctx, msg)
	/* to generate the struct of the message
	   not := &pb.NewPeerMetadataRequest{
	       CrawlerId: "crawler_1",
	       PeerId: peerMetrics.PeerId.String(),
	       NodeId: peerMetrics.NodeId,
	       ClientType: peerMetrics.ClientType,
	       // etc
	   }
	*/
	if err != nil {
		fmt.Println("error reporting peer metadata to server: ", err)
	}
	fmt.Printf("Response:", r)
}
