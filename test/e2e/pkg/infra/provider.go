package infra

// Provider defines an API for manipulating the infrastructure of a
// specific set of testnet infrastructure.
type Provider interface {

	// Setup generates any necessary configuration for the infrastructure
	// provider during testnet setup.
	Setup() error

	// UpdateVersion updates the infrastructure provider's configuration
	// for each node to the one specified
	UpdateVersion() error
}

// NoopProvider implements the provider interface by performing noops for every
// interface method. This may be useful if the infrastructure is managed by a
// separate process.
type NoopProvider struct {
}

func (NoopProvider) Setup() error         { return nil }
func (NoopProvider) UpdateVersion() error { return nil }

var _ Provider = NoopProvider{}
