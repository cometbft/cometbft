package client

import (
	"testing"

	"github.com/stretchr/testify/require"

	types "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

// The JSON-RPC 2.0 spec lets a server return the responses of a batch call in
// any order, so they have to be paired with the requests by ID.
func TestUnmarshalResponseBytesArrayOutOfOrder(t *testing.T) {
	first, second := new(string), new(string)

	results, err := unmarshalResponseBytesArray(
		[]byte(`[{"jsonrpc":"2.0","id":1,"result":"B"},{"jsonrpc":"2.0","id":0,"result":"A"}]`),
		[]types.JSONRPCIntID{0, 1},
		[]any{first, second},
	)
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, "A", *first)
	require.Equal(t, "B", *second)
}

func TestUnmarshalResponseBytesArrayUnsolicitedID(t *testing.T) {
	_, err := unmarshalResponseBytesArray(
		[]byte(`[{"jsonrpc":"2.0","id":0,"result":"A"},{"jsonrpc":"2.0","id":7,"result":"B"}]`),
		[]types.JSONRPCIntID{0, 1},
		[]any{new(string), new(string)},
	)
	require.Error(t, err)
}
