// Package protowire provides a minimal, allocation-free reader for walking
// raw protobuf wire bytes. Every read is bounds-checked and advances a cursor.
package protowire

import (
	"errors"
	"fmt"
)

// Wire types (the low 3 bits of a field tag).
const (
	WireVarint  = 0
	WireFixed64 = 1
	WireBytes   = 2
	WireFixed32 = 5
)

var (
	// ErrVarintOverflow is returned when a varint does not terminate within
	// the 10 bytes that can hold a 64-bit value.
	ErrVarintOverflow = errors.New("varint overflow")
	// ErrTruncatedVarint is returned when the buffer ends mid-varint.
	ErrTruncatedVarint = errors.New("truncated varint")
	// ErrOutOfBounds is returned when a length prefix or fixed-width field
	// would run past the end of the buffer.
	ErrOutOfBounds = errors.New("length out of bounds")
	// ErrIllegalFieldNumber is returned when a tag decodes to a field number
	// that is not a legal protobuf field number (>= 1).
	ErrIllegalFieldNumber = errors.New("illegal field number")
	// ErrUnsupportedWireType is returned when a tag carries a wire type the
	// cursor does not know how to skip.
	ErrUnsupportedWireType = errors.New("unsupported wire type")
)

// WireCursor walks a protobuf buffer left to right. Every read is bounds-checked
// and advances the cursor; a read past the end returns an error rather than
// panicking.
type WireCursor struct {
	buf []byte
	pos int
}

// NewWireCursor returns a cursor positioned at the start of buf.
func NewWireCursor(buf []byte) WireCursor {
	return WireCursor{buf: buf}
}

// AtEnd reports whether the cursor has consumed the whole buffer.
func (c *WireCursor) AtEnd() bool { return c.pos >= len(c.buf) }

// ReadVarint decodes a base-128 varint at the cursor and advances past it.
func (c *WireCursor) ReadVarint() (uint64, error) {
	var v uint64
	for shift := uint(0); ; shift += 7 {
		if shift >= 64 {
			return 0, ErrVarintOverflow
		}
		if c.pos >= len(c.buf) {
			return 0, ErrTruncatedVarint
		}
		b := c.buf[c.pos]
		c.pos++
		v |= uint64(b&0x7f) << shift
		if b < 0x80 {
			return v, nil
		}
	}
}

// ReadTag decodes a field tag, splitting it into field number and wire type.
func (c *WireCursor) ReadTag() (fieldNum, wireType int, err error) {
	v, err := c.ReadVarint()
	if err != nil {
		return 0, 0, err
	}
	fieldNum = int(int32(v >> 3))
	wireType = int(v & 0x7)
	if fieldNum <= 0 {
		return 0, 0, fmt.Errorf("%w: %d", ErrIllegalFieldNumber, fieldNum)
	}
	return fieldNum, wireType, nil
}

// ReadLengthDelimited reads a length-prefixed run of bytes (wire type 2) and
// returns it as a sub-slice of the underlying buffer, advancing past it.
func (c *WireCursor) ReadLengthDelimited() ([]byte, error) {
	length, err := c.ReadVarint()
	if err != nil {
		return nil, err
	}
	start := c.pos
	end := start + int(length)
	if end < start || end > len(c.buf) {
		return nil, ErrOutOfBounds
	}
	c.pos = end
	return c.buf[start:end], nil
}

// SkipField advances the cursor past the body of a field with the given wire
// type, used to ignore fields the caller does not care about.
func (c *WireCursor) SkipField(wireType int) error {
	switch wireType {
	case WireVarint:
		_, err := c.ReadVarint()
		return err
	case WireFixed64:
		return c.advance(8)
	case WireBytes:
		_, err := c.ReadLengthDelimited()
		return err
	case WireFixed32:
		return c.advance(4)
	default:
		return fmt.Errorf("%w: %d", ErrUnsupportedWireType, wireType)
	}
}

// RepeatedBytesEntrySize returns the number of wire bytes proto encodes for one
// entry in a repeated bytes field: 1-byte field tag + varint(dataLen).
// dataLen must be >= 0; Go's len() guarantees this for all practical callers.
func RepeatedBytesEntrySize(dataLen int) int {
	if dataLen < 0 {
		panic("protowire: negative dataLen")
	}
	n := 1 /* tag */ + 1 /* min varint byte */
	v := uint64(dataLen) >> 7
	for v > 0 {
		n++
		v >>= 7
	}
	return n
}

// advance moves the cursor forward by n bytes, bounds-checked.
func (c *WireCursor) advance(n int) error {
	end := c.pos + n
	if end < c.pos || end > len(c.buf) {
		return ErrOutOfBounds
	}
	c.pos = end
	return nil
}
