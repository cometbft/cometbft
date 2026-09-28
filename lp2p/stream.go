package lp2p

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"slices"
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

// streamReadIdleTimeout is how long StreamReadSized waits for the next bytes
// of a frame before giving up on the peer. Overridden in tests.
var streamReadIdleTimeout = TimeoutStream

// MaxStreamSize is the global maximum size of a stream.
// Protocols should configure their own maximum size.
const MaxStreamSize = 4 * (1 << 20)

// readChunkSize bounds how much payload is requested from the stream per read,
// so the receive buffer grows with the bytes actually delivered rather than
// with the size the peer declared in the header.
const readChunkSize = 64 * (1 << 10)

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
// Note: this method doesn't enforce any size limits! Use StreamReadSized instead.
func StreamRead(s network.Stream) ([]byte, error) {
	return StreamReadSized(s, MaxStreamSize)
}

// StreamReadSized reads payload from a stream with a maximum size.
// It doesn't control stream's lifecycle, so it's up to the caller to close the stream.
func StreamReadSized(s network.Stream, maxSize uint64) ([]byte, error) {
	if s.Conn().IsClosed() {
		return nil, fmt.Errorf("stream is closed")
	}

	// A peer that stops sending mid-frame must not pin this goroutine (and
	// the payload buffer) for as long as it keeps the stream open. The
	// deadline is extended whenever data arrives, so it bounds idle time,
	// not the total transfer.
	extendDeadline := func() error {
		return s.SetReadDeadline(time.Now().Add(streamReadIdleTimeout))
	}
	if err := extendDeadline(); err != nil {
		return nil, errors.Wrap(err, "failed to set read deadline")
	}

	reader := bufio.NewReader(s)

	// in bytes
	payloadSize, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read payload size")
	}

	// Honor the caller's limit; MaxStreamSize is a fallback, not a hard cap.
	payloadLimit := maxSize
	if payloadLimit == 0 {
		payloadLimit = MaxStreamSize
	}

	if payloadSize > payloadLimit {
		return nil, errors.Errorf("payload is too large (got %d, max %d)", payloadSize, payloadLimit)
	}

	payload, err := readExactly(reader, payloadSize, extendDeadline)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// StreamReadClose reads payload from a stream and closes it right after.
// Also, resets the stream on both ends in case of error.
func StreamReadClose(s network.Stream) (payload []byte, err error) {
	return StreamReadSizedClose(s, MaxStreamSize)
}

// StreamReadSizedClose reads payload from a stream and closes it right after with a maximum size.
// Also, resets the stream on both ends in case of error.
func StreamReadSizedClose(s network.Stream, maxSize uint64) (payload []byte, err error) {
	defer func() {
		if err != nil {
			// nukes broken stream on both ends
			_ = s.Reset()
		}
	}()

	payload, err = StreamReadSized(s, maxSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read payload")
	}

	if err := closeStream(s); err != nil {
		return nil, errors.Wrap(err, "closeStream")
	}

	return payload, nil
}

// readExactly reads exactly $size bytes from the reader. The buffer grows
// with the bytes received, and onProgress is invoked after every read that
// delivered data.
func readExactly(r io.Reader, size uint64, onProgress func() error) ([]byte, error) {
	out := make([]byte, 0, min(size, readChunkSize))

	for bytesRead := uint64(0); bytesRead < size; {
		want := min(size-bytesRead, readChunkSize)
		out = slices.Grow(out, int(want))

		n, err := r.Read(out[bytesRead : bytesRead+want])
		bytesRead += uint64(n)
		out = out[:bytesRead]

		if n > 0 && onProgress != nil {
			if err := onProgress(); err != nil {
				return nil, errors.Wrap(err, "failed to extend read deadline")
			}
		}

		switch {
		case errors.Is(err, io.EOF) && bytesRead == size:
			// no more bytes to read and size matches => all good!
			return out, nil
		case errors.Is(err, io.EOF):
			// no more bytes to read, but size doesn't match => partial read
			return nil, errors.Wrapf(err, "eof partial read (%d/%d bytes)", bytesRead, size)
		case err != nil:
			// just some error
			return nil, errors.Wrapf(err, "failed to read payload (read %d/%d bytes)", bytesRead, size)
		}
	}

	return out, nil
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
