## Overview


As of Q3 2023, the CometBFT team has dedicated significant resources addressing a number of storage related concerns:
1. Pruning not working: operators noticed that even when nodes prune data, the storage footprint is increasing.
2. Enabling pruning slows down nodes. Many chains disable pruning due to the impact on block processing time. 
3. CometBFT is addressing application level concerns such as transaction indexing. Furthermore, operators have very coarse grained control over what is stored on their node. 
4. Comet supports many database backends when ideally we should converge towards one. This requires understanding of the DB features but also the way CometBFT uses the database. We continued work starte on Tendermint 0.36, experimenting with a different database key layout that orders the DB data in a more friendly and efficient manner. 


By the end of Q3 we have addressed and documented the second problem by introducing a data companion API. The API allows node operators to extract data out of full nodes or validators, index them in whichever way they find suitable and instruct CometBFT to prune data at a much finer granularity:
- Blocks
- State
- ABCI Results 
- The transaction indexer
- The block indexer

For comparison, until then, CometBFT would only prune the block and state store (not including ABCI results), based on instructions from the application.

More details on the API itself and how it can be used can be found in the corresponding ADR and documentation (TODO LINK).

This report covers that changes and their impacts related to fixing and improving the pruning related points (1 and 3) as well as the last point. The results are obtained using `goleveldb` as the default backend unless stated otherwise. 

### Testing setup

We validated our results in a number of different settings:
 1. We can call this setup **Local-1node**:  Local runs on one node using a light kvstore application with almost no app state. This setup increases the changes that storage is the bottleneck and enabled us to evaluate the changes independent 
 of the demands of specific applications. Furthermore, we were able to create a larger footprint quicker
 thus speeding up the experimentation process. 

 To evaluate the impact of the compaction and the different key layouts on both compaction/pruning and performance, we run the following set of experiments on this setup:
 - Current layout - no pruning
 - Current layout - pruning, no forced compaction
 - Current layout - pruning and forced compaction
 - New layout - no pruning
 - New layout - pruning, no forced compaction
 - New layout - pruning and forced compaction

 We have also experimented with a third option, from which we initially expected the most: The new layout combined with insights into the access pattern of Comet to order together keys frequently accessed. In all
 our experiments Comet running this layout was less efficient than the other two and we therefore dismissed it. (TODO link PR). 

 We reduced the `timeout_commit` in this setup to 300ms to speed up execution

 Each experiment was repeated 3 times to make sure the results are deterministic. 

 2. **e2e-6 node**: CometBFT's e2e application run on a Digital Ocean cluster of 6 nodes. Each node had a different combination of changes we tested:
  - Pruning with and without compaction on the current database key layout vs. a new, better ordered key lauyout. 
  - No pruning using the current database key layout vs. a new key layout. 

  The nodes ran on top of a 10GB database to analize the effects of pruning but also potentially capture additional impact on performance depending on the key layout. 

3. **e2e - 200node**: This experiment runs the standard benchmark used by CometBFT to do QA. The only difference is that the nodes in the network have a mixed setup of nodes doing pruning with and without compaction on different database key layouts. The idea is to a) confirm having such a setup is not downgrading performance; and b) capture the same metrics as in the previous setups on a big network. 

3. **production-testing**: The validator team at Informal staking was kind enough to spend a lot of time with us trying to evaluate our changes on full nodes running on mainnet Injective. As their time was limited and we had reports that pruning, in addition to not working, slows down Injective nodes, we were interested to understand the impact our changes made on their network.  


#### **Metrics collected**

In addition to the storage footprint we collected information about the following system parameters:
- *RAM usage* 
- **Block processing time** (*cometbft_state_block_processing_time*)
- **Duration of individual consensus steps** (*cometbft_consensus_step_duration_seconds* aggregated by step)
- *consensus_total_txs*

During this work we extended CometBFT with two additional storage related metrics:
- *Block store access time* by each method accessing the block store
- *State store access time* by each method accessing the state store

### Pruning

Pruning the blockstore and statestore is a long supported feature by CometBFT. An application can set a retain height - the number of blocks which must be kept, and instruct CometBFT to prune the remaining blocks (taking into account some other constraints). 

