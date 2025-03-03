package conn

import "time"

// MCConnectionStream is just a wrapper around the original net.Conn.
type MConnectionStream struct {
	conn     *MConnection
	streamID byte

	readTimeout time.Duration
}

// Read reads bytes for the given stream from the internal read queue. Used in
// tests. Production code should use MConnection.OnReceive to avoid copying the
// data.
func (s *MConnectionStream) Read(b []byte) (n int, err error) {
	timeout := s.readTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return s.conn.readBytes(s.streamID, b, timeout)
}

// Write queues bytes to be sent onto the internal write queue.
// thread-safe.
func (s *MConnectionStream) Write(b []byte) (n int, err error) {
	if err := s.conn.sendBytes(s.streamID, b, true /* blocking */); err != nil {
		return 0, err
	}
	return len(b), nil
}

// TryWrite queues bytes to be sent onto the internal write queue.
// thread-safe.
func (s *MConnectionStream) TryWrite(b []byte) (n int, err error) {
	if err := s.conn.sendBytes(s.streamID, b, false /* non-blocking */); err != nil {
		return 0, err
	}
	return len(b), nil
}

// Close closes the stream.
// thread-safe.
func (s *MConnectionStream) Close() error {
	delete(s.conn.channelsIdx, s.streamID)
	return nil
}

// SetReadDeadline sets the deadline for future Read calls. Used only in tests.
func (s *MConnectionStream) SetReadDeadline(t time.Time) error {
	s.readTimeout = time.Until(t)
	return nil
}

// SetDeadline is analog of calling SetReadDeadline.
func (s *MConnectionStream) SetDeadline(t time.Time) error {
	return s.SetReadDeadline(t)
}
