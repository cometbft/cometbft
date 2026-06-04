package privval

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/lp2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsConnTimeoutForNonTimeoutErrors(t *testing.T) {
	assert.False(t, IsConnTimeout(fmt.Errorf("max retries exceeded: %w", ErrDialRetryMax)))
	assert.False(t, IsConnTimeout(errors.New("completely irrelevant error")))
}

func TestNewSignerListenerFromAddrNoise(t *testing.T) {
	nodeKey := ed25519.GenPrivKey()
	kmsPeer, err := lp2p.IDFromPrivateKey(ed25519.GenPrivKey())
	require.NoError(t, err)

	ep, err := NewSignerListenerFromAddr(
		"noise://"+kmsPeer.String()+"@127.0.0.1:0",
		nodeKey,
		log.TestingLogger(),
	)
	require.NoError(t, err)
	require.NotNil(t, ep)
	require.NoError(t, ep.Start())
	require.NoError(t, ep.Stop())
}

func TestNewSignerListenerFromAddrTCPStillWorks(t *testing.T) {
	ep, err := NewSignerListenerFromAddr("tcp://127.0.0.1:0", ed25519.GenPrivKey(), log.TestingLogger())
	require.NoError(t, err)
	require.NotNil(t, ep)
	require.NoError(t, ep.Start())
	require.NoError(t, ep.Stop())
}
