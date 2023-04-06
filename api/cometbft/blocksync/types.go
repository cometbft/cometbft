package blocksync

import (
	v1 "github.com/cometbft/cometbft/api/cometbft/blocksync/v1"
	v2 "github.com/cometbft/cometbft/api/cometbft/blocksync/v2"
)

type Message = v2.Message
type Message_BlockRequest = v2.Message_BlockRequest
type Message_BlockResponse = v2.Message_BlockResponse
type Message_NoBlockResponse = v2.Message_NoBlockResponse
type Message_StatusRequest = v2.Message_StatusRequest
type Message_StatusResponse = v2.Message_StatusResponse

type BlockRequest = v1.BlockRequest
type BlockResponse = v2.BlockResponse
type NoBlockResponse = v1.NoBlockResponse
type StatusRequest = v1.StatusRequest
type StatusResponse = v1.StatusResponse
