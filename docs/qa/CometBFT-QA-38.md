---
order: 1
parent:
  title: CometBFT QA Results v0.38.x
  description: This is a report on the results obtained when running CometBFT v0.38.x on testnets
  order: 5
---

# CometBFT QA Results v0.38.x

This iteration of the QA was run on CometBFT `v0.38.0-alpha.2`, the second
`v0.38.x` version from the CometBFT repository.

The changes with respect to the baseline, `v0.37.0-alpha.3` from Feb 21, 2023,
include the introduction of the `FinalizeBlock` method to complete the full
range of ABCI++ functionality (ABCI 2.0), and other several improvements
described in the
[CHANGELOG](https://github.com/cometbft/cometbft/blob/v0.38.0-alpha.2/CHANGELOG.md).

## Testbed

As in other iterations of our QA process, we have used a 200-node network as
testbed, plus nodes to introduce load and collect metrics.

### Saturation point

As in previous iterations of our QA experiments, we first find the transaction
load on which the system begins to show a degraded performance. Then we run the
experiments with the system subjected to a load slightly under the saturation
point. 

The method to identify the saturation point is explained
[here](CometBFT-QA-34.md#saturation-point) and its application to the baseline
is described [here](TMCore-QA-37.md#finding-the-saturation-point). 

We chose `c=1,r=400` as the configuration for our experiments, where `c` is the
number of connections created by the load runner process to the target node, and
`r` is the rate or number of transactions issued per second. We could have
choosen `c=2,r=200`, which has a transaction load equal to `c=1,r=400`.

## Latencies

The following figure plots the latencies of the experiment carried out with the
configuration `c=1,r=400`.

![latency-1-400](img38/200nodes/e_de676ecf-038e-443f-a26a-27915f29e312.png).

For reference, the following figure shows the latencies of one of the
experiments in the baseline, where `c=2,r=200` is equivalent to the same
configuration used in this experiment. As can be seen, latencies are quite
similar, and even in some cases the baseline has slightly higher latencies.

![latency-2-200-37](img37/200nodes_cmt037/e_75cb89a8-f876-4698-82f3-8aaab0b361af.png)

## Prometheus Metrics on the Chosen Experiment

This section further examines key metrics for this experiment extracted from
Prometheus data regarding the chosen experiment with configuration `c=1,r=400`.

### Mempool Size

The mempool size, a count of the number of transactions in the mempool, was
shown to be stable and homogeneous at all full nodes. It did not exhibit any
unconstrained growth. The plot below shows the evolution over time of the
cumulative number of transactions inside all full nodes' mempools at a given
time.

![mempoool-cumulative](img38/200nodes/mempool_size.png)

The following picture shows the evolution of the average mempool size over all
full nodes, which mostly oscilates between 1000 and 2500 outstanding
transactions.

![mempool-avg](img38/200nodes/avg_mempool_size.png)

The peaks observed coincide with the moments when some nodes reached round 1 of
consensus (see below).

The behavior is similar to the observed in the baseline, presented next.

![mempool-cumulative-baseline](img37/200nodes_cmt037/mempool_size.png)

![mempool-avg-baseline](img37/200nodes_cmt037/avg_mempool_size.png)


### Peers

The number of peers was stable at all nodes. It was higher for the seed nodes
(around 140) than for the rest (between 20 and 70 for most nodes). The red
dashed line denotes the average value.

![peers](img38/200nodes/peers.png)

Just as in the baseline, shown next, the fact that non-seed nodes reach more
than 50 peers is due to [\#9548].

![peers](img37/200nodes_cmt037/peers.png)


### Consensus Rounds per Height

Most heights took just one round, that is, round 0, but some nodes needed to
advance to round 1.

![rounds](img38/200nodes/rounds.png)

The following specific run of the baseline required some nodes to reach round 1.

![rounds](img37/200nodes_cmt037/rounds.png)


### Blocks Produced per Minute, Transactions Processed per Minute

The following plot shows the rate in which blocks were created, from the point
of view of each node. That is, it shows when each node learned that a new block
had been agreed upon.

![heights](img38/200nodes/block_rate.png)

For most of the time when load was being applied to the system, most of the
nodes stayed around 20 blocks/minute.

The spike to more than 100 blocks/minute is due to a slow node catching up.

The baseline experienced a similar behavior.

![heights-baseline](img37/200nodes_cmt037/block_rate.png)

The collective spike on the right of the graph marks the end of the load
injection, when blocks become smaller (empty) and impose less strain on the
network. This behavior is reflected in the following graph, which shows the
number of transactions processed per minute.

![total-txs](img38/200nodes/total_txs_rate.png)

The following is the transaction processing rate of the baseline, which is
similar to above.

![total-txs-baseline](img37/200nodes_cmt037/total_txs_rate.png)


### Memory Resident Set Size

The following graph shows the Resident Set Size of all monitored processes, with
maximum memory usage of 1.6GB, slighlty lower than the baseline shown after.

![rss](img38/200nodes/memory.png)

A similar behavior was shown in the baseline, with even a slightly higher memory
usage.

![rss](img37/200nodes_cmt037/memory.png)

The memory of all processes went down as the load is removed, showing no signs
of unconstrained growth.


### CPU utilization

The best metric from Prometheus to gauge CPU utilization in a Unix machine is
`load1`, as it usually appears in the [output of
`top`](https://www.digitalocean.com/community/tutorials/load-average-in-linux).

The load is contained below 5 on most nodes, as seen in the following graph.

![load1](img38/200nodes/cpu.png)

The baseline had a similar behavior.

![load1-baseline](img37/200nodes_cmt037/cpu.png)


## Test Results

The comparison against the baseline results show that both scenarios had similar
numbers and are therefore equivalent.

A conclusion of these tests is shown in the following table, along with the
commit versions used in the experiments.

| Scenario | Date       | Version                                                    | Result |
| -------- | ---------- | ---------------------------------------------------------- | ------ |
| CometBFT | 2023-02-14 | v0.37.0-alpha.3 (bef9a830e7ea7da30fa48f2cc236b1f465cc5833) | Pass   |
| CometBFT | 2023-05-21 | v0.38.0-alpha.2 (1f524d12996204f8fd9d41aa5aca215f80f06f5e) | Pass   |


[\#9548]: https://github.com/tendermint/tendermint/issues/9548
