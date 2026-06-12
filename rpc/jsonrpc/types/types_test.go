package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SampleResult struct {
	Value string
}

type responseTest struct {
	id       jsonrpcid
	expected string
}

var responseTests = []responseTest{
	{JSONRPCStringID("1"), `"1"`},
	{JSONRPCStringID("alphabet"), `"alphabet"`},
	{JSONRPCStringID(""), `""`},
	{JSONRPCStringID("àáâ"), `"àáâ"`},
	{JSONRPCIntID(-1), "-1"},
	{JSONRPCIntID(0), "0"},
	{JSONRPCIntID(1), "1"},
	{JSONRPCIntID(100), "100"},
}

func TestResponses(t *testing.T) {
	assert := assert.New(t)
	for _, tt := range responseTests {
		jsonid := tt.id
		a := NewRPCSuccessResponse(jsonid, &SampleResult{"hello"})
		b, _ := json.Marshal(a)
		s := fmt.Sprintf(`{"jsonrpc":"2.0","id":%v,"result":{"Value":"hello"}}`, tt.expected)
		assert.Equal(s, string(b))

		d := RPCParseError(errors.New("hello world"))
		e, _ := json.Marshal(d)
		f := `{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error. Invalid JSON","data":"hello world"}}`
		assert.Equal(f, string(e))

		g := RPCMethodNotFoundError(jsonid)
		h, _ := json.Marshal(g)
		i := fmt.Sprintf(`{"jsonrpc":"2.0","id":%v,"error":{"code":-32601,"message":"Method not found"}}`, tt.expected)
		assert.Equal(string(h), i)
	}
}

func TestUnmarshallResponses(t *testing.T) {
	assert := assert.New(t)
	for _, tt := range responseTests {
		response := &RPCResponse{}
		err := json.Unmarshal(
			[]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":%v,"result":{"Value":"hello"}}`, tt.expected)),
			response,
		)
		assert.Nil(err)
		a := NewRPCSuccessResponse(tt.id, &SampleResult{"hello"})
		assert.Equal(*response, a)
	}
	response := &RPCResponse{}
	err := json.Unmarshal([]byte(`{"jsonrpc":"2.0","id":true,"result":{"Value":"hello"}}`), response)
	assert.NotNil(err)
}

// TestIDFromInterface_FloatRange covers idFromInterface's float64 arm — the
// path the issue (#5846) was reported against. Without the bounds check an
// out-of-int64 input is silently saturated to math.MinInt by int(float),
// collapsing distinct large IDs onto the same negative value and breaking
// request/response correlation. The exact saturation result is
// architecture-dependent (Go spec §6.5 leaves float-to-int conversion of an
// unrepresentable value implementation-defined), so these assertions are
// written against explicit bounds rather than the saturation residue.
func TestIDFromInterface_FloatRange(t *testing.T) {
	t.Run("in_range_int_accepted", func(t *testing.T) {
		for _, v := range []float64{0, 1, -1, 100, -100, 1234567890} {
			id, err := idFromInterface(v)
			require.NoError(t, err, "input %v", v)
			require.Equal(t, JSONRPCIntID(int(v)), id)
		}
	})

	t.Run("fractional_rejected", func(t *testing.T) {
		_, err := idFromInterface(1.5)
		require.Error(t, err)
		require.Contains(t, err.Error(), "whole number")
	})

	t.Run("nan_rejected", func(t *testing.T) {
		_, err := idFromInterface(math.NaN())
		require.Error(t, err)
		require.Contains(t, err.Error(), "finite number")
	})

	t.Run("pos_inf_rejected", func(t *testing.T) {
		_, err := idFromInterface(math.Inf(1))
		require.Error(t, err)
		require.Contains(t, err.Error(), "finite number")
	})

	t.Run("neg_inf_rejected", func(t *testing.T) {
		_, err := idFromInterface(math.Inf(-1))
		require.Error(t, err)
		require.Contains(t, err.Error(), "finite number")
	})

	// 2^63 is exactly representable in float64 but does not fit in int64.
	// The closed >= upper bound in idFromInterface rejects it regardless of
	// how int(float) behaves on the host architecture.
	t.Run("two_to_63_rejected", func(t *testing.T) {
		_, err := idFromInterface(math.Pow(2, 63))
		require.Error(t, err)
		require.Contains(t, err.Error(), "out of int64 range")
	})

	// -2^63 is the int64 minimum and is exactly representable in float64;
	// anything strictly less than that must be rejected. We pick a value
	// well below -2^63 to dodge float64 precision rounding (~1024 between
	// adjacent representable values at this scale).
	t.Run("below_neg_two_to_63_rejected", func(t *testing.T) {
		_, err := idFromInterface(-math.Pow(2, 63) - 2049)
		require.Error(t, err)
		require.Contains(t, err.Error(), "out of int64 range")
	})

	t.Run("very_large_finite_rejected", func(t *testing.T) {
		_, err := idFromInterface(1e20)
		require.Error(t, err)
		require.Contains(t, err.Error(), "out of int64 range")
	})

	// Regression pin for #5846: five distinct oversized inputs that
	// previously collapsed onto the same JSONRPCIntID(math.MinInt) must now
	// each return an error.
	t.Run("oversized_ids_do_not_collide", func(t *testing.T) {
		inputs := []float64{1e19, 1e20, 1e21, 9.3e18, math.Pow(2, 63)}
		for _, v := range inputs {
			_, err := idFromInterface(v)
			require.Error(t, err, "input %v should error, not collide", v)
		}
	})

	// -2^63 itself must still be accepted: it is exactly representable in
	// float64 and is the legal int64 minimum.
	t.Run("min_int64_accepted", func(t *testing.T) {
		id, err := idFromInterface(-math.Pow(2, 63))
		require.NoError(t, err)
		require.Equal(t, JSONRPCIntID(math.MinInt64), id)
	})
}

func TestRPCError(t *testing.T) {
	assert.Equal(t, "RPC error 12 - Badness: One worse than a code 11",
		fmt.Sprintf("%v", &RPCError{
			Code:    12,
			Message: "Badness",
			Data:    "One worse than a code 11",
		}))

	assert.Equal(t, "RPC error 12 - Badness",
		fmt.Sprintf("%v", &RPCError{
			Code:    12,
			Message: "Badness",
		}))
}
