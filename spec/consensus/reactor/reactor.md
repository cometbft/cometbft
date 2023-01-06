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

The specification is divided in 3 documents
* reactor.md (this document): specification in English
* [reactor.qnt](./reactor.qnt): corresponding specifications in [Quint](https://github.com/informalsystems/quint)
* [implementation.md](./implementation.md): a description of what is currently implemented in Tendermint Core, in English.

# Status

> **Warning**
> This is a Work In Progress

> **Warning**
> Permalinks to excerpts of the Quint specification are provided for convenience throughout this document for convenience, but may outdated.


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

> **Eventual $\Delta$-Timely Communication**    
> There is a bound $\Delta$ and an instant GST (Global Stabilization Time) such that if a correct process $p$ sends message $m$ at time $t \geq \text{GST}$ to a correct process $q$, then $q$ will receive $m$ before $t + \Delta$.

Tendermint BFT also assumes that this property is used to provide **Gossip Communication**:

> **Gossip communication**    
> * (i) If a correct process $p$ sends some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max} \{t,\text{GST}\} + \Delta$.   
> * (ii) If a correct process $p$ receives some message $m$ at time $t$, all correct processes will receive $m$ before $\text{max}\{t,\text{GST}\} + \Delta$.

Because Gossip Communication requires even messages sent before GST to be reliably delivered between correct processes ($t$ may happen before GST in (i)) and because GST could take arbitrarily long to arrive, in practice, implementing this property would require an unbounded message buffer.

However, while Gossip Communication is a sufficient condition for Tendermint BFT to terminate, it is not strictly necessary that all messages from correct processes reach all other correct processes.
What is required is for either the message to be delivered or that, eventually, some newer message, with information that **supersedes** that in the first message, to be timely delivered.
In Tendermint BFT, this property is seen, for example, when nodes ignore proposal messages from prior rounds.

> **Supersession**     
> * Given messages $m1$ and $m2$, we say that **$m2$ supersedes $m1$** if after receiving $m2$ a process would make at least as much progress as it would by receiving $m1$ and we note it as $m2.\text{SSS}(m1)$.   
> * Supersession is transitive, i.e., if $m3.\text{SSS}(m2)$ and $m2.\text{SSS}(m1)$, then $m3.\text{SSS}(m1)$

> :clipboard: **TODO**    
> Better way of saying "would make at least as much progress"?

Therefore we formalize the requirements of Tendermint BFT in terms communication primitives that take supersession into account, providing a *best-effort* to deliver all messages but which may not deliver those that have been superseded, and are combined with GST outside of GOSSIP or P2P to ensure eventual progress.


> **Superseded communication**    
> If $m1$ and $m2$ are broadcast by any processes and $m2.\text{SSS}(m1)$, then the delivery of $m1$ is not required.

> **Note**    
> A process should not broadcast an already superseded message, so $m1$ must have been broadcast before $m2$.
>> :clipboard: **TODO**: should this be a requirement of GOSSIP?

To be useful, however, some legitimate effort has to be made to deliver messages.

> **Best-Effort Superseded communication**    
> If $m1$ and $m2$ are broadcast by any correct processes, $m2.\text{SSS}(m1)$, no other message that supersedes $m2$ is broadcast, and there are no process failures or network partitions, then eventually every correct process delivers at least $m2$.

In order to deliver messages even in the presence of failures, the network must be connected in such a way to allow routing messages around any malicious nodes and to provide redundant paths between correct ones.
This may not be feasible at all times, but should happen at least during periods in which the system is "stable".

In other words, if at some point in time messages are no longer superseded and GST is reached, then there should be a time interval $\Delta$ such that all messages from correct processes are delivered within $\Delta$ to all other correct processes.

> **Eventual $\Delta$-Timely Superseded communication**: 
> * (i) If a correct process $p$ broadcasts some message $m$ at time $t$ and $m$ is not superseded, then all correct processes will receive $m$ before $\text{max} \{t,\text{GST}\} + \Delta$.    
> * (ii) If a correct process $p$ receives some message $m$ at time $t$ and $m$ is not superseded, then all correct processes will receive $m$ before $\text{max}\{t,\text{GST}\} + \Delta$.

GST cannot be enforced but simply assumed to show that algorithms can make progress under good conditions.
In practice, observations of actual systems show that "long" stable periods are frequent and algorithms that depend on GST to progress can use these stable periods to make progress.

It is also assumed that, after GST, eventually timeouts do not expire precociously and therefore superseding votes for Nil are not broadcast, and Best-Effort Superseded communication leads to Eventual $\Delta$-Timely Superseded communication leads to termination.


> **TODO**    
> * Refine based on better definition of supersession.
> * Include "message is not superseded before max(t,gst)+Delta"?
> * Consider supersession due to original sender sending a new message or it happening en route?


