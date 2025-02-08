package fuzz

// Update import in mconn/connection.go to use this config
type FuzzConnConfig struct {
	Mode         int
	MaxDelay     int
	ProbDropRW   float64
	ProbDropConn float64
	ProbSleep    float64
}
