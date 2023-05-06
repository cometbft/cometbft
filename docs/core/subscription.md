---
order: 7
---

# Subscribing to events via Websocket

CometBFT emits different events, which you can subscribe to via
[Websocket](https://en.wikipedia.org/wiki/WebSocket). This can be useful
for third-party applications (for analysis) or for inspecting state.

[List of events](https://godoc.org/github.com/cometbft/cometbft/types#pkg-constants)

To connect to a node via websocket from the CLI, you can use a tool such as
[wscat](https://github.com/websockets/wscat) and run:

```sh
wscat -c ws://127.0.0.1:26657/websocket
```

NOTE: If your node's RPC endpoint is TLS-enabled, utilize the scheme `wss` instead of `ws`.

You can subscribe to any of the events above by calling the `subscribe` RPC
method via Websocket along with a valid query.

```json
{
    "jsonrpc": "2.0",
    "method": "subscribe",
    "id": 0,
    "params": {
        "query": "tm.event='NewBlock'"
    }
}
```

Check out [API docs](https://docs.cometbft.com/v0.34/rpc/) for
more information on query syntax and other options.

You can also use tags, given you had included them into DeliverTx
response, to query transaction results. See [Indexing
transactions](./indexing-transactions.md) for details.

## ValidatorSetUpdates

When validator set changes, ValidatorSetUpdates event is published. The
event carries a list of pubkey/power pairs. The list is the same
CometBFT receives from ABCI application (see [EndBlock
section](https://github.com/cometbft/cometbft/blob/v0.34.x/spec/abci/abci++_methods.md#endblock) in
the ABCI spec).

Response:

```json
{
    "jsonrpc": "2.0",
    "id": 0,
    "result": {
        "query": "tm.event='ValidatorSetUpdates'",
        "data": {
            "type": "tendermint/event/ValidatorSetUpdates",
            "value": {
              "validator_updates": [
                {
                  "address": "09EAD022FD25DE3A02E64B0FE9610B1417183EE4",
                  "pub_key": {
                    "type": "tendermint/PubKeyEd25519",
                    "value": "ww0z4WaZ0Xg+YI10w43wTWbBmM3dpVza4mmSQYsd0ck="
                  },
                  "voting_power": "10",
                  "proposer_priority": "0"
                }
              ]
            }
        }
    }
}
```
