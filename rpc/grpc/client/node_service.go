package client

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/grpc"

	nodesvc "github.com/cometbft/cometbft/api/cometbft/services/node/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/p2p"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
)

// ValidatoInfo contains information about the node's local validator.
type LocalValidatorInfo struct {
	PubKeyType  string
	Address     crypto.Address
	PubKeyBytes []byte
	VotingPower int64
}

// NodeStatus contains information about the node providing the gRPC interface.
type NodeStatus struct {
	NodeInfo      *p2p.DefaultNodeInfo // node's basic info
	SyncInfo      *ctypes.SyncInfo     // node's syncing state info
	ValidatorInfo *LocalValidatorInfo  // node's local validator info
}

// NodeHealth contains information about the health of the node providing the gRPC
// interface.
type NodeHealth struct {
	// Code is the status code of the health check. 0 means healthy.
	Code uint32
}

// NodeServiceClient exposes the functionalities that a client must implement to
// query the NodeService gRPC endpoint.
type NodeServiceClient interface {
	// GetStatus queries the node's current status, including node info, public key,
	// latest block hash, app hash, block height, and time.
	GetStatus(ctx context.Context) (*NodeStatus, error)

	// GetHealth queries the node's health.
	GetHealth(ctx context.Context) (*NodeHealth, error)
}

// nodeServiceClient is the gRPC client for the NodeService gRPC endpoint.
// nodeServiceClient implements NodeServiceClient.
type nodeServiceClient struct {
	client nodesvc.NodeServiceClient
}

// newNodeServiceClient returns a new NodeService gRPC client using the given grpc
// connection.
func newNodeServiceClient(conn grpc.ClientConn) NodeServiceClient {
	return &nodeServiceClient{
		client: nodesvc.NewNodeServiceClient(conn),
	}
}

// GetStatus is the gRPC endpoint serving requests for the node current status.
// Implements the NodeServiceClient interface.
func (c *nodeServiceClient) GetStatus(ctx context.Context) (*NodeStatus, error) {
	resp, err := c.client.GetStatus(ctx, &nodesvc.GetStatusRequest{})
	if err != nil {
		return nil, fmt.Errorf("gRPC call returned: %s", err)
	}

	nInfo := &p2p.DefaultNodeInfo{
		ProtocolVersion: p2p.ProtocolVersion{
			P2P:   resp.NodeInfo.ProtocolVersion.P2P,
			Block: resp.NodeInfo.ProtocolVersion.Block,
			App:   resp.NodeInfo.ProtocolVersion.App,
		},
		Other: p2p.DefaultNodeInfoOther{
			TxIndex:    resp.NodeInfo.Other.TxIndex,
			RPCAddress: resp.NodeInfo.Other.RpcAddress,
		},
		DefaultNodeID: p2p.ID(resp.NodeInfo.Id),
		ListenAddr:    resp.NodeInfo.ListenAddr,
		Network:       resp.NodeInfo.Network,
		Version:       resp.NodeInfo.Version,
		Channels:      resp.NodeInfo.Channels,
		Moniker:       resp.NodeInfo.Moniker,
	}

	sInfo := &ctypes.SyncInfo{
		LatestBlockHash:     resp.SyncInfo.LatestBlockHash,
		LatestAppHash:       resp.SyncInfo.LatestAppHash,
		LatestBlockHeight:   resp.SyncInfo.LatestBlockHeight,
		LatestBlockTime:     resp.SyncInfo.LatestBlockTime,
		EarliestBlockHash:   resp.SyncInfo.EarliestBlockHash,
		EarliestAppHash:     resp.SyncInfo.EarliestAppHash,
		EarliestBlockHeight: resp.SyncInfo.EarliestBlockHeight,
		EarliestBlockTime:   resp.SyncInfo.EarliestBlockTime,
		CatchingUp:          resp.SyncInfo.CatchingUp,
	}

	vInfo := &LocalValidatorInfo{
		Address:     resp.ValidatorInfo.Address,
		PubKeyType:  resp.ValidatorInfo.PubKeyType,
		PubKeyBytes: resp.ValidatorInfo.PubKeyBytes,
		VotingPower: resp.ValidatorInfo.VotingPower,
	}

	status := &NodeStatus{
		NodeInfo:      nInfo,
		SyncInfo:      sInfo,
		ValidatorInfo: vInfo,
	}

	return status, nil
}

// GetHealth is the gRPC endpoint serving requests for the node current health.
// Implements the NodeServiceClient interface.
// GetHealth serves as a ping to check if the node is responsive. A successful
// call (i.e., no error) indicates the node is responsive; the response itself
// is empty.
func (c *nodeServiceClient) GetHealth(ctx context.Context) (*NodeHealth, error) {
	resp, err := c.client.GetHealth(ctx, &nodesvc.GetHealthRequest{})
	if err != nil {
		return nil, fmt.Errorf("gRPC call returned: %s", err)
	}

	hs := &NodeHealth{
		Code: resp.Code,
	}

	return hs, nil
}

// disabledNodeServiceClient is a NodeServiceClient that panics when used.
// We use it when we don't create a NodeService gRPC client, thus making the
// NodeService API unavailable to users.
// It implements the NodeServiceClient interface.
type disabledNodeServiceClient struct{}

// newDisabledNodeServiceClient returns a disabled NodeService gRPC client that
// panics if the client uses it.
func newDisabledNodeServiceClient() NodeServiceClient {
	return &disabledNodeServiceClient{}
}

const nodeSvcPanicMsg = "node service client is disabled"

// GetStatus panics if called, because the node service client is disabled.
// Implements the NodeServiceClient interface.
func (*disabledNodeServiceClient) GetStatus(context.Context) (*NodeStatus, error) {
	panic(nodeSvcPanicMsg)
}

// GetHealth panics if called, because the node service client is disabled.
// Implements the NodeServiceClient interface.
func (*disabledNodeServiceClient) GetHealth(context.Context) (*NodeHealth, error) {
	panic(nodeSvcPanicMsg)
}
