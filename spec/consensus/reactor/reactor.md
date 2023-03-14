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

## Node's state as a CRDT

One way of looking at the information used by CometBFT nodes is as a distributed tuple space to which all nodes contribute;
to share a proposal, the proposer adds it to the tuple space and to vote for a proposal, the node adds the vote.
Each update is non-conflicting with any other updates, that is, no two nodes try to make the same update, by virtue of signing each update.

Since we are talking about an asynchronous distributed system, individual nodes can only maintain approximations of the tuple space.
Nodes may broadcast the update operations operations to all nodes, including themselves, and, if the communication is reliable, as guaranteed by Gossip Communication, the tuple space will eventually converge.

Otherwise, nodes may periodically compare their approximations with each other to identify and correct differences by adding missing entries, using some gossip/anti-entropy protocol.
In this approach, tuples with superseded information may be removed from the tuple space.

The two approaches just described correspond to an operation-based and a state-based **2 Phase-Set** CRDT, or 2P-Set.
This distributed data-structure is easily described as a combination of two sets, one in which elements are added to include them in the 2P-Set, $A$, and one in which elements are added to remove them from the 2P-Set, $R$; the actual membership of the 2P-Set is given by $A \setminus B$.

Updates are commutative from the point of view of the tuple space, even if they are not commutative from the Tendermint algorithm's point of view; however, nodes observing different approximations of the tuple space may decide at different point in time but cannot violate any correctness guarantees.
And the eventual convergence of the 2P-Set implies the eventual termination of the algorithm.

> Warning/TODO: a word about tombstones, that is, the $B$ set.    
> Tombstones are not gossiped; each node must be given information to realize by itself that an entry is no longer needed.
> Tombstones, if at all materialized, must be garbage collected.

## The Condition State

The condition state consists in a tuple space with information regarding steps taken by validators during possibly many rounds of possibly many heights. Each entry has form $\lang Height, Round, Step, Validator, Value \rang$ and corresponds to the message Validator sent in Step of Round of Height; Value is a tuple of the message contents.

A query to the tuple space has the same form of the entries, with parts replaced by values or by `*`, meaning any value is allowed.
For example, suppose the tuple space has the following values, here organized as a table for easier visualization:

| Height | Round | Step     | Validator | Value |
|--------|-------|----------|-----------|-------|
| H      | R     | Proposal | v         | pval  |
| H      | R     | PreVote  | v         | vval  |
| H      | R'    | PreCommit| v'        | cval  | 
| H'     | R     | Proposal | v         | pval' |
| H'     | R''   | PreVote  | v'        | vval' |
| H'     | R'''  | PreCommit| v'        | cval' | 

Query $\lang H, R, Proposal, v, * \rang$ returns $\{ \lang H, R, Proposal, v, pval \rang \}$ and query
 $\lang H, R, *, v, * \rang$ returns $\{ \lang H, R, Proposal, v, pval \rang,  \lang H, R, PreVote, v, vval \rang \}$.

### State Validity
 
Given that each validator can set this state at most once per round, a query that specifies height, round, step and validator must return empty, if the state has not been set, or a single tuple. 

- $\forall h \in \N, r \in \N, s \in \text{Proposal, PreVote, PreCommit}, v \in \text{ValSet}_{h,r}$,  $\lang h, r, s, v, * \rang$ returns at most one value.

In the specific case of the Proposal step, only the proposer of the round can have a matching entry. 

- $\lang H, R, Proposal, *, * \rang$ returns at most one value.

A violation of these rules is a proof of misbehavior.


### Local views

The Condition State is potentially infinite, given that the number of heights and rounds is infinite.
Hence each process $p$ keeps a local, limited view $L_p$ of the Condition State $C$.
Nodes only query their local views and queries subscripted with the node to indicate which local view is being consulted.
The local view approximates the the full state in the following ways:

- if $\exists p$ such that $\exists e \in L_p$, then $e \in C$
    - adding an entry to a local view adds it to $C$
    - removing an entry from a local view removes it from $C$, if the local view was the last one to contain the entry.

- If an entry is added to $C$ and then removed, it should not be added again. If an entry is added to a local view, then it may be added again, depending on the garbage collection used.
    - Algorithms should enforce that entries are only removed if they will not be needed again and therefore adding them back will not harm agreement.
    - They should strive not to add state again so it eventually gets removed from $C$.
    - $p$ knows how to differentiate a dropped entries from entries that have never been known;


