package rpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	cmttest "github.com/cometbft/cometbft/internal/test"
	lcmock "github.com/cometbft/cometbft/light/rpc/mocks"
	rpcmock "github.com/cometbft/cometbft/rpc/client/mocks"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
)

func TestTxBindsResponseToRequest(t *testing.T) {
	const height = int64(5)
	txs := types.Txs{types.Tx("included-a"), types.Tx("included-b")}
	proof := txs.Proof(0)
	lb := &types.LightBlock{
		SignedHeader: &types.SignedHeader{Header: &types.Header{Height: height, DataHash: txs.Hash()}},
	}

	testCases := []struct {
		name    string
		reqHash []byte
		resTx   types.Tx
		wantErr bool
	}{
		{"valid", txs[0].Hash(), txs[0], false},
		{"forged tx body", types.Tx("forged").Hash(), types.Tx("forged"), true},
		{"tx never included", types.Tx("not-included").Hash(), txs[0], true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next := &rpcmock.Client{}
			next.On("Tx", mock.Anything, tc.reqHash, true).Return(&ctypes.ResultTx{
				Hash:   tc.reqHash,
				Height: height,
				Tx:     tc.resTx,
				Proof:  proof,
			}, nil)
			lc := &lcmock.LightClient{}
			lc.On("VerifyLightBlockAtHeight", mock.Anything, height, mock.Anything).Return(lb, nil).Maybe()

			c := NewClient(next, lc)
			_, err := c.Tx(context.Background(), tc.reqHash, true)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHeaderByHashBindsResponseToRequest(t *testing.T) {
	trusted := cmttest.MakeHeader(t, &types.Header{Height: 5})
	other := cmttest.MakeHeader(t, &types.Header{Height: 6})
	lb := &types.LightBlock{SignedHeader: &types.SignedHeader{Header: other}}

	next := &rpcmock.Client{}
	next.On("HeaderByHash", mock.Anything, mock.Anything).Return(&ctypes.ResultHeader{Header: other}, nil)
	lc := &lcmock.LightClient{}
	lc.On("VerifyLightBlockAtHeight", mock.Anything, other.Height, mock.Anything).Return(lb, nil).Maybe()

	c := NewClient(next, lc)

	// The primary answers with a genuine header, but not the one that was asked for.
	_, err := c.HeaderByHash(context.Background(), trusted.Hash())
	require.Error(t, err)

	res, err := c.HeaderByHash(context.Background(), other.Hash())
	require.NoError(t, err)
	assert.Equal(t, other.Hash(), res.Header.Hash())
}
