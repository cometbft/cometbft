package conn

import "time"

// MCConnectionStream is just a wrapper around the original net.Conn.
type MConnectionStream struct {
	conn     *MConnection
	streamID byte
}

// Read reads bytes for the given stream from the internal read queue. Used in
// tests. Production code should use MConnection.OnReceive to avoid copying the
// data.
func (s *MConnectionStream) Read(b []byte) (n int, err error) {
	return s.conn.readBytes(s.streamID, b, 5*time.Second)
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
