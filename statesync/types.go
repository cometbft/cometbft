package statesync

import (
	ssproto "github.com/cometbft/cometbft/api/cometbft/statesync/v1"
	"github.com/cometbft/cometbft/v2/types"
)

var (
	_ types.Wrapper = &ssproto.ChunkRequest{}
	_ types.Wrapper = &ssproto.ChunkResponse{}
	_ types.Wrapper = &ssproto.SnapshotsRequest{}
	_ types.Wrapper = &ssproto.SnapshotsResponse{}
)
