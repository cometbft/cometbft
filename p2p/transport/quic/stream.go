package quic

import (
	"github.com/quic-go/quic-go"
)

// Stream implements the transport.Stream interface
type Stream struct {
	stream quic.Stream
}

func (s *Stream) Write(b []byte) (n int, err error) {
	return s.stream.Write(b)
}

func (s *Stream) Close() error {
	return s.stream.Close()
}

func (s *Stream) TryWrite(b []byte) (n int, err error) {
	// QUIC streams don't have a non-blocking write, so we just do a regular write
	return s.Write(b)
}
