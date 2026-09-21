package core

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

func TestUnsafeDialSeeds(t *testing.T) {
	sw := p2p.MakeSwitch(cfg.DefaultP2PConfig(), 1,
		func(n int, sw *p2p.Switch) *p2p.Switch { return sw })
	err := sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Error(err)
		}
	})

	env := &Environment{}
	env.Logger = log.TestingLogger()
	env.P2PPeers = sw

	testCases := []struct {
		seeds []string
		isErr bool
	}{
		{[]string{}, true},
		{[]string{"d51fb70907db1c6c2d5237e78379b25cf1a37ab4@127.0.0.1:41198"}, false},
		{[]string{"127.0.0.1:41198"}, true},
	}

	for _, tc := range testCases {
		res, err := env.UnsafeDialSeeds(&rpctypes.Context{}, tc.seeds)
		if tc.isErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, res)
		}
	}
}

func TestUnsafeDialPeers(t *testing.T) {
	sw := p2p.MakeSwitch(cfg.DefaultP2PConfig(), 1,
		func(n int, sw *p2p.Switch) *p2p.Switch { return sw })
	sw.SetAddrBook(&p2p.AddrBookMock{
		Addrs:        make(map[string]struct{}),
		OurAddrs:     make(map[string]struct{}),
		PrivateAddrs: make(map[string]struct{}),
	})
	err := sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Error(err)
		}
	})

	env := &Environment{}
	env.Logger = log.TestingLogger()
	env.P2PPeers = sw

	testCases := []struct {
		peers                               []string
		persistence, unconditional, private bool
		isErr                               bool
	}{
		{[]string{}, false, false, false, true},
		{[]string{"d51fb70907db1c6c2d5237e78379b25cf1a37ab4@127.0.0.1:41198"}, true, true, true, false},
		{[]string{"127.0.0.1:41198"}, true, true, false, true},
	}

	for _, tc := range testCases {
		res, err := env.UnsafeDialPeers(&rpctypes.Context{}, tc.peers, tc.persistence, tc.unconditional, tc.private)
		if tc.isErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, res)
		}
	}
}

func TestGenesisChunkedRejectsOutOfRangeIndex(t *testing.T) {
	env := &Environment{genChunks: []string{"a", "b"}}

	res, err := env.GenesisChunked(&rpctypes.Context{}, 1)
	require.NoError(t, err)
	assert.Equal(t, "b", res.Data)

	for _, chunk := range []uint{2, math.MaxUint, 1 << 63} {
		res, err := env.GenesisChunked(&rpctypes.Context{}, chunk)
		require.Error(t, err, "chunk %d", chunk)
		assert.Nil(t, res)
	}
}
