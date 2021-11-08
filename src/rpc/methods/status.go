package methods

import (
	"github.com/migalabs/armiarma/src/rpc/reqresp"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

var StatusRPCv1 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/status/1/ssz",
	RequestCodec:              reqresp.NewSSZCodec(func() reqresp.SerDes { return new(common.Status) }, common.StatusByteLen, common.StatusByteLen),
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(common.Status) }, common.StatusByteLen, common.StatusByteLen),
	DefaultResponseChunkCount: 1,
}
