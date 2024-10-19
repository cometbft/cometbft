package conn

import (
	"sync"
	"time"
)

// MCConnectionStream is just a wrapper around the original net.Conn.
type MConnectionStream struct {
	conn        *MConnection
	streadID    byte
	unreadBytes []byte

	mtx           sync.RWMutex
	deadline      time.Time
	readDeadline  time.Time
	writeDeadline time.Time
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
	ch, ok := s.conn.recvMsgsByStreamID[s.streadID]
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
				return 0, ErrNotRunning
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
			return 0, ErrNotRunning
		}
	}

	// No messages to read.
	return 0, nil
}

// Write queues bytes to be sent onto the internal write queue. It returns
// len(b), but it doesn't guarantee that the Write actually succeeds.
// thread-safe.
func (s *MConnectionStream) Write(b []byte) (n int, err error) {
	if err := s.conn.sendBytes(s.streadID, b, s.writeTimeout()); err != nil {
		return 0, err
	}
	return len(b), nil
}

// Close does nothing.
// thread-safe.
func (s *MConnectionStream) Close() error {
	s.conn.mtx.Lock()
	delete(s.conn.recvMsgsByStreamID, s.streadID)
	delete(s.conn.channelsIdx, s.streadID)
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

// SetWriteDeadline sets the write deadline for this stream. It does not set the
// write deadline on the underlying TCP connection! A zero value for t means
// Conn.Write will not time out.
//
// Only applies to new writes.
// thread-safe.
func (s *MConnectionStream) SetWriteDeadline(t time.Time) error {
	s.mtx.Lock()
	s.writeDeadline = t
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

func (s *MConnectionStream) writeTimeout() time.Duration {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	now := time.Now()
	switch {
	case s.writeDeadline.IsZero() && s.deadline.IsZero():
		return 0
	case s.writeDeadline.After(now):
		return s.writeDeadline.Sub(now)
	case s.deadline.After(now):
		return s.deadline.Sub(now)
	}
	return 0
}
