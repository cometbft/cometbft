package mconn

// FuzzConnConfig is a configuration for fuzzing the connection
type FuzzConnConfig struct {
	Mode         int
	MaxDelay     int
	ProbDropRW   float64
	ProbDropConn float64
	ProbSleep    float64
}
