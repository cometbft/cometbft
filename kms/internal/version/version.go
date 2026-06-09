// Package version exposes the cometkms build version.
package version

// Version is the cometkms semantic version. Overridable at build time via
// -ldflags "-X github.com/cometbft/cometbft/kms/internal/version.Version=...".
var Version = "0.1.0-dev"

// String returns the cometkms version string.
func String() string { return Version }
