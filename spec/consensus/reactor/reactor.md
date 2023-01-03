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
====================================ABCI=============================
  Mempool Reactor     Evidence Reactor    Consensus Reactor   PEX ...
- - - - - - - - - - - - - - - P2P-I - - - - - - - - - - - - - - - - - 
                                 P2P
=====================================================================
                            Network Stack
```

This document focuses on the interactions between the Consensus Reactor and the P2P layer.
The Consensus reactor is, itself, divided into two layers, State and Communication. 

The State layer, or **CONS**, keeps the state and transition functions described in the [The latest GOSSIP on BFT consensus](https://arxiv.org/abs/1807.0493).

The Communication layer, or **GOSSIP**, keeps state and transition functions related to the gossiping with other nodes.
GOSSIP reacts to messages received through the P2P layer by updating GOSSIP's internal state and, when conditions are met, calling into CONS.
It also handles the broadcasts made by CONS.
Exchanges between CONS and GOSSIP use multiple forms, but we will call them all **GOSSIP-I** here.


```
...
==========ABCI=========

  |      CONS       |
  |.....GOSSIP-I....| Consensus Reactor
  |     GOSSIP      |

- - - - - P2P-I - - - -
         P2P
=======================
    Network Stack
```

The goal here is to understand what the CONS requires from GOSSIP (GOSSIP-I) and what the GOSSIP requires from P2P (P2P-I) in order to satisfy the CONS needs.

# Status

This is a Work In Progress and is far from completion. 

Specifications are provided in english, here, and are accompanied by [Quint specifications](https://github.com/tendermint/tendermint/main/spec/consensus/reactor).
Permalinks are inserted throughout this document for convenience but may outdated.

# TODO
- Provide an outline
- Complete the TNT specs
- Update references to TNT
- Update permalinks
- Consider splitting the TNT spec?
    - Common vocabulary
    - CONS
    - GOSSIP
# Outline



# Part 1: Background

Here we assume that you understand the Tendermint BFT algorithm, which has been described in multiple places, such as [here](../).

The Tendermint BFT algorithm assumes that a **Global Stabilization Time (GST)** exists, after which communication is reliable and timely:

> **Eventual $\Delta$-Timely Communication**: There is a bound $\Delta$ and an instant GST (Global Stabilization Time) such that if a correct process $p$ sends message $m$ at time $t \geq \text{GST}$ to a correct process $q$, then $q$ will receive $m$ before $t + \Delta$.

Tendermint BFT also assumes that this property is used to provide **Gossip Communication**:

> **Gossip communication**: (i)If a correct process $p$ sends some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max} \{t,\text{GST}\} + \Delta$.   
Furthermore, (ii)if a correct process $p$ receives some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max}\{t,\text{GST}\} + \Delta$.

Because Gossip Communication requires even messages sent before GST to be reliably delivered between correct processes ($t$ may happen before GST in (i)) and because GST could take arbitrarily long to arrive, in practice, implementing this property would require an unbounded message buffer.

However, while Gossip Communication is a sufficient condition for Tendermint BFT to terminate, it is not strictly necessary that all messages from correct processes reach all other correct processes.
What is required is for either the message to be delivered or that, eventually, some newer message, with information that supersedes that in the first message, to be timely delivered.
In Tendermint BFT, this property is seen, for example, when nodes ignore proposal messages from prior rounds.

> **Supersession**:    
>Given messages $\text{m}$ and $\text{mn}$, we say that **$\text{mn}$ supersedes $\text{m}$** if after receiving $\text{mn}$ a process would make at least as much progress as it would by receiving $\text{m}$ and we note it as $\text{mn}.\text{SSS}(\text{m})$.   
>Supersession is transitive, i.e., if $mm.SSS(mn)$ and $mn.SSS(m)$, then $mm.SSS(m)$
> **TODO**: Better definition?

It seems reasonable, therefore, to formalize the requirements of Tendermint in terms communication primitives that make a *best-effort* to deliver all messages but may drop superseded messages, and that the best-effort guarantees are combined with GST outside of GOSSIP or P2P, to ensure eventual termination.


> **Superseded communication**:    
If $m$ and $mn$ are broadcast in this order by any processes and $mn.SSS(m)$, then the delivery of $m$ is not required.

To be useful, however, some legitimate effort has to be made to deliver messages.

> **Best-Effort Superseded communication**:   
If $m$ and $mn$ are broadcast in this order by any correct processes, $mn.SSS(m)$, no other message that supersedes $mn$ is broadcast, and there are no process failures or network partitions, then eventually every correct process delivers at least $mn$.

This implies that the network must be connected in such a way to allow routing messages around any malicious nodes and to provide redundant paths between processes.
This may not be feasible at all times, but should happen during periods in which the system is "stable".

In other words, if at some point in time messages are no longer superseded and GST is reached, then there should be a time interval $\Delta$ such that all messages are delivered within $\Delta$.

> **Eventual $\Delta$-Timely Superseded communication**: 
(i)If a correct process $p$ sends some message $m$ at time $t$ and $m$ is not superseded, then all correct processes will receive $m$ before $\text{max} \{t,\text{GST}\} + \Delta$.    
Furthermore, (ii)if a correct process $p$ receives some message $m$ at time $t$ and it is not superseded, then all correct processes will receive $m$ before $\text{max}\{t,\text{GST}\} + \Delta$.


#### Current implementation
It is also assumed that, after GST, timeouts do not expire precociously and therefore superseding votes for Nil are not sent, then Eventual $\Delta$-Timely Superseded communication will lead to termination.


> **TODO**
> Refine based on better definition of supersession.
> Messages are not superseded before max(t,gst)+Delta?
> Not superseded by the original sender? 
> En route? Handling a message may trigger a supersession...




# Part 2: Specifications
## The Consensus Reactor State Layer - CONS
CONS is where the actions of the Tendermint BFT are implemented.
Actions are executed once certain pre-conditions apply, such as timeout expirations or reception of information from particular subsets of the nodes in the system., neighbors or not.

An action may require communicating with applications and other reactors, for example to gather data to compose a proposal or to deliver decisions, and with the P2P layer, to communicate with other nodes.

### Northbound Interaction
Although CONS communicates with the Mempool reactor to build tentative proposals, actual proposals are defined by the application.
Hence we ignore other reactors here and refer to the northbound interaction as being only to Applications, which is covered by the [ABCI](../../abci/) specification.

#### Requires from Applications
For details on what CONS poses as requirements to applications, see [ABCI](../../abci/abci%2B%2B_app_requirements.md).

> **TODO**    
> Confirm that the following requirements are made to applications:
> * Timely creation and validation of proposals
> * Timely processing of decisions


#### Provides to Applications
For details on what CONS provides to applications, see [ABCI](../../abci/abci%2B%2B_tmint_expected_behavior.md)

> **TODO**    
> Confirm that these are properly captured: 
> * Fair proposal selection
>   * Let V be the set of validators
>   * $\forall v \in V$, let $V[v]$ be the voting power of $v$
>   * Let $m = \text{mcd}(\{V[v]: v \in V\})$
>   * In the absence of validator set changes, in any sequence of heights of length equal to $\sum_{v\in V} V[v]/m, v$ is the proposer $V[v]/m$ times.
>       * This spec seems to be broken based on Anca’s comment that the same proposer is elected for round 0 and 1, always.
>       * What to do with validator changes?



### Southbound Interaction
CONS interacts southbound only with GOSSIP, to broadcast and receive messages of predefined types.

#### GOSSIP-I Vocabulary
CONS uses GOSSIP to broadcast messages but it is not informed of specific message reception events. Instead, it is called back when the set of messages received, combined with CONS internal state, matches certain criteria.
Hence CONS and GOSSIP share a vocabulary of messages sent by CONS, of CONS state observed by GOSSIP, and of predicates over sets of messages received by GOSSIP.


[CONS-GOSSIP-VOCABULARY]

* Messages
    * proposal
    * prevote
    * precommit
* CONS state
    * height: Nat
    * round: Nat
    * step: {propose, prevote, precommit}
    * decision: List[Value]
* Predicates



### Requires from GOSSIP
CONS assumes a single method to broadcast messages, `broadcast`.
As per the discussion in Part I, CONS requires a **best-effort** in broadcasting messages, allowing GOSSIP to drop messages no longer useful or, in other words, which have been **superseded**.

<!--> Although, ideally, CONS should not be concerned about the implementation details of `broadcast`, it would be unreasonable to do so as it could impose unattainable behavior to the Communication layer.
Specifically, because the network is potentially not complete, message forwarding may be required to implement `broadcast` and, as discussed previously, this would require potentially infinite message buffers on the "relayers".-->

> **Warning**    
> Since it would be impossible for all the nodes to know immediately when a message is superseded, we use non-superseded as a synonym to "not known by the node to have been superseded".

**[REQ-CONS-GOSSIP-BROADCAST]**    
Eventual $\Delta$-Timely Superseded Communication, as defined in [#Part I].


Most Tendermint BFT actions are triggered when a set of messages received satisfy some criteria.
GOSSIP must, therefore, accumulate the messages received that might still be used to satisfy some condition and let CONS reevaluate conditions whenever a new message is received or a timeout expires.

> **Warning**    
> The paragraph above departs from what was stated in the vocabulary wrt to the application providing predicates for GOSSIP to evaluate and, instead, opens the state of GOSSIP for CONS to evaluate predicates itself.
Which of these two approaches will be the used in the end is still to be defined.

**[REQ-CONS-GOSSIP-KEEP_NON_SUPERSEDED]**    
For any message $m \in \text{DMsgs}[p]$ at time $t$, if there exists a time $t2, t \leq t2$, at which $m \notin \text{DMsgs}[p]$, then $\exists t1, t \leq t1 \leq t2$ at which there exists a message $m1 \in \text{DMsgs}[p], m1.\text{SSS}(m)$

### Provides to GOSSIP

**[PROV-CONS-GOSSIP-SUPERSESSION]**
In order to identify when a message has been superseded, the CONS must provide GOSSIP with a supersession operator `SSS(lhs,rhs)`, which returns true iff $\text{lhs}$ supersedes $\text{rhs}$

**[DEF-SUPERSESSION]**    
> :clipboard: TODO   
> Define

## Communication Layer (AKA GOSSIP)
GOSSIP provides the facilities for CONS to communicate with other nodes by sending messages and by receiving callbacks when conditions specified in the algorithm are met.

### Requires from CONS - GOSSIP-I
Since the network is not fully connected, communication happens at a **local** level, in which information is delivered to the neighbors of a node, and a **global level**, in which information is forwarded on to neighbors of neighbors and so on.

Since connections and disconnections may happen continuously and the total membership of the system is not knowable, reliably delivering messages in this scenario would require buffering messages indefinitely, in order to pass them on to any nodes that might be connected in the future.
Since buffering must be limited, GOSSIP needs to know which messages have been superseded and can be dropped, and that the number of non-superseded messages at any point in time is bounded.


**[REQ-GOSSIP-CONS-SUPERSESSION.1]**   
`SSS(lhs,rhs)` is provided.


**[REQ-GOSSIP-CONS-SUPERSESSION.2]**    
There exists a constant $c \in Int$ such that, at any point in time, for any process $p$, the subset of messages in Msgs[p] that have not been superseded is smaller than $c$.

> **Note**    
> Supersession allows dropping messages but does not require it.

https://github.com/tendermint/tendermint/blob/95e05b33b1ad95a88c6aac8eafc68421053bf0f2/spec/consensus/reactor/reactor.tnt#L45-L58


[REQ-GOSSIP-CONS-SUPERSESSION.2] implies that the messages being broadcast by the process itself and those being forwarded must be limited.
In Tendermint BFT this is achieved by virtue of only validators broadcasting messages and the set of validators being always limited.

Although only validators broadcast messages, even non-validators (including sentry nodes) must deliver them, because: 
* only the nodes themselves know if they are validators,
* non-validators may also need to decide to implement the state machine replication, and,
* the network is not fully connected and non-validators are used to forward information to the validators.

Non-validators that could support applications on top may be able to inform the communication layer about superseded messages (for example, upon decisions).

Non-validators that are deployed only to facilitate communication between peers (i.e., P2P only nodes, which implement the Communication layer but not the State layer) still need to be provided with a supersession operator in order to limit buffering.

#### Current implementation: P2P only nodes**
All nodes currently run Tendermint BFT, but desire to have lightweight, gossip only only, nodes has been expressed, e.g. in [ADR052](#references)

### Provides to the State layer (GOSSIP-I)

To broadcast as message $m$, process $p$ adds it to the set `BMsgs[p]` set.

**[DEF-BROADCAST]**    
https://github.com/tendermint/tendermint/blob/95e05b33b1ad95a88c6aac8eafc68421053bf0f2/spec/consensus/reactor/reactor.tnt#L60-L63



**[PROV-GOSSIP-CONS-NEIGHBOR_CAST.1]**   
For any message $m$ added to BMsgs[p] at instant $t$, let NePt be the value of Ne[$p$] at time $t$; for each process $q \in \text{Ne}[p]$, $m$ will be delivered to $q$ at some point in time $t1 > t$, or there exists a point in time $t2 > t$ at which $q$ disconnects from $p$, or a message $m1$ is added to $\text{BMsgs}[p]$ at some instant $t3 > t$ and $\text{SSS}(m1,m)$.

https://github.com/tendermint/tendermint/blob/95e05b33b1ad95a88c6aac8eafc68421053bf0f2/spec/consensus/reactor/reactor.tnt#L66-L83


**[PROV-GOSSIP-CONS-NEIGHBOR_CAST.2]**   
For every message received, either the message itself is forwarded or a superseding message is broadcast.

https://github.com/tendermint/tendermint/blob/95e05b33b1ad95a88c6aac8eafc68421053bf0f2/spec/consensus/reactor/reactor.tnt#L90-L106

Observe that the requirements from the State allow the Communication layer to provide these guarantees as a best effort and while bounding the memory used.

> [PROV-GOSSIP-CONS-NEIGHBOR_CAST.1] + [PROV-GOSSIP-CONS-NEIGHBOR_CAST.2] + [REQ-GOSSIP-CONS-SUPERSESSION.2] = Best effort communication + Bounded memory usage.

#### Current implementations

`broadcast`   
State does not directly broadcast messages; it changes its state and rely on the Communication layer to see the change in the state and propagate it to other nodes.

[PROV-GOSSIP-CONS-NEIGHBOR_CAST.1]   
For each of the neighbors of the node, looping go-routines continuously evaluate the conditions to send messages to other nodes.
If a message must be sent, it is enqueued for transmission using TCP and will either be delivered to the destination or the connection will be dropped.
New connections reset the state of Communication layer wrt the newly connected node (in case it is a reconnection) and any messages previously sent (but possibly not delivered) will be resent if the conditions needed apply. If the conditions no longer apply, it means that the message has been superseded and need no be retransmitted.

[PROV-GOSSIP-CONS-NEIGHBOR_CAST.2]   
Messages delivered either cause the State to be advanced, causing the message to be superseded, or are added to the Communication layer internal state to be checked for matching conditions in the future.
From the internal state it will affect the generation of new messages, which may have exactly the same contents of the original one or not, either way superseding the original.

### Requires from P2P (P2P-I)

The P2P layer must expose functionality to allow 1-1 communication at the Communication layer, for example to implement request/response (e.g., "Have you decided?").

**[REQ-GOSSIP-P2P-UNICAST.1]**   
Ability address messages to a single neighbor.

https://github.com/tendermint/tendermint/blob/95e05b33b1ad95a88c6aac8eafc68421053bf0f2/spec/consensus/reactor/reactor.tnt#L108-L113

**[REQ-GOSSIP-P2P-UNICAST.2]**   
Requirement for the unicast message to be delivered.

> **TODO**
> Is this a real requirement? If it is, is [PROV-GOSSIP-CONS-NEIGHBOR_CAST.1] a real requirement?
> How different are these 2?

https://github.com/tendermint/tendermint/blob/95e05b33b1ad95a88c6aac8eafc68421053bf0f2/spec/consensus/reactor/reactor.tnt#L115-L132


**[REQ-GOSSIP-P2P-NEIGHBOR_ID]**    
Ability to discern sources of messages received.


> **TODO**
> How to specify that something WAS true?
> If a message was received from 

Moreover, since the Communication layer must provide 1-to-many communication, the P2P layer must provide:

**[REQ-GOSSIP-P2P-CONCURRENT_CONN]**    
Support for connecting to multiple nodes concurrently.

> **TODO**    
> Is this useful, to state that the set of neighbors could have more than 2 values?
```scala
assume _ = Proc.forall(p => size(Ne[p]) >= 0)
```

**[REQ-GOSSIP-P2P-CHURN-DETECTION]**    
Support for tracking connections and disconnections from neighbors.

```scala
assume _ = Proc.forall(p => size(Ne[p]) >= 0)
```


**[REQ-GOSSIP-P2P-NON_REFUTABILITY]**     
Needed for authentication.


#### Current implementations

[REQ-GOSSIP-P2P-UNICAST]   
- `Send(Envelope)`/`TrySend(Envelope)`
    - Enqueue and forget. 
    - Disconnection and reconnection causes drop from queues.
    - Enqueuing may block for a while in `Send`, but not on `TrySend`

[REQ-GOSSIP-P2P-NEIGHBOR_ID]
- Node cryptographic IDs.
- IP Address

[REQ-GOSSIP-P2P-CONCURRENT_CONN]    
- Inherited from the network stack
- Driven by PEX and config parameters

[REQ-GOSSIP-P2P-CHURN-DETECTION]    
- `AddPeer`
- `RemovePeer`


[REQ-GOSSIP-P2P-NON_REFUTABILITY]    
- Cryptographic signing and authentication.




#### Non-requirements
- Non-duplication
    - GOSSIP itself can duplicate messages, so the State layer must be able to handle them, for example by ensuring idempotency.



## References
- [The latest gossip on BFT consensus](https://arxiv.org/abs/1807.0493)
- [ADR 052: Tendermint Mode](https://github.com/tendermint/tendermint/blob/master/docs/architecture/adr-052-tendermint-mode.md)
