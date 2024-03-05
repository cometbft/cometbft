- `[privval]` DO NOT require extension signature from privval if vote
  extensions are disabled. Remote signers must ONLY sign the extension if
  `sign_extension` flag in `SignVoteRequest` is true.
  [\#2496](https://github.com/cometbft/cometbft/pull/2496)
