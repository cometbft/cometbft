package blocksync

import (
	cmtbs "github.com/cometbft/cometbft/api/cometbft/blocksync/v2"
	"github.com/cometbft/cometbft/v2/types"
)

var (
	_ types.Wrapper = &cmtbs.StatusRequest{}
	_ types.Wrapper = &cmtbs.StatusResponse{}
	_ types.Wrapper = &cmtbs.NoBlockResponse{}
	_ types.Wrapper = &cmtbs.BlockResponse{}
	_ types.Wrapper = &cmtbs.BlockRequest{}
)
