# Tendermint Message sets as a CRDT

Here argue that the exchange of messages of the Tendermint algorithm can and should be seen as an eventually convergent Tuple implemented as a tuple space.

> :warning:
> We assume that you understand the Tendermint algorithm and therefore we will not review it here.
If this is not the case, please refer to the [consensus.md](../consensus.md) document.

Three kinds of messages are exchanged in the Tendermint algorithm: `PROPOSAL`, `PRE-VOTE`, and `PRE-COMMIT`.
The algorithm progresses when certain conditions are satisfied over the set of messages received.
For example, in order do decide on a value `v`, the set must include a `PROPOSAL` for `v` and `PRE-COMMIT` for the same `v` from more than two thirds of the validators for the same round.
Since processes are subject to failures, correct processes cannot wait indefinitely for messages since the sender may be faulty.
Hence, processes execute in rounds in which they wait for conditions to be met for some time but, if they timeout, send negative messages that will lead to new rounds.

## On the need for the Gossip Communication property

Progress and termination are only guaranteed in if there exists a **Global Stabilization Time (GST)** after which communication is reliable and timely (Eventual $\Delta$-Timely Communication).

| Eventual $\Delta$-Timely Communication|
|-----|
|There is a bound $\Delta$ and an instant GST (Global Stabilization Time) such that if a correct process $p$ sends a message $m$ at a time $t \geq \text{GST}$ to a correct process $q$, then $q$ will receive $m$ before $t + \Delta$.

Eventual $\Delta$-Timely Communication is used to provide the **Gossip Communication property**, which ensures that all messages sent by correct processes will be eventually delivered to all correct processes.

|Gossip Communication property|
|-----|
| (i) If a correct process $p$ sends some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max} (t,\text{GST}) + \Delta$.
| (ii) If a correct process $p$ receives some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max}(t,\text{GST}) + \Delta$.

This will, in turn, lead all correct processes to eventually be able to execute a round in which the conditions to decide are met, even if only after the GST is reached.

Even if Eventual $\Delta$-Timely Communication is assumed, implementing the Gossip Communication property would be unfeasible.
Given that all messages, even messages sent before the GST, need to be buffered to be reliably delivered between correct processes and that GST may take indefinitely long to arrive, implementing this primitive would require unbounded memory.

Fortunately, while the Gossip Communication property is a sufficient condition for the Tendermint algorithm to terminate, it is not strictly necessary:
i) the conditions to progress and terminate are evaluated over the messages of subsets of rounds executed, not all of them; ii) as new rounds are executed, messages in previous rounds may be become obsolete and be ignored and forgotten.
In other words, the algorithm does not require all messages to be delivered, only messages that advance the state of the processes.

## Node's state as a Tuple Space

One way of looking at the information used by CometBFT nodes is as a distributed tuple space (a set of tuples) to which all nodes contribute.
Entries are added by validators over possibly many rounds of possibly many heights.
Each entry has form $\lang h, r, s, v, p \rang$ and corresponds to the message validator node $v$ sent in step $s$ of round $r$ of height $h$; $p$ is a tuple with the message payload.
In the algorithm, whenever a message would be broadcast now a tuple is added to the tuple space.

Because of the asynchronous nature of distributed systems, what a node's view of what is in the tuple space, its **local view**, will differ from the other nodes.
There are essentially two ways of making converging the local views of nodes.

- **Approach One**: nodes broadcast all the updates they want to perform to all nodes, including themselves.
If using Reliable Broadcast/the Gossip Communication property, the tuple space will eventually converge to include all broadcast messages.
- **Approach Two**: nodes periodically compare their approximations with each other, 1-to-1, to identify and correct differences by adding missing entries, using some gossip/anti-entropy protocol.

These approaches work to reach convergence because the updates are commutative regarding the tuple space; each update simply adds an entry to a set.
From the Tendermint algorithm's point of view, convergence guarantees progress but is not a requirement for correctness.
In other words, nodes observing different approximations of the tuple space may decide at different points in time but cannot violate any correctness guarantees and the eventual convergence of tuple space implies the eventual termination of the algorithm.


### Tuple Removal and Garbage Collection

In both approaches for synchronization, the tuple space could grow indefinitely, given that the number of heights and rounds is infinite.
To save memory, entries should be removed from the tuple space as soon as they become stale, that is, they are no longer useful.
For example, if a new height is started, all entries corresponding to previous heights become stale.

