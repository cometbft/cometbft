package trace

import (
	"errors"
	"os"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
)

// Entry is an interface for all structs that are used to define the schema for
// traces.
type Entry interface {
	// Table defines which table the struct belongs to.
	Table() string
}

// Tracer defines the methods for a client that can write and read trace data.
type Tracer interface {
	Write(Entry)
	IsCollecting(table string) bool
	Stop()
}

func NewTracer(cfg *config.Config, logger log.Logger, chainID, nodeID string) (Tracer, error) {
	switch cfg.Instrumentation.TraceType {
	case "local":
		return NewLocalTracer(cfg, logger, chainID, nodeID)
	case "noop":
		return NoOpTracer(), nil
	default:
		logger.Error("unknown tracer type, using noop", "type", cfg.Instrumentation.TraceType)
		return NoOpTracer(), nil
	}
}

func NoOpTracer() Tracer {
	return &noOpTracer{}
}

type noOpTracer struct{}

func (n *noOpTracer) Write(_ Entry) {}
func (n *noOpTracer) ReadTable(_ string) (*os.File, error) {
	return nil, errors.New("no-op tracer does not support reading")
}
func (n *noOpTracer) IsCollecting(_ string) bool { return false }
func (n *noOpTracer) Stop()                      {}
