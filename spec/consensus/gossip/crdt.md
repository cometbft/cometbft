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

This will, in turn, lead all correct processes to eventually be able to execute a round in which the conditions to decide are met, even if only after GST is reached.

Even if Eventual $\Delta$-Timely Communication is assumed, implementing the Gossip Communication property would be unfeasible.
Given that all messages, even messages sent before the GST, need to be buffered to be reliably delivered between correct processes and that GST may take indefinitely long to arrive, implementing this primitive would require unbounded memory.

Fortunately, while the Gossip Communication property is a sufficient condition for the Tendermint algorithm to terminate, it is not strictly necessary:
i) the conditions to progress and terminate are evaluated over the messages of subsets of rounds executed, not all of them; ii) as new rounds are executed, messages in previous rounds may be become obsolete and be ignored and forgotten.
In other words, the algorithm does not require all messages to be delivered, only messages that advance the state of the processes.

## Node's state as a Tuple Space

One way of looking at the information used by CometBFT nodes is as a distributed tuple space (a set of tuples) to which all nodes contribute.
Entries are added by validators over possibly many rounds of possibly many heights.
Each entry has form $\lang h, r, s, v, p \rang$ and corresponds to the message validator node $v$ sent in step $s$ of round $r$ of height $h$; $p$ is a tuple with the message payload.
To propose a value, the proposer adds it to the tuple space, and to vote for a proposal, a validator adds the vote.

Each tuple is signed by the validator adding it to the tuple space.
Since each tuple includes the heigh, round, and step there is no room for forgery (and any attempt to add differing entries for the same heigh, round and step by the same validator will be seen as evidence of misbehavior).[^todo1]

Because nodes are are part of an asynchronous distributed system, individual nodes can only maintain approximations of the tuple space, to which they try to converge.
There are essentially two ways of making the tuple space converge.

- **Approach One**: nodes broadcast all the updates they want to perform to all nodes, including themselves.
If using Reliable Broadcast/the Gossip Communication property, the tuple space will eventually converge to include all broadcast messages.
- **Approach Two**: nodes periodically compare their approximations with each other, 1-to-1, to identify and correct differences by adding missing entries, using some gossip/anti-entropy protocol.

These approaches work to reach convergence because the updates are commutative regarding the tuple space; each update simply adds an entry to a set.
From the Tendermint algorithm's point of view, convergence guarantees progress but is not requirement for correctness.[^todo2]
In other words, nodes observing different approximations of the tuple space may decide at different point in time but cannot violate any correctness guarantees and the eventual convergence of tuple space implies the eventual termination of the algorithm.

