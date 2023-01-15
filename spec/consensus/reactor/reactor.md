# Consensus Reactor

Tendermint Core is a state machine replication framework and part of the stack used in the Cosmos ecosystem to build distributed applications.
Tendermint Core uses its northbound API, ABCI, to communicate with applications.
South of Tendermint Core, is the OS' network stack.

Tendermint Core is implemented in a modular way, separating protocol implementations in **reactors**.
Reactors communicate with their counterparts on other nodes using the P2P layer, through what we will call the **P2P-I**.


```
                                                      SDK Apps
                                                   ==================
 Applications                                        Cosmos SDK    
======================================ABCI============================ ┐
 [Mempool Reactor] [Evidence Reactor] [Consensus Reactor] [PEX] [...]  |
- - - - - - - - - - - - - - - - P2P-I - - - - - - - - - - - -- - - - - | Tendermint
                                  P2P                                  | Core
====================================================================== ┘
                            Network Stack
```


This document focuses on the interactions between the P2P layer and the Consensus Reactor, which is divided into two layers.
The first layer, **CONS**, keeps the state and transition functions described in the [Tendermint BFT paper][1].
Instances of CONS use gossiping to communicate with other nodes.
The second layer, **GOSSIP**, keeps state and transition functions needed to implement gossiping on top of the 1-to-1 communication facilities provided by the P2P layer.
Exchanges between CONS and GOSSIP use multiple forms, but we will call them all **GOSSIP-I** here.


```
...
==========ABCI=========
                         ┐
  |      CONS       |    |
  |.....GOSSIP-I....|    |  Consensus Reactor
  |     GOSSIP      |    |
                         ┘
- - - - - P2P-I - - - -
         P2P
=======================
    Network Stack
```

The goals of this document are to specify the following:
* what CONS requires from and provides to GOSSIP (GOSSIP-I);
* what GOSSIP requires from and provides to CONS (GOSSIP-I); and,
* GOSSIP requires from and provides to P2P (P2P-I) in order to satisfy CONS' needs.

The specification is divided in multiple documents
* reactor.md (this document): specification in English
* [reactor.qnt](./reactor.qnt): corresponding specifications in [Quint](https://github.com/informalsystems/quint)
* [implementation.md](./implementation.md): a description of what is currently implemented in Tendermint Core, in English.
* [implementation.qnt](./implementation.qnt): Quint model of current behavior, for model checking of provided properties.

# Conventions

* MUST, SHOULD, MAY...
* [X-Y-Z-W.C]
    * X: What
        * VOC: Vocabulary
        * DEF: Definition
        * REQ: Requires
        * PROV: Provides
    * Y-Z: Who-to whom
    * W.C: Identifier.Counter


# Status

> **Warning**    
> This is a Work In Progress

> **Warning**    
> Permalinks to excerpts of the Quint specification are provided throughout this document for convenience, but may outdated.

The following table summarizes the relationship between requirements and provisions on the GOSSIP-I, if they are formally defined in Quint, and if there is a discussion of how the current implementation of CometBFT :comet: matches the provisions.

| Requirement |Quint | Provision | Quint | Match | Implemented |
|----|----|----|----|----|----|
| [REQ-CONS-GOSSIP-BROADCAST.1]     | X | [PROV-GOSSIP-CONS-BROADCAST.1]        | X | X |  |
| [REQ-CONS-GOSSIP-DELIVERY.1]      | X | [PROV-GOSSIP-CONS-DELIVERY.1]         | X | X |  |
| [REQ-CONS-GOSSIP-BROADCAST.2]     |   | [PROV-GOSSIP-CONS-BROADCAST.2]        | P |   |  |
| [REQ-CONS-GOSSIP-DELIVERY.2]      | X | [PROV-GOSSIP-CONS-DELIVERY.2]         | X | X |  |
| [REQ-GOSSIP-CONS-SUPERSESSION.1]  | X | [PROV-CONS-GOSSIP-SUPERSESSION.1]     | X | X |  |
| [REQ-GOSSIP-CONS-SUPERSESSION.2]  | X | [PROV-CONS-GOSSIP-SUPERSESSION.2]     |   |   |  |
|                                   |   | [PROV-CONS-GOSSIP-SUPERSESSION.3]     | X |   |  |
| [REQ-GOSSIP-P2P-CONNECTION.1]     | X |                                       |   |   |  |
| [REQ-GOSSIP-P2P-UNICAST.1]        | X |                                       |   |   |  |
| [REQ-GOSSIP-P2P-UNICAST.2]        | X |                                       |   |   |  |
| [REQ-GOSSIP-P2P-CONCURRENT_CONN]  | X |                                       |   |   |  | 

## TODO

This is a high level TODO list.
Smaller items are spread throughout the document.

- Complete the QNT specs
- Update permalink references to QNT
- Update permalinks
- Consider splitting the QNT spec?
    - Common vocabulary
    - CONS
    - GOSSIP

# Outline

> **TODO**: Provide an outline


# Part 1: Background

We assume that you understand the Tendermint BFT algorithm and therefore will not review it here. If this is not the case, please review the algorithm [here](../).

The Tendermint BFT algorithm assumes that a **Global Stabilization Time (GST)** exists, after which communication is reliable and timely:

| Eventual $\Delta$-Timely Communication|
|-----|
|There is a bound $\Delta$ and an instant GST (Global Stabilization Time) such that if a correct process $p$ sends message $m$ at time $t \geq \text{GST}$ to a correct process $q$, then $q$ will receive $m$ before $t + \Delta$.

Tendermint BFT also assumes that this property is used to provide **Gossip Communication**:

|Gossip communication|
|-----|
| (i) If a correct process $p$ sends some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max} \{t,\text{GST}\} + \Delta$.
| (ii) If a correct process $p$ receives some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max}\{t,\text{GST}\} + \Delta$.

