package ethereum

import (
	comm "github.com/migalabs/armiarma/pkg/networks/common"
)

const (
	BeaconStatusRPC   comm.RPCRequestsName = "beacon-status"
	BeaconMetadataRPC comm.RPCRequestsName = "beacon-metadata"
)

type ethereumRPCs struct {
	// List of requestable functions that can be asked when
	// we stablish a connection to a node
	RPCRequestables map[comm.RPCRequestsName]comm.RPCRequest
}

var EthereumRPCs = initEthereumNode()

func initEthereumNode() ethereumRPCs {

	ethNode := ethereumRPCs{
		RPCRequestables: make(map[comm.RPCRequestsName]comm.RPCRequest),
	}

	// TODO: Still need to address this -
	// add the RPCs
	// ethNode.RPCRequestables[BeaconStatusRPC] = rpc.ReqBeaconStatus
	// ethNode.RPCRequestables[BeaconMetadataRPC] = rpc.ReqBeaconMetadata

	return ethNode
}
