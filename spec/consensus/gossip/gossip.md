# CONS/GOSSIP interactions

***This document is outdated, it belongs to a previous interaction of the specification work***

CONS interacts with GOSSIP to update the gossip state and to evaluate the current state to check for conditions that enable actions.
To update the state, CONS passes GOSSIP a tuple and to evaluate the conditions CONS queries the tuple space.
The exact mechanism by which tuples and query results are exchanged is implementation dependent[^options], but here we describe it as function calls.

[^options:] Some implementation options
>
> - check on conditions on a loop, starting from the highest known round of the highest known height and down the round numbers, sleeping on each iteration for some predefined amount of time;
> - set callbacks to inspect conditions on a (height,round) whenever a new message for such height and round is received;
> - provide GOSSIP with queries and predicates over the query results before hand, so GOSSIP will execute them according to its convenience and optimizing it, and with callbacks to be invoked when the predicates evaluate to true;
>
> All approaches should be equivalent and not impact the specification much, even if the corresponding implementations would be much different.







### Shared Vocabulary

CONS and GOSSIP share the type of tuples added/consulted to/from the tuple space.

```qnt reactor.gen.qnt
<<VOC-CONS-GOSSIP-TYPES>>
```

### Requires from GOSSIP

CONS is provided with functions to add and remove tuples from the space.[^removal]

[^removal]: removal of tuples has no equivalent in the Tendermint algorithm. **TODO** This is something to be added here.

```qnt reactor.gen.qnt
<<VOC-CONS-GOSSIP-ACTIONS>>
```

CONS is provided access to the local view.


```qnt reactor.gen.qnt
<<DEF-READ-TUPLE>>
```


As per the discussion in [Part I](#part-1-background), CONS requires GOSSIP to be a valid tuple space

```qnt reactor.gen.qnt
<<TS-VALIDTY>>
```

and to ensure Eventual $\Delta$-Timely Convergence** from GOSSIP

```qnt reactor.gen.qnt
<<REQ-CONS-GOSSIP-CONVERGENCE>>
```


### Provides to GOSSIP

> **TODO**





# Part III: GOSSIP requirements and provisions
GOSSIP, the Consensus Reactor Communication Layer, provides on its northbound interface the facilities for CONS to communicate with other nodes by adding and removing tuples and exposing the eventually converging tuple space.
On its southbound interface, GOSSIP relies on the P2P layer to implement the gossiping.

## Northbound Interaction - GOSSIP-I
Northbound interaction is performed through GOSSIP-I, whose vocabulary has been already [defined](#gossip-i-vocabulary).

Next we enumerate what is required and provided from the point of view of GOSSIP as a means to detect mismatches between CONS and GOSSIP.


### Requires from CONS
> **TODO**

### Provides to CONS
> **TODO**


## SouthBound Interaction

### P2P-I Vocabulary

Differently from the interaction between GOSSIP and CONS, in which GOSSIP understands CONS messages, P2P is oblivious to the contents of messages it transfers, which makes the P2P-I interface simple in terms of message types.

```qnt reactor.gen.qnt
<<VOC-GOSSIP-P2P-TYPES>>
```


P2P is free to establish connections to other nodes as long as it respect GOSSIP's restrictions, on the maximum number of connections to establish and on which nodes to not connect.

```qnt reactor.gen.qnt
<<VOC-CONS-GOSSIP-ACTIONS>>
```

GOSSIP needs to know to which other nodes it is connected.

```qnt reactor.gen.qnt
<<VOC-CONS-GOSSIP-ACTIONS>>
```

P2P must expose functionality to allow 1-1 communication with connected nodes.

```qnt reactor.gen.qnt
<<DEF-UNICAST>>
```

### Requires from P2P - P2P-I
Message to nodes that remain connected are reliably delivered.

```qnt reactor.gen.qnt
<<REQ-GOSSIP-P2P-UNICAST>>
```

The neighbor set of $p$ is never larger than `maxCon(p)`.
> TODO: can maxConn change in runtime?


```qnt reactor.gen.qnt
<<REQ-GOSSIP-P2P-CONCURRENT_CONN>>
```

Ignored processes should never belong to the neighbor set.

```qnt reactor.gen.qnt
<<REQ-GOSSIP-P2P-IGNORING>>
```




# Part IV: Closing

> :clipboard: **TODO** Anything else to add?









[1]: https://arxiv.org/abs/1807.0493 "The latest gossip on BFT consensus"
[2]: https://github.com/tendermint/tendermint/blob/master/docs/architecture/adr-052-tendermint-mode.md "ADR 052: Tendermint Mode"