Because Gossip Communication requires even messages sent before GST to be reliably delivered between correct processes ($t$ may happen before GST in (i)) and because GST could take arbitrarily long to arrive, in practice, implementing this property would require an unbounded message buffer.

However, while Gossip Communication is a sufficient condition for Tendermint BFT to terminate, it is not strictly necessary that all messages from correct processes reach all other correct processes.
What is required is for either the message to be delivered or that, eventually, some newer message, with information that **supersedes** that in the first message, to be timely delivered.
In Tendermint BFT, this property is seen, for example, when nodes ignore proposal messages from prior rounds.

|Supersession|
|----|
| Given messages $m1$ and $m2$, we say that **$m2$ supersedes $m1$** if after receiving $m2$ a process would make at least as much progress as it would by receiving $m1$ and we note it as $m2.\text{SSS}(m1)$.   
| Supersession is transitive, i.e., if $m3.\text{SSS}(m2)$ and $m2.\text{SSS}(m1)$, then $m3.\text{SSS}(m1)$

> :clipboard: **TODO**: Better way of saying "would make at least as much progress"?

Therefore we formalize the requirements of Tendermint BFT in terms communication primitives that take supersession into account, providing a *best-effort* to deliver all messages but which may not deliver those that have been superseded, and are combined with GST outside of GOSSIP or P2P to ensure eventual progress.


|Communication with Supersession|
|-----|
| If process $p$ broadcast message $m2$, then $p$ does not broadcast any message $m1$, $m2.\text{SSS}(m1)$.[^todo1]
| If $m1$ and $m2$ are broadcast by any processes and $m2.\text{SSS}(m1)$, then the delivery of $m1$ is not required.

[^todo1]: :clipboard: **TODO**: Is this really required? Should this be a requirement of GOSSIP?

To be useful, however, some legitimate effort has to be made to deliver messages.

|Best-Effort Communication with Supersession|
|-----|
| (i) If a correct process $p$ broadcasts some message $m$ at time $t$, $m$ is not superseded, and no failures or network partitions happen after $t$, then eventually every correct process delivers $m$.    
| (ii) If a correct process $p$ receives some message $m$ at time $t$, $m$ is not superseded, and no failures or network partitions happen after $t$, then eventually every correct process delivers $m$.    

> **TODO**: Not sure about the usage of $t$ in the previous definition.

<!-- |Best-Effort Superseded communication|
|-----|
| If $m1$ and $m2$ are broadcast by any correct processes, $m2.\text{SSS}(m1)$, $m2$ is not superseded, and there are no process failures or network partitions, then eventually every correct process delivers at least $m2$. -->

In order to deliver messages even in the presence of failures, the network must be connected in such a way to allow routing messages around any malicious nodes and to provide redundant paths between correct ones.
This may not be feasible at all times, but should happen at least during periods in which the system is "stable".

In other words, if at some point in time messages are no longer superseded and GST is reached, then there should be a time interval $\Delta$ such that all messages from correct processes are delivered within $\Delta$ to all other correct processes.

| Eventual $\Delta$-Timely Communication with Supersession |
|---|
| (i) If a correct process $p$ broadcasts some message $m$ at time $t$ and $m$ is not superseded, then all correct processes will receive $m$ before $\text{max} \{t,\text{GST}\} + \Delta$.    
| (ii) If a correct process $p$ receives some message $m$ at time $t$ and $m$ is not superseded, then all correct processes will receive $m$ before $\text{max}\{t,\text{GST}\} + \Delta$.

$\Delta$ encapsulates the assumption that, after GST, timeouts eventually do not expire precociously, given that they all can be adjusted to reasonable values and the steps needed to deliver a message can be accomplished within $\Delta$.
Without precocious timeouts, no superseding votes for Nil are broadcast, and Best-Effort Communication with Supersession leads to Eventual $\Delta$-Timely Communication with Supersession leads to termination.

While GST cannot be enforced but simply assumed to show that algorithms can make progress under good conditions, in practice, systems go through frequent "long" stable periods, which algorithms that depend on GST can use to make progress.

> :clipboard: **TODO**
> * Refine based on better definition of supersession.
> * Include "message is not superseded before max(t,gst)+Delta"?
> * Consider supersession due to original sender sending a new message or it happening en route?
> * Show that "best-effort superseded communication" + GST implies "Eventual delta timely superseded communication".

