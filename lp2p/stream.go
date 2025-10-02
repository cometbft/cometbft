package lp2p

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
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

	payload, err := readExactly(reader, payloadSize)
	if err != nil {
		return nil, err
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

// readExactly allocates & reads exactly $size bytes from the reader.
func readExactly(r io.Reader, size uint64) ([]byte, error) {
	var (
		out       = make([]byte, size)
		bytesRead uint64
		n         int
		err       error
		eof       bool
	)

	for {
		n, err = r.Read(out[bytesRead:])
		eof = errors.Is(err, io.EOF)

		bytesRead += uint64(n)

		switch {
		case eof && bytesRead == size:
			// no more bytes to read and size matches => all good!
			return out, nil
		case eof && bytesRead != size:
			// no more bytes to read, but size doesn't match => partial read
			return nil, errors.Wrapf(err, "eof partial read (%d/%d bytes)", bytesRead, size)
		case err != nil:
			// just some error
			return nil, errors.Wrapf(err, "failed to read payload (read %d/%d bytes)", bytesRead, size)
		case bytesRead < size:
			// not enough bytes to read => continue
			continue
		case bytesRead > size:
			// should not happen
			return nil, errors.Errorf("read more bytes than expected (%d/%d bytes)", bytesRead, size)
		default:
			// all good!
			return out, nil
		}
	}
}

func closeStream(s network.Stream) error {
	err := s.Close()
	switch {
	case isErrCancelled(err):
		// expected if peer canceled the stream
	case err != nil:
		return errors.Wrap(err, "failed to close stream")
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
