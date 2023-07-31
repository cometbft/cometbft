package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/cometbft/cometbft/libs/log"
	pbversionsvc "github.com/cometbft/cometbft/proto/tendermint/services/version/v1"
	"github.com/cometbft/cometbft/rpc/grpc/server/services/versionservice"
	"google.golang.org/grpc"
)

// Option is any function that allows for configuration of the gRPC server
// during its creation.
type Option func(*serverBuilder)

type serverBuilder struct {
	listener       net.Listener
	versionService pbversionsvc.VersionServiceServer
	logger         log.Logger
	grpcOpts       []grpc.ServerOption
}

func newServerBuilder(listener net.Listener) *serverBuilder {
	return &serverBuilder{
		listener: listener,
		logger:   log.NewNopLogger(),
		grpcOpts: make([]grpc.ServerOption, 0),
	}
}

// Listen starts a new net.Listener on the given address.
//
// The address must conform to the standard listener address format used by
// CometBFT, i.e. "<protocol>://<address>". For example,
// "tcp://127.0.0.1:26670".
func Listen(addr string) (net.Listener, error) {
	parts := strings.SplitN(addr, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf(
			"invalid listening address %s (use fully formed addresses, including the tcp:// or unix:// prefix)",
			addr,
		)
	}
	return net.Listen(parts[0], parts[1])
}

// WithVersionService enables the version service on the CometBFT server.
func WithVersionService() Option {
	return func(b *serverBuilder) {
		b.versionService = versionservice.New()
	}
}

// WithLogger enables logging using the given logger. If not specified, the
// gRPC server does not log anything.
func WithLogger(logger log.Logger) Option {
	return func(b *serverBuilder) {
		b.logger = logger.With("module", "grpc-server")
	}
}

// WithGRPCOption allows one to specify Google gRPC server options during the
// construction of the CometBFT gRPC server.
func WithGRPCOption(opt grpc.ServerOption) Option {
	return func(b *serverBuilder) {
		b.grpcOpts = append(b.grpcOpts, opt)
	}
}

// Serve constructs and runs a CometBFT gRPC server using the given listener
// and options.
//
// This function only returns upon error, otherwise it blocks the current
// goroutine.
func Serve(listener net.Listener, opts ...Option) error {
	b := newServerBuilder(listener)
	for _, opt := range opts {
		opt(b)
	}
	server := grpc.NewServer(b.grpcOpts...)
	if b.versionService != nil {
		pbversionsvc.RegisterVersionServiceServer(server, b.versionService)
		b.logger.Debug("Registered version service")
	}
	b.logger.Info("serve", "msg", fmt.Sprintf("Starting gRPC server on %s", listener.Addr()))
	return server.Serve(b.listener)
}
