*January 24, 2024*

This release fixes a problem introduced in `v0.38.3`: if an application
updates the value of ConsensusParam `VoteExtensionsEnableHeight` to the same value
(actually a "noop" update) this is accepted in `v0.38.2` but rejected under some
conditions in `v0.38.3` and `v0.38.4`. Even if rejecting a useless update would make sense
in general, in a point release we should not reject a set of inputs to
a function that was previuosly accepted (unless there is a good reason
for it). The goal of this release is to accept again all "noop" updates, like `v0.38.2` did.

