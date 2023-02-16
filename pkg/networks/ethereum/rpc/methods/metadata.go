package methods

import (
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"
	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
)

var MetaDataRPCv1 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/metadata/2/ssz",
	RequestCodec:              (*reqresp.SSZCodec)(nil), // no reqresp data, just empty bytes.
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(beacon.MetaData) }, beacon.MetadataByteLen, beacon.MetadataByteLen),
	DefaultResponseChunkCount: 1,
}
