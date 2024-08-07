package nodeservice

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	nodesvc "github.com/cometbft/cometbft/api/cometbft/services/node/v1"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/rpc/core"
)

type server struct {
	log     log.Logger
	nodeEnv *core.Environment
}

// New returns a gRPC server serving request for information about a CometBFT node.
func New(l log.Logger, env *core.Environment) nodesvc.NodeServiceServer {
	return &server{
		log:     l.With("service", "NodeService"),
		nodeEnv: env,
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
	l := s.log.With("endpoint", "GetStatus")

	if ctx.Err() != nil {
		err := ctx.Err()

		l.Error("exited early because of context cancellation", "err", err)

		formatStr := "client canceled request: %s"
		return nil, status.Errorf(codes.Canceled, formatStr, err)
	}

	info, ok := s.nodeEnv.P2PTransport.NodeInfo().(p2p.DefaultNodeInfo)
	if !ok {
		// this should never happen.
		// p2p.DefaultNodeInfo is the only concrete type implementing the
		// p2p.NodeInfo interface.
		// We do this check to be good citizens in case something changes in the p2p
		// package.
		errMsg := "p2p.NodeInfo concrete type != p2p.DefaultNodeInfo"
		l.Error(errMsg, "type", reflect.TypeOf(info).String())

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
	blkStore := s.nodeEnv.BlockStore

	// load the metadata of the oldest block that this node stores.
	if blkMeta := blkStore.LoadBaseMeta(); blkMeta != nil {
		syncInfo.EarliestAppHash = blkMeta.Header.AppHash
		syncInfo.EarliestBlockHash = blkMeta.BlockID.Hash
		syncInfo.EarliestBlockHeight = blkMeta.Header.Height
		syncInfo.EarliestBlockTime = blkMeta.Header.Time
	}

	// now load the metadata of the latest block that this node stores.
	blkHeight := blkStore.Height()
	if blkMeta := blkStore.LoadBlockMeta(blkHeight); blkMeta != nil {
		syncInfo.LatestAppHash = blkMeta.Header.AppHash
		syncInfo.LatestBlockHash = blkMeta.BlockID.Hash
		syncInfo.LatestBlockTime = blkMeta.Header.Time
	}
	syncInfo.LatestBlockHeight = blkHeight
	syncInfo.CatchingUp = s.nodeEnv.ConsensusReactor.WaitSync()

	power, err := localValidatorVotingPower(blkHeight, s.nodeEnv)
	if err != nil {
		errMsg := "unknown node validator's voting power"
		l.Error(errMsg, "err", err)
		return nil, status.Errorf(codes.Internal, "%s: %s", errMsg, err)
	}

	pubKey := s.nodeEnv.PubKey
	validatorInfo := &nodesvc.ValidatorInfo{
		Address:     pubKey.Address(),
		PubKeyBytes: pubKey.Bytes(),
		PubKeyType:  pubKey.Type(),
		VotingPower: power,
	}

	resp := &nodesvc.GetStatusResponse{
		NodeInfo:      nodeInfo,
		SyncInfo:      syncInfo,
		ValidatorInfo: validatorInfo,
	}

	return resp, nil
}

// localValidatorVotingPower returns the voting power of the node's local validator.
func localValidatorVotingPower(
	blkHeight int64,
	nodeEnv *core.Environment,
) (int64, error) {
	validatorSet, err := nodeEnv.StateStore.LoadValidators(blkHeight)
	if err != nil {
		return 0, fmt.Errorf("validator set unavailable: %s", err)
	}

	validatorAddr := nodeEnv.PubKey.Address()
	_, validator := validatorSet.GetByAddress(validatorAddr)
	if validator == nil {
		formatStr := "node's validator (addr: %s) isn't in the validator set"
		return 0, fmt.Errorf(formatStr, validatorAddr)
	}

	return validator.VotingPower, nil
}

func (*server) GetHealth(
	_ context.Context,
	_ *nodesvc.GetHealthRequest,
) (*nodesvc.GetHealthResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