In general, simply forgetting stale entries in the local view would save the most space.
However, if the second approach described [above](#nodes-state-as-a-tuple-space), it could lead to entries being added back and never being completely purged from the system.
Although stale entries do not affect the algorithm, or they would not be considered stale, not adding the entries back is important for performance and resource utilization sake.

One way to prevent re-adding entries is keeping _tombstones_ for the removed entries.
A tombstone is nothing but an that supersedes a specific other entry.
Let $\bar{e}$ be the tombstone for an entry $e$; if, during synchronization, a node is informed of $e$ but it already has $\bar{e}$, then it does not add $e$ to its local view.

However small tombstones may be (for example, they could contain just the hash of the entry it supersedes), with time they will accrue and need to be garbage collected, in which case the corresponding entry may be added again; again, this will not break correctness and as long as tombstones are kept for long enough, the risk of re-adding becomes minimal.

In the case of the Tendermint algorithm we note that staleness comes from adding newer entries (belonging to higher rounds and heights) to the tuple space.
If, as an optimization to Approach Two, these newer entries are exchanged first, then the stale entries can be excluded before being shared to other nodes that might have forgotten them and tombstones may not be needed at all.

While gossiping of tombstones themselves could be useful, it adds the risk of malicious nodes using them to disrupt the system.
This could be prevented by tombstones carrying the set of entries that led to the entry removal, but these messages need to be gossiped as well, allowing the counterpart in the gossip to do its own cleanup and without needing the bloated tombstone at all; hence the tombstones should be local only.

### Equivocation and forgery

The originator validator of each tuple signs it before adding it to the tuple space therefore making it unfeasible to forge entries from other validators unless they collude, which does not add to the power of attacks.
Equivocation attacks are not averted, but will be detected once two tuples differing only on the payload are added to the same local view and may be used as evidence of misbehavior.[^todo1]
The Tendermint algorithm itself tolerates equivocation attacks within certain bounds.

[^todo1]: 1) Do we need anything more in terms of making this data structure byzantine fault tolerant? If looking this structure independently, then a byzantine node could simply add more and entries. From TM point of though, those messages won't cause any harm. 2) How does TM prevent a byz node from "flooding" the network with nill votes today? it does not, so we are not making the problem worse. 3)If needed, look into [Making CRDTs Byzantine Fault Tolerant](https://martin.kleppmann.com/papers/bft-crdt-papoc22.pdf) for inspiration: "Many CRDTs, such as Logoot [44] and Treedoc [ 36], assign a unique identifier to each item (e.g. each element in a sequence); the data structure does not allow multiple items with the same ID, since then an ID would be ambiguous.” Can we use the payload itself as ID? However this work is focused on operation-based CRDT, not state-based.

### Querying the Tuple Space

The tuple space is consulted through queries, which have the same form as the entries.
Queries return all entries in their local views whose values match those in the query; `*` matches all values.
For example, suppose a node's local view of the tuple space has the following entries, here organized as rows of a table for easier visualization:

| Height | Round | Step      | Validator | Payload        |
| ------ | ----- | --------- | --------- | -------------- |
| 1      | 0     | Proposal  | v1        | pp1            |
| 1      | 0     | PreVote   | v1        | vp1            |
| 1      | 1     | PreCommit | v2        | cp1            |
| 2      | 0     | Proposal  | v1        | pp2            |
| 2      | 2     | PreVote   | v2        | vp2            |
| 2      | 3     | PreCommit | v2        | cp2   [^todo2] |
| 2      | 3     | PreCommit | v2        | cp2'  [^todo2] |

- Query $\lang 0, 0, Proposal, v1, * \rang$ returns $\{ \lang 0, 0, Proposal, v1, pp1 \rang \}$
- Query $\lang 0, 0, *, v1, * \rang$ returns $\{ \lang 0, 0, Proposal, v1, pp1 \rang,  \lang 0, 0, PreVote, v1, vp1 \rang \}$.

If needed for disambiguation, queries are subscripted with the node being queried.

[^todo2]: These tuples are evidence of an equivocation attack. It is not clear yet if we should keep both entries in the local view.

#### State Validity

Let $V_h \subseteq P$ be the set of validators of height $h$ and $\pi^h_r \in V_h$ be the proposer of round $r$ of height $h$.
When the context eliminates any ambiguity on the height number, we might write these values simply as $V$ and $\pi_r$.

Given that each validator can execute each step only once per round, a query that specifies height, round, step and validator SHOULD either return empty or a single tuple.

