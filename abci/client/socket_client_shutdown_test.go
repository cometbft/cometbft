package abcicli

import (
	"container/list"
	"testing"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
)

func TestSocketClientFlushQueueRemovesInFlightRequests(t *testing.T) {
	reqres := NewReqRes(types.ToRequestEcho("hello"))
	cli := &socketClient{
		reqQueue: make(chan *ReqRes, 1),
		reqSent:  list.New(),
	}
	cli.reqSent.PushBack(reqres)

	cli.flushQueue()

	err := cli.didRecvResponse(types.ToResponseEcho("hello"))
	var unexpected ErrUnexpectedResponse
	require.ErrorAs(t, err, &unexpected)
	require.Empty(t, cli.reqSent.Len())
}