[^todo1]: Formalize using [Making CRDTs Byzantine Fault Tolerant](https://martin.kleppmann.com/papers/bft-crdt-papoc22.pdf) as basis; "Many CRDTs, such as Logoot [44] and Treedoc [ 36], assign a unique identifier to each item (e.g. each element in a sequence); the data structure does not allow multiple items with the same ID, since then an ID would be ambiguous.”

[^todo2]: This should be trivial from the fact that the first approach is essentially Tendermint using the Gossip Communication property.

### Tuple Removal and Garbage Collection

In both approaches for synchronization, the tuple space could grow indefinitely, given that the number of heights and rounds is infinite.
To save memory, entries should be removed from the tuple space as soon as they become stale, that is, they are no longer useful.
For example, if a new height is started, all entries corresponding to previous heights become stale.

In general, simply forgetting stale entries in the local view would save the most space.
However, it could lead to entries being added back and never being completely purged from the system.
Although stale entries do not affect the algorithm, or they would not be considered stale, not adding the entries back for performance and resource utilization sake.

One approach to prevent re-adding entries may be prevented by keeping _tombstones_ for the removed entries.
With time, however, even small tombstones may accrue and need to be garbage collected, in which cause the corresponding entry may be added again; again, this will not break correctness and as long as tombstones are kept for long enough, the risk of re-adding is minimal.

In the case of the Tendermint algorithm we note that staleness comes from adding newer messages (belonging to higher rounds and heights) to the tuple space.
Hence, in Approach Two, if as an optimization these newer messages are exchanged first, then the stale messages can be excluded before being shared to other nodes that might have forgotten them and tombstones may not be needed at all.

#### Local and Global views

Let $T$ be the tuple space as seen by an external observer, also referred here as the **global view**, and $t_p$ be node $p$'s approximation of $T$, also referred as $p$'s **local view**.
Because of the asynchronous nature of distributed systems, local views may not include all entries in the global view or may still include entries no longer in the global view.
Formally, let $P$ be the set of validators and $\bar{e}$ be the tombstone for an entry $e$, if tombstones are used.

Let $P$ be the set of all processes in the system,

- $T = \cup_P t_p$
- $e \in T \Leftrightarrow \exists p \in P, e \in t_p \land \not\exists q \in P, \bar{e}\in t_q$


### Querying the Tuple Space

The tuple space is consulted through queries, which have the same form as the entries.
Queries return all entries whose values match those in the query, where a `*` matches all values.
Nodes can only query their own local views, not the global view $T$.
For example, suppose a node's local view of the tuple space has the following entries, here organized as rows of a table for easier visualization:

| Height | Round | Step      | Validator | Payload |
| ------ | ----- | --------- | --------- | ------- |
| 1      | 0     | Proposal  | v1        | pp      |
| 1      | 0     | PreVote   | v1        | vp      |
| 1      | 1     | PreCommit | v2        | cp      |
| 2      | 0     | Proposal  | v1        | pp'     |
| 2      | 2     | PreVote   | v2        | vp'     |
| 2      | 3     | PreCommit | v2        | cp'     |

- Query $\lang 0, 0, Proposal, v1, * \rang$ returns $\{ \lang 0, 0, Proposal, v1, pp \rang \}$
- Query $\lang 0, 0, *, v1, * \rang$ returns $\{ \lang 0, 0, Proposal, v1, pp \rang,  \lang 0, 0, PreVote, v1, vp \rang \}$.

If needed for disambiguation, queries are subscripted with the node being queried.

#### State Validity

Let $\text{ValSet}_h \subseteq P$ be the set of validators of height $h$ and $\text{Prop}_{h,r}$ be the proposer of round $r$ of height $h$.

Given that each validator can execute each step only once per round, a query that specifies height, round, step and validator must either return empty or a single tuple.

- $\forall h \in \N, r \in \N, s \in \{\text{Proposal, PreVote, PreCommit}\}, v \in \text{ValSet}_h$,  $\cup_{p \in P} \lang h, r, s, v, * \rang_p$ contains at most one element.

In the specific case of the Proposal step, only the proposer of the round can have a matching entry.

- $\forall h \in \N, r \in \N, \lang h, r, \cup_{p \in P} \text{Proposal}, *, * \rang_p$ contains at most one element and it also matches $\cup_{p\in P} \lang h, r, \text{Proposal}, \text{Prop}_{h,r}, * \rang_p$.

A violation of these rules is a misbehavior by the validator signing the offending entries.

### Eventual Convergence

Consider the following definition for **Eventual Convergence**.

|Eventual Convergence|
|-----|
| If there exists a correct process $p$ such that $e \in t_p$, then, eventually, for every correct process $q$, $e \in t_q$ or there exists a correct process $r$ such that $\bar{e} \in t_r$.|

> **Note**
> Nodes may learn of an entry deletion before learning of its addition.

In order to ensure convergence even in the presence of failures, the network must be connected in such a way to allow communication around any malicious nodes and to provide paths between correct ones.
This can be achieved if there is a GST, after which timeouts eventually do not expire precociously, given that they all can be adjusted to reasonable values, which implies that all communication will eventually happen timely, which implies that the tuple space will converge and keep converging.
Formally, if there is a GST then following holds true:

| Eventual $\Delta$-Timely Convergence |
|---|
| If $e\in t_p$, for some correct process $p$, at instant $t$, then by $\text{max}(t,\text{GST}) + \Delta$, either $e \in t_q$, for every correct process $q$ or $e$ is stale in \bar{e} \in t_p$.

Although GST may be too strong an expectation, in practice timely communication frequently happens within small stable periods, also leading to convergence.

> :warning:
> TODO: need to update to consider staleness and not only tombstones.

### Why use a Tuple Space

Let's recall why we are considering using a tuple space to propagate Tendermint's messages.
It should be straightforward to see that Reliable Broadcast may be used to achieve Eventual Convergence and the Gossip Communication property may be used to implement Eventual $\Delta$-Timely Convergence:
to add an entry to the tuple space, broadcast the entry;
once delivered, add the entry to the local view.[^proof]
If indeed we use the Gossip Communication property, then there are no obvious gains.

It should also be clear that if no entries are ever removed from the tuple space, then the inverse is also true:
to broadcast a message, add it to the local view;
once an entry is added to the local view, deliver it.

> :warning:
> TODO: The latter is probably not true.
> Assume a Byzantine process that equivocates. Lets  be a correct process that receives a version of the equivocating message, and  a correct process that receives the other version. From Gossip communication (ii), any correct process should deliver both versions. Eventual -Timely Convergence only allows us to deliver one of them.

However, if entries can be removed, then the Tuple Space is actually weaker, since some entries may never be seen by some nodes, and should be easier to implement.
We argue later that it can be implemented using Anti-Entropy or Epidemic protocols/Gossiping (not equal to the Gossip Communication property).
We pointed out [previously](#on-the-need-for-the-gossip-communication-property) that the Gossip Communication property is overkill for Tendermint because it requires even stale messages to be delivered.
Removing tuple is exactly how stale messages get removed.

[^proof]:  TODO: do we need to extend here?

### When to remove

> **TODO** Define conditions for tuple removal. Reference

