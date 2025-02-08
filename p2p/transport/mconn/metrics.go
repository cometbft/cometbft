package mconn

// Metrics contains metrics exposed by MConnection
type Metrics struct {
	// TODO: implement metrics
}

// NopMetrics returns no-op Metrics
func NopMetrics() *Metrics {
	return &Metrics{}
}
