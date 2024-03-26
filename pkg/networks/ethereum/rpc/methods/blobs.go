package methods

import (
	"encoding/hex"
	"fmt"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"
)

// https://github.com/ethereum/consensus-specs/blob/dev/specs/deneb/p2p-interface.md#blobsidecarsbyroot-v1
const (
	MAX_BLOBS_PER_BLOCK   int = 6
	MAX_BLOBS_PER_RPC_REQ int = MAX_REQUEST_BLOCKS_DENEB * MAX_BLOBS_PER_BLOCK
)

type BlobIdentifier struct {
	BlockRoot Root
	Index     view.Uint64View
}

func (blobId *BlobIdentifier) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&blobId.BlockRoot, &blobId.Index)
}

func (blobId *BlobIdentifier) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&blobId.BlockRoot, &blobId.Index)
}

func (blobId *BlobIdentifier) ByteLength() uint64 {
	return blobId.BlockRoot.FixedLength() + blobId.Index.FixedLength()
}

func (blobId *BlobIdentifier) FixedLength() uint64 {
	return blobId.BlockRoot.FixedLength() + blobId.Index.FixedLength()
}

func (blobId *BlobIdentifier) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&blobId.BlockRoot, &blobId.Index)
}

func (blobId *BlobIdentifier) String() string {
	return fmt.Sprintf("%v", *blobId)
}

type BlobByRootV1 []BlobIdentifier

func (b BlobByRootV1) Deserialize(dr *codec.DecodingReader) error {
	var idx int = 0
	return dr.List(
		func() codec.Deserializable {
			i := idx
			idx++
			return &b[i]
		},
		uint64(len(b)),
		uint64(MAX_BLOBS_PER_RPC_REQ))
}

func (b BlobByRootV1) Serialize(w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &b[i]
	},
		uint64(len(b)),
		uint64(MAX_BLOBS_PER_RPC_REQ))
}

func (b BlobByRootV1) ByteLength() uint64 {
	return uint64(len(b) * (32 + 8))
}

func (b BlobByRootV1) FixedLength() uint64 {
	return 0
}

func (b BlobByRootV1) String() string {
	if len(b) == 0 {
		return "empty blobs-by-root request"
	}
	out := make([]byte, 0, len(b)*66)
	for i, bId := range b {
		hex.Encode(out[i*66:], bId.BlockRoot[:])
		out[(i+1)*66-2] = ','
		out[(i+1)*66-1] = ' '
	}
	return "blobs-by-root requested: " + string(out[:len(out)-1])
}

type BlobsByRangeReqV1 struct {
	StartSlot Slot
	Count     view.Uint64View
}

func (b *BlobsByRangeReqV1) Data() map[string]interface{} {
	return map[string]interface{}{
		"start_slot": b.StartSlot,
		"count":      b.Count,
	}
}

func (b *BlobsByRangeReqV1) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&b.StartSlot, &b.Count)
}

func (b *BlobsByRangeReqV1) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&b.StartSlot, &b.Count)
}

const blobsByRangeReqBytes uint64 = 8 + 8

func (b BlobsByRangeReqV1) ByteLength() uint64 {
	return blobsByRangeReqBytes
}

func (b *BlobsByRangeReqV1) FixedLength() uint64 {
	return blobsByRangeReqBytes
}

func (b *BlobsByRangeReqV1) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&b.StartSlot, &b.Count)
}

func (b *BlobsByRangeReqV1) String() string {
	return fmt.Sprintf("%v", *b)
}

var BlobsByRangeRPCv1 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/blob_sidecars_by_range/1/ssz_snappy",
	RequestCodec:              reqresp.NewSSZCodec(func() reqresp.SerDes { return new(BlobsByRangeReqV1) }, blobsByRangeReqBytes, blobsByRangeReqBytes),
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(BlobsByRangeReqV1) }, 0, uint64(0)),
	DefaultResponseChunkCount: 20,
}

var BlobsByRootRPCv1 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/blob_sidecars_by_root/1/ssz_snappy",
	RequestCodec:              reqresp.NewSSZCodec(func() reqresp.SerDes { return new(BlobByRootV1) }, 0, uint64((32+8)*MAX_BLOBS_PER_RPC_REQ)),
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(BlobByRootV1) }, 0, uint64(0)),
	DefaultResponseChunkCount: 20,
}
