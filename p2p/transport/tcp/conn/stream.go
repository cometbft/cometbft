package conn

import (
	"sync"
	"time"
)

// MCConnectionStream is just a wrapper around the original net.Conn.
type MConnectionStream struct {
	conn        *MConnection
	streamID    byte
	unreadBytes []byte

	mtx          sync.RWMutex
	deadline     time.Time
	readDeadline time.Time
}

// Read reads whole messages from the internal map, which is populated by the
// global read loop.
// thread-safe.
func (s *MConnectionStream) Read(b []byte) (n int, err error) {
	// If there are unread bytes, read them first.
	if len(s.unreadBytes) > 0 {
		n = copy(b, s.unreadBytes)
		if n < len(s.unreadBytes) {
			s.unreadBytes = s.unreadBytes[n:]
		} else {
			s.unreadBytes = nil
		}
		return n, nil
	}

	s.conn.mtx.RLock()
	ch, ok := s.conn.recvMsgsByStreamID[s.streamID]
	s.conn.mtx.RUnlock()

	// If there are messages to read, read them.
	if ok {
		readTimeout := s.readTimeout()
		if readTimeout > 0 { // read with timeout
			select {
			case msgBytes := <-ch:
				n = copy(b, msgBytes)
				if n < len(msgBytes) {
					s.unreadBytes = msgBytes[n:]
				}
				return n, nil
			case <-s.conn.Quit():
				return 0, nil
			case <-time.After(readTimeout):
				return 0, ErrTimeout
			}
		}

		// read without timeout
		select {
		case msgBytes := <-ch:
			n = copy(b, msgBytes)
			if n < len(msgBytes) {
				s.unreadBytes = msgBytes[n:]
			}
			return n, nil
		case <-s.conn.Quit():
			return 0, nil
		}
	}

	// No messages to read.
	return 0, nil
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

// Close does nothing.
// thread-safe.
func (s *MConnectionStream) Close() error {
	s.conn.mtx.Lock()
	delete(s.conn.recvMsgsByStreamID, s.streamID)
	delete(s.conn.channelsIdx, s.streamID)
	s.conn.mtx.Unlock()
	return nil
}

// SetDeadline sets both the read and write deadlines for this stream. It does not set the
// read nor write deadline on the underlying TCP connection! A zero value for t means
// Conn.Read and Conn.Write will not time out.
//
// Only applies to new reads and writes.
// thread-safe.
func (s *MConnectionStream) SetDeadline(t time.Time) error {
	s.mtx.Lock()
	s.deadline = t
	s.mtx.Unlock()
	return nil
}

// SetReadDeadline sets the read deadline for this stream. It does not set the
// read deadline on the underlying TCP connection! A zero value for t means
// Conn.Read will not time out.
//
// Only applies to new reads.
// thread-safe.
func (s *MConnectionStream) SetReadDeadline(t time.Time) error {
	s.mtx.Lock()
	s.readDeadline = t
	s.mtx.Unlock()
	return nil
}

func (s *MConnectionStream) readTimeout() time.Duration {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	now := time.Now()
	switch {
	case s.readDeadline.IsZero() && s.deadline.IsZero():
		return 0
	case s.readDeadline.After(now):
		return s.readDeadline.Sub(now)
	case s.deadline.After(now):
		return s.deadline.Sub(now)
	}
	return 0
}
