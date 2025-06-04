package coretypes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cometbft/cometbft/v2/p2p"
)

func TestStatusIndexer(t *testing.T) {
	var status *ResultStatus
	assert.False(t, status.TxIndexEnabled())

	status = &ResultStatus{}
	assert.False(t, status.TxIndexEnabled())

	status.NodeInfo = p2p.NodeInfoDefault{}
	assert.False(t, status.TxIndexEnabled())

	cases := []struct {
		expected bool
		other    p2p.NodeInfoDefaultOther
	}{
		{false, p2p.NodeInfoDefaultOther{}},
		{false, p2p.NodeInfoDefaultOther{TxIndex: "aa"}},
		{false, p2p.NodeInfoDefaultOther{TxIndex: "off"}},
		{true, p2p.NodeInfoDefaultOther{TxIndex: "on"}},
	}

	for _, tc := range cases {
		status.NodeInfo.Other = tc.other
		assert.Equal(t, tc.expected, status.TxIndexEnabled())
	}
}
