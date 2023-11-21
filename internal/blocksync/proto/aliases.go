//nolint:revive,stylecheck
package proto

import (
	"github.com/cometbft/cometbft/api/cometbft/blocksync/v1beta1"
	"github.com/cometbft/cometbft/api/cometbft/blocksync/v1beta2"
)

type (
	Message                 = v1beta2.Message
	Message_BlockRequest    = v1beta2.Message_BlockRequest
	Message_BlockResponse   = v1beta2.Message_BlockResponse
	Message_NoBlockResponse = v1beta2.Message_NoBlockResponse
	Message_StatusRequest   = v1beta2.Message_StatusRequest
	Message_StatusResponse  = v1beta2.Message_StatusResponse
)

type (
	BlockRequest    = v1beta1.BlockRequest
	BlockResponse   = v1beta2.BlockResponse
	NoBlockResponse = v1beta1.NoBlockResponse
	StatusRequest   = v1beta1.StatusRequest
	StatusResponse  = v1beta1.StatusResponse
)
