package protowire

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWireCursor_ReadVarint(t *testing.T) {
	// 300 encodes as 0xac 0x02 in base-128 varint.
	c := NewWireCursor([]byte{0xac, 0x02})
	v, err := c.ReadVarint()
	require.NoError(t, err)
	require.Equal(t, uint64(300), v)
	require.True(t, c.AtEnd())
}

func TestWireCursor_ReadVarint_Truncated(t *testing.T) {
	c := NewWireCursor([]byte{0xff})
	_, err := c.ReadVarint()
	require.ErrorIs(t, err, ErrTruncatedVarint)
}

func TestWireCursor_ReadVarint_Overflow(t *testing.T) {
	// 11 continuation bytes never terminate within 64 bits.
	c := NewWireCursor([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	_, err := c.ReadVarint()
	require.ErrorIs(t, err, ErrVarintOverflow)
}

func TestWireCursor_ReadTag(t *testing.T) {
	// Field 1, wire type 2 (length-delimited) => tag byte 0x0a.
	c := NewWireCursor([]byte{0x0a})
	fieldNum, wireType, err := c.ReadTag()
	require.NoError(t, err)
	require.Equal(t, 1, fieldNum)
	require.Equal(t, WireBytes, wireType)
}

func TestWireCursor_ReadTag_IllegalFieldNumber(t *testing.T) {
	// Tag 0x00 => field number 0, which is illegal.
	c := NewWireCursor([]byte{0x00})
	_, _, err := c.ReadTag()
	require.ErrorIs(t, err, ErrIllegalFieldNumber)
}

func TestWireCursor_ReadLengthDelimited(t *testing.T) {
	c := NewWireCursor([]byte{0x03, 'a', 'b', 'c'})
	b, err := c.ReadLengthDelimited()
	require.NoError(t, err)
	require.Equal(t, []byte("abc"), b)
	require.True(t, c.AtEnd())
}

func TestWireCursor_ReadLengthDelimited_OutOfBounds(t *testing.T) {
	// Declares length 5 but only 1 byte follows.
	c := NewWireCursor([]byte{0x05, 'a'})
	_, err := c.ReadLengthDelimited()
	require.ErrorIs(t, err, ErrOutOfBounds)
}

func TestWireCursor_SkipField(t *testing.T) {
	testCases := []struct {
		name     string
		buf      []byte
		wireType int
	}{
		{"varint", []byte{0xac, 0x02}, WireVarint},
		{"fixed64", []byte{1, 2, 3, 4, 5, 6, 7, 8}, WireFixed64},
		{"bytes", []byte{0x02, 'x', 'y'}, WireBytes},
		{"fixed32", []byte{1, 2, 3, 4}, WireFixed32},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewWireCursor(tc.buf)
			require.NoError(t, c.SkipField(tc.wireType))
			require.True(t, c.AtEnd())
		})
	}
}

func TestWireCursor_SkipField_Unsupported(t *testing.T) {
	c := NewWireCursor([]byte{0x00})
	// Wire type 3 (start group) is not supported.
	require.ErrorIs(t, c.SkipField(3), ErrUnsupportedWireType)
}

func TestWireCursor_SkipField_OutOfBounds(t *testing.T) {
	// Fixed64 needs 8 bytes; only 2 are present.
	c := NewWireCursor([]byte{1, 2})
	require.ErrorIs(t, c.SkipField(WireFixed64), ErrOutOfBounds)
}

func TestRepeatedBytesEntrySize(t *testing.T) {
	// Verified against proto.Size on Txs{Txs: [][]byte{make([]byte, n)}}.
	for _, tt := range []struct {
		dataLen int
		want    int
	}{
		{0, 2},   // tag(1) + varint(0)=1
		{1, 2},   // tag(1) + varint(1)=1
		{127, 2}, // last 1-byte varint
		{128, 3}, // first 2-byte varint
		{1000, 3},
		{16383, 3}, // last 2-byte varint
		{16384, 4}, // first 3-byte varint
		{1048576, 4},
	} {
		require.Equal(t, tt.want, RepeatedBytesEntrySize(tt.dataLen), "dataLen=%d", tt.dataLen)
	}
}
