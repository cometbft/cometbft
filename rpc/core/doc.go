/*
Package core defines the CometBFT RPC endpoints.

CometBFT ships with its own JSONRPC library -
https://github.com/cometbft/cometbft/tree/main/rpc/jsonrpc.

## Get the list

An HTTP Get request to the root RPC endpoint shows a list of available endpoints.

```bash
curl "http://localhost:26657"  | textutil -stdin -convert txt -stdout | sed 's/\/\/localhost:26657//g'
```

> Response:

```plain
Available endpoints:
/v1/abci_info
/v1/dump_consensus_state
/v1/genesis
/v1/health
/v1/net_info
/v1/num_unconfirmed_txs
/v1/status
/v1/unsafe_flush_mempool
/v1/unsubscribe_all?

Endpoints that require arguments:
/v1/abci_query?path=_&data=_&height=_&prove=_
/v1/block?height=_
/v1/block_by_hash?hash=_
/v1/block_results?height=_
/v1/block_search?query=_&page=_&per_page=_&order_by=_
/v1/blockchain?minHeight=_&maxHeight=_
/v1/broadcast_evidence?evidence=_
/v1/broadcast_tx_async?tx=_
/v1/broadcast_tx_commit?tx=_
/v1/broadcast_tx_sync?tx=_
/v1/check_tx?tx=_
/v1/commit?height=_
/v1/consensus_params?height=_
/v1/consensus_state?
/v1/genesis_chunked?chunk=_
/v1/header?height=_
/v1/header_by_hash?hash=_
/v1/subscribe?query=_
/v1/tx?hash=_&prove=_
/v1/tx_search?query=_&prove=_&page=_&per_page=_&order_by=_
/v1/unconfirmed_txs?limit=_
/v1/unsubscribe?query=_
/v1/validators?height=_&page=_&per_page=_
```
*/
package core
