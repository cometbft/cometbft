# Part 1: Background

> **Warning**
> We assume that you understand the Tendermint algorithm and therefore we will not review it here.
If this is not the case, please refer to [here](../).

Three kinds of messages are exchanged in the Tendermint algorithm: `PROPOSAL`, `PRE-VOTE`, and `PRE-COMMIT`.
The algorithm progresses when certain conditions are satisfied over the set of messages received.
For example, in order do decide on a value `v`, the set must include a `PROPOSAL` for `v` and more than two thirds of the number processes in `PRE-COMMIT` for the same `v`, during the same round.

To ensure progress and termination, the algorithm assumes a **Global Stabilization Time (GST)**, after which communication is reliable and timely (Eventual $\Delta$-Timely Communication), which is used to provide **Gossip Communication**.

| Eventual $\Delta$-Timely Communication|
|-----|
|There is a bound $\Delta$ and an instant GST (Global Stabilization Time) such that if a correct process $p$ sends a message $m$ at a time $t \geq \text{GST}$ to a correct process $q$, then $q$ will receive $m$ before $t + \Delta$.

|Gossip communication|
|-----|
| (i) If a correct process $p$ sends some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max} \{t,\text{GST}\} + \Delta$.
| (ii) If a correct process $p$ receives some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max}\{t,\text{GST}\} + \Delta$.

Roughly speaking, if Gossip Communication is guaranteed, then all messages sent by correct processes will be eventually delivered to all correct processes and allow all correct processes to realize that the same conditions are.

However, since processes are subject to failures, it is not guaranteed that proposals are ever sent and correct processes cannot wait indefinitely.
Hence, processes execute in rounds in which they wait for conditions to be met for sometime and, if they timeout, send negative messages that will lead to new rounds.
Again, Gossip Communication guarantees that eventually the conditions for deciding are met, even if only after GST is reached.

Implementing Gossip Communication, however, is hard because even messages sent before the GST need to be buffered to be reliably delivered between correct processes, but since GST may take indefinitely long to arrive, in practice implementing this property would require unbounded memory.

Fortunately, while Gossip Communication is a sufficient condition for the Tendermint algorithm to terminate, it is not strictly necessary, because the conditions to progress and terminate are evaluated over the messages of subsets of rounds executed, not all of them, and as new rounds are executed, messages in old rounds may be become obsolete and be ignored and forgotten, saving memory.
In other words, the algorithm does not require all messages to be delivered, only messages that advance the state of the processes.

## Node's state as a Tuple Space

One way of looking at the information used by CometBFT nodes is as a distributed tuple space to which all nodes contribute with signed entries;
to share a proposal, the proposer adds it to the tuple space and to vote for a proposal, the node adds the vote.
Each update includes the author name and therefore they are all non-conflicting with each other and update forgery is prevented with signing each update.

Since we are talking about an asynchronous distributed system, individual nodes can only maintain approximations of the tuple space.
Nodes may broadcast the update operations to all nodes, including themselves, and, if the communication is reliable, as guaranteed by Gossip Communication, the tuple space will eventually converge.

Otherwise, nodes may periodically compare their approximations with each other to identify and correct differences by adding missing entries, using some gossip/anti-entropy protocol.

Updates are commutative from the point of view of the tuple space, even if they are not commutative from the Tendermint algorithm's point of view; however, nodes observing different approximations of the tuple space may decide at different point in time but cannot violate any correctness guarantees.
And the eventual convergence of tuple space implies the eventual termination of the algorithm.

The tuple space could grow indefinitely, given that the number of heights and rounds is infinite.
Hence entries should be removed once they are no longer useful.[^deletion]
For example, if a new height is started, information from smaller heights may be discarded.

[^deletion]: Implementations should enforce that entries are only removed if they will not be needed again by CometBFT.

Even though, depending on the implementation, removed entries might get added again, this does not compromise CometBFT correctness properties.
Even so, for efficiency, entries that are removed should not be added again, which may be prevented by keeping "tombstones" for the entries removed, which must be smaller than the entries themselves.
We note the tombstone for an entry $e$ as $\bar{e}$.

With time, even small tombstones may accrue and need to be garbage collected, which may cause the entry being added again, which continues to be safe.


### Querying the Tuple Space

The tuple space contains information regarding steps taken by validators during possibly many rounds of possibly many heights.
Each entry has form $\lang Height, Round, Step, Validator, Value \rang$ and corresponds to the message Validator sent in Step of Round of Height; Value is a tuple of the message contents.

A query to the tuple space has the same form as the entries, with parts replaced by values, that must match the values in the entries, or by `*`, which matches any value.
For example, suppose the tuple space has the following entries, here organized as rows of a table for easier visualization:

| Height | Round | Step     | Validator | Value |
|--------|-------|----------|-----------|-------|
| H      | R     | Proposal | v         | pval  |
| H      | R     | PreVote  | v         | vval  |
| H      | R'    | PreCommit| v'        | cval  |
| H'     | R     | Proposal | v         | pval' |
| H'     | R''   | PreVote  | v'        | vval' |
| H'     | R'''  | PreCommit| v'        | cval' |

- Query $\lang H, R, Proposal, v, * \rang$ returns $\{ \lang H, R, Proposal, v, pval \rang \}$
- Query $\lang H, R, *, v, * \rang$ returns $\{ \lang H, R, Proposal, v, pval \rang,  \lang H, R, PreVote, v, vval \rang \}$.


### When to remove

