package clob

import "github.com/gogo/protobuf/proto"

//nolint:stylecheck,revive // Match variable formats in dydx v4 repo
const (
	OrderIdFlags_ShortTerm   = uint32(0)
	OrderIdFlags_Conditional = uint32(32)
	OrderIdFlags_LongTerm    = uint32(64)
)

// IsShortTermOrder returns true if this order ID is for a short-term order, false if
// not (which implies the order ID is for a long-term or conditional order).
// Note that all short-term orders will have the `OrderFlags` field set to 0.
func (o *OrderId) IsShortTermOrder() bool {
	return o.OrderFlags == OrderIdFlags_ShortTerm
}

// IsConditionalOrder returns true if this order ID is for a conditional order, false if
// not (which implies the order ID is for a short-term or long-term order).
func (o *OrderId) IsConditionalOrder() bool {
	// If the third bit in the first byte is set and no other bits are set,
	// this is a conditional order.
	// Note that 32 in decimal == 0x20 in hex == 0b00100000 in binary.
	return o.OrderFlags == OrderIdFlags_Conditional
}

// IsLongTermOrder returns true if this order ID is for a long-term order, false if
// not (which implies the order ID is for a short-term or conditional order).
func (o *OrderId) IsLongTermOrder() bool {
	// If the second bit in the first byte is set and no other bits are set,
	// this is a long-term order.
	// Note that 64 in decimal == 0x40 in hex == 0b01000000 in binary.
	return o.OrderFlags == OrderIdFlags_LongTerm
}

// IsStatefulOrder returns whether this order is a stateful order, which is true for Long-Term
// and conditional orders and false for Short-Term orders.
func (o *OrderId) IsStatefulOrder() bool {
	return o.IsLongTermOrder() || o.IsConditionalOrder()
}

// GetOrderTextString returns the JSON representation of this order.
func (o *Order) GetOrderTextString() string {
	return proto.MarshalTextString(o)
}
