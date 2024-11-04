package core

// _testGenesis is a GenesisDoc used in the `Environment.InitGenesisChunks` tests.
// It is the genesis that the ci.toml e2e tests uses.
const _testGenesis = `
{
  "genesis_time": "2024-10-02T11:53:14.181969Z",
  "chain_id": "ci",
  "initial_height": "1000",
  "consensus_params": {
    "block": {
      "max_bytes": "4194304",
      "max_gas": "10000000"
    },
    "evidence": {
      "max_age_num_blocks": "14",
      "max_age_duration": "1500000000",
      "max_bytes": "1048576"
    },
    "validator": {
      "pub_key_types": [
        "ed25519"
      ]
    },
    "version": {
      "app": "1"
    },
    "synchrony": {
      "precision": "500000000",
      "message_delay": "2000000000"
    },
    "feature": {
      "vote_extensions_enable_height": "0",
      "pbts_enable_height": "0"
    }
  },
  "validators": [
    {
      "address": "75C02D9AC4DB1A1F802CECF9EADB4CC4CB952AE6",
      "pub_key": {
        "type": "tendermint/PubKeyEd25519",
        "value": "01E+NeFFiH8D2uQHJ+X45wePtfVPs9pncpBnv/g9DQs="
      },
      "power": "100",
      "name": "validator01"
    }
  ],
  "app_hash": "",
  "app_state": {
    "initial01": "a",
    "initial02": "b",
    "initial03": "c"
  }
}`
