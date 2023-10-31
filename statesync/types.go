package statesync

import (
	ssproto "github.com/cometbft/cometbft/api/cometbft/statesync/v1beta1"
	"github.com/cometbft/cometbft/p2p"
)

var (
	_ p2p.Wrapper = &ssproto.ChunkRequest{}
	_ p2p.Wrapper = &ssproto.ChunkResponse{}
	_ p2p.Wrapper = &ssproto.SnapshotsRequest{}
	_ p2p.Wrapper = &ssproto.SnapshotsResponse{}
)
