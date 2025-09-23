package lp2p

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/pkg/errors"
)

// ProtocolIDPrefix is the prefix for all protocol IDs.
const ProtocolIDPrefix = "/p2p/cometbft/1.0.0"

// TimeoutStream is the timeout for a stream.
const TimeoutStream = 10 * time.Second

// ProtocolID returns the protocol ID for a given channel
// Byte is used for compatibility with the original CometBFT implementation.
func ProtocolID(channelID byte) protocol.ID {
	return protocol.ID(
		fmt.Sprintf("%s/channel/0x%02x", ProtocolIDPrefix, channelID),
	)
}

// StreamWrite sends payload over a stream w/o waiting for a response.
// Only guarantees that the recipient will receive the bytes (no "message processed" guarantee).
// It doesn't control stream's lifecycle, so it's up to the caller to close the stream.
func StreamWrite(s network.Stream, data []byte) (int, error) {
	switch {
	case len(data) == 0:
		// noop
		return 0, nil
	case s.Conn().IsClosed():
		return 0, fmt.Errorf("stream is closed")
	}

	// [header(content_len) | payload]
	var (
		header  = uint64ToUvarint(uint64(len(data)))
		payload = append(header, data...)
	)

	bytesWritten, err := s.Write(payload)
	if err != nil {
		err = errors.Wrapf(err, "failed to write payload (%d/%d sent)", bytesWritten, len(payload))
	}

	return bytesWritten, err
}

// StreamWriteClose sends payload over a stream and closes it right after.
// The caller doesn't expect a response in this case.
// Also, resets the stream on both ends in case of error.
func StreamWriteClose(s network.Stream, data []byte) (err error) {
	defer func() {
		if err != nil {
			// nukes broken stream on both ends
			_ = s.Reset()
		}
	}()

	// todo timeouts, size limits, etc...

	if _, err := StreamWrite(s, data); err != nil {
		return errors.Wrap(err, "send failed")
	}

	if err := closeStream(s); err != nil {
		return errors.Wrap(err, "closeStream")
	}

	return nil
}

// StreamRead reads payload from a stream.
// It doesn't control stream's lifecycle, so it's up to the caller to close the stream.
func StreamRead(s network.Stream) ([]byte, error) {
	if s.Conn().IsClosed() {
		return nil, fmt.Errorf("stream is closed")
	}

	reader := bufio.NewReader(s)

	payloadSize, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read payload size")
	}

	payload := make([]byte, payloadSize)

	bytesRead, err := reader.Read(payload)
	switch {
	case err != nil:
		return nil, errors.Wrapf(err, "failed to read payload (read %d/%d bytes)", bytesRead, payloadSize)
	case uint64(bytesRead) != payloadSize:
		return nil, errors.Errorf("partial read (%d/%d bytes)", bytesRead, payloadSize)
	}

	return payload, nil
}

// StreamReadClose reads payload from a stream and closes it right after.
// Also, resets the stream on both ends in case of error.
func StreamReadClose(s network.Stream) (payload []byte, err error) {
	defer func() {
		if err != nil {
			// nukes broken stream on both ends
			_ = s.Reset()
		}
	}()

	payload, err = StreamRead(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read payload")
	}

	if err := closeStream(s); err != nil {
		return nil, errors.Wrap(err, "closeStream")
	}

	return payload, nil
}

func closeStream(s network.Stream) error {
	errCloseWrite := s.CloseWrite()
	switch {
	case isErrCancelled(errCloseWrite):
		// expected if peer canceled the stream
		return nil
	case errCloseWrite != nil:
		return errors.Wrap(errCloseWrite, "failed to close stream for write")
	}

	errCloseRead := s.CloseRead()
	switch {
	case isErrCancelled(errCloseRead):
		// expected if peer canceled the stream
	case errCloseRead != nil:
		return errors.Wrap(errCloseRead, "failed to close stream for read")
	}

	return nil
}

// go-libp2p doesn't have a sentinel error for this!
func isErrCancelled(err error) bool {
	if err == nil {
		return false
	}

	const pattern = "close called for canceled stream"

	return strings.Contains(err.Error(), pattern)
}

func uint64ToUvarint(len uint64) []byte {
	out := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(out, len)

	return out[:bytesWritten]
}
