package methods

import (
	"github.com/migalabs/armiarma/src/rpc/reqresp"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

var MetaDataRPCv1 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/metadata/1/ssz",
	RequestCodec:              (*reqresp.SSZCodec)(nil), // no request data, just empty bytes.
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(common.MetaData) }, common.MetadataByteLen, common.MetadataByteLen),
	DefaultResponseChunkCount: 1,
}
