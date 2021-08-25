package nodemetadata

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	gcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/libp2p/go-libp2p/config"
	tcp "github.com/libp2p/go-tcp-transport"
	"github.com/protolambda/rumor/p2p/custom"
)

func Test_ReqBeaconStatus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Generate the first testing Node Metadata Exchange

}

func makeNode(port int) (*host.Host, error) {
	// generate the options needed to generate a host
	hostOptions := custom.Config{}

	// Generate the private key for the host
	priv, err := ecdsa.GenerateKey(gcrypto.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	// Set the tcp transport port
	tcpt, err := config.TransportContructor(tcp.NewTCPTrasnport)
	if err != nil {
		return nil, err
	}
	hostOptions.Transports = append(hostOptions.Transports, tptc)
	// Set the Muxer
	// set ymux
	mtpt, err := config.MuxerConstructor(yamux.DefaultTransport)
	if err != nil {
		return nil, err
	}
	hostOptions.Muxers = append(hostOptions.Muxers, config.MsMuxC{MuxC: mtpt, ID: "/yamux/1.0.0"})
	// set
}
