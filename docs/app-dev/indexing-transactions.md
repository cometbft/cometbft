---
order: 6
---

# Indexing Transactions

CometBFT allows you to index transactions and blocks and later query or
subscribe to their results. Transactions are indexed by `TxResult.Events` and
blocks are indexed by `Response(Begin|End)Block.Events`. However, transactions
are also indexed by a primary key which includes the transaction hash and maps
to and stores the corresponding `TxResult`. Blocks are indexed by a primary key
which includes the block height and maps to and stores the block height, i.e.
the block itself is never stored.

Each event contains a type and a list of attributes, which are key-value pairs
denoting something about what happened during the method's execution. For more
details on `Events`, see the
[ABCI](https://github.com/cometbft/cometbft/blob/v0.34.x/spec/abci/abci.md#events)
documentation.

An `Event` has a composite key associated with it. A `compositeKey` is
constructed by its type and key separated by a dot.

For example:

```json
"jack": [
  "account.number": 100
]
```

would be equal to the composite key of `jack.account.number`.

By default, CometBFT will index all transactions by their respective hashes
and height and blocks by their height.

CometBFT allows for different events within the same height to have 
equal attributes.

## Configuration

Operators can configure indexing via the `[tx_index]` section. The `indexer`
field takes a series of supported indexers. If `null` is included, indexing will
be turned off regardless of other values provided.

```toml
[tx-index]

# The backend database to back the indexer.
# If indexer is "null", no indexer service will be used.
#
# The application will set which txs to index. In some cases a node operator will be able
# to decide which txs to index based on configuration set in the application.
#
# Options:
#   1) "null"
#   2) "kv" (default) - the simplest possible indexer, backed by key-value storage (defaults to levelDB; see DBBackend).
#     - When "kv" is chosen "tx.height" and "tx.hash" will always be indexed.
#   3) "psql" - the indexer services backed by PostgreSQL.
# indexer = "kv"
```

### Supported Indexers

#### KV

The `kv` indexer type is an embedded key-value store supported by the main
underlying CometBFT database. Using the `kv` indexer type allows you to query
for block and transaction events directly against CometBFT's RPC. However, the
query syntax is limited and so this indexer type might be deprecated or removed
entirely in the future.

**Implementation and data layout**

The kv indexer stores each attribute of an event individually, by creating a composite key 
of the *event type*, *attribute key*, *attribute value*, *height* and *event sequence*. 

For example the following events:
 
```
Type: "transfer",
  Attributes: []abci.EventAttribute{
   {Key: []byte("sender"), Value: []byte("Bob"), Index: true},
   {Key: []byte("recipient"), Value: []byte("Alice"), Index: true},
   {Key: []byte("balance"), Value: []byte("100"), Index: true},
   {Key: []byte("note"), Value: []byte("nothing"), Index: true},
   },
 
```
 
```
Type: "transfer",
  Attributes: []abci.EventAttribute{
   {Key: []byte("sender"), Value: []byte("Tom"), Index: true},
   {Key: []byte("recipient"), Value: []byte("Alice"), Index: true},
   {Key: []byte("balance"), Value: []byte("200"), Index: true},
   {Key: []byte("note"), Value: []byte("nothing"), Index: true},
   },
```

will be represented as follows in the store: 

```
Key                                 value
transferSenderBobEndBlock1           1
transferRecipientAliceEndBlock11     1
transferBalance100EndBlock11         1
transferNodeNothingEndblock11        1
---- event2 ------
transferSenderTomEndBlock12          1
transferRecipientAliceEndBlock12     1
transferBalance200EndBlock12         1
transferNodeNothingEndblock12        1
 
```
The key is thus formed of the event type, the attribute key and value, the event the attribute belongs to (`EndBlock` or `BeginBlock`),
the height and the event number. The event number is a local variable kept by the indexer and incremented when a new event is processed. 

It is an `int64` variable and has no other semantics besides being used to associate attributes belonging to the same events within a height. 
This variable is not atomically incremented as event indexing is deterministic. **Should this ever change**, the event id generation
will be broken. 

#### PostgreSQL

The `psql` indexer type allows an operator to enable block and transaction event
indexing by proxying it to an external PostgreSQL instance allowing for the events
to be stored in relational models. Since the events are stored in a RDBMS, operators
can leverage SQL to perform a series of rich and complex queries that are not
supported by the `kv` indexer type. Since operators can leverage SQL directly,
searching is not enabled for the `psql` indexer type via CometBFT's RPC -- any
such query will fail.

Note, the SQL schema is stored in `state/indexer/sink/psql/schema.sql` and operators
must explicitly create the relations prior to starting CometBFT and enabling
the `psql` indexer type.

Example:

```shell
$ psql ... -f state/indexer/sink/psql/schema.sql
```

## Default Indexes

The CometBFT tx and block event indexer indexes a few select reserved events
by default.

### Transactions

The following indexes are indexed by default:

- `tx.height`
- `tx.hash`

### Blocks

The following indexes are indexed by default:

- `block.height`

## Adding Events

Applications are free to define which events to index. CometBFT does not
expose functionality to define which events to index and which to ignore. In
your application's `DeliverTx` method, add the `Events` field with pairs of
UTF-8 encoded strings (e.g. "transfer.sender": "Bob", "transfer.recipient":
"Alice", "transfer.balance": "100").

Example:

```go
func (app *KVStoreApplication) DeliverTx(req types.RequestDeliverTx) types.Result {
    //...
    events := []abci.Event{
        {
            Type: "transfer",
            Attributes: []abci.EventAttribute{
                {Key: []byte("sender"), Value: []byte("Bob"), Index: true},
                {Key: []byte("recipient"), Value: []byte("Alice"), Index: true},
                {Key: []byte("balance"), Value: []byte("100"), Index: true},
                {Key: []byte("note"), Value: []byte("nothing"), Index: true},
            },
        },
    }
    return types.ResponseDeliverTx{Code: code.CodeTypeOK, Events: events}
}
```

If the indexer is not `null`, the transaction will be indexed. Each event is
indexed using a composite key in the form of `{eventType}.{eventAttribute}={eventValue}`,
e.g. `transfer.sender=bob`.

## Querying Transactions Events

You can query for a paginated set of transaction by their events by calling the
`/tx_search` RPC endpoint:

```bash
curl "localhost:26657/tx_search?query=\"message.sender='cosmos1...'\"&prove=true"
```
If the conditions are related to transaction events and the user wants to make sure the
conditions are true within the same events, the `match_events` keyword should be used, 
as described [below](#querying_block_events)

Check out [API docs](https://docs.cometbft.com/v0.34/rpc/#/Info/tx_search)
for more information on query syntax and other options.

## Subscribing to Transactions

Clients can subscribe to transactions with the given tags via WebSocket by providing
a query to `/subscribe` RPC endpoint.

```json
{
  "jsonrpc": "2.0",
  "method": "subscribe",
  "id": "0",
  "params": {
    "query": "message.sender='cosmos1...'"
  }
}
```

Check out [API docs](https://docs.cometbft.com/v0.34/rpc/#subscribe) for more information
on query syntax and other options.

## Querying Block Events

You can query for a paginated set of blocks by their events by calling the
`/block_search` RPC endpoint:

```bash
curl "localhost:26657/block_search?query=\"block.height > 10 AND val_set.num_changed > 0\""
```

## `match_events` keyword 

The query results in the height number(s) (or transaction hashes when querying transactions) which contain events whose attributes match the query conditions. 
However, there are two options to query the indexers. To demonstrate the two modes, we reuse the two events
where Bob and Tom send money to Alice and query the block indexer. We issue the following query:

```bash
curl "localhost:26657/block_search?query=\"sender=Bob AND balance = 200\""
```

The result will return height 1 even though the attributes matching the conditions in the query 
occurred in different events. 

If we wish to retrieve only heights where the attributes occurred within the same event,
the query syntax is as follows:

```bash
curl "localhost:26657/block_search?query=\"sender=Bob AND balance = 200\"&match_events=true"
```
Currently the default behavior is if `match_events` is set  to false.

Check out [API docs](https://docs.cometbft.com/v0.34/rpc/#/Info/block_search)
for more information on query syntax and other options.

**Backwards compatibility**

Storing the event sequence was introduced in CometBFT 0.34.25. As there are no previous releases of CometBFT,
all nodes running CometBFT will include the event sequence. However, mixed networks running CometBFT v0.34.25 and greater
and Tendermint Core versions before v0.34.25 are possible. On nodes running Tendermint Core, the `match_events` keyword
is ignored and the data is retrieved as if `match_events=false`.

Additionally, if a node that was running Tendermint Core 
when the data was first indexed, and switched to CometBFT, is queried, it will retrieve this previously indexed
data as if `match_events=false` (attributes can match the query conditions across different events on the same height).
