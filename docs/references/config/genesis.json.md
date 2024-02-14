---
order: 1
parent:
  title: genesis.json
  description: The network genesis file
  order: 1
---
# genesis.json
On first start, the network parameters are read from the `genesis.json` file into a freshly created database.

On subsequent starts, the `genesis.json` file is ignored. If you want to re-read the file, the CometBFT database has
to be deleted. (Run `cometbft unsafe-reset-all --help` for more information.)

## genesis_time
TBD

## chain_id
TBD

## initial_height
TBD

## consensus_params.block.max_bytes
TBD

## consensus_params.block.max_gas
TBD

## consensus_params.evidence.max_age_num_blocks
TBD

## consensus_params.evidence.max_age_duration
TBD

## consensus_params.evidence.max_bytes
TBD

## consensus_params.validator.pub_key_types
TBD

## consensus_params.version.app
TBD

## consensus_params.abci.vote_extensions_enable_height
TBD

## validators
TBD

## app_hash
TBD
