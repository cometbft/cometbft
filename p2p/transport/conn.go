package transport

import (
	"io"
	"net"
	"time"
)

// Conn is a multiplexed connection that can send and receive data
// on multiple streams.
type Conn interface {
	// OpenStream opens a new stream on the connection with an optional
	// description. If you're using tcp.MultiplexTransport, all streams must be
	// registered in advance.
	OpenStream(streamID byte, desc any) (Stream, error)

	// LocalAddr returns the local network address, if known.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address, if known.
	RemoteAddr() net.Addr

	// Close closes the connection.
	// If the protocol supports it, a reason will be sent to the remote.
	// Any blocked Read operations will be unblocked and return errors.
	Close(reason string) error

	// FlushAndClose flushes all the pending bytes and closes the connection.
	// If the protocol supports it, a reason will be sent to the remote.
	// Any blocked Read operations will be unblocked and return errors.
	FlushAndClose(reason string) error

	// ConnState returns basic details about the connection.
	// Warning: This API should not be considered stable and might change soon.
	ConnState() ConnState

	// ErrorCh returns a channel that will receive errors from the connection.
	ErrorCh() <-chan error

	// HandshakeStream returns the stream to be used for the handshake.
	HandshakeStream() HandshakeStream
}

// Stream is the interface implemented by QUIC streams or multiplexed TCP connection.
type Stream interface {
	SendStream
}

// A SendStream is a unidirectional Send Stream.
type SendStream interface {
	// Write writes data to the stream.
	// It blocks until data is sent or the stream is closed.
	io.Writer
	// Close closes the write-direction of the stream.
	// Future calls to Write are not permitted after calling Close.
	// It must not be called concurrently with Write.
	// It must not be called after calling CancelWrite.
	io.Closer
	// TryWrite attempts to write data to the stream.
	// If the send queue is full, the error satisfies the WriteError interface, and Full() will be true.
	TryWrite(b []byte) (n int, err error)
}

// WriteError is returned by TryWrite when the send queue is full.
type WriteError interface {
	error
	Full() bool // Is the error due to the send queue being full?
}

// HandshakeStream is a stream that is used for the handshake.
type HandshakeStream interface {
	SetDeadline(t time.Time) error
	io.ReadWriter
}
