package mconn

// MStream represents a multiplexed stream within an MConnection
type MStream struct {
	conn *MConnection
	chID byte
}

func (s *MStream) Write(b []byte) (n int, err error) {
	err = s.conn.Send(s.chID, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (s *MStream) Close() error {
	// Remove channel from connection
	delete(s.conn.channels, s.chID)
	return nil
}

func (s *MStream) TryWrite(b []byte) (n int, err error) {
	err = s.conn.TrySend(s.chID, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}