> **TODO**: Expand on when to remove, based on the algorithm.


### Local views

The tuple space information is distributed among nodes that keep local views of the whole space.
Because of the asynchronous nature of the system, $t_p$ may not include entries in the space or may still include entries no longer in the space.
Formally, let $T$ be the tuple space and $t_p$ be node $p$'s view of $T$; $T = \cup_p t_p$.

- $e \in T \Leftrightarrow \exists p, e \in t_p \land \not\exists q, \bar{e}\in t_q$

Nodes can only query local views, not $T$.
Queries are subscripted with the node being queried.


### State Validity

Given that each validator can execute each step only once per round, a query that specifies height, round, step and validator must either return empty or a single tuple.

- $\forall h \in \N, r \in \N, s \in \text{Proposal, PreVote, PreCommit}, v \in \text{ValSet}_{h,r}$,  $\lang h, r, s, v, * \rang$ returns at most one value.

In the specific case of the Proposal step, only the proposer of the round can have a matching entry.

- $\lang H, R, Proposal, *, * \rang$ returns at most one value.

A violation of these rules is a proof of misbehavior.


### Convergence Requirements

We now formalize the requirements of the Tendermint algorithm in terms of an eventually consistent Tuple Space, which may be implemented using reliable communication (Gossip Communication), some best *best-effort* communication primitive that only delivers messages that are still useful, or some Gossip/Epidemic/Anti-Entropy approach for state convergence.
All of these can be made to progress after GST but should also progress during smaller stability periods.

|Eventual Convergence|
|-----|
| If there exists a correct process $p$ such that $e \in t_p$, then, eventually, for every correct process $q$, $e \in t_q$ or there exists a correct process $r$ such that $\bar{e} \in t_r$.|

> **Note**
> Nodes may learn of an entry deletion before learning of its addition.

In order to ensure convergence even in the presence of failures, the network must be connected in such a way to allow communication around any malicious nodes and to provide redundant paths between correct ones.
This may not be feasible at all times, but should happen at least during periods in which the system is stable.

In other words, during periods in which messages are timely delivered between correct processes long enough for multiple communication rounds to succeed, processes will identify and solve differences in their states.
We call "long enough" $\Delta$.

| Eventual $\Delta$-Timely Convergence |
|---|
| If $e\in t_p$, for some correct process $p$, at instant $t$, then by $\text{max} \{t,\text{GST}\} + \Delta$, either $e \in t_q$, for every correct process $q$ or $\bar{e} \in t_p$.

$\Delta$ encapsulates the assumption that, during stable periods, timeouts eventually do not expire precociously, given that they all can be adjusted to reasonable values, and the steps needed to converge on entries can be accomplished within $\Delta$.
Without precocious timeouts, all new entries should help algorithms progress.
In the Tendermint algorithm, for example, no votes for Nil are added and the round should end if a proposal was made.



## The tuple space as a CRDT

The two approaches described in the [previous section](#nodes-state-as-a-tuple-space), without the deletion of entries, correspond to operation-based and state-based [Grow-only SET](https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type#G-Set_(Grow-only_Set)) CRDT (G-Set).
This distributed data-structure is easily described as a set per process in which elements are added to include them G-Set; the sets kept by processes are approximations of the G-Set.

The [2 Phase-Set](https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type#2P-Set_(Two-Phase_Set)) (2P-Set) is a variation that allows removals.
It combines two sets, one in which elements are added to include them in the 2P-Set, $A$, and one in which elements are added to remove them from the 2P-Set, $D$; the actual membership of the 2P-Set is given by $A \setminus D$.




> Warning/TODO: A word about tombstones, that is $D$
> - Only state that is not required should be deleted/tombstone'd.
> - Instead of tombstones, add new entries that trigger removal of other entries (for example, state about a new height); each node must be given information to realize by itself that an entry is no longer needed.
> - Tombstones are an optimization, kept to prevent data recreation and redeletion.
> - Tombstones should be garbage collected at some point; imprecision shouldn't affect correctness/termination, since this is an optimization (as long as deleted state is never required again).
> - Tombstones are not to be gossiped; if they were, they would need to carry proof for the reason they were created, defeating their point.





# Part 2: CONS/GOSSIP interaction

CONS, the Consensus Reactor State Layer, is where the actions of the Tendermint algorithm are implemented.
Actions are executed once certain pre-conditions apply, such as timeout expirations or reception of information from validators in the system, neighbors or not.

An action may require communicating with applications and other reactors, for example to gather data to compose a proposal or to deliver decisions, and with the P2P layer, to communicate with other nodes.

## Northbound Interaction - ABCI

This specification focuses on the southbound interactions of CONS, with the GOSSIP and through GOSSIP with P2P.

For those interested in the interactions of CONS with applications and other reactors, we redirect the readers to the [Application Blockchain Interface (ABCI)](../../abci/) specification, which covers most of such communication.
ABCI specifies both what CONS [requires from the applications](../../abci/abci%2B%2B_app_requirements.md) and on what CONS [provides to Applications](../../abci/abci%2B%2B_tmint_expected_behavior.md).

Interactions with other reactors, such as with the Mempool reactor to build tentative proposals, will be covered elsewhere.



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



- [1]: https://arxiv.org/abs/1807.0493 "The latest gossip on BFT consensus"
- [2]: https://github.com/tendermint/tendermint/blob/master/docs/architecture/adr-052-tendermint-mode.md "ADR 052: Tendermint Mode"
