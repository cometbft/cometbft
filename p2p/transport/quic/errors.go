package quic

import "errors"

var (
	// ErrTransportNotListening is returned when trying to accept connections before listening
	ErrTransportNotListening = errors.New("transport not listening")

	// ErrTransportClosed is returned when the transport has been closed
	ErrTransportClosed = errors.New("transport closed")

	// ErrInvalidAddress is returned when an invalid address is provided
	ErrInvalidAddress = errors.New("invalid address")
)