- $\forall h \in \N, \forall r \in \N, \forall s \in \{\text{Proposal, PreVote, PreCommit}\}, \forall v \in V_h$,  $\cup_{p \in P} \lang h, r, s, v, * \rang_p$ contains at most one element.

In the specific case of the Proposal step, only the proposer of the round can have a matching entry.

- $\forall h \in \N, \forall r \in \N$, $\cup_{p \in P} \lang h, r, \text{Proposal}, *, * \rang_p$ contains at most one element and it also matches $\cup_{p\in P} \lang h, r, \text{Proposal}, \pi^h_r, * \rang_p$

A violation of these rules is a misbehavior by the validator signing the offending entries.

### Eventual Convergence

Consider the following definition for **Eventual Convergence**.

|Eventual Convergence|
|-----|
| If there exists a correct process $p \in P$ such that $e \in t_p$, then, eventually, for every correct process $q \in P$, either $e \in t_q$ or $e$ is stale in $t_q$.

In order to ensure convergence even in the presence of failures, the network must be connected in such a way to allow communication around any malicious nodes, that is, to provide paths connecting correct nodes.
Even if paths connecting correct nodes exist, effectively using them requires timeouts to not expire precociously and abort communication attempts.
Timeout values can be guaranteed to eventually be enough for communication after a GST is reached, which implies that all communication between correct processes will eventually happen timely, which implies that the tuple space will converge and keep converging.
Formally, if there is a GST then following holds true:

| Eventual $\Delta$-Timely Convergence |
|---|
| If $e\in t_p$, for some correct process $p \in P$, at instant $t$, then by $\text{max}(t,\text{GST}) + \Delta$, for every correct process $q \in P$, either $e \in t_q$ or $e$ is stale in $t_p$.

Although GST may be too strong an expectation, in practice timely communication frequently happens within small stable periods, also leading to convergence.

### Why use a Tuple Space

Let's recall why we are considering using a tuple space to propagate Tendermint's messages.
It should be straightforward to see that Reliable Broadcast may be used to achieve Eventual Convergence and the Gossip Communication property may be used to implement Eventual $\Delta$-Timely Convergence:

- to add an entry to the tuple space, broadcast the entry;
- once delivered, add the entry to the local view.[^proof]

But if indeed we use the Gossip Communication property, then there are no obvious gains with respect to simply using broadcasts directly.

It should also be clear that if no entries are ever removed from the tuple space, then the inverse is also true:

- to broadcast a message, add it to the local view;
- once an entry is added to the local view, deliver it.

