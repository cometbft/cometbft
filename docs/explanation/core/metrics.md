---
order: 5
---

# Metrics

CometBFT can report and serve the Prometheus metrics, which in their turn can
be consumed by Prometheus collector(s).

This functionality is disabled by default.

To enable the Prometheus metrics, set `instrumentation.prometheus=true` in your
config file. Metrics will be served under `/metrics` on 26660 port by default.
Listen address can be changed in the [config file](./configuration.md) (see
`instrumentation.prometheus_listen_addr`).

## List of available metrics

The following metrics are available:

| **Name**                                                | **Type**  | **Tags**           | **Description**                                                                                                                        |
| ------------------------------------------------------- | --------- | ------------------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| abci\_connection\_method\_timing\_seconds               | Histogram | method, type       | Timings for each of the ABCI methods                                                                                                   |
| blocksync\_syncing                                      | Gauge     |                    | Either 0 (not block syncing) or 1 (syncing)                                                                                            |
| consensus\_height                                       | Gauge     |                    | Height of the chain                                                                                                                    |
| consensus\_validators                                   | Gauge     |                    | Number of validators                                                                                                                   |
| consensus\_validators\_power                            | Gauge     | validator\_address | Total voting power of all validators                                                                                                   |
| consensus\_validator\_power                             | Gauge     | validator\_address | Voting power of the node if in the validator set                                                                                       |
| consensus\_validator\_last\_signed\_height              | Gauge     | validator\_address | Last height the node signed a block, if the node is a validator                                                                        |
| consensus\_validator\_missed\_blocks                    | Gauge     |                    | Total amount of blocks missed for the node, if the node is a validator                                                                 |
| consensus\_missing\_validators                          | Gauge     |                    | Number of validators who did not sign                                                                                                  |
| consensus\_missing\_validators\_power                   | Gauge     |                    | Total voting power of the missing validators                                                                                           |
| consensus\_byzantine\_validators                        | Gauge     |                    | Number of validators who tried to double sign                                                                                          |
| consensus\_byzantine\_validators\_power                 | Gauge     |                    | Total voting power of the byzantine validators                                                                                         |
| consensus\_block\_interval\_seconds                     | Histogram |                    | Time between this and last block (Block.Header.Time) in seconds                                                                        |
| consensus\_rounds                                       | Gauge     |                    | Number of rounds                                                                                                                       |
| consensus\_num\_txs                                     | Gauge     |                    | Number of transactions                                                                                                                 |
| consensus\_total\_txs                                   | Gauge     |                    | Total number of transactions committed                                                                                                 |
| consensus\_block\_parts                                 | Counter   | peer\_id           | Number of blockparts transmitted by peer                                                                                               |
| consensus\_latest\_block\_height                        | Gauge     |                    | /status sync\_info number                                                                                                              |
| consensus\_block\_size\_bytes                           | Gauge     |                    | Block size in bytes                                                                                                                    |
| consensus\_step\_duration\_seconds                      | Histogram | step               | Histogram of durations for each step in the consensus protocol                                                                         |
| consensus\_round\_duration\_seconds                     | Histogram |                    | Histogram of durations for all the rounds that have occurred since the process started                                                 |
| consensus\_block\_gossip\_parts\_received               | Counter   | matches\_current   | Number of block parts received by the node                                                                                             |
| consensus\_quorum\_prevote\_delay                       | Gauge     | proposer\_address  | Interval in seconds between the proposal timestamp and the timestamp of the earliest prevote that achieved a quorum                    |
| consensus\_full\_prevote\_delay                         | Gauge     | proposer\_address  | Interval in seconds between the proposal timestamp and the timestamp of the latest prevote in a round where all validators voted       |
| consensus\_vote\_extension\_receive\_count              | Counter   | status             | Number of vote extensions received                                                                                                     |
| consensus\_proposal\_receive\_count                     | Counter   | status             | Total number of proposals received by the node since process start                                                                     |
| consensus\_proposal\_create\_count                      | Counter   |                    | Total number of proposals created by the node since process start                                                                      |
| consensus\_round\_voting\_power\_percent                | Gauge     | vote\_type         | A value between 0 and 1.0 representing the percentage of the total voting power per vote type received within a round                  |
| consensus\_late\_votes                                  | Counter   | vote\_type         | Number of votes received by the node since process start that correspond to earlier heights and rounds than this node is currently in. |
| consensus\_duplicate\_vote                              | Counter   |                    | Number of times we received a duplicate vote.                                                                                          |
| consensus\_duplicate\_block\_part                       | Counter   |                    | Number of times we received a duplicate block part.                                                                                    |
| consensus\_proposal\_timestamp\_difference              | Histogram | is\_timely         | Difference between the timestamp in the proposal message and the local time of the validator at the time it received the message.      |
| p2p\_message\_send\_bytes\_total                        | Counter   | message\_type      | Number of bytes sent to all peers per message type                                                                                     |
| p2p\_message\_receive\_bytes\_total                     | Counter   | message\_type      | Number of bytes received from all peers per message type                                                                               |
| p2p\_peers                                              | Gauge     |                    | Number of peers node's connected to                                                                                                    |
| p2p\_peer\_pending\_send\_bytes                         | Gauge     | peer\_id           | Number of pending bytes to be sent to a given peer                                                                                     |
| p2p\_recv\_rate\_limiter\_delay                         | Counter   | peer\_id           | Time in seconds spent sleeping by the receive rate limiter, in seconds.                                                                |
| p2p\_send\_rate\_limiter\_delay                         | Counter   | peer\_id           | Time in seconds spent sleeping by the send rate limiter, in seconds.                                                                   |
| mempool\_lane\_size                                     | Counter   | lane               | Number of uncommitted transactions per lane                                                                                            |
| mempool\_lane\_bytes                                    | Counter   | lane               | Number of used bytes per lane                                                                                                          |
| mempool\_size                                           | Gauge     |                    | Number of uncommitted transactions in the mempool                                                                                      |
| mempool\_size\_bytes                                    | Gauge     |                    | Total size of the mempool in bytes                                                                                                     |
| mempool\_tx\_size\_bytes                                | Histogram |                    | Histogram of transaction sizes in bytes                                                                                                |
| mempool\_evicted\_txs                                   | Counter   |                    | Number of transactions that make it into the mempool and were later evicted for being invalid                                          |
| mempool\_failed\_txs                                    | Counter   |                    | Number of transactions that failed to make it into the mempool for being invalid                                                       |
| mempool\_rejected\_txs                                  | Counter   |                    | Number of transactions that failed to make it into the mempool due to resource limits                                                  |
| mempool\_recheck\_times                                 | Counter   |                    | Number of times transactions are rechecked in the mempool                                                                              |
| mempool\_already\_received\_txs                         | Counter   |                    | Number of times transactions were received more than once                                                                              |
| mempool\_active\_outbound\_connections                  | Gauge     |                    | Number of connections being actively used for gossiping transaction (experimental)                                                     |
| mempool\_recheck\_duration\_seconds                     | Gauge     |                    | Cumulative time spent rechecking transactions                                                                                          |
| state\_consensus\_param\_updates                        | Counter   |                    | Number of consensus parameter updates returned by the application since process start                                                  |
| state\_validator\_set\_updates                          | Counter   |                    | Number of validator set updates returned by the application since process start                                                        |
| state\_pruning\_service\_block\_retain\_height          | Gauge     |                    | Accepted block retain height set by the data companion                                                                                 |
| state\_pruning\_service\_block\_results\_retain\_height | Gauge     |                    | Accepted block results retain height set by the data companion                                                                         |
| state\_pruning\_service\_tx\_indexer\_retain\_height    | Gauge     |                    | Accepted transactions indices retain height set by the data companion                                                                  |
| state\_pruning\_service\_block\_indexer\_retain\_height | Gauge     |                    | Accepted blocks indices retain height set by the data companion                                                                        |
| state\_application\_block\_retain\_height               | Gauge     |                    | Accepted block retain height set by the application                                                                                    |
| state\_block\_store\_base\_height                       | Gauge     |                    | First height at which a block is available                                                                                             |
| state\_abciresults\_base\_height                        | Gauge     |                    | First height at which ABCI results are available                                                                                       |
| state\_tx\_indexer\_base\_height                        | Gauge     |                    | First height at which tx indices are available                                                                                         |
| state\_block\_indexer\_base\_height                     | Gauge     |                    | First height at which block indices are available                                                                                      |
| state\_store\_access\_duration\_seconds                 | Histogram | method             | Duration of accesses to the state store labeled by which method was called on the store                                                |
| state\_fire\_block\_events\_delay\_seconds              | Gauge     |                    | Duration of event firing related to a new block                                                                                        |
| statesync\_syncing                                      | Gauge     |                    | Either 0 (not state syncing) or 1 (syncing)                                                                                            |

## Useful queries

Percentage of missing + byzantine validators:

```md
((consensus\_byzantine\_validators\_power + consensus\_missing\_validators\_power) / consensus\_validators\_power) * 100
```
