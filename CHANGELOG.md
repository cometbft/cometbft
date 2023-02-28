# CHANGELOG

## Unreleased

### BREAKING CHANGES

- `[tools/tm-signer-harness]` Set OS home dir to instead of the hardcoded PATH.
  ([\#6498](https://github.com/tendermint/tendermint/pull/6498))
- `[state]` Move pruneBlocks from node/state to state/execution.
  ([\#6541](https://github.com/tendermint/tendermint/pull/6541))
- `[p2p]` Remove unused p2p/trust package
  ([\#9625](https://github.com/tendermint/tendermint/pull/9625))
- `[rpc]` Remove global environment and replace with constructor
  ([\#9655](https://github.com/tendermint/tendermint/pull/9655))
- `[node]` Move DBContext and DBProvider from the node package to the config
  package. ([\#9655](https://github.com/tendermint/tendermint/pull/9655))
- `[inspect]` Add a new `inspect` command for introspecting
  the state and block store of a crashed tendermint node.
  ([\#9655](https://github.com/tendermint/tendermint/pull/9655))
- `[metrics]` Move state-syncing and block-syncing metrics to
  their respective packages. Move labels from block_syncing
  -> blocksync_syncing and state_syncing -> statesync_syncing
  ([\#9682](https://github.com/tendermint/tendermint/pull/9682))

### BUG FIXES

- `[docker]` Ensure Docker image uses consistent version of Go.
  ([\#9462](https://github.com/tendermint/tendermint/pull/9462))
- `[abci-cli]` Fix broken abci-cli help command.
  ([\#9717](https://github.com/tendermint/tendermint/pull/9717))

### FEATURES

- `[config]` Introduce `BootstrapPeers` to the config to allow
  nodes to list peers to be added to the addressbook upon start up.
  ([\#9680](https://github.com/tendermint/tendermint/pull/9680))
- `[proxy]` Introduce `NewUnsyncLocalClientCreator`, which allows local ABCI
  clients to have the same concurrency model as remote clients (i.e. one
  mutex per client "connection", for each of the four ABCI "connections").
  ([\#9830](https://github.com/tendermint/tendermint/pull/9830))

### IMPROVEMENTS

- `[crypto/merkle]` Improve HashAlternatives performance
  ([\#6443](https://github.com/tendermint/tendermint/pull/6443))
- `[p2p/pex]` Improve addrBook.hash performance
  ([\#6509](https://github.com/tendermint/tendermint/pull/6509))
- `[crypto/merkle]` Improve HashAlternatives performance
  ([\#6513](https://github.com/tendermint/tendermint/pull/6513))
- `[pubsub]` Performance improvements for the event query API
  ([\#7319](https://github.com/tendermint/tendermint/pull/7319))
- `[rpc]` Enable caching of RPC responses
  ([\#9650](https://github.com/tendermint/tendermint/pull/9650))

---

CometBFT is a fork of [Tendermint Core](https://github.com/tendermint/tendermint) as of late December 2022.

## Bug bounty

Friendly reminder, we have a [bug bounty program](https://hackerone.com/cosmos).

## Previous changes

For changes released before the creation of CometBFT, please refer to the Tendermint Core [CHANGELOG.md](https://github.com/tendermint/tendermint/blob/a9feb1c023e172b542c972605311af83b777855b/CHANGELOG.md).