# Part 2: CONS/GOSSIP interaction
CONS, the Consensus Reactor State Layer, is where the actions of the Tendermint BFT are implemented.
Actions are executed once certain pre-conditions apply, such as timeout expirations or reception of information from particular subsets of the nodes in the system, neighbors or not.

An action may require communicating with applications and other reactors, for example to gather data to compose a proposal or to deliver decisions, and with the P2P layer, to communicate with other nodes.

## Northbound Interaction - ABCI
Although CONS communicates with the Mempool reactor to build tentative proposals, actual proposals are defined by the Applications (see PrepareProposal).
Hence we ignore other reactors here and refer to the northbound interaction as being only to Applications, which is covered by the [ABCI](../../abci/) specification.

For details on what CONS poses as requirements to applications, see [ABCI](../../abci/abci%2B%2B_app_requirements.md).

> **TODO**    
> Confirm that the following requirements are made to applications:
> * Timely creation and validation of proposals
> * Timely processing of decisions

For details on what CONS provides to applications, see [ABCI](../../abci/abci%2B%2B_tmint_expected_behavior.md)

> **TODO**    
> * Confirm that the following is properly captured:
>   * Fair proposal selection
>     * Let $V$ be the set of validators
>     * Let $v^p$ be the voting power of a validator $v$
>     * Let $m = \text{mcd}(\{v^p: v \in V\})$
>     * In the absence of validator set changes, in any sequence of heights of length equal to $\sum_{v\in V} v^p/m$, $v$ appears in the sequence $v^p/m$ times.
> * There is a comment by Anca that the same proposer is elected for round 0 and 1, always. Does this break fairness?
> * How to ensure fairness when validator set changes?


## Southbound Interaction
CONS interacts southbound only with GOSSIP, to broadcast and receive messages of predefined types.

### GOSSIP-I Vocabulary
CONS uses GOSSIP to broadcast messages but it is not informed of specific message reception events. 
Instead, it is called back when the set of messages received, combined with CONS internal state, matches certain criteria.
Hence CONS and GOSSIP share a vocabulary of messages sent by CONS, of CONS state observed by GOSSIP, and of predicates over sets of messages received by GOSSIP.


[CONS-GOSSIP-VOCABULARY]

* Messages Types
    * proposal
    * prevote
    * precommit
* bMsgs: set of messages broadcast by CONS


<!-- >
* CONS state
    * height[p]: Nat
    * round[p]: Nat
    * step[p]: {propose, prevote, precommit}
    * decision[p]: List
* GOSSIP state
    * DMsgs[p]: set of messages delivered

* Predicates
-->


> **TODO**    
> Complete vocabulary.



### Requires from GOSSIP

[REQ-CONS-GOSSIP-BROADCAST.1]
There is an API to include messages in $\text{bMsgs}$.

