package methods

import (
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"

	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
)

var StatusRPCv1 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/status/1/ssz",
	RequestCodec:              reqresp.NewSSZCodec(func() reqresp.SerDes { return new(beacon.Status) }, beacon.StatusByteLen, beacon.StatusByteLen),
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(beacon.Status) }, beacon.StatusByteLen, beacon.StatusByteLen),
	DefaultResponseChunkCount: 1,
}