## Communication Requirements

We now formalize the requirements of the Tendermint algorithm in terms of an eventually consistent Condition State, which may be implemented using reliable communication (Gossip Communication), some best *best-effort* communication primitive that only delivers messages that are still useful, or some Gossip/Epidemic/Anti-Entropy approach for state convergence.
All of these can be made to progress after GST but should also progress during smaller stability periods.

|Eventual Consistency|
|-----|
| If there exists a correct process $p$ such that $e \in L_p$, then, eventually, for every correct process $q$, $e \in L_q$ or there exists a correct process $r$ that removes $e$ from $L_r$.|

> **Note**
> 1. Nodes may learn of an entry deletion before learning of its addition.

In order to ensure convergence even in the presence of failures, the network must be connected in such a way to allow communication around any malicious nodes and to provide redundant paths between correct ones.
This may not be feasible at all times, but should happen at least during periods in which the system is stable.

In other words, during periods without network partition, in which messages are timely delivered, and that are long enough for the multiple communication rounds to succeed, processes will identify and solve differences in their states.
We call "long enough" $\Delta$.

| Eventual $\Delta$-Timely Convergence |
|---|
| If $e\in L_p$, for some correct process $p$, at instant $t$, then by $\text{max} \{t,\text{GST}\} + \Delta$, either $e \notin L_p$ or $e \in L_q$, for every correct process $q$.

$\Delta$ encapsulates the assumption that, during stable periods, timeouts eventually do not expire precociously, given that they all can be adjusted to reasonable values, and the steps needed to converge on entries can be accomplished within $\Delta$.
Without precocious timeouts, all new entries in the state should help algorithms progress.
In the Tendermint algorithm, for example, no votes for Nil are added and the round should end if a proposal was made.


# Part 2: CONS/GOSSIP interaction

CONS, the Consensus Reactor State Layer, is where the actions of the Tendermint algorithm are implemented.
Actions are executed once certain pre-conditions apply, such as timeout expirations or reception of information from particular subsets of the nodes in the system, neighbors or not.

An action may require communicating with applications and other reactors, for example to gather data to compose a proposal or to deliver decisions, and with the P2P layer, to communicate with other nodes.

## Northbound Interaction - ABCI
Here we assume that all communication with the Application and other reactors are performed through the Application Blockchain Interface, or [ABCI](../../abci/).
We make such assumption based on the example of how proposals are created; although CONS interacts with with the Mempool reactor to build tentative proposals, actual proposals are defined by the Applications (see PrepareProposal), and therefore the communication with Mempool could be ignored.

ABCI specifies both what CONS [requires from the applications](../../abci/abci%2B%2B_app_requirements.md) and on what CONS [provides to Applications](../../abci/abci%2B%2B_tmint_expected_behavior.md).






## Southbound Interaction - GOSSIP-I
CONS interacts southbound only with GOSSIP, to broadcast messages.

CONS does not handle individual message delivery but, instead, is given conditions check if the set of already received and non-superseded messages match criteria needed to trigger actions.
These conditions may be expressed in different ways:
* the set of messages may be queried by CONS or is directly exposed to CONS by GOSSIP
* CONS provides GOSSIP with predicates to be evaluated over the set of delivered messages and with "callbacks" to be invoked when the predicates evaluate to true.

Both approaches should be equivalent and not impact the specification much, even if the corresponding implementations would be much different.
For now, we follow the first approach by having CONS read the sets directly.[^setsorpred]

[^setsorpred]: **TODO**: should we not specify these shared variables and instead pass predicates to GOSSIP from consensus? Variables make it harder to separate the CONS from GOSSIP, as the the variables are shared, but is highly efficient. Predicates are cleaner, but harder to implement efficiently. For example, when a vote arrives, multiple predicates may have to be tested independently, while with variables the tests may collaborate with each other.

CONS and GOSSIP also share a vocabulary of CONS messages and an operator to test for message supersession.

[VOC-CONS-GOSSIP]    
* Message Types[^messagetypes]
    * ProposalMessage
    * Prevote
    * Precommit
