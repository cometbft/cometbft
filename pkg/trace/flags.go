package trace

const (
	FlagTracePushConfig             = "trace-push-url"
	FlagTracePullAddress            = "trace-pull-address"
	FlagTracePushConfigDescription  = "URL of the trace push server"
	FlagTracePullAddressDescription = "address to listen on for pulling trace data"

	FlagPyroscopeURL              = "pyroscope-url"
	FlagPyroscopeURLDescription   = "URL of the Pyroscope instance to use for continuous profiling. If not specified, profiling will not be enabled"
	FlagPyroscopeTrace            = "pyroscope-trace"
	FlagPyroscopeTraceDescription = "enable adding trace data to pyroscope profiling"
)
