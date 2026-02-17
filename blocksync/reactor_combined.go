package blocksync

func (r *Reactor) poolCombinedModeRoutine() {
	r.Logger.Info("Starting blocksync pool routine (combined mode)")

	// todo: write loop similarly to poolRoutine
	// todo: fetch two blocks
	// todo: check consensus height (maybe drop the block)
	// todo: fetch latest state
	// todo: light client verification
	// todo: if ok, call consensus.IngestVerifiedBlock(...) error
	// todo: handle error: noop, already processed, failure, etc...
}
