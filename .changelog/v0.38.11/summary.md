*August 12, 2024*

This release fixes a panic in consensus where CometBFT would previously panic
if there's no extension signature in non-nil Precommit EVEN IF vote extensions
themselves are disabled.

It also includes a few other bug fixes and performance improvements.
