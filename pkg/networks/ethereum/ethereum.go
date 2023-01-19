package ethereum

import (
	eth "github.com/ethereum/go-ethereum/p2p/enode"
)

const (
	AllRPCRequests []RPCRequestsName = []RPCRequestsName{
		// TODO: Add all the RPC requests for BeaconStatusRPC and MetadataRPC
	}
)

type EthereumNode struct {
	// TODO: add all kind of missing
	ENR eth.Enode

	// List of requestable functions that can be asked when
	// we stablish a connection to a node
	RPCRequestables map[RPCRequestsName]RPCRequest
}
