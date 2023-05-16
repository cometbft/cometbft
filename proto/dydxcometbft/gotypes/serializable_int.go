package gotypes

import (
	"encoding"
	"encoding/json"
	"math/big"
)

// SerializableInt is basically copied from cosmos-sdk/types/Int but:
// - doesn’t have a bit-length restriction
// - uses GobEncode/GobDecode instead of serializing to an ascii string
// - removes superfluous functions to do `big.Int` math on the underlying value
type SerializableInt struct {
	i *big.Int
}

// BigInt converts Int to big.Int
func (i SerializableInt) BigInt() *big.Int {
	if i.IsNil() {
		return nil
	}
	return new(big.Int).Set(i.i)
}

// IsNil returns true if Int is uninitialized
func (i SerializableInt) IsNil() bool {
	return i.i == nil
}

// NewInt constructs Int from int64
func NewInt(n int64) SerializableInt {
	return SerializableInt{big.NewInt(n)}
}

// NewIntFromUint64 constructs an Int from a uint64.
func NewIntFromUint64(n uint64) SerializableInt {
	b := big.NewInt(0)
	b.SetUint64(n)
	return SerializableInt{b}
}

// NewIntFromBigInt constructs Int from big.Int. If the provided big.Int is nil,
func NewIntFromBigInt(i *big.Int) SerializableInt {
	if i == nil {
		return SerializableInt{}
	}

	return SerializableInt{i}
}

// ZeroInt returns Int value with zero
func ZeroInt() SerializableInt {
	return SerializableInt{big.NewInt(0)}
}

// Marshal implements the gogo proto custom type interface.
func (i SerializableInt) Marshal() ([]byte, error) {
	i.ensureNonNil()
	return i.i.GobEncode()
}

// MarshalTo implements the gogo proto custom type interface.
func (i *SerializableInt) MarshalTo(data []byte) (n int, err error) {
	bz, err := i.Marshal()
	if err != nil {
		return 0, err
	}

	n = copy(data, bz)
	return n, nil
}

// Unmarshal implements the gogo proto custom type interface.
func (i *SerializableInt) Unmarshal(data []byte) error {
	i.ensureNonNil()

	if err := i.i.GobDecode(data); err != nil {
		return err
	}

	return nil
}

// Size implements the gogo proto custom type interface.
func (i *SerializableInt) Size() int {
	i.ensureNonNil()
	n := i.i.BitLen()
	return 1 + ((n + 7) / 8)
}

// MarshalJSON defines custom encoding scheme
func (i SerializableInt) MarshalJSON() ([]byte, error) {
	i.ensureNonNil()
	return marshalJSON(i.i)
}

// UnmarshalJSON defines custom decoding scheme
func (i *SerializableInt) UnmarshalJSON(bz []byte) error {
	i.ensureNonNil()
	return unmarshalJSON(i.i, bz)
}

// MarshalJSON for custom encoding scheme
// Must be encoded as a string for JSON precision
func marshalJSON(i encoding.TextMarshaler) ([]byte, error) {
	text, err := i.MarshalText()
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(text))
}

// UnmarshalJSON for custom decoding scheme
// Must be encoded as a string for JSON precision
func unmarshalJSON(i *big.Int, bz []byte) error {
	var text string
	if err := json.Unmarshal(bz, &text); err != nil {
		return err
	}

	return unmarshalText(i, text)
}

func unmarshalText(i *big.Int, text string) error {
	if err := i.UnmarshalText([]byte(text)); err != nil {
		return err
	}

	return nil
}

func (i *SerializableInt) ensureNonNil() {
	if i.i == nil {
		i.i = new(big.Int)
	}
}
