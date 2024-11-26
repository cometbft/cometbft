package transport

import (
	"io"
	"net"
	"time"
)

// Connection is a multiplexed connection that can send and receive data
// on multiple streams.
type Connection interface {
	// OpenStream opens a new stream on the connection with an optional
	// description.
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

	// ConnectionState returns basic details about the connection.
	// Warning: This API should not be considered stable and might change soon.
	ConnectionState() any

	// ErrorCh returns a channel that will receive errors from the connection.
	ErrorCh() <-chan error
}

// Stream is the interface implemented by QUIC streams or multiplexed TCP connection.
type Stream interface {
	ReceiveStream
	SendStream
	// SetDeadline sets the read and write deadlines associated with the connection. It is equivalent to calling both
	// SetReadDeadline and SetWriteDeadline.
	SetDeadline(t time.Time) error
}

// A ReceiveStream is a unidirectional Receive Stream.
type ReceiveStream interface {
	// Read reads data from the stream.
	// Read can be made to time out and return a net.Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetReadDeadline.
	// If the stream was canceled by the peer, the error is a StreamError and
	// Remote == true.
	// If the connection was closed due to a timeout, the error satisfies
	// the net.Error interface, and Timeout() will be true.
	io.Reader
	// SetReadDeadline sets the deadline for future Read calls and
	// any currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(t time.Time) error
}

// A SendStream is a unidirectional Send Stream.
type SendStream interface {
	// Write writes data to the stream.
	// Write can be made to time out and return a net.Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetWriteDeadline.
	// If the stream was canceled by the peer, the error is a StreamError and
	// Remote == true.
	// If the connection was closed due to a timeout, the error satisfies
	// the net.Error interface, and Timeout() will be true.
	io.Writer
	// Close closes the write-direction of the stream.
	// Future calls to Write are not permitted after calling Close.
	// It must not be called concurrently with Write.
	// It must not be called after calling CancelWrite.
	io.Closer
	// SetWriteDeadline sets the deadline for future Write calls
	// and any currently-blocked Write call.
	// Even if write times out, it may return n > 0, indicating that
	// some data was successfully written.
	// A zero value for t means Write will not time out.
	SetWriteDeadline(t time.Time) error
}
