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
### tablename.property_key
This is the two-sentence summary of the parameter. It does all kinds of stuff.
```toml
tablename.property_key = "default value"
```

| Value type          | string/bool/int/etc                |
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
