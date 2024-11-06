*November 6, 2024*

This release supersedes [`v0.38.14`](#v03814), which mistakenly updated the Go version to
`1.23`, introducing an unintended breaking change. It sets the Go version back
to `1.22.7` by reverting [\#4297](https://github.com/cometbft/cometbft/pull/4297).

The release includes the bug fixes, performance improvements, and importantly,
the fix for the security vulnerability in the vote extensions (VE) validation
logic that were part of `v0.38.14`. For more details, please refer to [ASA-2024-011](https://github.com/cometbft/cometbft/security/advisories/GHSA-p7mv-53f2-4cwj).
