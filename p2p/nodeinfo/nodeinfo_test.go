package nodeinfo

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtnet "github.com/cometbft/cometbft/internal/net"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/nodekey"
)

const testCh = 0x01

type mockNodeInfo struct {
	addr *na.NetAddr
}

func (ni mockNodeInfo) ID() nodekey.ID                                   { return ni.addr.ID }
func (ni mockNodeInfo) NetAddr() (*na.NetAddr, error)                    { return ni.addr, nil }
func (mockNodeInfo) Validate() error                                     { return nil }
func (mockNodeInfo) CompatibleWith(NodeInfo) error                       { return nil }
func (mockNodeInfo) Handshake(net.Conn, time.Duration) (NodeInfo, error) { return nil, nil }

func testNodeInfo(id nodekey.ID) NodeInfo {
	return Default{
		ProtocolVersion: NewProtocolVersion(0, 0, 0),
		DefaultNodeID:   id,
		ListenAddr:      fmt.Sprintf("127.0.0.1:%d", getFreePort()),
		Network:         "testing",
		Version:         "1.2.3-rc0-deadbeef",
		Channels:        []byte{testCh},
		Moniker:         "testing",
		Other: DefaultOther{
			TxIndex:    "on",
			RPCAddress: fmt.Sprintf("127.0.0.1:%d", getFreePort()),
		},
	}
}

func getFreePort() int {
	port, err := cmtnet.GetFreePort()
	if err != nil {
		panic(err)
	}
	return port
}

func TestNodeInfoValidate(t *testing.T) {
	// empty fails
	ni := Default{}
	require.Error(t, ni.Validate())

	channels := make([]byte, maxNumChannels)
	for i := 0; i < maxNumChannels; i++ {
		channels[i] = byte(i)
	}
	dupChannels := make([]byte, 5)
	copy(dupChannels, channels[:5])
	dupChannels = append(dupChannels, testCh) //nolint:makezero // huge errors when we don't do it the "wrong" way

	nonASCII := "¢§µ"
	emptyTab := "\t"
	emptySpace := "  "

	testCases := []struct {
		testName         string
		malleateNodeInfo func(*Default)
		expectErr        bool
	}{
		{
			"Too Many Channels",
			func(ni *Default) { ni.Channels = append(channels, byte(maxNumChannels)) }, //nolint: makezero
			true,
		},
		{"Duplicate Channel", func(ni *Default) { ni.Channels = dupChannels }, true},
		{"Good Channels", func(ni *Default) { ni.Channels = ni.Channels[:5] }, false},

		{"Invalid NetAddr", func(ni *Default) { ni.ListenAddr = "not-an-address" }, true},
		{"Good NetAddr", func(ni *Default) { ni.ListenAddr = "0.0.0.0:26656" }, false},

		{"Non-ASCII Version", func(ni *Default) { ni.Version = nonASCII }, true},
		{"Empty tab Version", func(ni *Default) { ni.Version = emptyTab }, true},
		{"Empty space Version", func(ni *Default) { ni.Version = emptySpace }, true},
		{"Empty Version", func(ni *Default) { ni.Version = "" }, false},

		{"Non-ASCII Moniker", func(ni *Default) { ni.Moniker = nonASCII }, true},
		{"Empty tab Moniker", func(ni *Default) { ni.Moniker = emptyTab }, true},
		{"Empty space Moniker", func(ni *Default) { ni.Moniker = emptySpace }, true},
		{"Empty Moniker", func(ni *Default) { ni.Moniker = "" }, true},
		{"Good Moniker", func(ni *Default) { ni.Moniker = "hey its me" }, false},

		{"Non-ASCII TxIndex", func(ni *Default) { ni.Other.TxIndex = nonASCII }, true},
		{"Empty tab TxIndex", func(ni *Default) { ni.Other.TxIndex = emptyTab }, true},
		{"Empty space TxIndex", func(ni *Default) { ni.Other.TxIndex = emptySpace }, true},
		{"Empty TxIndex", func(ni *Default) { ni.Other.TxIndex = "" }, false},
		{"Off TxIndex", func(ni *Default) { ni.Other.TxIndex = "off" }, false},

		{"Non-ASCII RPCAddress", func(ni *Default) { ni.Other.RPCAddress = nonASCII }, true},
		{"Empty tab RPCAddress", func(ni *Default) { ni.Other.RPCAddress = emptyTab }, true},
		{"Empty space RPCAddress", func(ni *Default) { ni.Other.RPCAddress = emptySpace }, true},
		{"Empty RPCAddress", func(ni *Default) { ni.Other.RPCAddress = "" }, false},
		{"Good RPCAddress", func(ni *Default) { ni.Other.RPCAddress = "0.0.0.0:26657" }, false},
	}

	nodeKey := nodekey.NodeKey{PrivKey: ed25519.GenPrivKey()}

	// test case passes
	ni = testNodeInfo(nodeKey.ID()).(Default)
	ni.Channels = channels
	require.NoError(t, ni.Validate())

	for _, tc := range testCases {
		ni := testNodeInfo(nodeKey.ID()).(Default)
		ni.Channels = channels
		tc.malleateNodeInfo(&ni)
		err := ni.Validate()
		if tc.expectErr {
			require.Error(t, err, tc.testName)
		} else {
			require.NoError(t, err, tc.testName)
		}
	}
}

func TestNodeInfoCompatible(t *testing.T) {
	nodeKey1 := nodekey.NodeKey{PrivKey: ed25519.GenPrivKey()}
	nodeKey2 := nodekey.NodeKey{PrivKey: ed25519.GenPrivKey()}

	var newTestChannel byte = 0x2

	// test NodeInfo is compatible
	ni1 := testNodeInfo(nodeKey1.ID()).(Default)
	ni2 := testNodeInfo(nodeKey2.ID()).(Default)
	require.NoError(t, ni1.CompatibleWith(ni2))

	// add another channel; still compatible
	ni2.Channels = append(ni2.Channels, newTestChannel)
	assert.True(t, ni2.HasChannel(newTestChannel))
	require.NoError(t, ni1.CompatibleWith(ni2))

	// wrong NodeInfo type is not compatible
	_, netAddr := na.CreateRoutableAddr()
	ni3 := mockNodeInfo{netAddr}
	require.Error(t, ni1.CompatibleWith(ni3))

	testCases := []struct {
		testName         string
		malleateNodeInfo func(*Default)
	}{
		{"Wrong block version", func(ni *Default) { ni.ProtocolVersion.Block++ }},
		{"Wrong network", func(ni *Default) { ni.Network += "-wrong" }},
		{"No common channels", func(ni *Default) { ni.Channels = []byte{newTestChannel} }},
	}

	for _, tc := range testCases {
		ni := testNodeInfo(nodeKey2.ID()).(Default)
		tc.malleateNodeInfo(&ni)
		require.Error(t, ni1.CompatibleWith(ni))
	}
}
