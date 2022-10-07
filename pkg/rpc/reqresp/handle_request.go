package reqresp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

const reqrespBufferSize = 2048

// RequestPayloadHandler processes a reqresp (decompressed if previously compressed), read from r.
// The handler can respond by writing to w. After returning the writer will automatically be closed.
// If the input is already known to be invalid, e.g. the reqresp size is invalid, then `invalidInputErr != nil`, and r will not read anything more.
type RequestPayloadHandler func(ctx context.Context, peerId peer.ID, reqrespLen uint64, r io.Reader, w io.Writer, comp Compression, invalidInputErr error)

type StreamCtxFn func() context.Context

// MakeStreamHandler startReqRPC registers a reqresp handler for the given protocol. Compression is optional and may be nil.
func (handle RequestPayloadHandler) MakeStreamHandler(newCtx StreamCtxFn, comp Compression, minreqrespContentSize, maxreqrespContentSize uint64) network.StreamHandler {
	return func(stream network.Stream) {
		peerId := stream.Conn().RemotePeer()
		ctx, cancel := context.WithCancel(newCtx())
		defer cancel()

		go func() {
			<-ctx.Done()
			// TODO: should this be a stream reset?
			_ = stream.Close() // Close stream after ctx closes.
		}()

		w := io.WriteCloser(stream)
		// If no reqresp data, then do not even read a length from the stream.
		if maxreqrespContentSize == 0 {
			handle(ctx, peerId, 0, nil, w, comp, nil)
			return
		}

		var invalidInputErr error

		// TODO: pool this
		blr := NewBufLimitReader(stream, reqrespBufferSize, 0)
		blr.N = 1 // var ints need to be read byte by byte
		blr.PerRead = true
		reqLen, err := binary.ReadUvarint(blr)
		blr.PerRead = false
		switch {
		case err != nil:
			invalidInputErr = err
		case reqLen < minreqrespContentSize:
			// Check against raw content size minimum (without compression applied)
			invalidInputErr = fmt.Errorf("reqresp length %d is unexpectedly small, reqresp size minimum is %d", reqLen, minreqrespContentSize)
		case reqLen > maxreqrespContentSize:
			// Check against raw content size limit (without compression applied)
			invalidInputErr = fmt.Errorf("reqresp length %d exceeds reqresp size limit %d", reqLen, maxreqrespContentSize)
		case comp != nil:
			// Now apply compression adjustment for size limit, and use that as the limit for the buffered-limited-reader.
			s, err := comp.MaxEncodedLen(maxreqrespContentSize)
			if err != nil {
				invalidInputErr = err
			} else {
				maxreqrespContentSize = s
			}
		}
		switch {
		case invalidInputErr != nil: // If the input is invalid, never read it.
			maxreqrespContentSize = 0
		case comp == nil:
			blr.N = int(maxreqrespContentSize)
		default:
			v, err := comp.MaxEncodedLen(maxreqrespContentSize)
			if err != nil {
				blr.N = int(maxreqrespContentSize)
			} else {
				blr.N = int(v)
			}
		}
		r := io.Reader(blr)
		handle(ctx, peerId, reqLen, r, w, comp, invalidInputErr)
	}
}
