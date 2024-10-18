package coretypes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	ni "github.com/cometbft/cometbft/p2p/nodeinfo"
)

func TestStatusIndexer(t *testing.T) {
	var status *ResultStatus
	assert.False(t, status.TxIndexEnabled())

	status = &ResultStatus{}
	assert.False(t, status.TxIndexEnabled())

	status.NodeInfo = ni.Default{}
	assert.False(t, status.TxIndexEnabled())

	cases := []struct {
		expected bool
		other    ni.DefaultOther
	}{
		{false, ni.DefaultOther{}},
		{false, ni.DefaultOther{TxIndex: "aa"}},
		{false, ni.DefaultOther{TxIndex: "off"}},
		{true, ni.DefaultOther{TxIndex: "on"}},
	}

	for _, tc := range cases {
		status.NodeInfo.Other = tc.other
		assert.Equal(t, tc.expected, status.TxIndexEnabled())
	}
}
