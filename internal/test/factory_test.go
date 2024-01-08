package test

import (
	"testing"

	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
)

func TestMakeHeader(t *testing.T) {
	header := MakeHeader(t, &types.Header{})
	require.NotNil(t, header)

	require.NoError(t, header.ValidateBasic())
}