# Part 2: CONS/GOSSIP interaction
CONS, the Consensus Reactor State Layer, is where the actions of the Tendermint BFT are implemented.
Actions are executed once certain pre-conditions apply, such as timeout expirations or reception of information from particular subsets of the nodes in the system, neighbors or not.

An action may require communicating with applications and other reactors, for example to gather data to compose a proposal or to deliver decisions, and with the P2P layer, to communicate with other nodes.

## Northbound Interaction - ABCI
Here we assume that all communication with the Application and other reactor are performed through the Application Blockchain Interface, or [ABCI](../../abci/).
Make make such assumption based on the example creating proposals; although CONS interacts with with the Mempool reactor to build tentative proposals, actual proposals are defined by the Applications (see PrepareProposal), and therefore the communication with Mempool can be ignored.

For details on what CONS poses as requirements to Applications, see [ABCI](../../abci/abci%2B%2B_app_requirements.md), and on what CONS provides to Applications, see [ABCI](../../abci/abci%2B%2B_tmint_expected_behavior.md).

> **TODO**    
> * Confirm that the following requirements are made to applications:
>     * Timely creation and validation of proposals
>     * Timely processing of decisions
> 
> * Confirm that the following is properly captured:
>   * Fair proposal selection
>     * Let $V$ be the set of validators
>     * Let $v^p$ be the voting power of a validator $v$
>     * Let $m = \text{mcd}(\{v^p: v \in V\})$
>     * In the absence of validator set changes, in any sequence of heights of length equal to $\sum_{v\in V} v^p/m$, $v$ appears in the sequence $v^p/m$ times.
> * There is a comment by Anca that the same proposer is elected for round 0 and 1, always. Does this break fairness?
> * How to ensure fairness when validator set changes?

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
* Message sets
    * $\text{bMsgs}$: set of messages broadcast by CONS
    * $\text{dMsgs}$: set of messages delivered by CONS through gossiping.
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
| A process $p$ receives messages broadcast by itself and other processes.


As per the discussion in [Part I](#part-1-background), CONS requires a **best-effort** in delivering broadcasting messages.

| [REQ-CONS-GOSSIP-BROADCAST.2] |
|----|
| **Best-Effort Communication with Supersession** is provided.

Best effort implies that non superseded messages are delivered.
For practical reasons, they cannot be kept in $\text{dMsgs}$ *ad eternum* and may be dropped after delivery.
However, CONS requires that non-superseded messages are kept for as long as needed.

|[REQ-CONS-GOSSIP-DELIVERY.2]|
|----|
|For any message $m1 \in \text{dMsgs}[p]$ at time $t1$, if there exists a time $t3, t1 \leq t3$, at which $m1 \notin \text{dMsgs}[p]$, then there exists a time $t2, t1 \leq t2 \leq t3$ at which there exists a message $m2 \in \text{DMsgs}[p], m2.\text{SSS}(m1)$

### Provides to GOSSIP

In order to identify when a message has been superseded, GOSSIP must be provided with a supersession operator.

|[PROV-CONS-GOSSIP-SUPERSESSION.1]|
|----|
|`SSS(lhs,rhs)` returns true if and only if $\text{lhs}.\text{SSS}(\text{rhs})$



|[PROV-CONS-GOSSIP-SUPERSESSION.2]|
|----|
| The number of non-superseded messages broadcast by a process is limited by some constant.

And does not broadcast messages superseded at creation (TODO: at all?)

|[PROV-CONS-GOSSIP-SUPERSESSION.3]|
|----|
| If $p$ broadcast $m2$ at time $t1$, then $p$ does not broadcast any $m1$, $m2.\text{SSS}(m1)$ at any point in time $t2 > t1$.

> :clipboard: **TODO**   
> * Define supersession for messages in the GOSSIP-I vocabulary.

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
| There exists a constant $c \in Int$ such that, at any point in time, for any process $p$, the subset of messages in bMsgs[p] that have not been superseded is smaller than $c$.





### Provides to CONS

| [PROV-GOSSIP-CONS-BROADCAST.1]|
|----|
| To broadcast a message $m$ to its neighbors, process $p$ adds $m$ to set $\text{bMsgs}[p]$.

| [PROV-GOSSIP-CONS-DELIVERY.1]|
|----|
| $\text{dMsgs}[p]$ is the set of messages received by $p$ through gossiping.


|[PROV-GOSSIP-CONS-BROADCAST.2]|
|-----|
| TODO
| [PROV-CONS-GOSSIP-SUPERSESSION.3] + ?? implies [REQ-CONS-GOSSIP-BROADCAST.2] is satisfied.

Observe that the requirements from CONS allows GOSSIP to provide broadcast guarantees as a best effort and while bounding the memory used. That is, 

|[PROV-GOSSIP-CONS-DELIVERY.2]|
|----|
| [REQ-CONS-GOSSIP-DELIVERY.2]

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
| Adding message $m$ to $\text{uMsgs}[p][q]$, q \in \text{nes}[p]$, unicasts $m$ from $p$ to $q$.


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
