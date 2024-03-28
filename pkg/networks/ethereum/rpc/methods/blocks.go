package methods

import (
	"encoding/hex"
	"fmt"
	"github.com/migalabs/armiarma/pkg/networks/ethereum/rpc/reqresp"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"
)

const MAX_REQUEST_BLOCKS_DENEB int = 128

type BlocksByRangeReqV2 struct {
	StartSlot Slot
	Count     view.Uint64View
	Step      view.Uint64View
}

func (r *BlocksByRangeReqV2) Data() map[string]interface{} {
	return map[string]interface{}{
		"start_slot": r.StartSlot,
		"count":      r.Count,
		"step":       r.Step,
	}
}

func (d *BlocksByRangeReqV2) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&d.StartSlot, &d.Count, &d.Step)
}

func (d *BlocksByRangeReqV2) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&d.StartSlot, &d.Count, &d.Step)
}

const blocksByRangeReqByteLen = 8 + 8 + 8

func (d *BlocksByRangeReqV2) ByteLength() uint64 {
	return blocksByRangeReqByteLen
}

func (*BlocksByRangeReqV2) FixedLength() uint64 {
	return blocksByRangeReqByteLen
}

func (d *BlocksByRangeReqV2) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&d.StartSlot, &d.Count, &d.Step)
}

func (r *BlocksByRangeReqV2) String() string {
	return fmt.Sprintf("%v", *r)
}

const MAX_REQUEST_BLOCKS_BY_ROOT = 1024

type BlocksByRootReq []Root

func (a *BlocksByRootReq) Deserialize(dr *codec.DecodingReader) error {
	return tree.ReadRootsLimited(dr, (*[]Root)(a), MAX_REQUEST_BLOCKS_BY_ROOT)
}

func (a BlocksByRootReq) Serialize(w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, a)
}

func (a BlocksByRootReq) ByteLength() (out uint64) {
	return uint64(len(a)) * 32
}

func (a *BlocksByRootReq) FixedLength() uint64 {
	return 0 // it's a list, no fixed length
}

func (r BlocksByRootReq) Data() []string {
	out := make([]string, len(r), len(r))
	for i := range r {
		out[i] = hex.EncodeToString(r[i][:])
	}
	return out
}

func (r BlocksByRootReq) String() string {
	if len(r) == 0 {
		return "empty blocks-by-root request"
	}
	out := make([]byte, 0, len(r)*66)
	for i, root := range r {
		hex.Encode(out[i*66:], root[:])
		out[(i+1)*66-2] = ','
		out[(i+1)*66-1] = ' '
	}
	return "blocks-by-root requested: " + string(out[:len(out)-1])
}

// methods

var BlocksByRangeRPCv2 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/beacon_blocks_by_range/2/ssz_snappy",
	RequestCodec:              reqresp.NewSSZCodec(func() reqresp.SerDes { return new(BlocksByRangeReqV2) }, blocksByRangeReqByteLen, blocksByRangeReqByteLen),
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(BlocksByRangeReqV2) }, 0, uint64(0)),
	DefaultResponseChunkCount: 20,
}

var BlocksByRootRPCv2 = reqresp.RPCMethod{
	Protocol:                  "/eth2/beacon_chain/req/beacon_blocks_by_root/2/ssz_snappy",
	RequestCodec:              reqresp.NewSSZCodec(func() reqresp.SerDes { return new(BlocksByRootReq) }, 0, 32*MAX_REQUEST_BLOCKS_BY_ROOT),
	ResponseChunkCodec:        reqresp.NewSSZCodec(func() reqresp.SerDes { return new(BlocksByRootReq) }, 0, uint64(0)),
	DefaultResponseChunkCount: 20,
}
