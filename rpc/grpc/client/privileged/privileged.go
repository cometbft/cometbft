package privileged

import (
	"context"
	"fmt"
	"net"

	"github.com/cosmos/gogoproto/grpc"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cmtnet "github.com/cometbft/cometbft/libs/net"
)

type Option func(*clientBuilder)

// PrivilegedClient defines the full client interface for interacting with
// a CometBFT node via the privileged gRPC server.
type PrivilegedClient interface {
	PruningServiceClient
}

type clientBuilder struct {
	dialerFunc func(context.Context, string) (net.Conn, error)
	grpcOpts   []ggrpc.DialOption

	pruningServiceEnabled bool
}

func newClientBuilder() *clientBuilder {
	return &clientBuilder{
		dialerFunc:            defaultDialerFunc,
		grpcOpts:              make([]ggrpc.DialOption, 0),
		pruningServiceEnabled: true,
	}
}

func defaultDialerFunc(ctx context.Context, addr string) (net.Conn, error) {
	return cmtnet.ConnectContext(ctx, addr)
}

type client struct {
	conn grpc.ClientConn

	PruningServiceClient
}

// WithInsecure disables transport security for the underlying client
// connection.
//
// A shortcut for using grpc.WithTransportCredentials and
// insecure.NewCredentials from google.golang.org/grpc.
func WithInsecure() Option {
	return WithGRPCDialOption(ggrpc.WithTransportCredentials(insecure.NewCredentials()))
}

// WithPruningServiceEnabled allows control of whether or not to create a
// client for interacting with the pruning service of a CometBFT node.
//
// If disabled and the client attempts to access the pruning service API, the
// client will panic.
func WithPruningServiceEnabled(enabled bool) Option {
	return func(b *clientBuilder) {
		b.pruningServiceEnabled = enabled
	}
}

// WithGRPCDialOption allows passing lower-level gRPC dial options through to
// the gRPC dialer when creating the client.
func WithGRPCDialOption(opt ggrpc.DialOption) Option {
	return func(b *clientBuilder) {
		b.grpcOpts = append(b.grpcOpts, opt)
	}
}

// New constructs a client for interacting with a CometBFT node via its
// privileged gRPC server.
//
// Makes no assumptions about whether or not to use TLS to connect to the given
// address. To connect to a gRPC server without using TLS, use the WithInsecure
// option.
//
// To connect to a gRPC server with TLS, use the WithGRPCDialOption option with
// the appropriate gRPC credentials configuration. See
// https://pkg.go.dev/google.golang.org/grpc#WithTransportCredentials
func New(ctx context.Context, addr string, opts ...Option) (PrivilegedClient, error) {
	builder := newClientBuilder()
	for _, opt := range opts {
		opt(builder)
	}
	conn, err := ggrpc.DialContext(ctx, addr, builder.grpcOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
	}
	pruningServiceClient := newDisabledPruningServiceClient()
	if builder.pruningServiceEnabled {
		pruningServiceClient = newPruningServiceClient(conn)
	}
	client := &client{
		conn:                 conn,
		PruningServiceClient: pruningServiceClient,
	}
	return client, nil
}