However, if entries can be removed, then the Tuple Space is actually weaker, since some entries may never be seen by some nodes, and should be easier to implement.
We argue later that it can be implemented using Anti-Entropy or Epidemic protocols/Gossiping (not equal to the Gossip Communication property).
We pointed out [previously](#on-the-need-for-the-gossip-communication-property) that the Gossip Communication property is overkill for Tendermint because it requires even stale messages to be delivered.
We remove corresponding to stale messages and never deliver them them.

[^proof]:  TODO: do we need to extend here?

### When to remove

> **TODO** Define conditions for tuple removal. Reference

## The tuple space as a CRDT

Conflict-free Replicated Data Types (CRDT) are distributed data structures that explore commutativity in update operations to achieve [Strong Eventual Consistency](https://en.wikipedia.org/wiki/Eventual_consistency#Strong_eventual_consistency).
As an example of CRDT, consider a counter updated by increment operations, known as Grown only counter (G-Counter): as long as the same set of operations are executed by two replicas, their views of the counter will be the same, irrespective of the execution order.

More relevant CRDT are the Grown only Set (G-Set), in which operations add elements to a set, and the [2-Phase Set](https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type#2P-Set_(Two-Phase_Set)) (2P-Set), which combines two G-Set to collect inclusions and exclusions to a set.

CRDT may be defined in two ways, operation- and state-based.
Operation-based CRDT use reliable communication to ensure that all updates (operations) are delivered to all replicas.
If the reliable communication primitive precludes duplication, then applying all operations will lead to the same state, irrespective of the delivery order since operations are commutative.
If duplications are allowed, then the operations must be made idempotent somehow.

State-based CRDT do not rely on reliable communication.
Instead they assumes that replicas will compare their states and converge two-by-two using a merge function;
as long as the function is commutative, associative and idempotent, the states will converge.
For example, in the G-Set case, the merge operator is simply the union of the sets.

The two approaches for converging the message sets in the Tendermint algorithm described [earlier](#nodes-state-as-a-tuple-space), without the deletion of entries, correspond to the operation- and state-based [Grow-only Set](https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type#G-Set_(Grow-only_Set)) CRDT;
if removals must be handled, then using a 2P-Set is an option.
To the best of our knowledge, no existing CRDT supports superseding of elements, until now.

### About supersession

Let

- `e1` and `e2` be `Entry`;
- `v1` and `v2` be a set of `Entry`, also referred to as a view;

We say that `e1` is superseded by `v1` if subset of `v1` makes `e1` stale.
We say that `v1` is superseded by `v2` all elements of `v1` are superseded by `v2`.

Supersession must respect the following properties:

1. Transitive: if `e1` is superseded by `v1` and `v1` is superseded by `v2`, then `e1` is superseded by `v2`;
1. Reflexive: `v1` is superseded by `v1`
1. Anti-symmetric: if `v1` is superseded by `v2` and `v2` is superseded by `v1` then `v1 == v2`
1. if `e1` is superseded by `v1` then `e1` is superseded by any sets obtained by replacing entries in `v1` by their corresponding tombstones;

> :warning:
> this should ensure no cycles.

> :warning: TODO
> Prove that `merge` operator is:

> - associative:
> - commutative
> - idempotent.

> :warning: TODO
> `view.exists(e => e.isStale(view)).implies(e.isStale(removeStale(view.addEntry(e)))`
> wrong usage of `exists` in the previous definition. `e` is not quantified in the `implies`



### Set with Superseding Elements - CRDT

We define here the Set with Superseding Elements CRDT (SSE) as a set in which the containing of some elements may render other elements stale or superseded.
Stale elements are irrelevant from the point of view of the set's users and therefore may be removed from the set.
Our definition of SSE is as a state-based CRDT, but an equivalent operations-based definition must exist.
The corresponding Quint definitions are in [sse.qnt](sse.qnt).

- `Entry`
    - a tuple or record;
    - application specific;

- `EntryOrTs`
    - a wrapper around an `Entry` to also represent tombstones;
    - a whapper that is not a tombstone is alive;

- `View`:
    - the information maintained by replicas;
    - a set of `EntryOrTs`
    - the empty `View` is the empty set.
    - TODO: what about a set with just tombstones?

- `isSupersededBy(ets: EntryOrTs, view: View): bool`:
    - returns true if the `Entry` is superseded in the `View`;
    - is application specific;

- `removeStale(view: View): (View,View)`
    - returns a new view without all the stale entries and a view with just the stale entries;


- `hasEntry(v: View, e:Entry):bool` returns `true` iff the view contains a live `EntryOrTs` for the entry;

- `addEntry(v:View, e: Entry): View`
    - constructs a new `View` with the entry in it, if the original does not have a tombstone preventing it;

- `merge(lhs: View, rhs: View): View`
    - combines two `View` into a new `View` that is a superset of the inputs
    - stale entries are removed;

- `delEntry(v:View, e: Entry): View`
    - constructs a new `View` without the `Entry` and with the corresponding tombstone;




## TODO

- Tombstones
    - implicit staleness - regular entries
    - explicit staleness - special tombstone entry
        - Gossiping
            - If tombstones are not to be gossiped, they do not need to be signed
            - If they are to be gossiped
                - either they need to carry proof for the reason they were created in the form of other entries, defeating the purpose of gossiping the tombstone
                - they need to be signed to allow detection of misbehavior
        - Tombstones removal may lead to entry revival in the local views
            - breaks the CRDT properties
                - still useful, for example as a mempool
                - Although the Tendermint algorithm shouldn't require tombstones at all, there may be cases in which their addition for messages already staled by others helps clean up the tuple space quickly, but their removal will not break correctness.




- @josef-widder We should try to understand what we need to do about Byzantine validators. In principle, we could say that the tuple space only contains entries for correct validators, and factor in Byzantine behavior by adapting semantics of the query belows.
- @josef-widder Here (when querying) we could say if p is faulty, query can give any result (allowing for two-faces perceptions of Byzantine faults).
- @lasarojc $T$ is the actual tuple space, the global view if you will, and combines the values of all local views.
No node can see it and I've defined it here expecting to be useful from a formalization point of view. I may be exactly the case for dealing with byzantine validators. That is, byzantine nodes can poison local views with tuples not with the global view (because the collide). But I will need to think more about it before committing to a solution.
