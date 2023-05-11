# CONS/GOSSIP interactions


## Southbound Interaction - GOSSIP-I
CONS interacts southbound only with GOSSIP, to update the gossip state and to evaluate the current state to check for conditions that enable actions.

To update the state, CONS passes a tuple GOSSIP a tuple with the exact content to be added to the tuple space through a functional call.
Implementation as free to do this through message calls, IPC or any other means.

To check for conditions, we assume that CONS constantly evaluates all conditions by directly accessing the GOSSIP state, as this keeps the specification simpler.
The exact mechanism of how conditions are evaluated is implementation specific, but some high level examples would be:
- check on conditions on a loop, starting from the hightest known round of the hightest known height and down the round numbers, sleeping on each iteration for some predefined amount of time;
- set callbacks to inspect conditions on a (height,round) whenever a new message for such height and round is received;
- provide GOSSIP with evaluation predicates that GOSSIP will execute according to its convenience and with callbacks to be invoked when the predicates evaluate to true.

All approaches should be equivalent and not impact the specification much, even if the corresponding implementations would be much different.[^setsorpred]

[^setsorpred]: **TODO**: should we not specify these shared variables and instead pass predicates to GOSSIP from consensus? Variables make it harder to separate the CONS from GOSSIP, as the the variables are shared, but is highly efficient. Predicates are cleaner, but harder to implement efficiently. For example, when a vote arrives, multiple predicates may have to be tested independently, while with variables the tests may collaborate with each other.

The state accessed by CONS is assumed to be valid.
However this is achieved is a concern of the GOSSIP and P2P layers. [^todo-validity]

[^todo-validity]: **TODO**: ensure that this requirement is mentioned in Gossip/P2P

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

> **Note**
> If you read previous versions of this draft, you will recall GOSSIP was aware of supersession. In this version, I am hiding supersession in REQ-CONS-GOSSIP-REMOVE and initially attributing the task of identifying superseded entries to CONS, which then removes what has been superseded. A a later refined version of this spec will clearly specify how supersession is handled and translated into removals.


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
