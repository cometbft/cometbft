package blocksync

import (
	cmtbs "github.com/cometbft/cometbft/api/cometbft/blocksync"
	"github.com/cometbft/cometbft/p2p"
)

var (
	_ p2p.Wrapper = &cmtbs.StatusRequest{}
	_ p2p.Wrapper = &cmtbs.StatusResponse{}
	_ p2p.Wrapper = &cmtbs.NoBlockResponse{}
	_ p2p.Wrapper = &cmtbs.BlockResponse{}
	_ p2p.Wrapper = &cmtbs.BlockRequest{}
)
