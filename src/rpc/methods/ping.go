package methods

import (
	"github.com/migalabs/armiarma/src/rpc/reqresp"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

var PingRPCv1 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/ping/1/ssz",
	RequestCodec:              reqresp.NewSSZCodec(func() reqresp.SerDes { return new(common.Ping) }, 8, 8),
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(common.Ping) }, 8, 8),
	DefaultResponseChunkCount: 1,
}
