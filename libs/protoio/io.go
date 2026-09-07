// Protocol Buffers for Go with Gadgets
//
// Copyright (c) 2013, The GoGo Authors. All rights reserved.
// http://github.com/gogo/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// Modified to return number of bytes written by Writer.WriteMsg(), and added byteReader.

package protoio

import (
	"io"

	"github.com/cosmos/gogoproto/proto"
)

type Writer interface {
	WriteMsg(proto.Message) (int, error)
}

type WriteCloser interface {
	Writer
	io.Closer
}

type Reader interface {
	ReadMsg(msg proto.Message) (int, error)
}

type ReadCloser interface {
	Reader
	io.Closer
}

type marshaler interface {
	MarshalTo(data []byte) (n int, err error)
}

func getSize(v any) (int, bool) {
	if sz, ok := v.(interface {
		Size() (n int)
	}); ok {
		return sz.Size(), true
	} else if sz, ok := v.(interface {
		ProtoSize() (n int)
	}); ok {
		return sz.ProtoSize(), true
	}
	return 0, false
}

// byteReader wraps an io.Reader and implements io.ByteReader, required by
// binary.ReadUvarint(). Reading one byte at a time is extremely slow, but this
// is what Amino did previously anyway, and the caller can wrap the underlying
// reader in a bufio.Reader if appropriate.
type byteReader struct {
	reader    io.Reader
	buf       []byte
	bytesRead int // keeps track of bytes read via ReadByte()
}

func newByteReader(r io.Reader) *byteReader {
	return &byteReader{
		reader: r,
		buf:    make([]byte, 1),
	}
}

// maxConsecutiveEmptyReads bounds how many (0, nil) reads ReadByte will
// tolerate before giving up. Mirrors the analogous guard in the standard
// library's bufio.Reader (bufio.maxConsecutiveEmptyReads), which exists for
// the same reason: an io.Reader is allowed to legally return (0, nil), but a
// reader that does so forever must not hang its caller indefinitely.
const maxConsecutiveEmptyReads = 100

func (r *byteReader) ReadByte() (byte, error) {
	for i := 0; i < maxConsecutiveEmptyReads; i++ {
		n, err := r.reader.Read(r.buf)
		r.bytesRead += n
		if n == 1 {
			// A conforming io.Reader may legally return n=1 alongside a
			// non-nil error (e.g. io.EOF) in the same call; the byte itself
			// is still valid and must not be discarded along with the error.
			return r.buf[0], nil
		}
		if err != nil {
			return 0x00, err
		}
		// n == 0, err == nil is explicitly legal per the io.Reader doc
		// ("callers should treat a return of 0 and nil as indicating that
		// nothing happened"). Retry rather than returning r.buf[0], which
		// was never written this call and would otherwise surface as a
		// fabricated zero byte or a stale repeat of the previous read.
	}
	return 0x00, io.ErrNoProgress
}

func (r *byteReader) resetBytesRead() {
	r.bytesRead = 0
}
