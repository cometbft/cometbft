// CometBFT service.Service interface modelling.
// Code: https://github.com/cometbft/cometbft/blob/main/libs/service/service.go

// Service defines a service that can be started, stopped, and reset.
module service {

	/*
	 * Implementation of a Service.
	 * Modelling is very simple: whether it succeeds on starting and stopping.
	 */

	var startSuccess: bool
	var stopSuccess: bool

	// Implementation startup procedure.
	def OnStart(): bool = all {
		startSuccess
	}

	// Implementation shutdown procedure.
	def OnStop(): bool = all {
		stopSuccess
	}

	// Implementation reset procedure.
	def OnReset(): bool = all {
		false // Panics by default - must be overwritten to enable reset.
	}

	// Helper for simplifying Service modelling.
	action implementationUnchanged = all {
		startSuccess' = startSuccess,
		stopSuccess' = stopSuccess,
	}

	/*
	 * BaseService is a generic Service.
	 * Implementations should extend BaseService.
         */

	var started: bool
	var stopped: bool

	action initBaseService = all {
		started' = false,
		stopped' = false,
	}

	// Return true if the service is running.
	def IsRunning(): bool = {
		started and not(stopped)
	}

	// Start the service.
	// If it's already started or stopped, will return an error.
	// If OnStart() returns an error, it's returned by Start()
	action Start(): bool = all {
		not(started),
		not(stopped),
		OnStart,
		started' = true,
		stopped' = stopped, // FIXME: if we get here, stopped == false
		implementationUnchanged,
	}

	// Stop the service.
	// If it's already stopped, will return an error.
	// OnStop must never error.
	action Stop(): bool = all {
		not(stopped),
		started,
		OnStop or true, // We don't care about OnStop output
		stopped' = true,
		started' = started, // FIXME: if we get here, started == true
		implementationUnchanged,
	}

	// Reset the service.
	action Reset(): bool = all {
		stopped,
		OnReset,
		started' = false,
		stopped' = false,
		implementationUnchanged,
	}

	/*
	 * Setup and debugging helpers
	 */

	// Default setup.
	action init = all {
		initBaseService,
		startSuccess' = true,
		stopSuccess' = true,
	}

	action initStartFails = all {
		initBaseService,
		startSuccess' = false,
		stopSuccess' = true,
	}

	action initStopFails = all {
		initBaseService,
		startSuccess' = true,
		stopSuccess' = false,
	}

	type State = {
		isStarted: bool,
		isStopped: bool,
		isRunning: bool,
	}

	def state(): State = {
		{isStarted: started, isStopped: stopped, isRunning: IsRunning}
	}

}
