package methods

import (
	"github.com/migalabs/armiarma/src/rpc/reqresp"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

var GoodbyeRPCv1 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/goodbye/1/ssz",
	RequestCodec:              reqresp.NewSSZCodec(func() reqresp.SerDes { return new(common.Goodbye) }, 8, 8),
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(common.Goodbye) }, 8, 8),
	DefaultResponseChunkCount: 0,
}
