package nodeservice

import (
	"context"
	"reflect"

	nodesvc "github.com/cometbft/cometbft/api/cometbft/services/node/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SyncStatusChecker interface {
	// WaitSync returns true if the node is waiting for state/block sync.
	WaitSync() bool
}

type server struct {
	logger          log.Logger
	info            p2p.NodeInfo
	blkStore        state.BlockStore
	syncChecker     SyncStatusChecker
	validatorPubKey crypto.PubKey
}

// New returns a gRPC server serving request for information about a CometBFT node.
func New(
	l log.Logger,
	info p2p.NodeInfo,
	store *store.BlockStore,
	syncChecker SyncStatusChecker,
	vPubKey crypto.PubKey,
) nodesvc.NodeServiceServer {
	return &server{
		logger:          l.With("service", "NodeService"),
		info:            info,
		blkStore:        store,
		syncChecker:     syncChecker,
		validatorPubKey: vPubKey,
	}
}

// GetStatus is the gRPC endpoint serving requests for the node current status.
// The request object isn't used in the current implementation, and the function
// doesn't expect clients to provide any data in it.
// TODO: refactor into smaller functions.
func (s *server) GetStatus(
	ctx context.Context,
	_ *nodesvc.GetStatusRequest,
) (*nodesvc.GetStatusResponse, error) {
	l := s.logger.With("endpoint", "GetStatus")

	if ctx.Err() != nil {
		err := ctx.Err()

		l.Error("exited early because of context cancellation", "err", err)

		formatStr := "client canceled request: %s"
		return nil, status.Errorf(codes.Canceled, formatStr, err)
	}

	info, ok := s.info.(p2p.DefaultNodeInfo)
	if !ok {
		// this should never happen.
		// p2p.DefaultNodeInfo is the only concrete type implementing the
		// p2p.NodeInfo interface.
		// We do this check to be good citizens in case something changes in the p2p
		// package.
		errMsg := "p2p.NodeInfo concrete type != p2p.DefaultNodeInfo"
		l.Error(errMsg, "type", reflect.TypeOf(s.info).String())

		errMsg = "node's basic information unavailable"
		return nil, status.Error(codes.Internal, errMsg)
	}

	nodeInfo := &nodesvc.NodeInfo{
		ProtocolVersion: &nodesvc.NodeInfo_ProtocolVersion{
			App:   info.ProtocolVersion.App,
			Block: info.ProtocolVersion.Block,
			P2P:   info.ProtocolVersion.P2P,
		},
		Id:         string(info.DefaultNodeID),
		ListenAddr: info.ListenAddr,
		Network:    info.Network,
		Version:    info.Version,
		Channels:   info.Channels,
		Moniker:    info.Moniker,
		Other: &nodesvc.NodeInfo_NodeInfoOther{
			TxIndex:    info.Other.TxIndex,
			RpcAddress: info.Other.RPCAddress,
		},
	}

	syncInfo := &nodesvc.SyncInfo{}

	// load the metadata of the oldest block that this node stores.
	if blkMeta := s.blkStore.LoadBaseMeta(); blkMeta != nil {
		syncInfo.EarliestAppHash = blkMeta.Header.AppHash
		syncInfo.EarliestBlockHash = blkMeta.BlockID.Hash
		syncInfo.EarliestBlockHeight = blkMeta.Header.Height
		syncInfo.EarliestBlockTime = blkMeta.Header.Time
	}

	// now load the metadata of the latest block that this node stores.
	lastKnownHeight := s.blkStore.Height()
	if blkMeta := s.blkStore.LoadBlockMeta(lastKnownHeight); blkMeta != nil {
		syncInfo.LatestAppHash = blkMeta.Header.AppHash
		syncInfo.LatestBlockHash = blkMeta.BlockID.Hash
		syncInfo.LatestBlockTime = blkMeta.Header.Time
	}
	syncInfo.LatestBlockHeight = lastKnownHeight
	syncInfo.CatchingUp = s.syncChecker.WaitSync()

	resp := &nodesvc.GetStatusResponse{
		NodeInfo: nodeInfo,
		SyncInfo: syncInfo,
	}

	return resp, nil
}

func (*server) GetHealth(
	_ context.Context,
	_ *nodesvc.GetHealthRequest,
) (*nodesvc.GetHealthResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