#### **Storage footprint is not reduced**
Unfortunately, many users have noticed that, despite this feature being enabled, the growth of both the state and block store does not stop. This leads operators to copy the database, enforce compaction of the delete items manually and copy it back. We have talked to operators and some have to do this weekly or every two weeks. 

After some reasearch, we found that some of the database backends can be forced to compact the data. We experimented on it and confirmed those findings.

That is why we extended `cometbft-db`, adding an API to instruct the database to compact the files (TODO link PR). Then we made sure that Comet calls this function after blocks are pruned (TODO link PR). 

To evaluate whether this was really benefitial, we ran a couple of experiments to validate our findings:
- Local 1 node run of a dummy app that grows the DB to 20GB:

![local-run-compaction](img/impact_compaction_local.png)


- Running CometBFT's e2e application in a mixed network of 4 nodes. 
(TODO)

- Injective mainnet

![injective-no-compaction](img/injective_no_compaction.png "Injective -  pruning without compaction")

![injective-compaction](img/injective_compaction.png "Injective - pruning with compaciton")

#### *Pruning is slowing nodes down*

While the previous results confirm the storage footprint can be reduced, it is important that this is not impacting the performance of the entire system. 

The most impactful change we have made with regards to that is moving block and state pruning into a background process. Up until v1.x, pruning was done before a node moves on to the next height, blocking 
consensus from proceeding. In Q3 2023, we changed this by launching a pruning service that checks in fixed intervals, whether there are blocks to be pruned. This interval is configurable and is 10s by default. 

The impact of this changes is best demonstrated by reporting from Informal staking comparing 4 Injective nodes with the following setup:

1. *injective-sentry0* comet="v0.37" , pruning="default", keylayout=old
2. *injective-sentry1* comet="v0.37" , pruning="none" , keylayout=old
3. *injective-pruning* comet="modified" , pruning="600blocks" , keylayout=old
4. *injective-newlayout* comet="modified" , pruning="600blocks" , keylayout=new

We report the time to execute Commit along with the time to process blocks:
![injective-commit](img/injective_commit.png "Injective - commit")

![injective-bpt](img/injective_block_processing_time.png "Injective - average block processing time")


#### *Database key layout and pruning*

These results clearly show that pruning is not impacting the nodes performance anymore and could be turned on. However, while running the same set of experiments locally, we obtained contradicting results on the impact of the key layout on these numbers. 

Namely when running experiments in the **1-node-local** setup, we came to the conclusion that, if pruning is turned on, only the version of CometBFT using the new database key layout was not impacted by it. The throughput of CometBFT (meaured by num of txs processed within 1h), decreased with pruning (with and without compaction) usng the current layout - 500txs/s vs 700 txs/s with the new layout. The duration of the compaction operation itself was also much lower than with the old key layout. The block processing time difference is between 100 and 200ms which for some chains can be significant. 
The same was true for additional parameters such as RAM usage (200-300MB). 

We collected locally periodic heap usage samples via `pprof` and noticed that compaction for the old layout would take ~80MB of RAM vs ~30MB for with the new layout. 

When backporting these changes to the 0.37.x based branch we gave to Informal staking, we obtained similar results. However, this is not what they observed on mainnet.

(TODO add table showcasing this)


(TODO add results from e2e if they add anything here)

We have therefore decided to release v1.x with the support for both key layouts. They are not interchange-able, thus once one is used, a node cannot switch to the other. The version to be used is set in the `config.toml` and defaults to `v1` - the current layout. We will also release a migration script that offline converts the old layout to the new layout. The main reasons we have not addressed DB migration in more detail are:
- For nodes that do not do pruning, the new layout did not show great benefits
- When nodes prune, their DBs to migrate are presumably smaller. They could also statesync using `v2` as their desired key layout. 

The support for both layouts will allow users to benchmark their applications. If they determine that the new layout is boosting their performance, we can think of smarter DB migration scripts that will prevent nodes from being offline. 


## Pebble

`PebbleDB` was recently added to `cometbft-db` by Notional and based on their benchmarks it was superior to goleveldDB. 

We repeated our tests done in **1-node-local** using PebbleDB as the underlying database. While the difference in performance it self was slightly better, the most impressive difference is that PebbleDB seemed to handle compaction itself very well. 

In the graph below, we see the old layout without any compaction and the new layout with and without compaction on the same workload that generated 20GB of data when no pruning is active. 


![pebble](img/pebble.png "Pebble")

