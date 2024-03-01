---
order: 1
parent:
  title: config.toml
  description: CometBFT general configuration
  order: 3
---
<!---
The current CometBFT documentation template is not perfect for reference manuals.
These style suggestions make it more readable.
--->
<style>
.PageContent table {margin: 0;}           /* tables are left-aligned */
</style>

<!--- Entry template: Use this as a template to add more parameter descriptions.
### tablename.property_key <- full path to property in the file
This is the two-sentence summary of the parameter. It does all kinds of stuff.
```toml
tablename.property_key = "default value"
```

| Value type          | string/bool/int/etc                | <- can be further clarified (duration, hex, base64-encoded)
|:--------------------|:-----------------------------------|
| **Possible values** | `"default value or exact example"` |
|                     | `generic description`              |

Lorem ipsum and everything else that you can think about this value and stuff.
This is the enhanced description and additional notes that is worth talking about.
Description about the possible values too.
--->

# config.toml
The `config.toml` file is a standard [TOML](https://toml.io/en/v1.0.0) file that configures the basic functionality
of CometBFT, including the configuration of the reactors.

All relative paths in the configuration are relative to `$CMTHOME`.
(See [the HOME folder](README.md#the-home-folder) for more details.)

## Base configuration
The root table defines generic node settings. It is implemented in a struct called `BaseConfig`, hence the name.

### version
The version of the CometBFT binary that created or last modified the config file.
```toml
version = "1.0.0"
```

| Value type          | string                  |
|:--------------------|:------------------------|
| **Possible values** | semantic version string |
|                     | `""`                    |

This string validates the configuration file for the binary. The string has to be either a
[valid semver](https://semver.org) string or an empty string. In any other case, the binary halts with an
`ERROR: error in config file: invalid version string` error.

In the future, the code might make restrictions on what version of the file is compatible with what version of the
binary. There is no such check in place right now. Configuration and binary versions are interchangeable.

### proxy_app
The TCP or UNIX socket of the ABCI application or the name of the ABCI application compiled in with the CometBFT
library.
```toml
proxy_app = "tcp://127.0.0.1:26658"
```

| Value type          | string                                             |
|:--------------------|:---------------------------------------------------|
| **Possible values** | TCP Stream socket (`"tcp://127.0.0.1:26658"`)      |
|                     | Unix domain socket (`"unix:///var/run/abci.sock"`) |
|                     | `"kvstore"`                                        |
|                     | `"persistent_kvstore"`                             |
|                     | `"noop"`                                           |

When the ABCI application is written in a different language than Golang, (for example the
[Nomic binary](https://github.com/nomic-io/nomic) is written in Rust) the application can open a TCP port or create a
UNIX domain socket to communicate with CometBFT, while CometBFT runs as a separate process.

The [abci](#abci) parameter is used in conjunction with this parameter to define the protocol used for communication.

In other cases, (for example in the [Gaia binary](https://github.com/cosmos/gaia)) CometBFT is imported as a library
and the configuration entry is unused.

For development and testing, the built-in ABCI applications can be used without additional processes running.

<!--- Todo: describe the built-in applications and their use --->

### moniker
A custom human-readable name for this node.
```toml
moniker = "my.host.name"
```

| Value type          | string                                                   |
|:--------------------|:---------------------------------------------------------|
| **Possible values** | any human-readable string                                |

The main use of this entry is to keep track of the different nodes in a local environment. For example the `/status` RPC
endpoint will return the node moniker in the `.result.moniker` key.

Monikers do not need to be unique. They are for local administrator use and troubleshooting.

Nodes on the peer-to-peer network are identified by `nodeID@IP:port` as discussed in the
[node_key.json](node_key.json.md) section.

### db_backend
The chosen database backend for the node.
```toml
db_backend = "goleveldb"
```

| Value type          | string        | dependencies  | GitHub                                           |
|:--------------------|:--------------|:--------------|:-------------------------------------------------|
| **Possible values** | `"goleveldb"` | pure Golang   | [goleveldb](https://github.com/syndtr/goleveldb) |
|                     | `"cleveldb"`  | requires gcc  | [leveldb](https://github.com/google/leveldb)     |
|                     | `"boltdb"`    | pure Golang   | [bbolt](https://github.com/etcd-io/bbolt)        |
|                     | `"rocksdb"`   | requires gcc  | [grocksdb](https://github.com/linxGnu/grocksdb)  |
|                     | `"badgerdb"`  | pure Golang   | [badger](https://github.com/dgraph-io/badger)    |
|                     | `"pebbledb"`  | pure Golang   | [pebble](https://github.com/cockroachdb/pebble)  |

During the build process, only the `goleveldb` library is built into the binary.
To add support for alternative databases, you need to add them in the build tags.
For example: `go build -tags cleveldb,rocksdb`.

The RocksDB fork has API changes from the upstream RocksDB implementation. All other databases claim a stable API.

The CometBFT team tests rely on the GoLevelDB implementation. All other implementations are considered experimental from
a CometBFT perspective. However, the GoLevelDB library has not been maintained in the past two years.
Choose your poison.

### db_dir
The directory path where the database is stored.
```toml
db_dir = "data"
```

| Value type          | string                                                      |
|:--------------------|:------------------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME` |
|                     | absolute directory path                         |

The default relative path translates to `$CMTHOME/data`. In case `$CMTHOME` is unset, it defaults to
`$HOME/.cometbft/data`.

### log_level
A comma-separated list of `module:level` pairs that describe the log level of each module. Alternatively, a single word
can be set which will apply that log level to all modules.
```toml
log_level = "info"
```

| Value type     | string          |                                        |
|:---------------|:----------------|----------------------------------------|
| **Modules**    | `"main"`        | CometBFT main application logs         |
|                | `"consensus"`   | consensus reactor logs                 |
|                | `"p2p"`         | p2p reactor logs                       |
|                | `"pex"`         | Peer Exchange logs                     |
|                | `"proxy"`       | ABCI proxy service (MultiAppConn) logs |
|                | `"abci-client"` | ABCI client service logs               |
|                | `"rpc-server"`  | RPC server logs                        |
|                | `"txindex"`     | Indexer service logs                   |
|                | `"events"`      | Events service logs                    |
|                | `"pubsub"`      | PubSub service logs                    |
|                | `"evidence"`    | Evidence reactor logs                  |
|                | `"statesync"`   | StateSync reactor logs                 |
|                | `"mempool"`     | Mempool reactor logs                   |
|                | `"blocksync"`   | BlockSync reactor logs                 |
|                | `"state"`       | Pruner service logs                    |
|                | `"*"`           | All modules                            |
| **Log levels** | `"debug"`       |                                        |
|                | `"info"`        |                                        |
|                | `"error"`       |                                        |
|                | `"none"`        |                                        |

At the end of a `module:level` list, a default log level can be set for modules with no level set. Use `*` instead of a
module name to set a default log level. The default is `*:info`.

Examples:

Set the consensus reactor to `none` log level and the `p2p` reactor to `debug`. Everything else should be set to error:
```toml
log_level = "consensus:none,p2p:debug,*:error"
```
Set RPC server logs to `debug` and leave everything else at `info`:
```toml
log_level = "rpc-server:debug"
```

### log_format
Define the output format of the logs.
```toml
log_format = "plain"
```

| Value type          | string    |
|:--------------------|:----------|
| **Possible values** | `"plain"` |
|                     | `"json"`  |

`plain` provides ANSI color-coded plain-text logs.

`json` provides JSON objects (one per line, not prettified) using the following (incomplete) schema:
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://cometbft.com/log.schema.json",
  "title": "JSON log",
  "description": "A log entry in JSON object format",
  "type": "object",
  "properties": {
    "level": {
      "description": "log level",
      "type": "string"
    },
    "ts": {
      "description": "timestamp in RFC3339Nano format; the trailing zeroes are removed from the seconds field",
      "type": "string"
    },
    "_msg": {
      "description": "core log message",
      "type": "string"
    },
    "module": {
      "description": "module name that emitted the log",
      "type": "string"
    },
    "impl": {
      "description": "some modules point out specific areas or tasks in log entries",
      "type": "string"
    },
    "msg": {
      "description": "some entries have more granular messages than just the core _msg",
      "type": "string"
    },
    "height": {
      "description": "some entries happen at a specific height",
      "type": "integer",
      "exclusiveMinimum": 0
    },
    "app_hash": {
      "description": "some entries happen at a specific app_hash",
      "type": "string"
    }
  },
  "required": [ "level", "ts", "_msg", "module" ]
}
```
> Note: The list of properties is not exhaustive. When implementing log parsing, check your logs and update the schema.

<!--- Todo: Probably we should create separate schemas for the different log levels or modules. --->

### genesis_file
Path to the JSON file containing the initial validator set and other meta-data.
```toml
genesis_file = "config/genesis.json"
```

| Value type          | string                                          |
|:--------------------|:------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME` |
|                     | absolute directory path                         |

The default relative path translates to `$CMTHOME/config/genesis.json`. In case `$CMTHOME` is unset, it defaults to
`$HOME/.cometbft/config/genesis.json`.

### priv_validator_key_file
Path to the JSON file containing the private key to use as a validator in the consensus protocol.
```toml
priv_validator_key_file = "config/priv_validator_key.json"
```

| Value type          | string                                          |
|:--------------------|:------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME` |
|                     | absolute directory path                         |

The default relative path translates to `$CMTHOME/config/priv_validator_key.json`. In case `$CMTHOME` is unset, it
defaults to `$HOME/.cometbft/config/priv_validator_key.json`.


### priv_validator_state_file
Path to the JSON file containing the last sign state of a validator.
```toml
priv_validator_state_file = "data/priv_validator_state.json"
```

| Value type          | string                                          |
|:--------------------|:------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME` |
|                     | absolute directory path                         |

The default relative path translates to `$CMTHOME/data/priv_validator_state.json`. In case `$CMTHOME` is unset, it
defaults to `$HOME/.cometbft/data/priv_validator_state.json`.

### priv_validator_laddr
TCP or UNIX socket listen address for CometBFT that allows external consensus signing processes to connect.
```toml
priv_validator_laddr = ""
```

| Value type          | string                                                |
|:--------------------|:------------------------------------------------------|
| **Possible values** | TCP Stream socket (`"tcp://127.0.0.1:26658"`)         |
|                     | Unix domain socket (`"unix:///var/run/privval.sock"`) |

When consensus signing is outsourced from CometBFT (typically to a Hardware Security Module, like a
[YubiHSM](https://www.yubico.com/product/yubihsm-2) device), this address is opened by CometBFT for incoming connections
from the signing service.

Make sure the port is available on the host machine and firewalls allow the signing service to connect to it.

More information on a supported signing service can be found in the [TMKMS](https://github.com/iqlusioninc/tmkms)
documentation.

### node_key_file
Path to the JSON file containing the private key to use for node authentication in the p2p protocol.
```toml
node_key_file = "config/node_key.json"
```

| Value type          | string                                          |
|:--------------------|:------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME` |
|                     | absolute directory path                         |

The default relative path translates to `$CMTHOME/config/node_key.json`. In case `$CMTHOME` is unset, it defaults to
`$HOME/.cometbft/config/node_key.json`.

### abci
The mechanism used to connect to the ABCI application.
````toml
abci = "socket"
````

| Value type          | string     |
|:--------------------|:-----------|
| **Possible values** | `"socket"` |
|                     | `"grpc"`   |

This mechanism is used when connecting to the ABCI application over the [proxy_app](#proxy_app) socket.

### filter_peers
When connecting to a new peer, filter the connection through an ABCI query to decide, if the connection should be kept.
```toml
filter_peers = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

When this setting is `true`, the ABCI application has to implement a query that returns `true` or `false` depending on
if the ABCI application wants to keep the connection or drop it.

To filter the peers by remote address, the query `/p2p/filter/addr/(peer address)` has to return `true` or `false`.
Remote addresses have to be fully qualified, for example: `tcp://1.2.3.4:26656`.

To filter the peers by [node ID](node_key.json.md), the query `/p2p/filter/id/(peer ID)` has to return `true` or `false`
.

## RPC Server
These configuration options change the behaviour of the built-in RPC server.

### rpc.laddr
TCP or UNIX socket address for the RPC server to listen on.
```toml
laddr = "tcp://127.0.0.1:26657"
```

| Value type          | string                                            |
|:--------------------|:--------------------------------------------------|
| **Possible values** | TCP Stream socket (`"tcp://127.0.0.1:26657"`)     |
|                     | Unix domain socket (`"unix:///var/run/rpc.sock"`) |

The RPC server endpoints have OpenAPI specification definitions through [Swagger UI](../../rpc).
<!---
NOTE: The OpenAPI reference (../../rpc) is injected into the documentation during
the CometBFT docs build process. See https://github.com/cometbft/cometbft-docs/
for details.
--->

You can also read the [RPC specification](../../spec/rpc) for more information.

### rpc.cors_allowed_origins
A list of origins a cross-domain request can be executed from.
```toml
cors_allowed_origins = []
```

| Value type          | array of string                            |                      |
|:--------------------|:-------------------------------------------|----------------------|
| **Possible values** | `[]`                                       | disable CORS support |
|                     | `["*"]`                                    | allow any origin     |
|                     | array of strings containing domain origins |                      |

Domain origins are fully qualified domain names with protocol prefixed, for example `"https://cometbft.com"` or
they can contain exactly one wildcard to extend to multiple subdomains, for example: `"https://*.myapis.com"`.

Example:

Allow only some subdomains for CORS requests:
```toml
cors_allowed_origins = ["https://www.cometbft.com", "https://*.apis.cometbft.com"]
```

### rpc.cors_allowed_methods
A list of methods the client is allowed to use with cross-domain requests.
```toml
cors_allowed_methods = ["HEAD", "GET", "POST", ]
```

| Value type                              | array of string |
|:----------------------------------------|:----------------|
| **Possible string values in the array** | `"HEAD"`        |
|                                         | `"GET"`         |
|                                         | `"POST"`        |

You can read more about the methods in the
[Mozilla CORS documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS).

### rpc.cors_allowed_headers
A list of headers the client is allowed to use with cross-domain requests.
```toml
cors_allowed_headers = ["Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time", ]
```

| Value type                              | array of string      |
|:----------------------------------------|:---------------------|
| **Possible string values in the array** | `"Accept"`           |
|                                         | `"Accept-Language"`  |
|                                         | `"Content-Language"` |
|                                         | `"Content-Type"`     |
|                                         | `"Range"`            |

The list of possible values are from the [Fetch spec](https://fetch.spec.whatwg.org/#cors-safelisted-request-header)
which defines `Origin` as a forbidden value. Read the
[Mozilla CORS documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) and do your own tests, if you want
to use this key.

<!--- Possibly, we should clarify the allowed values better. --->

### rpc.unsafe
Activate unsafe RPC endpoints.
```toml
unsafe = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

| Unsafe RPC endpoints    | Description                                                                           |
|:------------------------|---------------------------------------------------------------------------------------|
| `/dial_seeds`           | dials the given seeds (comma-separated id@IP:port)                                    |
| `/dial_peers`           | dials the given peers (comma-separated id@IP:port), optionally making them persistent |
| `/unsafe_flush_mempool` | removes all transactions from the mempool                                             |

Keep this `false` on production systems.

### rpc.max_open_connections
Maximum number of simultaneous open connections. This includes WebSocket connections.
```toml
max_open_connections = 900
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt; 0  |

If you want to accept a larger number of connections than the default 900, make sure that you increase the maximum
number of open connections in the operating system. Usually, the `ulimit` command can help with that.

This value can be estimated by the following calculation:
```
$(ulimit -Sn) - {p2p.max_num_inbound_peers} - {p2p.max_num_outbound_peers} - {number of WAL, DB and other open files}
```

Estimating the number of WAL, DB and other files at `50`, and using the default soft limit of Debian Linux (`1024`):
```
1024 - 40 - 10 - 50 = 924 (~900)
```

Note, that macOS has a default soft limit of `256`. Make sure you calculate this value for the operating system CometBFT
runs on.

### rpc.max_subscription_clients
Maximum number of unique clientIDs that can subscribe to events at the `/subscribe` RPC endpoint.
```toml
max_subscription_clients = 100
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

If you are using the `/broadcast_tx_commit` RPC endpoint, set this to the estimated maximum number of
`broadcast_tx_commit` calls per block.

### rpc.max_subscriptions_per_client
Maximum number of unique queries a given client can subscribe to at the `/subscribe` RPC endpoint.
```toml
max_subscriptions_per_client = 5
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

If you are using the `/broadcast_tx_commit` RPC endpoint, set this to the estimated maximum number of
`broadcast_tx_commit` calls per block.

### rpc.experimental_subscription_buffer_size
Experimental parameter to specify the maximum number of events a node will buffer, per subscription, before returning
an error and closing the subscription.
```toml
experimental_subscription_buffer_size = 200
```

| Value type          | integer   |
|:--------------------|:----------|
| **Possible values** | &gt;= 100 |

Higher values will accommodate higher event throughput rates (and will use more memory).

### rpc.experimental_websocket_write_buffer_size
Experimental parameter to specify the maximum number of RPC responses that can be buffered per WebSocket client.
```toml
experimental_websocket_write_buffer_size = 200
```

| Value type          | integer                                         |
|:--------------------|:------------------------------------------------|
| **Possible values** | &gt;= rpc.experimental_subscription_buffer_size |


If clients cannot read from the WebSocket endpoint fast enough, they will be disconnected, so increasing this parameter
may reduce the chances of them being disconnected (but will cause the node to use more memory).

If set lower than `rpc.experimental_subscription_buffer_size`, connections could be dropped unnecessarily. This value
should ideally be somewhat higher to accommodate non-subscription-related RPC responses.

### rpc.experimental_close_on_slow_client
Close the WebSocket client in case it cannot read events fast enough. Allows greater predictability in subscription
behaviour.
```toml
experimental_close_on_slow_client = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

The default behaviour for WebSocket clients is to silently drop events, if they cannot read them fast enough. This does
not cause an error and creates unpredictability.
Enabling this setting creates a predictable outcome by closing the WebSocket connection in case it cannot read events
fast enough.

### rpc.timeout_broadcast_tx_commit
Timeout waiting for a transaction to be committed when using the `/broadcast_tx_commit` RPC endpoint.
```toml
timeout_broadcast_tx_commit = "10s"
```

| Value type          | string (duration)          |
|:--------------------|:---------------------------|
| **Possible values** | &gt; `"0s"`; &lt;= `"10s"` |

Using a value larger than `"10s"` will result in increasing the global HTTP write timeout, which applies to all connections
and endpoints. There is an old developer discussion about this [here](https://github.com/tendermint/tendermint/issues/3435).

### rpc.max_body_bytes
Maximum size of request body, in bytes.
```toml
max_body_bytes = 1000000
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

### rpc.max_header_bytes
Maximum size of request header, in bytes.
```toml
max_header_bytes = 1048576
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

### rpc.tls_cert_file
TLS certificates file path for HTTPS server use.
```toml
tls_cert_file = ""
```

| Value type          | string                                                 |
|:--------------------|:-------------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME/config` |
|                     | absolute directory path                                |

The default relative path translates to `$CMTHOME/config`. In case `$CMTHOME` is unset, it defaults to
`$HOME/.cometbft/config`.

If the certificate is signed by a certificate authority, the certificate file should be the concatenation of the
server certificate, any intermediate certificates, and the Certificate Authority certificate.

The [rpc.tls_key_file](#rpctls_key_file) property also has to be set with the matching private key.

If this property is not set, the HTTP server launches.

### rpc.tls_key_file
TLS private key file path for HTTPS server use.
```toml
tls_key_file = ""
```

| Value type          | string                                                 |
|:--------------------|:-------------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME/config` |
|                     | absolute directory path                                |

The default relative path translates to `$CMTHOME/config`. In case `$CMTHOME` is unset, it defaults to
`$HOME/.cometbft/config`.

The [rpc.tls_cert_file](#rpctls_cert_file) property also has to be set with the matching server certificate.

If this property is not set, the HTTP server launches.

### rpc.pprof_laddr
Profiling data listen address and port. Without protocol prefix.
```toml
pprof_laddr = ""
```

| Value type          | string                       |
|:--------------------|:-----------------------------|
| **Possible values** | IP:port (`"127.0.0.1:6060"`) |
|                     | :port (`":6060"`)            |
|                     | `""`                         |

HTTP is always assumed as the protocol.

See the Golang [profiling](https://golang.org/pkg/net/http/pprof) documentation for more information.

## gRPC Server
These configuration options change the behaviour of the built-in gRPC server.

Each gRPC service can be turned on/off, and in some cases configured, individually.
If the gRPC server is not enabled, all individual services' configurations are ignored.

The gRPC server is exposed without any kind of security control or authentication. Do NOT expose this server
on the public Internet without appropriate precautions. Make sure it is secured, authenticated, load-balanced, etc.

### grpc.laddr
TCP or UNIX socket address for the gRPC server to listen on.
```toml
laddr = ""
```

| Value type          | string                                             |
|:--------------------|:---------------------------------------------------|
| **Possible values** | TCP Stream socket (`"tcp://127.0.0.1:26658"`)      |
|                     | Unix domain socket (`"unix:///var/run/abci.sock"`) |
|                     | `""`                                               |

If not specified, the gRPC server will be disabled.

### grpc.version_service.enabled
The gRPC version service provides version information about the node and the protocols it uses.
```toml
enabled = true
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `true`  |
|                     | `false` |

If [`grpc.laddr`](#grpcladdr) is empty, this setting is ignored and the service is not enabled.

### grpc.block_service.enabled
The gRPC block service returns block information.
```toml
enabled = true
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `true`  |
|                     | `false` |

If [`grpc.laddr`](#grpcladdr) is empty, this setting is ignored and the service is not enabled.

### grpc.block_results_service.enabled
The gRPC block results service returns block results for a given height. If no height is given, it will return the block
results from the latest height.
```toml
enabled = true
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `true`  |
|                     | `false` |

If [`grpc.laddr`](#grpcladdr) is empty, this setting is ignored and the service is not enabled.

### grpc.privileged.laddr
Configuration for privileged gRPC endpoints, which should **never** be exposed to the public internet.
```toml
laddr = ""
```

| Value type          | string                                             |
|:--------------------|:---------------------------------------------------|
| **Possible values** | TCP Stream socket (`"tcp://127.0.0.1:26658"`)      |
|                     | Unix domain socket (`"unix:///var/run/abci.sock"`) |
|                     | `""`                                               |

If not specified, the gRPC privileged endpoints will be disabled.

### grpc.privileged.pruning_service
Configuration specifically for the gRPC pruning service, which is considered a privileged service.
```toml
enabled = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

Only controls whether the pruning service is accessible via the gRPC API - not whether a previously set pruning service
retain height is honored by the node. See the [storage.pruning](#storagepruninginterval) section for control over pruning.

If [`grpc.laddr`](#grpcladdr) is empty, this setting is ignored and the service is not enabled.

## Peer-to-peer
These configuration options change the behaviour of the peer-to-peer protocol.

### p2p.laddr
TCP socket address for the P2P service to listen on.
```toml
laddr = "tcp://0.0.0.0:26656"
```

| Value type          | string                                            |
|:--------------------|:--------------------------------------------------|
| **Possible values** | TCP Stream socket (`"tcp://127.0.0.1:26657"`)     |

### p2p.external_address
TCP address to use as identification with peers. Useful when the node is running on a non-routable address or when the
node does not have the capabilities to figure out its IP address.
```toml
external_address = ""
```

| Value type          | string                      |
|:--------------------|:----------------------------|
| **Possible values** | IP:port (`"1.2.3.4:26656"`) |
|                     | `""`                        |

The port has to point to the node's P2P port.

Example with a node on a non-routable network:
- Node has local IP address `10.10.10.10` and uses port `10000` for P2P communication, set in [`p2p.laddr`](#p2pladdr).
- The network gateway has the public IP `1.2.3.4` and we want to use publicly open port `26656` on the IP address.
- A redirection has to be set up from `1.2.3.4:26656` to `10.10.10.10:1000`
- The `external_address` has to be set to `1.2.3.4:26656`.

### p2p.seeds
Comma-separated list of seed nodes that the node will try to connect.
```toml
seeds = ""
```

| Value type                        | string (comma-separated list)           |
|:----------------------------------|:----------------------------------------|
| **Possible values within commas** | nodeID@IP:port (`"abcd@1.2.3.4:26656"`) |
|                                   | `""`                                    |

Seed nodes will transmit a P2P address book for the node and then disconnect. No blocks or consensus communication takes
place.

Example:
```toml
seeds = "abcd@1.2.3.4:26656,deadbeef@5.6.7.8:10000"
```

### p2p.persistent_peers
Comma-separated list of nodes to keep persistent connections to.
```toml
persistent_peers = ""
```

| Value type                        | string (comma-separated list)           |
|:----------------------------------|:----------------------------------------|
| **Possible values within commas** | nodeID@IP:port (`"abcd@1.2.3.4:26656"`) |
|                                   | `""`                                    |

Persistent peers do not count toward [`p2p.max_num_inbound_peers`](#p2pmax_num_inbound_peers) or
[`p2p.max_num_outbound_peers`](#p2pmax_num_outbound_peers).

Example:
```toml
persistent_peers = "fedcba@11.22.33.44:26656,beefdead@55.66.77.88:20000"
```

### p2p.addr_book_file
Path to the address book file.
```toml
addr_book_file = "config/addrbook.json"
```

| Value type          | string                                                      |
|:--------------------|:------------------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME` |
|                     | absolute directory path                         |

The default relative path translates to `$CMTHOME/config/addrbook.json`. In case `$CMTHOME` is unset, it defaults to
`$HOME/.cometbft/config/addrbook.json`.

### p2p.addr_book_strict
Strict address routability rules disallow non-routable IP addresses in the address book. When `false`, private network
IP addresses are enabled to be stored in the address book.
```toml
addr_book_strict = true
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `true`  |
|                     | `false` |

Set it to `false` for testing on private network. Most production nodes can keep it at `true`.

### p2p.max_num_inbound_peers
Maximum number of inbound peers.
```toml
max_num_inbound_peers = 40
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

The [`p2p.max_num_inbound_peers`](#p2pmax_num_inbound_peers) and
[`p2p.max_num_outbound_peers`](#p2pmax_num_outbound_peers) values together define how many P2P connections the node will
maintain at maximum capacity, not including list of nodes in the [`p2p.persistent_peers`](#p2ppersistent_peers) setting.

The connections are bidirectional, so any connection can send or receive blocks and other data. The separation into
inbound and outbound setting only distinguishes the initial setup of the connection: outbound connections are initiated
by the node while inbound connections are initiated by a remote party.

Nodes on non-routable networks have to set their gateway to port-forward the P2P port for inbound connections to reach
the node. Outbound connections can be initiated as long as the node has generic Internet access. (Using NAT or methods.)

### p2p.max_num_outbound_peers
Maximum number of outbound peers.
```toml
max_num_outbound_peers = 10
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

The [`p2p.max_num_inbound_peers`](#p2pmax_num_inbound_peers) and
[`p2p.max_num_outbound_peers`](#p2pmax_num_outbound_peers) values together define how many P2P connections the node will
maintain at maximum capacity, not including list of nodes in the [`p2p.persistent_peers`](#p2ppersistent_peers) setting.

The connections are bidirectional, so any connection can send or receive blocks and other data. The separation into
inbound and outbound setting only distinguishes the initial setup of the connection: outbound connections are initiated
by the node while inbound connections are initiated by a remote party.

Nodes on non-routable networks have to set their gateway to port-forward the P2P port for inbound connections to reach
the node. Outbound connections can be initiated as long as the node has generic Internet access. (Using NAT or methods.)

### p2p.unconditional_peer_ids
List of node IDs, to which a connection will be (re)established ignoring any existing limits.
```toml
unconditional_peer_ids = ""
```

| Value type          | string (comma-separated)         |
|:--------------------|:---------------------------------|
| **Possible values** | comma-separated list of node IDs |
|                     | `""`                             |

If a peer listed in this property requests a connection, it will be accepted, even if the
[`p2p.max_num_inbound_peers`](#p2pmax_num_inbound_peers) or the
[`p2p.max_num_outbound_peers`](#p2pmax_num_outbound_peers) are reached.

Contrary to other settings, only the node ID has to be defined here, not the IP:port of the remote node.

### p2p.persistent_peers_max_dial_period
Maximum pause when redialing a persistent peer.
```toml
persistent_peers_max_dial_period = "0s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

If `"0s"` is set, then exponential backoff is applied to re-dial the remote node over and over.

<!--- Todo: add if there are any limitations to the exponential back-off. --->

### p2p.flush_throttle_timeout
Time to wait before flushing messages out on the connection.
```toml
flush_throttle_timeout = "100ms"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0ms"`     |

Write (flush) any buffered data to the connection. The flush is throttled, so if multiple triggers come in within the
duration, only one flush is executed.

Setting the value to `0ms` is possible, but the behaviour is undefined.

<!--- Todo: trace the code and update what happens if this is set to 0. --->

### p2p.max_packet_msg_payload_size
Maximum size of a message packet payload, in bytes.
```toml
max_packet_msg_payload_size = 1024
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt; 0  |

The value represents the maximum size of a message payload (or message data) in bytes.

### p2p.send_rate
Rate at which packets can be sent, in bytes/second.
```toml
send_rate = 5120000
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt; 0  |

The value represents the amount of packet bytes that can be sent per second.

### p2p.recv_rate
Rate at which packets can be received, in bytes/second.
```toml
recv_rate = 5120000
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt; 0  |

The value represents the amount of packet bytes that can be received per second.

### p2p.pex
Enable peer exchange reactor.

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `true`  |
|                     | `false` |

The peer exchange reactor is responsible for exchanging peer addresses among different nodes. If this is disabled, the
node can only connect to addresses preconfigured in [`p2p.seeds`](#p2pseeds) or
[`p2p.persistent_peers`](#p2ppersistent_peers).

### p2p.seed_mode
In seed mode, the node crawls the network and looks for peers. Any incoming connections will provide the gathered
addresses and then it disconnects without providing any other information.
```toml
seed_mode = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

In seed mode, the node becomes an online address book. Any incoming connections can receive the gathered addresses but
no other information (for example blocks or consensus data) is provided. The node simply disconnects after sending the
addresses.

The [`p2p.pex`](#p2ppex) option has to be set to `true` for the seed mode to work.

### p2p.private_peer_ids
Comma separated list of peer IDs to keep private (will not be gossiped to other peers)
```toml
private_peer_ids = ""
```

| Value type                        | string (comma-separated list)     |
|:----------------------------------|:----------------------------------|
| **Possible values within commas** | nodeID (`"abcdef0123456789abcd"`) |
|                                   | `""`                              |

The addresses with the listed node IDs will not be sent to other peers, when the peer exchange reactor
([`p2p.pex`](#p2ppex)) is enabled. This allows a more granular setting instead of completely disabling the peer exchange
reactor.

For example, sentry nodes in the
[Sentry Node Architecture](https://forum.cosmos.network/t/sentry-node-architecture-overview/454) on the Cosmos Hub can
use this setting to make sure they do not gossip the node ID of the validator node, while they can still accept node
addresses from the Internet.

### p2p.allow_duplicate_ip
Toggle to disable guard against peers connecting from the same ip.
```toml
allow_duplicate_ip = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

When this setting is set to `true`, multiple connections are allowed from the same IP address (for example on different
ports).

### p2p.handshake_timeout
Timeout duration for protocol handshake (or secret connection negotiation).
```toml
handshake_timeout = "20s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

This high-level timeout value is applied when the TCP connection has been made and the peers are negotiating an upgrade
to secret connection.

The value `"0s"` is undefined, and it can lead to unexpected behaviour.

### p2p.dial_timeout
Timeout duration for the low-level dialer that connects to the remote address on the TCP network.
```toml
dial_timeout = "3s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

This parameter is the timeout value for dialing on TCP networks. If a hostname is used instead of an IP address and the
hostname resolves to multiple IP addresses, the timeout is spread over each consecutive dial, such that each is given an
appropriate fraction of the time to connect.

Setting the value to `"0s"` disables timeout.

## Mempool
Mempool allows gathering and broadcasting uncommitted transactions among nodes.

The **mempool** is a storage for uncommitted transactions; the **mempool cache** is an internal storage within the
mempool, for invalid transactions. The mempool cache provides a list of known invalid transactions to filter out
incoming duplicate transactions without running a full validation on them.

### mempool.type
The type of mempool this node will use.
```toml
type = "flood"
```

| Value type          | string    |
|:--------------------|:----------|
| **Possible values** | `"flood"` |
|                     | `"nop"`   |

`"flood"` is the original mempool implemented for CometBFT. It is a concurrent linked list with flooding gossip
protocol.

`"nop"` is a "no operation" or disabled mempool, where the ABCI application is responsible storing, disseminating and
proposing transactions. Note, that it requires empty blocks to be created:
[`consensus.create_empty_blocks = true`](#consensuscreate_empty_blocks) has to be set.

### mempool.recheck
Transaction validity check.
```toml
recheck = true
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `true`  |
|                     | `false` |

Committing a block affects the application state, hence the remaining transactions in the mempool after a block commit
might become invalid. Setting `recheck = true` will go through the remaining transactions and remove invalid ones.

### mempool.broadcast
Broadcast the mempool content (uncommitted transactions) to other nodes.
```toml
broadcast = true
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `true`  |
|                     | `false` |

This ensures that uncommitted transactions have a chance to reach multiple validators and get committed by one of them.

Setting this to `false` will stop the mempool from relaying transactions to other peers until they are included in a
block. Only the peer you send the tx to will see it until it is included in a block.

### mempool.wal_dir
Mempool write-ahead log folder path.
```toml
wal_dir = ""
```

| Value type          | string                                          |
|:--------------------|:------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME` |
|                     | absolute directory path                         |
|                     | `""`                                            |

In case `$CMTHOME` is unset, it defaults to `$HOME/.cometbft`.

Configures the location of the Write Ahead Log (WAL) for the mempool. `""`  disables the WAL.

### mempool.size
Maximum number of transactions in the mempool.
```toml
size = 5000
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

If the mempool is full, incoming transactions are dropped.

The value `0` is undefined.

### mempool.max_txs_bytes
The maximum size of all transactions accepted in the mempool.
```toml
max_txs_bytes = 1073741824
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

This is the raw, total transaction size. Given 1MB transactions and a 5MB maximum transaction size, mempool will only
accept five transactions.

The default value is 1 Gibibytes (2^30 bytes).

### mempool.cache_size
Mempool internal cache size for invalid transactions.
```toml
cache_size = 10000
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

The mempool cache is an internal store for invalid transactions. Storing invalid transactions help in filtering incoming
transactions: we can compare incoming transactions to known invalid transactions and filter them out without going
through the process of validating the incoming transaction.

### mempool.keep-invalid-txs-in-cache
Invalid transactions might become valid in the future, hence they are regularly removed from the mempool cache of
invalid transactions. Turning this setting on will keep them in the list of invalid transactions forever.
# Do not remove invalid transactions from the cache (default: false)
# Set to true if it's not possible for any invalid transaction to become valid
# again in the future.
```toml
keep-invalid-txs-in-cache = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

Invalid transactions might become valid at a later time. The mempool cache clears invalid transactions regularly so the
transactions can be re-validated.

If this setting is set to `true`, the mempool cache will NOT remove the invalid transactions. It is useful in cases when
invalidated transactions can never become valid again.

This setting can be used by operators to lower the impact of some spam transactions: when a large number of duplicate
spam transactions are noted on the network, temporarily turning this setting to `true` will filter out the duplicates
quicker than validating each transaction one-by-one. It will also filter out transactions that are supposed to become
valid at a later date, so it is not advised to keep this `true` for a long time on regular networks as it can lead to
valid transactions failing.

### mempool.max_tx_bytes
Maximum size of a single transaction accepted into the mempool.
```toml
max_tx_bytes = 1048576
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

This is the maximum size of a transaction allowed to be transmitted over the network.

### mempool.experimental_max_gossip_connections_to_persistent_peers
> EXPERIMENTAL parameter!

Limit the number of persistent peer nodes that get mempool transaction broadcasts.
```toml
experimental_max_gossip_connections_to_persistent_peers = 0
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

When set to `0`, the mempool is broadcasting to all the nodes listed in the
[`p2p.persistent_peers`](#p2ppersistent_peers) list. If the number is above `0`, the number of nodes that get broadcasts
will be limited to this setting.

Unconditional peers and peers not listed in the [`p2p.persistent_peers`](#p2ppersistent_peers) list are not affected by
this parameter.

See
[`mempool.experimental_max_gossip_connections_to_non_persistent_peers`](#mempoolexperimental_max_gossip_connections_to_persistent_peers)
to limit mempool broadcasts that are not in the list of [`p2p.persistent_peers`](#p2ppersistent_peers).

### mempool.experimental_max_gossip_connections_to_non_persistent_peers
> EXPERIMENTAL parameter!

Limit the number of peer nodes that get mempool transaction broadcasts. This parameter does not limit nodes that are
in the [`p2p.persistent_peers`](#p2ppersistent_peers) list.
```toml
experimental_max_gossip_connections_to_non_persistent_peers = 0
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

When set to `0`, the mempool is broadcasting to all the nodes. If the number is above `0`, the number of nodes that get
broadcasts will be limited to this setting.

Unconditional peers and peers listed in the [`p2p.persistent_peers`](#p2ppersistent_peers) list are not affected by
this parameter.

See
[`mempool.experimental_max_gossip_connections_to_persistent_peers`](#mempoolexperimental_max_gossip_connections_to_persistent_peers)
to limit broadcasts to persistent peer nodes.

For non-persistent peers, if enabled, a value of 10 is recommended based on experimental performance results using the
default P2P configuration.

## State synchronization
State sync rapidly bootstraps a new node by discovering, fetching, and restoring a state machine snapshot from peers
instead of fetching and replaying historical blocks. It requires some peers in the network to take and serve state
machine snapshots. State sync is not attempted if the node has any local state (LastBlockHeight > 0).

The node will have a truncated block history, starting from the height of the snapshot.

### statesync.enable
Enable state synchronization.
```toml
enable = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

Enable state synchronization on first start.

### statesync.rpc_servers
Comma-separated list of RPC servers for light client verification of the synced state machine,
and retrieval of state data for node bootstrapping.
```toml
rpc_servers = ""
```

| Value type                        | string (comma-separated list)      |
|:----------------------------------|:-----------------------------------|
| **Possible values within commas** | nodeID@IP:port (`"1.2.3.4:26657"`) |
|                                   | `""`                               |

At least two RPC servers have to be defined for state synchronization to work.

### statesync.trust_height
The height of the trusted header hash.
```toml
trust_height = 0
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

`0` is only allowed when state synchronization is disabled.

### statesync.trust_hash
Header hash obtained from a trusted source.
```toml
trust_hash = ""
```

| Value type          | string             |
|:--------------------|:-------------------|
| **Possible values** | hex-encoded number |
|                     | ""                 |

`""` is only allowed when state synchronization is disabled.

This is the header hash value obtained from the trusted source at height
[statesync.trust_height](#statesynctrust_height).

### statesync.trust_period
The period during which validators can be trusted.
```toml
trust_period = "168h0m0s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

For Cosmos SDK-based chains, `statesync.trust_period` should usually be about 2/3rd of the unbonding period
(about 2 weeks) during which they can be financially punished (slashed) for misbehavior.

### statesync.discovery_time
Time to spend discovering snapshots before initiating a restore.
```toml
discovery_time = "15s"
```

<!--- What happens when this time expires? --->

### statesync.temp_dir
Temporary directory for state sync snapshot chunks.
```toml
temp_dir = ""
```

| Value type          | string                  |
|:--------------------|:------------------------|
| **Possible values** | undefined               |

This value is unused by CometBFT. It was not hooked up to the state sync reactor.

The codebase will always revert to `/tmp/<random_name>` for state snapshot chunks. Make sure you have enough space on
your drive that holds `/tmp`.

### statesync.chunk_request_timeout
The timeout duration before re-requesting a chunk, possibly from a different peer.
```toml
chunk_request_timeout = "10s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"5s"`      |

If a smaller duration is set when state syncing is enabled, an error message is raised.

### statesync.chunk_fetchers
The number of concurrent chunk fetchers to run.

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

`0` is only allowed when state synchronization is disabled.

## Block synchronization
Block synchronization configuration is limited to defining a version of block synchronization to use.

### blocksync.version
Block Sync version to use.
```toml
version = "v0"
```

| Value type          | string  |
|:--------------------|:--------|
| **Possible values** | `"v0"`  |

All other versions are deprecated.

## Consensus
Consensus parameters define how the consensus protocol should behave.

### consensus.wal_file
```toml
wal_file = "data/cs.wal/wal"
```

| Value type          | string                                          |
|:--------------------|:------------------------------------------------|
| **Possible values** | relative directory path, appended to `$CMTHOME` |
|                     | absolute directory path                         |

The default relative path translates to `$CMTHOME/data/cs.wal/wal`. In case `$CMTHOME` is unset, it defaults to
`$HOME/.cometbft/data/cs.wal/wal`.

### consensus.timeout_propose
How long we wait for a proposal block before prevoting nil.
```toml
timeout_propose = "3s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

### consensus.timeout_propose_delta
How much timeout_propose increases with each round.
```toml
timeout_propose_delta = "500ms"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0ms"`     |

### consensus.timeout_prevote
How long we wait after receiving +2/3 prevotes for “anything” (ie. not a single block or nil).
```toml
timeout_prevote = "1s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

### consensus.timeout_prevote_delta
How much the timeout_prevote increases with each round.
```toml
timeout_prevote_delta = "500ms"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0ms"`     |

### consensus.timeout_precommit
How long we wait after receiving +2/3 precommits for “anything” (ie. not a single block or nil).
```toml
timeout_precommit = "1s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

### consensus.timeout_precommit_delta
How much the timeout_precommit increases with each round.
```toml
timeout_precommit_delta = "500ms"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0ms"`     |

### consensus.timeout_commit
How long we wait after committing a block, before starting on the new height.
This gives us a chance to receive some more precommits, even though we already have +2/3.
```toml
timeout_commit = "1s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

### consensus.double_sign_check_height
How many blocks to look back to check existence of the node's consensus votes before joining consensus.
```toml
double_sign_check_height = 0
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

When non-zero, the node will panic upon restart if the same consensus key was used to sign {double_sign_check_height}
last blocks. So, validators should stop the state machine, wait for some blocks, and then restart the state machine to
avoid panic.

### consensus.skip_timeout_commit
Make progress as soon as we have all the precommits.
```toml
skip_timeout_commit = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

This is similar to setting [`consensus.timeout_commit`](#consensustimeout_commit) to `"0s"`.

### consensus.create_empty_blocks
If there are no transactions in the mempool, empty blocks are proposed to indicate that the chain is still running.
```toml
create_empty_blocks = true
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `true`  |
|                     | `false` |

This is more relevant to networks with a low number of transactions.

### consensus.create_empty_blocks_interval
If there are no transactions in the mempool, empty blocks are proposed to indicate that the chain is still running at
this interval.
```toml
create_empty_blocks_interval = "0s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

### consensus.peer_gossip_sleep_duration
Consensus reactor internal sleep duration when waiting for the next piece of calculation.
```toml
peer_gossip_sleep_duration = "100ms"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0ms"`     |

This is a generic sleep duration for the consensus reactor that allows other threads to run when the reactor has no work
to do.

### consensus.peer_gossip_intraloop_sleep_duration
Consensus reactor upper bound for a random sleep duration.
```toml
peer_gossip_intraloop_sleep_duration = "0s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`     |

Random sleeps are inserted in the consensus reactor when it is waiting for HasProposalBlockPart messages and HasVote
messages, so it can reduce the amount of these messages sent. This parameter sets an upper bound for the random value
that is used for these sleep commands.

### consensus.peer_query_maj23_sleep_duration
Consensus reactor `queryMaj23Routine` function sleep time.
```toml
peer_query_maj23_sleep_duration = "2s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

Sleep time for the `queryMaj23Routine` function.

## Storage
Storage parameters are important in production settings as it can make the difference between a 2GB data folder and a
20GB one.

Storage pruning sets a flag in the CometBFT database for data that has expired. The database only deletes this data from
the file system, when data compaction is called.

### storage.discard_abci_responses
Discard ABCI responses from the state store, which can save a considerable amount of disk space.
```toml
discard_abci_responses = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

Set to `false` to ensure ABCI responses are kept.

ABCI responses are required for the `/block_results` RPC queries, and to reindex events in the command-line tool.

### storage.pruning.interval
The time period between automated background pruning operations.
```toml
interval = "10s"
```

| Value type          | string (duration) |
|:--------------------|:------------------|
| **Possible values** | &gt;= `"0s"`      |

### storage.pruning.data_companion.enabled
Tell the automatic pruning function to respect values set by the data companion.

```toml
enabled = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

If disabled, only the application retain height will influence block pruning (but not block results pruning).

Only enabling this at a later stage will potentially mean that blocks below the application-set retain height at the
time will not be available to the data companion.

### storage.pruning.data_companion.initial_block_retain_height
The initial value for the data companion block retain height if the data companion has not yet explicitly set one.
If the data companion has already set a block retain height, this is ignored.
```toml
double_sign_check_height = 0
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

### storage.pruning.data_companion.initial_block_results_retain_height
The initial value for the data companion block results retain height if the data companion has not yet explicitly set
one. If the data companion has already set a block results retain height, this is ignored.
```toml
initial_block_results_retain_height = 0
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

### storage.pruning.data_companion.genesis_hash
Hash of the Genesis file, passed to CometBFT via the command line.
```toml
genesis_hash = ""
```

| Value type          | string             |
|:--------------------|:-------------------|
| **Possible values** | hex-encoded number |
|                     | `""`               |

If this hash mismatches the hash that CometBFT computes on the genesis file, the node is not able to boot.

## Transaction indexer
Transaction indexer settings.

The application will set which txs to index.
In some cases a node operator will be able to decide which txs to index based on configuration set in the application.

### tx_index.indexer
What indexer to use for transactions.
```toml
indexer = "kv"
```

| Value type          | string   |
|:--------------------|:---------|
| **Possible values** | `"kv"`   |
|                     | `"null"` |
|                     | `"psql"` |

`"null"` indexer indexes nothing.

`"kv"` is the simplest possible indexer, backed by a key-value storage.
The key-value storage database backend is defined in [`db_backend`](#db_backend).

`"psql"` indexer is backed by an external PostgreSQL server.
The server connection string is defined in [`tx_index.psql-conn`](#tx_indexpsql-conn).

The transaction height and transaction hash is always indexed, except with the `"null"` indexer.

### tx_index.psql-conn
The PostgreSQL connection configuration.
```toml
psql-conn = ""
```

| Value type          | string                                                       |
|:--------------------|:-------------------------------------------------------------|
| **Possible values** | `"postgresql://<user>:<password>@<host>:<port>/<db>?<opts>"` |
|                     | `""`                                                         |

## Prometheus Instrumentation
An extensive amount of Prometheus metrics are built into CometBFT.

### instrumentation.prometheus
Enable or disabled presenting the Prometheus metrics at an endpoint.
```toml
prometheus = false
```

| Value type          | boolean |
|:--------------------|:--------|
| **Possible values** | `false` |
|                     | `true`  |

When enabled, metrics are served under the `/metrics` endpoint on the
[instrumentation.prometheus_listen_addr](#instrumentationprometheus_listen_addr) address.

### instrumentation.prometheus_listen_addr
Address to listen for Prometheus collector(s) connections.
```toml
prometheus_listen_addr = ":26660"
```

| Value type          | string                                |
|:--------------------|:--------------------------------------|
| **Possible values** | Network address (`"127.0.0.1:26657"`) |

The metrics endpoint only supports HTTP.

### instrumentation.max_open_connections
Maximum number of simultaneous connections.
```toml
max_open_connections = 3
```

| Value type          | integer |
|:--------------------|:--------|
| **Possible values** | &gt;= 0 |

`0` allows unlimited connections.

### instrumentation.namespace
Instrumentation namespace
```toml
namespace = "cometbft"
```

| Value type          | string                    |
|:--------------------|:--------------------------|
| **Possible values** | Prometheus namespace name |

