package protoio

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// zeroThenDataReader returns (0, nil) once -- legal per the io.Reader
// contract, though discouraged -- before returning real data. Real-world
// readers that wrap rate limiters, pipes, or buffering layers can do this.
type zeroThenDataReader struct {
	data     []byte
	pos      int
	toldZero bool
}

func (r *zeroThenDataReader) Read(p []byte) (int, error) {
	if !r.toldZero {
		r.toldZero = true
		return 0, nil
	}
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// TestByteReaderIgnoresZeroByteNilErrorReturn pins that a single (0, nil)
// return from the wrapped io.Reader is treated as "nothing happened yet",
// not as a successfully-read (and fabricated) zero byte.
func TestByteReaderIgnoresZeroByteNilErrorReturn(t *testing.T) {
	src := &zeroThenDataReader{data: []byte{0x2a}}
	br := newByteReader(src)

	b, err := br.ReadByte()
	require.NoError(t, err)
	assert.Equal(t, byte(0x2a), b, "ReadByte must return the real data byte, not a zero fabricated from an unwritten buffer")
	assert.Equal(t, 1, br.bytesRead, "bytesRead must count only the byte that was actually read")
}

// dataWithEOFReader returns its one remaining byte together with io.EOF in
// the same Read call -- explicitly legal per the io.Reader doc ("It may
// return the (non-nil) error from the same call").
type dataWithEOFReader struct {
	b    byte
	done bool
}

func (r *dataWithEOFReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	r.done = true
	p[0] = r.b
	return 1, io.EOF
}

// TestByteReaderDoesNotDropByteReturnedAlongsideEOF pins that a byte
// returned in the same call as io.EOF is not discarded -- it must be
// surfaced to the caller before the stream is treated as exhausted.
func TestByteReaderDoesNotDropByteReturnedAlongsideEOF(t *testing.T) {
	src := &dataWithEOFReader{b: 0x7f}
	br := newByteReader(src)

	b, err := br.ReadByte()
	require.NoError(t, err, "a byte delivered alongside io.EOF must still be returned, not discarded as an error")
	assert.Equal(t, byte(0x7f), b)
	assert.Equal(t, 1, br.bytesRead)

	// The stream is now genuinely exhausted.
	_, err = br.ReadByte()
	assert.ErrorIs(t, err, io.EOF)
}
