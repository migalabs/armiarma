package networks

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/utils"
)

// RPC related Requestable calls
type RPCRequest func(
	context.Context,
	*sync.WaitGroup,
	host.Host,
	network.Conn,
	peer.ID,
	RPCResult,
	*error,
)
type RPCRequestsName string

// RPCResults is a simple interface that could be casted
type RPCResult interface{}

// NetworkNode is the interface that any nodes info from any network should satisfy
// General so that the Crawler only knows about how to store or to load it
type NetworkNode interface {
	SQLStore()
	SQLLoad()
	SQLRemove()
	SQLQuery()
	SQLDrop()
}

type NetworkAttribute interface{}

// Network interface compiles all the list of
type Network interface {
	Type() utils.NetworkType
	Node() NetworkNode
	RPCRequests() []RPCRequest
	GetAttr(string) (NetworkAttribute, error)
	SetAttr(string, NetworkAttribute)
}
