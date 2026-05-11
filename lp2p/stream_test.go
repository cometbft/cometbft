package lp2p

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"testing"

	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/require"
)

func TestStream(t *testing.T) {
	t.Run("StreamWrite", func(t *testing.T) {
		t.Run("Write", func(t *testing.T) {
			// ARRANGE
			suite := newStreamTestSuite(t)

			payload := []byte("hello stream")
			actualPayload := make(chan []byte, 1)
			readErr := make(chan error, 1)

			suite.hostA.SetStreamHandler(suite.protoID, func(stream network.Stream) {
				bz, err := StreamReadClose(stream)
				if err != nil {
					readErr <- err
					return
				}

				actualPayload <- bz
				readErr <- nil
			})

			stream := suite.newStream(t)

			expectedFrame := append(uint64ToUvarint(uint64(len(payload))), payload...)

			// ACT
			bytesWritten, err := StreamWrite(stream, payload)
			require.NoError(t, err)

			// ASSERT
			require.Equal(t, len(expectedFrame), bytesWritten)
			require.NoError(t, <-readErr)
			require.Equal(t, payload, <-actualPayload)
		})

		t.Run("Noop", func(t *testing.T) {
			// ARRANGE
			writeCalls := 0
			stream := &streamStub{
				conn: &connStub{closed: false},
				writeFn: func(p []byte) (int, error) {
					writeCalls++
					return len(p), nil
				},
			}

			// ACT
			bytesWritten, err := StreamWrite(stream, nil)

			// ASSERT
			require.NoError(t, err)
			require.Equal(t, 0, bytesWritten)
			require.Equal(t, 0, writeCalls)
		})

		t.Run("ClosedStream", func(t *testing.T) {
			// ARRANGE
			stream := &streamStub{
				conn: &connStub{closed: true},
				writeFn: func(_ []byte) (int, error) {
					t.Fatal("write must not be called when conn is closed")
					return 0, nil
				},
			}

			// ACT
			bytesWritten, err := StreamWrite(stream, []byte("ignored"))

			// ASSERT
			require.Error(t, err)
			require.ErrorContains(t, err, "stream is closed")
			require.Equal(t, 0, bytesWritten)
		})

		t.Run("PartialWrite", func(t *testing.T) {
			// ARRANGE
			payload := []byte("abcde")
			expectedFrameLen := len(uint64ToUvarint(uint64(len(payload)))) + len(payload)

			stream := &streamStub{
				conn: &connStub{closed: false},
				writeFn: func(_ []byte) (int, error) {
					return 2, io.ErrClosedPipe
				},
			}

			// ACT
			bytesWritten, err := StreamWrite(stream, payload)

			// ASSERT
			require.Equal(t, 2, bytesWritten)
			require.Error(t, err)
			require.ErrorContains(t, err, "failed to write payload")
			require.ErrorContains(t, err, "(2/"+strconv.Itoa(expectedFrameLen)+" sent)")
		})
	})

	t.Run("StreamReadSized", func(t *testing.T) {
		t.Run("Read", func(t *testing.T) {
			// ARRANGE
			payload := []byte("read me")
			frame := append(uint64ToUvarint(uint64(len(payload))), payload...)
			reader := io.MultiReader(
				bytes.NewReader(frame[:1]),
				bytes.NewReader(frame[1:2]),
				bytes.NewReader(frame[2:]),
			)

			stream := &streamStub{
				conn:   &connStub{closed: false},
				readFn: reader.Read,
			}

			// ACT
			out, err := StreamReadSized(stream, 1024)

			// ASSERT
			require.NoError(t, err)
			require.Equal(t, payload, out)
		})

		t.Run("ClosedStream", func(t *testing.T) {
			// ARRANGE
			stream := &streamStub{
				conn: &connStub{closed: true},
			}

			// ACT
			out, err := StreamReadSized(stream, 1024)

			// ASSERT
			require.Error(t, err)
			require.ErrorContains(t, err, "stream is closed")
			require.Nil(t, out)
		})

		t.Run("TooLargePayload", func(t *testing.T) {
			// ARRANGE
			const maxSize = uint64(3)
			headerOnly := uint64ToUvarint(maxSize + 1)

			stream := &streamStub{
				conn:   &connStub{closed: false},
				readFn: bytes.NewReader(headerOnly).Read,
			}

			// ACT
			out, err := StreamReadSized(stream, maxSize)

			// ASSERT
			require.Error(t, err)
			require.ErrorContains(t, err, "payload is too large")
			require.Nil(t, out)
		})

		t.Run("InvalidHeader", func(t *testing.T) {
			// ARRANGE
			invalidHeader := []byte{0x80}
			stream := &streamStub{
				conn:   &connStub{closed: false},
				readFn: bytes.NewReader(invalidHeader).Read,
			}

			// ACT
			out, err := StreamReadSized(stream, 1024)

			// ASSERT
			require.Error(t, err)
			require.ErrorContains(t, err, "failed to read payload size")
			require.Nil(t, out)
		})

		t.Run("TruncatedPayload", func(t *testing.T) {
			// ARRANGE
			header := uint64ToUvarint(4)
			truncatedPayload := []byte("ab")
			reader := io.MultiReader(bytes.NewReader(header), bytes.NewReader(truncatedPayload))

			stream := &streamStub{
				conn:   &connStub{closed: false},
				readFn: reader.Read,
			}

			// ACT
			out, err := StreamReadSized(stream, 1024)

			// ASSERT
			require.Error(t, err)
			require.ErrorContains(t, err, "eof partial read")
			require.Nil(t, out)
		})
	})
}

