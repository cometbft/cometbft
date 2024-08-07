package nodeservice

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	nodesvc "github.com/cometbft/cometbft/api/cometbft/services/node/v1"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/rpc/core"
)

// server implements nodesvc.NodeServiceServer.
// The node service is a gRPC endpoint serving requests for information about
// the CometBFT node provinding the gRPC interface.
type server struct {
	log     log.Logger
	nodeEnv *core.Environment
}

// New returns a gRPC server serving requests for information about a CometBFT node.
func New(l log.Logger, env *core.Environment) nodesvc.NodeServiceServer {
	return &server{
		log:     l.With("service", "NodeService"),
		nodeEnv: env,
	}
}

// GetStatus is the gRPC endpoint serving requests for the node current status.
// This includes the node's basic info, such as public key, latest block hash, app
// hash, block height, and time.
// The request object isn't used in the current implementation, and the function
// doesn't expect clients to provide any data in it.
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

	nodeInfo, err := s.collectNodeInfo()
	if err != nil {
		l.Error(err.Error(), "err", err)

		clientErrMSg := "node's basic information unavailable"
		return nil, status.Error(codes.Internal, clientErrMSg)
	}

	syncInfo, err := s.collectSyncInfo()
	if err != nil {
		l.Error(err.Error(), "err", err)

		clientErrMSg := "node's syncing state information unavailable"
		return nil, status.Error(codes.Internal, clientErrMSg)
	}

	validatorInfo, err := s.collectLocalValidatorInfo()
	if err != nil {
		l.Error(err.Error(), "err", err)

		clientErrMSg := "node's local validator information unavailable"
		return nil, status.Error(codes.Internal, clientErrMSg)
	}

	resp := &nodesvc.GetStatusResponse{
		NodeInfo:      nodeInfo,
		SyncInfo:      syncInfo,
		ValidatorInfo: validatorInfo,
	}

	return resp, nil
}

// GetHealth is the gRPC endpoint serving requests for the node current health.
// The request object isn't used in the current implementation, and the function
// doesn't expect clients to provide any data in it.
func (*server) GetHealth(
	_ context.Context,
	_ *nodesvc.GetHealthRequest,
) (*nodesvc.GetHealthResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// collectNodeInfo collects and returns the node's basic information.
func (s *server) collectNodeInfo() (*nodesvc.NodeInfo, error) {
	info, ok := s.nodeEnv.P2PTransport.NodeInfo().(p2p.DefaultNodeInfo)
	if !ok {
		// this should never happen with the current implementation.
		// p2p.DefaultNodeInfo is the only concrete type implementing the
		// p2p.NodeInfo interface.
		// We do this check to be good citizens in case something changes in the p2p
		// package.
		var (
			formatStr      = "p2p.NodeInfo concrete type != p2p.DefaultNodeInfo: %s"
			unexpectedType = reflect.TypeOf(info).String()
		)
		return nil, fmt.Errorf(formatStr, unexpectedType)
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

	return nodeInfo, nil
}

// collectSyncInfo collects and returns the node's syncing state information.
func (s *server) collectSyncInfo() (*nodesvc.SyncInfo, error) {
	var (
		blkStore  = s.nodeEnv.BlockStore
		blkHeight = blkStore.Height()
		syncInfo  = &nodesvc.SyncInfo{
			LatestBlockHeight: blkHeight,
			CatchingUp:        s.nodeEnv.ConsensusReactor.WaitSync(),
		}
	)

	// load the metadata of the oldest block that this node stores.
	blkMeta := blkStore.LoadBaseMeta()
	if blkMeta == nil {
		return nil, errors.New("node's base block metadata unavailable")
	}
	syncInfo.EarliestAppHash = blkMeta.Header.AppHash
	syncInfo.EarliestBlockHash = blkMeta.BlockID.Hash
	syncInfo.EarliestBlockHeight = blkMeta.Header.Height
	syncInfo.EarliestBlockTime = blkMeta.Header.Time

	// now load the metadata of the latest block that this node stores.
	blkMeta = blkStore.LoadBlockMeta(blkHeight)
	if blkMeta == nil {
		return nil, errors.New("node's latest block metadata unavailable")
	}
	syncInfo.LatestAppHash = blkMeta.Header.AppHash
	syncInfo.LatestBlockHash = blkMeta.BlockID.Hash
	syncInfo.LatestBlockTime = blkMeta.Header.Time

	return syncInfo, nil
}

// collectLocalValidatorInfo returns the node's local validator's information.
func (s *server) collectLocalValidatorInfo() (*nodesvc.ValidatorInfo, error) {
	power, err := localValidatorVotingPower(s.nodeEnv)
	if err != nil {
		return nil, fmt.Errorf("unknown node validator's voting power: %s", err)
	}

	pubKey := s.nodeEnv.PubKey
	validatorInfo := &nodesvc.ValidatorInfo{
		Address:     pubKey.Address(),
		PubKeyBytes: pubKey.Bytes(),
		PubKeyType:  pubKey.Type(),
		VotingPower: power,
	}

	return validatorInfo, nil
}

// localValidatorVotingPower returns the voting power of the node's local validator.
func localValidatorVotingPower(nodeEnv *core.Environment) (int64, error) {
	blkHeight := nodeEnv.BlockStore.Height()
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