As per the discussion in [Part I](#part-1-background), CONS requires a **best-effort** in broadcasting messages, allowing GOSSIP to drop messages no longer useful or, in other words, which have been **superseded**.

> **Warning**    
> Since it would be impossible for all the nodes to know immediately when a message is superseded, we use non-superseded as a synonym for "not yet known by the node to have been superseded".

**[REQ-CONS-GOSSIP-BROADCAST.2]**    
> **TODO**:  Best effort communication that, if combined with GST, leads to Eventual $\Delta$-Timely Superseded Communication, as defined in [Part I](#part-1-background).


Most Tendermint BFT actions are triggered when a set of messages received satisfy some criteria.
GOSSIP must, therefore, accumulate the messages received that might still be used to satisfy some condition and let CONS reevaluate conditions whenever a new message is received or a timeout expires.

> **Warning**    
> The paragraph above departs from what was stated in the vocabulary wrt to the application providing predicates for GOSSIP to evaluate and, instead, opens the state of GOSSIP for CONS to evaluate predicates itself. Which of these two approaches will be the used in the end is still to be defined.

**[REQ-CONS-GOSSIP-KEEP_NON_SUPERSEDED]**    
For any message $m1 \in \text{DMsgs}[p]$ at time $t1$, if there exists a time $t3, t1 \leq t3$, at which $m1 \notin \text{DMsgs}[p]$, then there exists a time $t2, t1 \leq t2 \leq t3$ at which there exists a message $m2 \in \text{DMsgs}[p], m2.\text{SSS}(m1)$

### Provides to GOSSIP

**[PROV-CONS-GOSSIP-SUPERSESSION.1]**    
In order to identify when a message has been superseded, CONS must provide GOSSIP with a supersession operator `SSS(lhs,rhs)`, which returns true if and only if $\text{lhs}.\text{SSS}(\text{rhs})$

> :clipboard: **TODO**   
> * Define supersession for messages in the GOSSIP-I vocabulary.


## Problem Statement (TODO: better title)

> **TODO**: a big, TODO.

Here we show that "Best-Effort Superseded communication" + GST implies "Eventual $\Delta$-Timely Superseded communication", needed by the consensus protocol to make progress. In other words we show that 

[REQ-CONS-GOSSIP-BROADCAST.1] + [REQ-CONS-GOSSIP-BROADCAST.2] + [REQ-CONS-GOSSIP-KEEP_NON_SUPERSEDED] implies "Eventual $\Delta$-Timely Superseded communication"



# Part III: GOSSIP requirements and provisions 
GOSSIP, the Consensus Reactor Communication Layer, provides on its northbound interface the facilities for CONS to communicate with other nodes by sending messages and by receiving callbacks when conditions specified in the algorithm are met.
On its southbound interface, GOSSIP relies on the P2P layer to transfer messages to other nodes.

## Northbound Interaction
Northbound interaction is performed through GOSSIP-I, whose vocabulary has been already [defined](#gossip-i-vocabulary).

Next we enumerate what is required and provided from the point of view of GOSSIP as a means to detect mismatches between CONS and GOSSIP.

### Requires from CONS - GOSSIP-I
Because connections and disconnections may happen continuously and the total membership of the system is not knowable, reliably delivering messages in this scenario would require buffering messages indefinitely, in order to pass them on to any nodes that might be connected in the future.
Since buffering must be limited, GOSSIP needs to know which messages have been superseded and can be dropped, and that the number of non-superseded messages at any point in time is bounded.


**[REQ-GOSSIP-CONS-SUPERSESSION.1]**   
`SSS(lhs,rhs)` is provided.


**[REQ-GOSSIP-CONS-SUPERSESSION.2]**    
There exists a constant $c \in Int$ such that, at any point in time, for any process $p$, the subset of messages in bMsgs[p] that have not been superseded is smaller than $c$.

> **Note**    
> Supersession allows dropping messages but does not require it.


> **TODO**:Add permalink to qnt.

> **TODO**: Add predicates to evaluate GOSSIP's state or let CONS do it directly?



### Provides to CONS - GOSSIP-I

**[PROV-GOSSIP-CONS-BROADCAST.1]**    
To broadcast a message $m$, process $p$ adds it to the $\text{bMsgs}[p]$ set.

Observe that the requirements from CONS allows GOSSIP to provide broadcast guarantees as a best effort and while bounding the memory used. That is, 

**[PROV-GOSSIP-CONS-BROADCAST.2]**    
> **TODO**: Should match **[REQ-CONS-GOSSIP-BROADCAST.2]**

## SouthBound Interaction
Differently from the interaction between GOSSIP and CONS, in which GOSSIP understands CONS messages, P2P is oblivious to the contents of messages it transfers.
Hence, the P2P-I interface is very simple.

### P2P-I Vocabulary

[GOSSIP-P2P-VOCABULARY]
* Peer Management
    * AddPeer(Proc): Inform of peer connection
    * RemovePeer(Proc): Inform of peer disconnection
* Communication
    * Message
    * Send(Peer,Message): Send message to Peer
    * Receive(Peer,Message): Receive message to Peer
    * MaxConn: Nat - maximum number of connections.
    * Disconnect(Peer): forces P2P to disconnect from peer
    * Ignore(Peer): forces P2P to ignore any communication attempts from Peer.

### Requires from P2P - P2P-I
P2P must expose functionality to allow 1-1 communication by GOSSIP, for example to implement request/response protocols.

**[REQ-GOSSIP-P2P-UNICAST.1]**   
Ability to address messages to a connected peer.

> **TODO**: Add permalink


**[REQ-GOSSIP-P2P-UNICAST.2]**   
Unicast message is eventually received if the peers remain connected.

> **TODO**: Add permalink

**[REQ-GOSSIP-P2P-NEIGHBOR_ID]**    
Ability to discern sources of messages received.

Moreover, since GOSSIP must provide 1-to-many communication, P2P must provide:

**[REQ-GOSSIP-P2P-CONCURRENT_CONN]**    
Support for connecting to multiple nodes concurrently, but limited to some maximum value.

> **TODO**    
> Is this useful, to state that the set of neighbors could have more than 0 values?
```scala
assume _ MaxConn: Nat
assume _ = Proc.forall(p => size(Ne[p]) >= 0 && size(Ne[p]) <= MaxConn)
```

**[REQ-GOSSIP-P2P-CHURN-DETECTION]**    
Support for tracking connections and disconnections from neighbors.


**[REQ-GOSSIP-P2P-NON_REFUTABILITY]**     
Needed for authentication.

### Non-requirements
- Non-duplication
    - GOSSIP itself can duplicate messages, so the State layer must be able to handle them, for example by ensuring idempotency.

## Problem statement

> **TODO**: a big, TODO.

Here we show that whatever is required from the P2P layer is enough to implement whatever is required by CONS.









# Part IV: Closing

## References
- [1]: https://arxiv.org/abs/1807.0493 "The latest gossip on BFT consensus"
- [2]: https://github.com/tendermint/tendermint/blob/master/docs/architecture/adr-052-tendermint-mode.md "ADR 052: Tendermint Mode"
