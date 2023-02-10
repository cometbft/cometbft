---
order: 1
parent:
  title: CometBFT Quality Assurance Results for v0.34.x
  description: This is a report on the results obtained when running v0.34.x on testnets
  order: 2
---

# v0.34.x - From Tendermint Core to CometBFT

This section reports on the QA process we followed before releasing the first `v0.34.x` version
from our CometBFT repository.

The changes with respect to the last version of `v0.34.x`
(namely `v0.34.26`, released from the Informal Systems' Tendermint Core fork)
are minimal, and focus on rebranding our fork of Tendermint Core to CometBFT at places
where there is no substantial risk of breaking compatibility
with earlier Tendermint Core versions of `v0.34.x`.

Indeed, CometBFT versions of `v0.34.x` (`v0.34.27` and subsequent) should fulfill
the following compatibility-related requirements.

* Operators can easily upgrade a `v0.34.x` version of Tendermint Core to CometBFT.
* Upgrades from Tendermint Core to CometBFT can be uncoordinated for versions of the `v0.34.x` branch.
* Nodes running CometBFT must be interoperable with those running Tendermint Core in the same chain,
  as long as all are running a `v0.34.x` version.

These QA tests focus on the third bullet, whereas the first two bullets are tested using our _e2e tests_.

It would be prohibitively time consuming to test mixed networks of all combinations of existing `v0.34.x`
versions, combined with the CometBFT release candidate under test.
Therefore our testing focuses on the last Tendermint Core version (`v0.34.26`) and the CometBFT release
candidate under test.

We only run the _200 node test_, and not the _rotating node test_. The effort of running the latter
is not justified given the amount and nature of the changes we are testing with respect to the
full QA cycle run previously on `v0.34.x`.
Since the changes to the system's logic are minimal, we are interested in these performance requirements:

* The CometBFT release candidate under test performs similarly to Tendermint Core (i.e., the baseline)
    * when used at scale (i.e., in a large network of CometBFT nodes)
    * when used at scale in a mixed network (i.e., some nodes are running CometBFT
      and others are running an older Tendermint Core version)

Therefore we carry out a complete run of the _200-node test_ on the following networks:

* A homogeneous 200-node testnet, where all nodes are running the CometBFT release candidate under test.
* A mixed network where 1/3 of the nodes are running the CometBFT release candidate under test,
  and the rest are running Tendermint Core `v0.34.26`.
* A mixed network where 2/3 of the nodes are running the CometBFT release candidate under test,
  and the rest are running Tendermint Core `v0.34.26`.

## Saturation Point

As the CometBFT release candidate under test has minimal changes
with respect to Tendermint Core `v0.34.26`, other than the rebranding changes,
we can confidently reuse the results from the `v0.34.x` baseline test regarding
the [saturation point](#finding-the-saturation-point).

Therefore, we will simply use a load of `r=200,c=2`
(see the explanation [here](#finding-the-saturation-point)).

## Examining latencies

In this section and the remaining, we provide the results of the _200 node test_.
Each section is divided into three parts,
reporting on the homogeneous network (all CometBFT nodes),
mixed network with 1/3 of Tendermint Core nodes,
and mixed network with 2/3 of Tendermint Core nodes.

On each of the three networks, the test consists of 4 experiments, with the goal
ensuring the data obtained is consistent across experiments.
On each of the networks, we pick only one representative run to present and discuss the results.

### CometBFT Homogeneous network

The figure below plots the four experiments carried out with this network.
We can see that the latencies follow a comparable pattern across experiments.

![latencies](./img/v034_200node_homog_latencies.png)

### 1/3 Tendermint Core - 2/3 CometBFT

![latencies](./img/cmt2tm1/latencies.png)

### 2/3 Tendermint Core - 1/3 CometBFT

![latencies_all_tm2_3_cmt1_3](img/v034_200node_tm2cmt1/latency_all.png)

![latencies_run1_tm2_3_cmt1_3](img/v034_200node_tm2cmt1/latency_run1.png){width=250}

## Prometheus Metrics

This section reports on the key prometheus metrics extracted from the experiments.

* For the CometBFT homogeneous network, we choose to present the
  experiment with UUID starting with `be8c` (see the latencies section above),
  as its latency data is representative,
  and   it contains the maximum latency of all runs (worst case scenario).
* For the mixed network with 1/3 of nodes running Tendermint Core `v0.34.26`
  and 2/3 running CometBFT.
  TODO
* For the mixed network with 2/3 of nodes running Tendermint Core `v0.34.26`
  and 1/3 running CometBFT.
  TODO

### Mempool Size

For reference, the plots below correspond to the baseline results.
The first shows the evolution over time of the cumulative number of transactions
inside all full nodes' mempools at a given time.

![mempool-cumulative](./img/v034_r200c2_mempool_size.png)

The second one shows evolution of the average over all full nodes, which oscillates between 1500 and 2000
outstanding transactions.

![mempool-avg](./img/v034_r200c2_mempool_size_avg.png)

#### CometBFT Homogeneous network

The mempool size was as stable at all full nodes as in the baseline.
These are the corresponding plots for the homogeneous network test.

![mempool-cumulative-homogeneous](./img/v034_homog_mempool_size.png)

![mempool-avg-homogeneous](./img/v034_homog_mempool_size_avg.png)

#### 1/3 Tendermint Core - 2/3 CometBFT

![mempool size](./img/cmt2tm1/mempool_size.png)

![average mempool size](./img/cmt2tm1/avg_mempool_size.png)

#### 2/3 Tendermint Core - 1/3 CometBFT

![mempool_tm2_3_cmt_1_3](./img/v034_200node_tm2cmt1/mempool_size.png)

![mempool-avg_tm2_3_cmt_1_3](./img/v034_200node_tm2cmt1/avg_mempool_size.png)

### Peers

The plot below corresponds to the baseline results, for reference.
It shows the stability of peers throughout the experiment.
Seed nodes typically have a higher number of peers.
The fact that non-seed nodes reach more than 50 peers is due to
[#9548](https://github.com/tendermint/tendermint/issues/9548).

![peers](./img/v034_r200c2_peers.png)

#### CometBFT Homogeneous network

The plot below shows the result for the homogeneous network.
It is very similar to the baseline. The only difference being that
the seed nodes seem to loose peers in the middle of the experiment.
However this cannot be attributed to the differences in the code,
which are mainly rebranding.

![peers-homogeneous](./img/v034_homog_peers.png)

#### 1/3 Tendermint Core - 2/3 CometBFT

![peers](./img/cmt2tm1/peers.png)

#### 2/3 Tendermint Core - 1/3 CometBFT

![peers-tm2_3_cmt1_3](./img/v034_200node_tm2cmt1/peers.png)

### Consensus Rounds per Height

TODO Move this under mempool as we refer to it


For reference, this is the baseline plot.

![rounds](./img/v034_r200c2_rounds.png)


#### CometBFT Homogeneous network

Most heights took just one round, some nodes needed to advance to round 1 at various moments,
and a few nodes even needed to advance to the third round at one point.
This coincides with the time at which we observed the biggest peak in mempool size
on the corresponding plot, shown above.

![rounds-homogeneous](./img/v034_homog_rounds.png)

#### 1/3 Tendermint Core - 2/3 CometBFT

![peers](./img/cmt2tm1/rounds.png)

#### 2/3 Tendermint Core - 1/3 CometBFT

![rounds-tm2_3_cmt1_3](./img/v034_200node_tm2cmt1/rounds.png)

### Blocks Produced per Minute, Transactions Processed per Minute

The following plot shows the rate of block creation, with a sliding window of 20 seconds,
throughout the experiment.

![heights](./img/v034_r200c2_heights_rate.png)

The next plot is the rate of transactions delivered, with a sliding window of 20 seconds,
throughout the experiment.

![total-txs](./img/v034_r200c2_total-txs_rate.png)

Both plots correspond to the baseline results.

#### CometBFT Homogeneous network

The plot for the homogeneous network shows the rate oscillates around 20 blocks/minute.

![heights-homogeneous-rate](./img/v034_homog_heights_rate.png)

The plot showing the transaction rate shows the rate stays around 20000 transactions per minute.

![txs-homogeneous-rate](./img/v034_homog_total-txs_rate.png)

#### 1/3 Tendermint Core - 2/3 CometBFT

Height rate
![heights](./img/cmt2tm1/block_rate.png)


Transaction rate
![transaction rate](./img/cmt2tm1/total_txs_rate.png)


#### 2/3 Tendermint Core - 1/3 CometBFT

![blocks_min_run1_tm2_3_cmt1_4](./img/v034_200node_tm2cmt1/block_rate.png)

In two minutes the height goes from 32 to 90 which gives an average of 29 blocks per minutes.

![tx_min_run1_tm2_3_cmt1_3](./img/v034_200node_tm2cmt1/total_txs_rate.png)

In 1 minutes and 30 seconds the system processes 35600 transactions which amounts to 23000 transactions per minute.

### Memory Resident Set Size

Reference plot for Resident Set Size (RSS) of all monitored processes.

![rss](./img/v034_r200c2_rss.png)

And this is the baseline average plot.

![rss-avg](./img/v034_r200c2_rss_avg.png)

#### CometBFT Homogeneous network

This is the plot for the homogeneous network, which slightly more stable than the baseline over
the time of the experiment.

![rss-homogeneous](./img/v034_homog_rss.png)

And this is the average plot. It oscillates around 560 MiB, which is noticeably lower than the baseline.

![rss-avg-homogeneous](./img/v034_homog_rss_avg.png)

#### 1/3 Tendermint Core - 2/3 CometBFT
![memory](./img/cmt2tm1/memory.png)

Here
![average memory](./img/cmt2tm1/avg_memory.png)


#### 2/3 Tendermint Core - 1/3 CometBFT

![rss_run1_tm2_3_cmt1_3](./img/v034_200node_tm2cmt1/memory.png)

![rss_avg_run1_tm2_3_cmt1_3](./img/v034_200node_tm2cmt1/avg_memory.png)

### CPU utilization

This is the baseline `load1` plot, for reference.

![load1](./img/v034_r200c2_load1.png)

#### CometBFT Homogeneous network

![load1-homogeneous](./img/v034_homog_load1.png)

Similarly to the baseline, it is contained in most cases below 5.

#### 1/3 Tendermint Core - 2/3 CometBFT

Total
![cpu](./img/cmt2tm1/cpu.png)

Average
![average cpu](./img/cmt2tm1/avg_cpu.png)

#### 2/3 Tendermint Core - 1/3 CometBFT

![cpu](./img/v034_200node_tm2cmt1/cpu.png)

Average
![average cpu](./img/v034_200node_tm2cmt1/avg_cpu.png)

## Test Results

| Scenario | Date | Version | Result |
|--|--|--|--|
|CometBFT Homogeneous network | 2023-02-08 | 3b783434f26b0e87994e6a77c5411927aad9ce3f | Pass
|1/3 Tendermint Core <br> 2/3 CometBFT | 2023-02-08 | CometBFT: 3b783434f26b0e87994e6a77c5411927aad9ce3f <br>Tendermint Core: 66c2cb63416e66bff08e11f9088e21a0ed142790 | Pass|
|2/3 Tendermint Core <br> 1/3 CometBFT | 2023-02-08 | CometBFT: 3b783434f26b0e87994e6a77c5411927aad9ce3f <br>Tendermint Core: 66c2cb63416e66bff08e11f9088e21a0ed142790  | Pass |