* Actions and predicates
    * broadcast(p, m): p broadcasts message m.
    * delivered(p): messages delivered $p$.
* Supersession
    * $\text{SSS}(\_,\_)$: the supersession operator

[^messagetypes]: **TODO**: specify message contents as they are needed to specify SSS, below.
 

### Requires from GOSSIP

A process needs to be able to broadcast and receive messages broadcast by itself and others.

| [REQ-CONS-GOSSIP-BROADCAST.1]|
|----|
| A process $p$ can broadcast a message $m$ to all of its neighbors.

|[REQ-CONS-GOSSIP-DELIVERY.1] |
|----|
| A process $p$ delivers messages broadcast by itself and other processes.


As per the discussion in [Part I](#part-1-background), CONS requires a **Best-Effort Communication with Supersession** from the pair broadcast/delivered.
Best effort implies that, under good conditions, broadcast and non superseded messages are delivered.

| [REQ-CONS-GOSSIP-BROADCAST.2] |
|----|
| (a) For all processes $p,q$ and message $m1$, if broadcast(p,m1), no failures or network partitions happen afterwards, and there is no message $m2$ broadcast such that m2.SSS(m1), then eventually, for every process $q$, $m1 \in delivered(q)$.|
| (b) If, for some correct process $p$, $m1 \in delivered(q)$, no failures or network partitions happen afterwards, and there is no message $m2$ broadcast such that m2.SSS(m1), 
then eventually, for every process $q$, $m \in delivered(q)$.

Messages delivered must remain available at least while they are not superseded.

|[REQ-CONS-GOSSIP-DELIVERY.2]|
|----|
|For any message $m1 \in superDelivered(p)$ at time $t1$, if there exists a time $t3, t1 \leq t3$, at which $m1 \notin superDelivered(p)$, then there exists a time $t2, t1 \leq t2 \leq t3$ at which there exists a message $m2 \in superDelivered(p), m2.\text{SSS}(m1)$

> :clipboard: **TODO**: time can be replaced by leads to.

### Provides to GOSSIP

In order to identify when a message has been superseded, GOSSIP must be provided with a supersession operator.

|[PROV-CONS-GOSSIP-SUPERSESSION.1]|
|----|
|`SSS(lhs,rhs)` returns true if and only if $\text{lhs}.\text{SSS}(\text{rhs})$

> :clipboard: **TODO**: Define supersession for messages in the GOSSIP-I vocabulary.


|[PROV-CONS-GOSSIP-SUPERSESSION.2]|
|----|
| The number of non-superseded messages broadcast by a process is limited by some constant.

And does not broadcast messages superseded at creation (TODO: at all?)

|[PROV-CONS-GOSSIP-SUPERSESSION.3]|
|----|
| If $p$ broadcast $m2$ at time $t1$, then $p$ does not broadcast any $m1$, $m2.\text{SSS}(m1)$ at any point in time $t2 > t1$.

## Problem Statement (TODO: better title)

> **TODO**: a big, TODO. 

Here we show that "Best-Effort Superseded communication" + GST implies "Eventual $\Delta$-Timely Superseded communication", needed by the consensus protocol to make progress. In other words we show that 

[REQ-CONS-GOSSIP-BROADCAST.1] + [REQ-CONS-GOSSIP-BROADCAST.2] + [REQ-CONS-GOSSIP-DELIVERY.1] + [REQ-CONS-GOSSIP-DELIVERY.2] + GST implies "Eventual $\Delta$-Timely Superseded communication"



# Part III: GOSSIP requirements and provisions 
GOSSIP, the Consensus Reactor Communication Layer, provides on its northbound interface the facilities for CONS to communicate with other nodes by sending gossiping the messages broadcast by CONS and accumulating the gossiped messages while they have not been superseded.
On its southbound interface, GOSSIP relies on the P2P layer to implement the gossiping.

## Northbound Interaction - GOSSIP-I
Northbound interaction is performed through GOSSIP-I, whose vocabulary has been already [defined](#gossip-i-vocabulary).

Next we enumerate what is required and provided from the point of view of GOSSIP as a means to detect mismatches between CONS and GOSSIP.

### Requires from CONS
Because connections and disconnections may happen continuously and the total membership of the system is not knowable, reliably delivering messages in this scenario would require buffering messages indefinitely, in order to pass them on to any nodes that might be connected in the future.
Since buffering must be limited, GOSSIP needs to know which messages have been superseded and can be dropped, and that the number of non-superseded messages at any point in time is bounded.[^drop]

[^drop]: Supersession allows dropping messages but does not require it.


|[REQ-GOSSIP-CONS-SUPERSESSION.1]|
|----|
|`SSS(lhs,rhs)` is provided.


|[REQ-GOSSIP-CONS-SUPERSESSION.2]|
|----|
| There exists a constant $c \in Int$ such that, at any point in time, for any process $p$, the subset of messages broadcast by $m$ that have not been superseded, by the $p$'s knowledge, is smaller than $c$.





### Provides to CONS

| [PROV-GOSSIP-CONS-BROADCAST.1]|
|----|
| To broadcast a message $m$ to its neighbors, process $p$ executes $\text{broadcast}(p,m)$.

| [PROV-GOSSIP-CONS-DELIVERY.1]|
|----|
| $\text{delivered}(p)$ is a set of messages received by $p$.


|[PROV-GOSSIP-CONS-BROADCAST.2]|
|-----|
| TODO
| [PROV-CONS-GOSSIP-SUPERSESSION.3] + ?? implies [REQ-CONS-GOSSIP-BROADCAST.2] is satisfied.

Observe that the requirements from CONS allows GOSSIP to provide broadcast guarantees as a best effort and while bounding the memory used. That is, 

|[PROV-GOSSIP-CONS-DELIVERY.2]|
|----|
| If some some message $m1$ is ever contained in delivered(p), then $m$ will be in any successive calls to delivered(p) at least until some $m2$, $m2.SSS(m1)$, is in delivered(p).

## SouthBound Interaction
Differently from the interaction between GOSSIP and CONS, in which GOSSIP understands CONS messages, P2P is oblivious to the contents of messages it transfers, which simplifies the P2P-I interface in terms of message types.

### P2P-I Vocabulary

[VOC-GOSSIP-P2P]
* nes[p]: sets of current connections of $p$
* igs[p]: set of processes not to establish connections
* uMsgs[p][q]: set of messages sent by $p$ to $q \in \text{nes}[p]$ not yet acknowledged as received by $q$.
* rMsgs[p][q]: set of messages received by $q$ from $p$
* maxConn[p]: maximum number of connections for $p$

### Requires from P2P - P2P-I
GOSSIP on a node needs to know to which other nodes it is connected.

| [REQ-GOSSIP-P2P-CONNECTION.1]|
|----|
| $\text{nes}[p]$ is the set of nodes to which $p$ is currently connected.

> **TODO**: Add permalink


P2P must expose functionality to allow 1-1 communication with connected nodes.

| [REQ-GOSSIP-P2P-UNICAST.1]|
|----|
| Adding message $m$ to $\text{uMsgs}[p][q], q \in \text{nes}[p]$, unicasts $m$ from $p$ to $q$.


Message to nodes that remain connected are reliably delivered.

| [REQ-GOSSIP-P2P-UNICAST.2] |
|----|
| A message added to $\text{uMsgs}[p][q]$ is only removed from $\text{uMsgs}[p][q]$ once it has been added to $\text{rMsgs}[q][p]$ or if q is removed $\text{nes}[p]$.

> **TODO**: Add permalink

|[REQ-GOSSIP-P2P-CONCURRENT_CONN] |
|----|
| The size of nes[p] should never exceed maxConn[p]

|[REQ-GOSSIP-P2P-IGNORING] |
|----|
| Processes in igs[p] should never belong to nes[p].

> **TODO**: Add permalink

### Non-requirements
- Non-duplication
    - GOSSIP itself can duplicate messages, so the State layer must be able to handle them, for example by ensuring idempotency.
- Non-refutation
    - It is assumed that all communication is authenticated at the gossip level.




# Part IV: Closing

> :clipboard: **TODO** Anything else to add?

## References
- [1]: https://arxiv.org/abs/1807.0493 "The latest gossip on BFT consensus"
- [2]: https://github.com/tendermint/tendermint/blob/master/docs/architecture/adr-052-tendermint-mode.md "ADR 052: Tendermint Mode"