func TestProtocolID(t *testing.T) {
	for _, tt := range []struct {
		channel  byte
		expected string
	}{
		{channel: 0x00, expected: "/p2p/cometbft/1.0.0/channel/0x00"},
		{channel: 0x01, expected: "/p2p/cometbft/1.0.0/channel/0x01"},
		{channel: 0x10, expected: "/p2p/cometbft/1.0.0/channel/0x10"},
		{channel: 0xff, expected: "/p2p/cometbft/1.0.0/channel/0xff"},
	} {
		require.Equal(t, protocol.ID(tt.expected), ProtocolID(tt.channel))
	}
}

func TestStreamReadSizedClose(t *testing.T) {
	t.Run("ReadTooLargePayload", func(t *testing.T) {
		// ARRANGE
		const maxSize = 100

		var (
			ctx     = context.Background()
			protoID = ProtocolID(0xAA)
			ports   = utils.GetFreePorts(t, 2)
			host1   = makeTestHost(t, ports[0], withLogging())
			host2   = makeTestHost(t, ports[1], withLogging())
		)

		// connect hosts
		require.NoError(t, host2.Connect(ctx, host1.AddrInfo()))

		readErr := make(chan error, 1)
		host1.SetStreamHandler(protoID, func(stream network.Stream) {
			defer stream.Close()

			_, err := StreamReadSizedClose(stream, maxSize)
			readErr <- err
		})

		// create stream
		stream, err := host2.NewStream(ctx, host1.ID(), protoID)
		require.NoError(t, err)
		t.Cleanup(func() { _ = stream.Close() })

		tooLargeHeader := uint64ToUvarint(maxSize + 1)

		// ACT
		_, err = stream.Write(tooLargeHeader)
		require.NoError(t, err)
		_ = stream.Close()

		// ASSERT
		err = <-readErr

		require.Error(t, err)
		require.ErrorContains(t, err, "payload is too large")
	})
}

type streamTestSuite struct {
	ctx     context.Context
	protoID protocol.ID
	hostA   *Host
	hostB   *Host
}

type connStub struct {
	network.Conn
	closed bool
}

func newStreamTestSuite(t *testing.T) *streamTestSuite {
	t.Helper()

	ports := utils.GetFreePorts(t, 2)

	hostA := makeTestHost(t, ports[0], withLogging())
	hostB := makeTestHost(t, ports[1], withLogging())
	protoID := ProtocolID(0xAA)

	// Register a default handler on both sides so protocol negotiation succeeds
	// even in tests that do not care about receiving data.
	noop := func(stream network.Stream) { _ = stream.Close() }
	hostA.SetStreamHandler(protoID, noop)
	hostB.SetStreamHandler(protoID, noop)

	ctx := context.Background()
	require.NoError(t, hostB.Connect(ctx, hostA.AddrInfo()))

	return &streamTestSuite{
		ctx:     ctx,
		protoID: protoID,
		hostA:   hostA,
		hostB:   hostB,
	}
}

// newStream creates a new stream from hostB to hostA.
func (ts *streamTestSuite) newStream(t *testing.T) network.Stream {
	t.Helper()

	stream, err := ts.hostB.NewStream(ts.ctx, ts.hostA.ID(), ts.protoID)
	require.NoError(t, err)

	return stream
}

type streamStub struct {
	network.Stream

	conn    network.Conn
	readFn  func([]byte) (int, error)
	writeFn func([]byte) (int, error)
}

func (s *streamStub) Conn() network.Conn {
	if s.conn != nil {
		return s.conn
	}

	return s.Stream.Conn()
}

func (s *streamStub) Read(p []byte) (int, error) {
	if s.readFn != nil {
		return s.readFn(p)
	}

	return s.Stream.Read(p)
}

func (s *streamStub) Write(p []byte) (int, error) {
	if s.writeFn != nil {
		return s.writeFn(p)
	}

	return s.Stream.Write(p)
}

func (c *connStub) IsClosed() bool { return c.closed }
