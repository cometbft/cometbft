# Current Implementation
GOSSIP-I is not clearly defined in the current implementation.
Several methods inspect and manipulate both CONS and GOSSIP internal state.
For example, the `enterPropose` function, explained further ahead, creates and broadcasts `ProposalMessage`s, which is a CONS behavior, but also ProposalPartMessage`, which is a GOSSIP behavior.
We will note CONS and GOSSIP behavior, according to our understanding, in the discussions below.


## [VOC-CONS-GOSSIP]

* `ProposalMessage` carries a CometBFT proposal. Type `ProposalMessage` simply embeds a [`Proposal`](./files.go/#proposal) so we refer to both as `ProposalMessage`.


## Broadcast and delivery.

Different messages are broadcast in different situations and need to be analyzed individually with respect to how this property is provided.

Observe that both `receiveRoutine` and `gossipDataRoutine` run on their own go routines and, therefore, concurrently to each other and other parts of the code.

### ProposalMessage
`ProposalMessage` broadcast is triggered by `enterPropose` upon different conditions.
Either way, the following ensues:

1. `enterPropose` calls `decidePropose (defaultDecideProposal)` to
    1. create a `ProposalMessage` (CONS)
    2. create `ProposalPartMessage`s (GOSSIP)
    3. call `sendInternalMessage` to (GOSSIP-I)
        1. publish `ProposalMessage` to `internalQueue`
        2. publish `ProposalPartMessage` to `internalMsgQueue`
2. `enterPropose` calls `newStep` to
    1. publish `RoundStateEvent` on eventBus (general purpose)
    2. publish `EventNewRoundStep` on eventSwitch (internal communication between CONS and GOSSIP)

3. Concurrently, `receiveRoutine` pols messages from the `internalMsgQueue` and passes to `handleMessage` which
    1. if message is of type `ProposalMessage`, passes it to `setProposal (defaultSetProposal)`to 
        1. store the message in `state.Proposal`
        2. create a set of block parts `state.ProposalBlockParts` corresponding to the proposal.
    2. if message is of type `ProposalPartMessage`
        1. calls `addProposalBlockPart` to accumulates block parts for corresponding proposal
        2. if all parts for a message have been accumulated, call `handleCompleteProposal`, meaning that the `ProposalMessage` is delivered.

4. Concurrently, `gossipDataRoutine` sends messages to connected peers.
    1. If a `ProposalMessage`, 

* [REQ-CONS-GOSSIP-BROADCAST.1.1]: 1.3.1 + 1.3.1 imply that the broadcast of a `ProposalMessage` is performed at least locally.
* [REQ-CONS-GOSSIP-DELIVERY.1.1]: [REQ-CONS-GOSSIP-BROADCAST.1.1] + 3.1 + 3.2 imply that the broadcast of a `ProposalMessage` implies that it is delivered locally, if the process is correct.
* [REQ-CONS-GOSSIP-BROADCAST.1.2]: [REQ-CONS-GOSSIP-DELIVERY.1.1] + 4 implies that the broadcast of a `ProposalMessage` will send the message to all neighbors, as long as the `ProposalMessage` is not superseded.


* [REQ-CONS-GOSSIP-BROADCAST.1]: Implied by [REQ-CONS-GOSSIP-BROADCAST.1.1] + [REQ-CONS-GOSSIP-BROADCAST.1.2]
* [REQ-CONS-GOSSIP-DELIVERY.1]: Implied by [REQ-CONS-GOSSIP-DELIVERY.1.1] + [REQ-CONS-GOSSIP-DELIVERY.1.2]



## "delivery"
Most Tendermint BFT actions are triggered when a set of messages received satisfy some criteria.

GOSSIP must, therefore, accumulate the messages received that might still be used to satisfy some condition and let CONS reevaluate conditions whenever a new message is received or a timeout expires.


## [REQ-CONS-GOSSIP-DELIVERY.2]
GOSSIP reacts to CONS messages by adding them to sets of similar messages, within the GOSSIP internal state, and then evaluating if Tendermint conditions are met and triggering changes to CONS, or by itself reacting to implement the gossip communication.

CONS messages are removed from the accumulating sets only when a new round/height is started, which effectively triggers the sending of new messages, which supersedes the ones from the previous round/height.


## [PROV-CONS-GOSSIP-SUPERSESSION] 
Currently the knowledge of message supersession is embedded in GOSSIP, which decides which messages to retransmit based on the CONS' state and the GOSSIP's state, and not provided as an operator.
 
Even though there is no specific superseding operator implemented, superseding happens by advancing steps, rounds and heights.

> @josef-wider
> In the past we looked a bit into communication closure w.r.t. consensus. Roughly, the lexicographical order over the tuples (height, round, step) defines a notion of logical time, and when I am in a certain height, round and step, I don't care about messages from "the past". Tendermint consensus is mostly communication-closed in that we don't care about messages from the past. An exception is line 28 in the arXiv paper where we accept prevote messages from previous rounds vr for the same height.
> 
> I guess a precise constructive definition of "supersession" can be done along these lines.

## [DEF-SUPERSESSION]
> **TODO**    
> Show that the rules used in the code map to the rules provided in the spec.

## [REQ-GOSSIP-CONS-SUPERSESSION.1]
[See](#prov-cons-gossip-supersession).

## [REQ-GOSSIP-CONS-SUPERSESSION.2]
1. Each process only actively participates in one height at a time (proposes/votes);
1. Each process maintains a maximum number of connections (see configuration parameters).
1. Broadcast messages are enqueued at most once for each connection on each iteration of the gossip routines.
1. Unless a message is broadcast multiple times, it will only be sent once on each connection.
1. GOSSIP keeps track of nodes to which it has sent certain kinds of messages, so resending such messages won't happen even if the message is rebroadcast.

> **TODO** .   
> Discuss each kind of message and make the previous itemize a proper argument, which will allow showing that $c$ exists, even if high.


## P2P-Only nodes
[REQ-GOSSIP-CONS-SUPERSESSION.2] implies that the messages being broadcast by the process itself and those being forwarded must be limited.
In Tendermint BFT this is achieved by virtue of only validators broadcasting messages and the set of validators being always limited.

Although only validators broadcast messages, even non-validators (including sentry nodes) must deliver them, because: 
* only the nodes themselves know if they are validators,
* non-validators may also need to decide to implement the state machine replication, and,
* the network is not fully connected and non-validators are used to forward information to the validators.

Non-validators that could support applications above them may be able to inform GOSSIP about superseded messages (for example, upon decisions).

Non-validators that are deployed only to facilitate communication between peers, that is, P2P only nodes that implement GOSSIP but not the CONS, still need to be provided with a supersession operator in order to limit buffering.


All nodes currently run Tendermint BFT, but desire to have lightweight, gossip only only, nodes has been expressed, e.g. in [ADR052](#references)


## [DEF-BROADCAST]
1. BMsgs is implemented by the gossip routines encountering certain conditions that will trigger the broadcast of messages.
1. Broadcast messages are put in outgoing queues for current connections.
1. If the condition persists, then they are put in the queue again on later iterations, for certain messages.
1. If the condition persists, then they are put in the queues for new connections in later iterations.
1. If the condition does not persist, then some superseding messages must have been sent and the superseded one may be dropped (same as remaining in BMsgs but never being actually sent again on the network).

1. Superseding is not explored to drop superseded messages from the outgoing queues.

## [LOCAL_CAST.1]
For each of the neighbors of the node, looping go-routines continuously evaluate the conditions to send messages to other nodes.
If a message must be sent, it is enqueued for transmission using TCP and will either be delivered to the destination or the connection will be dropped.
New connections reset the state of Communication layer wrt the newly connected node (in case it is a reconnection) and any messages previously sent (but possibly not delivered) will be resent if the conditions needed apply. If the conditions no longer apply, it means that the message has been superseded and need no be retransmitted.

1. Every message added to BCast is enqueued or will be enqueued later.
2. Every message enqueued will be delivered using TCP or the connection will break.

## [GLOBAL-CAST.1]
Messages delivered either cause the State to be advanced, causing the message to be superseded, or are added to the Communication layer internal state to be checked for matching conditions in the future.
From the internal state it will affect the generation of new messages, which may have exactly the same contents of the original one or not, either way superseding the original.

1. A received message is processed by the node, validator or not.
2. The updated state causes the "same" message to be sent or a superseding one.











### The inner workings

CONS uses `broadcast` to send messages to all processes, but since the network may not be fully connected, communication happens at a **local** level, in which information is delivered to the neighbors of a node, and a **global level**, in which information is forwarded on to neighbors of neighbors and so on.
While CONS is not aware of this distinction, GOSSIP is, since it is where the global level is implemented.

* Messages
    * proposal
    * prevote
    * precommit
    * has-block-part
    * has-vote
* CONS state
    * height[p]: Nat
    * round[p]: Nat
    * step[p]: {propose, prevote, precommit}
    * decision[p]: List
* GOSSIP state
    * DMsgs[p]: set of messages delivered

* Predicates

> **TODO**
> * Ne[p]: Neighbor set
> * 



**[PROV-GOSSIP-LOCAL_CAST.1]**   
For any message $m1$ added to $\text{BMsgs}[p]$ at instant $t1$, let $\text{NePt1}$ be the value of $\text{Ne}[p]$ at time $t1$.
For each process $q \in NePt1$, $m1$ will be delivered to $q$ at some point in time $t2 > t1$, or there exists a point in time $t3 > 1$ at which $q$ disconnects from $p$, or a message $m3$ is added to $\text{BMsgs}[p]$ at some instant $t3 > t1$ and $m3.\text{SSS}(m1)$.

> **TODO**: Add permalink

**[PROV-GOSSIP-GLOBAL_CAST.1]**   
For every message received, either the message itself is forwarded or a superseding message is broadcast.

> **TODO**: Add permalink


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

