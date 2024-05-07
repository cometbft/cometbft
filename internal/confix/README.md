# Confix

`Confix` is a configuration management tool that allows you to manage your configuration via CLI.

It is based on the [CometBFT RFC 019](https://github.com/cometbft/cometbft/blob/5013bc3f4a6d64dcc2bf02ccc002ebc9881c62e4/docs/rfc/rfc-019-config-version.md).

## Usage

```shell
cometbft config fix --help
```

### Get

Get a configuration value, e.g.:

```shell
cometbft config get pruning # gets the value pruning
cometbft config get chain-id # gets the value chain-id
```

### Set

Set a configuration value, e.g.:

```shell
cometbft config set pruning "enabled" # sets the value pruning
cometbft config set chain-id "foo-1" # sets the value chain-id
```
### Migrate

Migrate a configuration file to a new version:

```shell
cometbft config migrate v0.38 # migrates defaultHome/config/config.toml to the latest v0.38 config
```

### Diff

Get the diff between a given configuration file and the default configuration
file, e.g.:

```shell
cometbft config diff v0.38 # gets the diff between defaultHome/config/config.toml and the latest v0.38 config
```

### View

View a configuration file, e.g:

```shell
cometbft config view # views the current config
```

## Credits

This project is based on the [CometBFT RFC 019](https://github.com/cometbft/cometbft/blob/5013bc3f4a6d64dcc2bf02ccc002ebc9881c62e4/docs/rfc/rfc-019-config-version.md) and their own implementation of [confix](https://github.com/cometbft/cometbft/blob/v0.36.x/scripts/confix/confix.go).
Most of the code is copied over from [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/tree/main/tools/confix).
