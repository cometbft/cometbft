# Upgrading CometBFT

This guide provides instructions for upgrading to specific versions of CometBFT.

## Unreleased

### Config Changes

* A new config field, `BootstrapPeers` has been introduced as a means of
  adding a list of addresses to the addressbook upon initializing a node. This is an
  alternative to `PersistentPeers`. `PersistentPeers` shold be only used for
  nodes that you want to keep a constant connection with i.e. sentry nodes
* The field `Version` in the mempool section has been removed. The priority
  mempool (what was called version `v1`) has been removed (see below), thus
  there is only one implementation of the mempool available (what was called
  `v0`).
* Config fields `TTLDuration` and `TTLNumBlocks`, which were only used by the priority
  mempool, have been removed.

### ABCI Changes

* The `ABCIVersion` is now `2.0.0`.
* Added new ABCI methods `ExtendVote`, and `VerifyVoteExtension`.
  Applications upgrading to v0.38.0 must implement these methods as described
  [here](./spec/abci/abci++_app_requirements.md)
* Removed methods `BeginBlock`, `DeliverTx`, `EndBlock`, and replaced them by
  method `FinalizeBlock`. Applications upgrading to v0.38.0 must refactor
  the logic handling the methods removed to handle `FinalizeBlock`.
* The Application's hash (or any data representing the Application's current state)
  is known by the time `FinalizeBlock` finishes its execution.
  Accordingly, the `app_hash` parameter has been moved from `ResponseCommit`
  to `ResponseFinalizeBlock`.
* For details, please see the updated [specification](spec/abci/README.md)


### Mempool Changes

* The priority mempool (what was referred in the code as version `v1`) has been
  removed. There is now only one mempool (what was called version `v0`), that
  is, the default implementation as a queue of transactions. 
* In the protobuf message `ResponseCheckTx`, fields `sender`, `priority`, and
  `mempool_error`, which were only used by the priority mempool, were removed
  but still kept in the message as "reserved".
