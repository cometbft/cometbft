# Consensus Parameters

Consensus Parameters are global parameters that apply to all nodes in a blockchain.

They enforce certain limits in the blockchain, like the maximum size of blocks,
amount of gas used in a block, and the maximum acceptable age of evidence.
They can also enable some features (e.g., vote extensions or Proposer-Based
Timestamps) as well as configuring the operation of CometBFT nodes or a
specific feature.

Consensus parameters are configured (set or updated) by the ABCI application.
The configuration of consensus parameters must be deterministic, so that
all full nodes have the same value of parameters at a given height.


## Reference

The `ConsensusParams` type defines the consensus parameters adopted by CometBFT.

The table below overviews the existing categories of consensus parameters.
Notice that it presents two names for the fields: the name adopted by the
Protocol Buffers definition and the names used in the Go implementation:

| Name                    | Type                           | Description                                                             | Field Number |
|-------------------------|--------------------------------|-------------------------------------------------------------------------|:------------:|
| block / `Block`         | [BlockParams](#block)          | Parameters limiting the block and gas.                                  | 1            |
| evidence / `Evidence`   | [EvidenceParams](#evidence)    | Parameters determining the validity of evidences of Byzantine behavior. | 2            |
| validator / `Validator` | [ValidatorParams](#validator)  | Parameters limiting the types of public keys validators can use.        | 3            |
| version   / `Version`   | [VersionParams](#version)      | The version of specific components of CometBFT.                         | 4            |
| synchrony / `Synchrony` | [SynchronyParams](#synchrony)  | Parameters determining the validity of block timestamps.                | 6            |
| feature / `Feature`     | [FeatureParams](#feature) | Parameters for configuring the height from which features are enabled.  | 7            |

The following sections define each type of consensus parameters, with the same format as the above table.

### Block

The `BlockParams` parameters define limits on the block size and gas:

| Name                   | Type  | Description                                             | Field Number |
|------------------------|-------|---------------------------------------------------------|:------------:|
| max_bytes / `MaxBytes` | int64 | Maximum size of a block, in bytes.                      | 1            |
| max_gas / `MaxGas`     | int64 | Maximum gas wanted by transactions included in a block. | 2            |

#### MaxBytes

The max_bytes / `MaxBytes` parameter determines the maximum size in bytes of a
complete Protocol-Buffers encoded block.
This configured maximum block size is enforced by CometBFT.

This implies that the maximum size of transactions included in a block is the
block maximum size minus the expected size of the block header, the validator
set, and any included evidence in the block.

The application should be aware that validators _may_ produce and broadcast
blocks with up to the configured maximum block size.
As a result, the consensus
[timeout parameters](../../docs/explanation/core/configuration.md#consensus-timeouts-explained)
adopted by nodes should be configured so as to account for the worst-case
latency for the delivery of a full block with maximum size to all validators.

It can be set to any positive value that is lower than the hard-coded maximum
size for a block, which is 100MB.

It can also be set to `-1`.
In this case, the hard-coded maximum size for a block, which is 100MB, is adopted.
This setup has implications on the interaction of CometBFT with the ABCI
application, as detailed [here]().

#### MaxGas

CometBFT does not enforce the maximum wanted gas for committed blocks.

CometBFT does use this parameter to limit the number of transactions included
in a proposed block.

This parameter should be enforced by the ABCI application.
Blocks that violate configured maximum wanted gas have potentially been
proposed by Byzantine validators.
It is responsibility of the ABCI application handling blocks whose wanted gas
exceeds the configured maximum gas when processing the block.

The max_gas / `MaxGas` parameter must be greater or equal to -1.
If set to -1, no limit is enforced.

### Evidence

The `EvidenceParams` parameters determine the validity of evidences of Byzantine behavior:

| Name                                   | Type                                       | Description                                                          | Field Number |
|----------------------------------------|--------------------------------------------|----------------------------------------------------------------------|:------------:|
| max_age_num_blocks / `MaxAgeNumBlocks` | int64                                      | Max age of evidence, in blocks.                                      | 1            |
| max_age_duration / `MaxAgeDuration`    | [google.protobuf.Duration][proto-duration] | Max age of evidence, in time.                                        | 2            |
| max_bytes / `MaxBytes`                 | int64                                      | Maximum size in bytes of evidence allowed to be included in a block. | 3            |

The evidence parameters are enforced by CometBFT.

A correct validator cannot propose, accept, or commit a block containing:

- More that max_bytes / `MaxBytes` of evidence payload
- Expired evidences, produced at a height smaller than
  max_age_num_blocks / `MaxAgeNumBlocks` from the last block's heigh
  AND at a time earlier than max_age_duration / `MaxAgeDuration` from the last block's time.

#### MaxAgeNumBlocks

The recommended method for calculating the value max_age_num_blocks /
`MaxAgeNumBlocks` parameter to divide the max_age_duration / `MaxAgeDuration`
parameter by the average block time in the blockchain.

It must be positive, i.e., larger than 0.

#### MaxAgeDuration

The recommended value of max_age_duration / `MaxAgeNumBlocks` parameter should correspond to
the application's "unbonding period" or other similar mechanism for handling
[Nothing-At-Stake attacks](https://github.com/ethereum/wiki/wiki/Proof-of-Stake-FAQ#what-is-the-nothing-at-stake-problem-and-how-can-it-be-fixed).

It must be positive, i.e., larger than 0.

#### MaxBytes

Its value must not exceed the maximum size of a block minus the size of the
block's header and validator set.

It must be positive, i.e., larger than 0.

### Validator

The `ValidatorParams` parameters restrict the public key types validators can use:

| Name                          | Type            | Description                                                           | Field Number |
|-------------------------------|-----------------|-----------------------------------------------------------------------|:------------:|
| pub_key_types / `PubKeyTypes` | repeated string | List of accepted public key types. Uses same naming as `PubKey.Type`. | 1            |

The pub_key_types / `PubKeyTypes` parameter uses ABCI public keys naming, not Amino names.

### Version

The `VersionParams` parameters contain the version of specific components of CometBFT:

| Name        | Type   | Description                   | Field Number |
|-------------|--------|-------------------------------|:------------:|
| app / `App` | uint64 | The ABCI application version. | 1            |

The app / `App` parameter was named app_version / `AppVersion` in CometBFT 0.34.

### ABCI (deprecated)

| Name                                                         | Type  | Description                                       | Field Number |
|--------------------------------------------------------------|-------|---------------------------------------------------|:------------:|
| vote_extensions_enable_height / `VoteExtensionsEnableHeight` | int64 | The height where vote extensions will be enabled. | 1            |

The `ABCIParams` type has been **deprecated** from CometBFT `v1.0`.

The vote_extensions_enable_height / `VoteExtensionsEnableHeight` parameter is now part of `FeatureParams`.

### Synchrony

The `SynchronyParams` parameters determine the validity of block timestamps:

| Name                           | Type                                       | Description                                                                                                             | Field Number |
|--------------------------------|--------------------------------------------|-------------------------------------------------------------------------------------------------------------------------|:------------:|
| precision / `Precision`        | [google.protobuf.Duration][proto-duration] | Bound for how skewed a proposer's clock may be from any validator on the network while still producing valid proposals. | 1            |
| message_delay / `MessageDelay` | [google.protobuf.Duration][proto-duration] | Bound for how long a proposal message may take to reach all validators on a network and still be considered valid.      | 2            |

These parameters are part of the Proposer-Based Timestamps (PBTS) algorithm.
For more information on the relationship of the synchrony parameters to
block timestamps validity, refer to the [PBTS specification][pbts].

### Feature

The `FeatureParams` parameters configure the height from which features of CometBFT are enabled:

| Name                                                         | Type  | Description                                                       | Field Number |
|--------------------------------------------------------------|-------|-------------------------------------------------------------------|:------------:|
| vote_extensions_enable_height / `VoteExtensionsEnableHeight` | int64 | First height during which vote extensions will be enabled.        | 1            |
| pbts_enable_height / `PbtsEnableHeight`                      | int64 | Height at which Proposer-Based Timestamps (PBTS) will be enabled. | 2            |

If a feature parameter is set to 0, which is the default, then the
corresponding feature is disabled.

From the configured enable height, and for all subsequent heights, the
corresponding feature will be enabled.

The feature enable height parameters cannot be set to heights lower or equal to
the current blockchain height.

Once a feature enable height is set and reached by the blockchain, the feature
enable height cannot be updated.

## Initial Parameters

TODO: genesis

## Updating Parameters

The ABCI application can configure values for consensus parameters
using the [`InitChain`](./abci%2B%2B_methods.md#initchain) and
[`FinalizeBlock`](./abci%2B%2B_methods.md#finalizeblock) ABCI methods,
whose ABCI responses include a field for a `ConsensusParams` instance.

If the `ConsensusParams` instance received by CometBFT is empty,
no consensus parameter is updated.

Otherwise, each field that is set (not empty) is applied in full, namely, every
parameter defined in that category of consensus parameters is updated.
For example, if the `BlockParams` field is set because the application wants to
update the configured `MaxBytes` consensus parameter, the application MUST also
set every other parameter of `BlockParams` (e.g., `MaxGas`), even if the
application does not want to update the value of those parameters.

**Important:** If the application does not set values for every parameter of a category of
consensus parameters that is set (not empty) in the returned `ConsensusParams`
instance, the **zero** values for those parameters will be considered.
This can be very problematic, as some of the zero values for parameters are
invalid, and the whole update is rejected.

### Update Mechanism

TODO

[pbts]: ../consensus/proposer-based-timestamp/README.md
[bfttime]: ../consensus/bft-time.md
[proto-duration]: https://protobuf.dev/reference/protobuf/google.protobuf/#duration
