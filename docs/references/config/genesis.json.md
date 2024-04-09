---
order: 1
parent:
  title: genesis.json
  description: The network genesis file
  order: 1
---
# genesis.json
It is **crucial** that all nodes in a network must have _exactly_ the same contents in their `genesis.json` file.

On first start, the network parameters are read from the `genesis.json` file.
On subsequent starts (node recovery), the `genesis.json` file is ignored.

### Example
```json
{
  "genesis_time": "2024-03-01T20:22:57.532998Z",
  "chain_id": "test-chain-HfdKnD",
  "initial_height": "0",
  "consensus_params": {
    "block": {
      "max_bytes": "4194304",
      "max_gas": "10000000"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000",
      "max_bytes": "1048576"
    },
    "validator": {
      "pub_key_types": [
        "ed25519"
      ]
    },
    "version": {
      "app": "0"
    },
    "feature": {
      "vote_extensions_enable_height": "1"
      "pbts_enable_height": "1"
    }
  },
  "validators": [
    {
      "address": "E74FBE24164CFC4F88E311C3AC92E63D0DC310D8",
      "pub_key": {
        "type": "tendermint/PubKeyEd25519",
        "value": "UjxDQgVTlHJOZ7axpMl/iczMIJXiQpFxCFjwKGvzYqE="
      },
      "power": "10",
      "name": ""
    }
  ],
  "app_hash": ""
}
```

For a production example, you can see [here](https://github.com/cosmos/mainnet/tree/master/genesis)
the history of genesis files for the Cosmos Hub network.

## genesis_time
Timestamp of the genesis file creation.

| Value type          | string                     |
|:--------------------|:---------------------------|
| **Possible values** | RFC3339-formatted datetime |

RFC3339 has multiple representation. The one we use here has 6 digits for sub-second representation.

## chain_id
The chain ID of the blockchain network.

| Value type          | string                            |
|:--------------------|:----------------------------------|
| **Possible values** | usually in `"name-number"` format |

Cannot be empty.

Can be maximum 50 UTF-8-encoded character.

The `number` part is typically a revision number of the blockchain, starting at `1` and incrementing each time the network
undergoes a hard fork.

## initial_height
Initial height at genesis.

| Value type          | string      |
|:--------------------|:------------|
| **Possible values** | &gt;= `"0"` |

When a hard fork happens, a new chain can start from a higher initial height by setting this parameter.

> Notes:

>> If a Height `"0"` is specified in `initial_heigth`, then CometBFT during the genesis file validation, will change the
> initial height parameter to `"1"`.

>> Note: A height in CometBFT is an `int64` integer therefore its maximum value is `9223372036854775807`

## consensus_params.block.max_bytes
Maximum block size in bytes.

| Value type          | string     |
|:--------------------|:-----------|
| **Possible values** | &gt; `"0"` |
| **Possible values** | `"-1"`     |

`"-1"` means the hard-wired maximum of 104857600 bytes.
This is typically used by applications that want to control the maximum block size
in `PrepareProposal` (and validate it in `ProcessProposal`).
In this scenario, CometBFT will always send all transactions in its mempool in `PrepareProposalRequest`.

This parameter cannot be `0`.

## consensus_params.block.max_gas
Maximum gas allowed.

| Value type          | string      |
|:--------------------|:------------|
| **Possible values** | &gt;= `"0"` |
| **Possible values** | `"-1"`      |

`"-1"` means unlimited gas.

## consensus_params.evidence.max_age_num_blocks
Limit evidence of misbehaviour based on how old the evidence is. This parameter limits the evidence based on its block
number.

| Value type          | string     |
|:--------------------|:-----------|
| **Possible values** | &gt; `"0"` |

Setting this parameter limits accepting evidence of malfeasance to blocks with height that are between the last block
height, and the last block height minus this parameter. Older evidence is discarded.

To clarify: if this parameter is `5` and the last block height is `10`, any evidence with block numbers `6,7,8 or 9` are
accepted. Evidence with earlier block numbers are discarded.

## consensus_params.evidence.max_age_duration
Limit evidence of malfeasance based on how old the evidence is. This parameter limits the evidence based on its
timestamp.

| Value type          | string     |
|:--------------------|:-----------|
| **Possible values** | &gt; `"0"` |
|                     | `""`       |

Setting this parameter limits accepting evidence of malfeasance to blocks with a timestamp that are between the last
block's timestamp, and the last block's timestamp minus this parameter. Older evidence is discarded.

To clarify: if this parameter is `86400` (one day in seconds) and the last block's timestamp is
`2024-03-01T00:00:00.000000Z`, any evidence with block timestamp starting with `2024-02-29T...` are
accepted. Evidence with earlier timestamps are discarded.

## consensus_params.evidence.max_bytes
Limit on the size in bytes devoted to evidence information in a block.

| Value type          | string                                                                                 |
|:--------------------|:---------------------------------------------------------------------------------------|
| **Possible values** | &gt;= `"0"`; &lt;= [consensus_params.block.max_bytes](#consensus_paramsblockmax_bytes) |

## consensus_params.validator.pub_key_types
List of public key types accepted for consensus validation.

| Value type                           | array of string |
|:-------------------------------------|:----------------|
| **Possible values of array strings** | `"ed25519"`     |
|                                      | `"secp256k1"`   |

Note that the restriction to the specified types is only checked when the network is already running and there are
changes in the validator set. This means that in the genesis file's [validators](#validators) section, you can define
both types of consensus keys without restriction.

## consensus_params.version.app
Consensus parameter set version.

| Value type                           | string |
|:-------------------------------------|:-------|
| **Possible values of array strings** | `"0"`  |

This is used by the ABCI application on the chain. It can also be updated on chain by the ABCI application.

## consensus_params.feature_params.vote_extensions_enable_height
If you enable vote extensions in the genesis file, you have to decide at which height are the extensions enabled.

> Note: If a value `0` is used, it means the extension will not be enabled. In order for it to be enabled, a
> value greater than `0` needs to be specified. This value can also be set on chain by the ABCI application.

| Value type                           | string      |
|:-------------------------------------|:------------|
| **Possible values of array strings** | &gt;= `"0"` |

This can be enabled by governance if it is not enabled in the genesis file. Once active, vote extensions cannot be
disabled. (Vote extensions can be still disabled by governance, if they are supposed to be enabled at a future height.)

## validators
List of initial validators for consensus.

| Value type                              | array of objects |                                                                                         |
|:----------------------------------------|:-----------------|-----------------------------------------------------------------------------------------|
| **Mandatory keys of each array object** | address          | See [address](priv_validator_key.json.md#address) in priv_validator_key.json            |
|                                         | pub_key          | See [pub_key.type](priv_validator_key.json.md#pub_keytype)                              |
|                                         |                  | and [pub_key.value](priv_validator_key.json.md#pub_keyvalue) in priv_validator_key.json |
|                                         | power            | &gt;  `"0"`                                                                             |
|                                         | name             | string or `""`                                                                          |

## app_hash
The initial AppHash, represented by the state embedded in the genesis file.

| Value type          | string             |
|:--------------------|:-------------------|
| **Possible values** | hex-encoded number |
|                     | ""                 |

## app_state
A raw encoded JSON value that has the application state encoded in it.

| Value type          | string                 |
|:--------------------|:-----------------------|
| **Possible values** | raw bytes JSON-encoded |
|                     | ""                     |
