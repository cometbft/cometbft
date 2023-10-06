package versionservice

import (
	context "context"

	v1 "github.com/cometbft/cometbft/api/cometbft/services/version/v1"
	"github.com/cometbft/cometbft/version"
)

type versionServiceServer struct{}

// New creates a new CometBFT version service server.
func New() v1.VersionServiceServer {
	return &versionServiceServer{}
}

// GetVersion implements v1.VersionServiceServer
func (s *versionServiceServer) GetVersion(context.Context, *v1.GetVersionRequest) (*v1.GetVersionResponse, error) {
	return &v1.GetVersionResponse{
		Node:  version.TMCoreSemVer,
		Abci:  version.ABCIVersion,
		P2P:   version.P2PProtocol,
		Block: version.BlockProtocol,
	}, nil
}
