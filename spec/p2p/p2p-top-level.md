*** This is the beginning of an unfinished draft. Don't continue reading! ***

# << Tendermint P2P >>

<!--- > Rough outline of what the component is doing and why. 2-3 paragraphs 
--->

> Perhaps this should go into a separate document on Tendermint architecture. 

The Tendermint consists of multiple protocols, namely,
- Consensus
- Mempool
- Evidence
- Blocksync
- Statesync

that each plays a role in making sure that validators can produce blocks. These protocols are implemented in so-called reactors (one for each protocol) that encode two functionalities:

- Protocol logic (controlling the local state of the protocols and deciding what messages to send to others, e.g., the rules we find in the arXiv paper)

- Communication. Implement the communication abstractions needed by the protocol on top of the p2p system (e.g., Gossip)
> perhaps we should clarify nomenclature: the Consensus gossip service actually is not implemented by a gossip algorithm but a peer-to-peer system

The p2p system maintains an overlay network that should satisfy a list of requirements (connectivity, stability, diversity in geographical peers) that are inherited from the communication needs of the reactors.





GOAL: what are the consequences of the choices in the config file
 
<!---
# Outline

> Table of content with rough outline for the parts

- [Part I](#part-i---tendermint-blockchain): Introduction of
 relevant terms of the Tendermint
blockchain.

- [Part II](#part-ii---sequential-definition-problem): 
    - [Informal Problem
      statement](#Informal-Problem-statement): For the general
      audience, that is, engineers who want to get an overview over what
      the component is doing from a bird's eye view.
    - [Sequential Problem statement](#Sequential-Problem-statement):
      Provides a mathematical definition of the problem statement in
      its sequential form, that is, ignoring the distributed aspect of
      the implementation of the blockchain.

- [Part III](#part-iii---as-distributed-system): Distributed
  aspects, system assumptions and temporal
  logic specifications.

  - [Incentives](#incentives): how faulty full nodes may benefit from
    misbehaving and how correct full nodes benefit from cooperating.
  
  - [Computational Model](#Computational-Model):
      timing and correctness assumptions.

  - [Distributed Problem Statement](#Distributed-Problem-Statement):
      temporal properties that formalize safety and liveness
      properties in the distributed setting.

- [Part IV](#part-iv---Protocol):
  Specification of the protocols.

     - [Definitions](#Definitions): Describes inputs, outputs,
       variables used by the protocol, auxiliary functions

     - [Protocol](#core-verification): gives an outline of the solution,
       and details of the functions used (with preconditions,
       postconditions, error conditions).

     - [Liveness Scenarios](#liveness-scenarios): when the light
       client makes progress depends heavily on the changes in the
       validator sets of the blockchain. We discuss some typical scenarios.

- [Part V](#part-v---supporting-the-ibc-relayer): Additional
  discussions and analysis


In this document we quite extensively use tags in order to be able to
reference assumptions, invariants, etc. in future communication. In
these tags we frequently use the following short forms:

- TMBC: Tendermint blockchain
- SEQ: for sequential specifications
- LIVE: liveness
- SAFE: safety
- FUNC: function
- INV: invariant
- A: assumption

--->

# Part I - A Tendermint node 

We survey the communication protocols within the reactors. These should inform the requirements of the p2p layer

## Consensus reactor

The consensus protocol is based on the "gossip" property: every message delivered by a correct process (after stabilization time) is received by every correct process within bounded time. In practice, this requirement is sufficient but not necessary. 


#### [TM-REQ-CONS-GOSSIP.0]
If there is a messages *m* emitted by a correct process, such that

- *m* has not been received by correct process *p*
- correct process *p* is in a state where *m* would help *p* "to make progress"

then eventually it MUST BE that
- *p* receives *m*, OR
- *p* transitions into a state where *m* is not needed anymore to make progress.



#### [TM-PROTOCOL-CR-COMM.0]
Gossiping is done by surveying the peers' states within the consensus protocol. The information of peer *p* tells us what consensus messages *p* is waiting for. We then send these messages to *p*.

> In the current information this information is coarse. Peer *p* informs us about consensus height and round, and we send all votes for that height and round to *p*.

TODO: proposal.

The question is under which conditions [[TM-REQ-CR-COMM.0]] is sufficient to implement [[TM-REQ-CONS-GOSSIP.0]]. These conditions translate into requirements for the p2p layer

### Requirements on the p2p layer

There are **local requirements** that express needs as connections to neighbors

#### [TM-REQ-CR+P2P-STABLE.0]
In order to make sure that we can help a peer to make progress, the p2p layer MUST ensure that we need to stay connected to that peer sufficiently long to get a good view of its state, and to act on it by sending messages.

#### [TM-REQ-CR+P2P-ENCRYPTED.0]
In order  to make sure that we can help a peer to make progress, we need to be sure that we can trust the messages from a correct peer (IOW no masquerading). This is done by end-to-end encrypted channels. 


> The previous properties can be used to solve a local version of [[TM-REQ-CONS-GOSSIP.0]], that is, between neighbors. However, [[TM-REQ-CONS-GOSSIP.0]] is global, that is, whoever emits *m*, potentially we require that *m* is received by all correct processes in a potentially dynamic distributed system:
- nodes may join an leave physically (connection)
- validators join and leave logically (the validator set)

This translates into **global requirements** for the p2p layer.

#### [TM-REQ-CR+P2P-CONNECT.0]
TODO:
- all-to-all communication
- eclypse attack
- system can autonomously from failure (self-healing)
- proposers are not disconnected?

#### [TM-REQ-CR+P2P-STABILITY.0]
TODO: stay connected to good peers for some time

#### [TM-REQ-CR+P2P-OPENNESS.0]
TODO: New nodes can join / new validators can join


Since the systems we are building are decentralized and distributed, the global requirements can only be ensured by local actions of the distributed nodes. For instance, openness has been ensured by distinguishing inbound and outbound connections, and making sure that there are always nodes with open inbound connections.


## Mempool

Considering the whole CometBFT system, intuitively, there should be end-to-end 
requirements regarding transaction processing, e.g., "every transaction submitted
should be eventually put into a block" or "every transaction should be put in
at most one block". Note that these are just simple examples that are neither ensured
by current implementation nor are actually required by the system (e.g., typically we 
do not care about "all" transaction but rather "all valid"). 
Which of such requirements are actually ensured by the 
system depends on the guarantees by the concerned parts of the system.

The mempool is a distributed pool of pending transactions.
A pending transaction is a valid transaction that has been submitted by a
client of the blockchain but has not yet been committed to the blockchain.
The mempool is thus fed with client transactions,
that a priori can be submitted to any node in the network.
And it is consumed by the consensus protocol, more specifically by validator nodes,
which retrieve from the mempool transactions to be included in proposed blocks.

As 
- full nodes add transactions to the mempool following client requests, and
- validators read transactions from the mempool to compose proposed blocks,

the mentioned end-to-end requirements on transactions translate into requirements 
on the mempool, e.g., every transaction that is submitted to a fullnode should 
eventually be present at the mempool of a validator/proposer. 

For the purpose of this document (specifying what the mempool needs from p2p), 
we do not strive to precisely state the end-to-end requirements. Rather, we 
will show that the arguably strongest requirement on the mempool do translate roughly
into requirements on the p2p layer that are similar to those of the
consensus reactor.

- e-t-e requirement: "every transaction submitted should be eventually put into a block"
- requirement for consensus: for every submitted transaction from some hight on, 
the transaction is proposed by every correct validator in every height until it is written
into a block (assuming infinitely often correct proposers get their proposal through)  
- current solution to address this requirement: 
    - the mempool of correct validators is organized as a list. 
    > This list can be understood as a shared data structure with relaxed semantics, that is,
    not all nodes need to agree on the order of the elements, but we have a property that 
    approximately says "if at time *t* a transaction tx1 is in the list of all nodes, and a 
    client submits transaction tx2 after time t, then it is never the case the tx2 is before
    tx1 in a list of a correct validator.
    - to prepare a proposal,  a validator requests from the mempool a list of
pending transactions that respects certain limits, in terms of the number of
transactions returned, their total size in bytes, and their required gas.
The mempool then returns the longest **prefix** of its local list of pending
transactions that respects the limits established by the consensus protocol. 
> As a result of these two properties, once a transaction is known to all validators, it 
cannot be overtaken by infinitely many transaction that arrived in the system later (there
is still some possible overtaking due to faulty proposers); cf. no starvation.
- requirement on the mempool: 
   1. every transaction submitted to a correct full node must eventually be included in the mempool of a correct validator
   2. if a transaction is included in the mempool of a correct validator eventually it is included in the mempool of all correct validators 

> Point 1. appears slightly stronger that the requirement of the consensus reactor, as here all full nodes are sources of transaction, while in consensus only validators are source of consensus messages.

> Point 2. has potential to be weakened in practice, as it might be sufficient for a transaction to reach one correct validator as in practice CometBFT decides in one round. 

## Evidence reactor

All that we found for the mempool also applies for the evidence reactor. 
Evidence are specific transactions that are used by the application (e.g., the staking module) to punish misbehavior. Typical applications (CosmosSDK chains) thus would benefit from timeliness: due to proof-of-stake and unbonding periods, a specific evidence transaction has an expiration date after which the application cannot act upon it.

There are several crucial things that are outside of the control of the evidence reactor
1. (application level) unbonding period
2. BFT time (written into the blocks) is under control of consensus
3. when evidence is put into a block (similar to the mempool we may assume that correct proposers always put all the evidence they are aware of in a block, but CometBFT cannot enforce that)
4. when evidence is submitted

In face of this uncertainty, there are only few reasonable requirement we can postulate

A. an evidence transaction reaches all validator nodes as "fast as possible".
B. evidence is not written into a block after it has expired

> With more control over the four points, in principle we might define a real-time property which formalizes "if evidence is submitted at least a day before it expires, then evidence is written into a block before it expires". 

> To ensure B, the unbonding period (and number of blocks) are given to CometBFT as consensus parameters (in genesis; they can be changed by the application via consensus).

<!---
https://github.com/cometbft/cometbft/blob/af3bc47df982e271d4d340a3c5e0d773e440466d/evidence/pool.go#L265

https://github.com/cometbft/cometbft/blob/af3bc47df982e271d4d340a3c5e0d773e440466d/evidence/reactor.go#L193
--> 


 ## The rest


    - blocksync and statesync: mainly request/response protocols
- What does p2p expect from the reactors? (don't falsely report bad nodes; this puts requirements on the reactors and perhaps/likely also on the application running on top of ABCI)

Unclassified expectations (need to figure out where they come from)
- establish connections
- end-to-end encrypted
- prioritization of messages
- non blocking
- don't trust peers (DDOS-resistant)
- manual configuration possible

## Context of this document

> mention other components and or specifications that are relevant for this
spec. Possible interactions, possible use cases, etc. 

> should give the reader the understanding in what environment this component
will be used. 

TODO:
- Reactor API
- Network? How do we communicate with other nodes
- Discuss that validators run special set-up, and manage their own neighborhood (hide behind sentry nodes).
   - As a result: the distributed system is composed of
       - (correct) nodes that follow the protocol described here
       - (potentially) adversarial nodes whose behavior deviates to harm the system
       - (correct) nodes that don't follow the protocol to shield themselves but behave in a "nice way"



# Part II - Sequential Definition of the  Problem


##  Informal Problem statement


The p2p layer, specified here, manages the connections of a Tendermint node with other Tendermint nodes. It continuously provides a list of peers ensuring
1. Connectivity. The overlay network induced by the correct nodes in the local neighborhoods (defined by the lists of peers) is sufficiently connected to the remainder of the network so that the reactors can implement communication on top of it that is sufficient for their needs
    > There is the design decision that the same overlay is used by all reactors. It seems that consensus has the strongest requirements regarding connectivity and this defines the required properties
 
    > The overlay network shall be robust agains eclipse attacks. Apparently the current p2p was designed to mixed geographically close and far away neighbors to achieve that.
2. Stability. Typically, connections between correct peers should be stable
    > Even if at every time *t* we satisfy Point 1, if the overlays at times *t* and *t+1* are totally different, it might be hard to implement decent communication on top of it. E.g., Consensus gossip requires a neighbor to know its neighbors *k* state so that it can send the message to *k* that help *k* to advance. If *k* is connected only one second per hour, this is not feasible.
3. Openness. It is always the case that new nodes can be added to the system
    > Assuming 1. and 2. holds, this means, there must always be nodes that are willing to add connections to new peers.
4. Self-healing. The overlay network recovers swiftly from node crashes, partitions, unstable periods, etc. 


## Sequential Problem statement

> should be English and precise. will be accompanied with a TLA spec.

TODO: This seems to be a research question. Perhaps we can find some simple properties by looking at the peer-to-peer systems academic literature from several years ago?

# Part III - Distributed System

> Introduce distributed aspects 

> Timing and correctness assumptions. Possibly with justification that the
assumptions make sense, e.g., it is in the interest of a full node to behave
correctly 

> should have clear formalization in temporal logic.

## Incentives

TODO: 
- who will follow the protocol who won't
- validators hiding behind sentries (they have an incentive to not run it)
- what can be incentives/strategies of bad nodes 
     - DNS
     - filling up all your connections and then disconnecting you
     - feeding your reactors with garbage
     - corrupt overlay to harm protocols running on top, e.g., isolating validators to prevent them from being proposers, but using them to vote for proposals from the bad nodes

general question (is it likely? do we care)

## Computational Model

TODO: 
- partially synchronous systems?
- nodes maintain  long-term persistent identity (public key)
- nodes interact by exchanging messages via encrypted point-to-point communication channels (connections?)
- deployment flexibility: deployment among multiple administrative domains; administrators may decide whether to expose nodes to the public network; not completely connected

## Distributed Problem Statement

TODO
- peer discovery
    - seed nodes
    - persistent peers (provided by operator; configuration?)
    - peer exchange protocol
- address book
- establishing and managing connections

TODO: notation
- connection vs. channel

### Design choices

> input/output variables used to define the temporal properties. Most likely they come from an ADR

The p2p layer is
    - running the peer exchange protocol PEX (in a reactor)
    - using input from the operator (addresses)
    - responding to other peers wishing to connect
> the latter might just be the result of the first two points on the other peer


TODO: The following two points seem to be implementation details/legacy design decisions
- communicate to the reactors over the reactor API
- I/O
   - dispatch messages incoming from the network to the reactors
   - send messages incoming from the reactors to the network (the peers the messages should go to) 
- number of connections is bounded by constants, say 10 to 50

### Temporal Properties

> safety specifications / invariants in English 

TODO: In a good period, *p* should stay connected with *q*.

> liveness specifications in English. Possibly with timing/fairness requirements:
e.g., if the component is connected to a correct full node and communication is
reliable and timely, then something good happens eventually. 

should have clear formalization in temporal logic.


### Solving the sequential specification

> How is the problem statement linked to the "Sequential Problem statement". 
Simulation, implementation, etc. relations 


# Part IV - Protocol

> Overview


## Definitions

### Data Types

### Inputs


### Configuration Parameters

### Variables

### Assumptions

### Invariants

### Used Remote Functions / Exchanged Messages

## <<Core Protocol>>

### Outline

> Describe solution (in English), decomposition into functions, where communication to other components happens.


### Details of the Functions

> Function signatures followed by pseudocode (optional) and a list of features (required):
> - Implementation remarks (optional)
>   - e.g. (local/remote) function called in the body of this function
> - Expected precondition
> - Expected postcondition
> - Error condition


### Solving the distributed specification

> Proof sketches of why we believe the solution satisfies the problem statement.
Possibly giving inductive invariants that can be used to prove the specifications
of the problem statement 

> In case the specification describes an existing protocol with known issues,
e.g., liveness bugs, etc. "Correctness Arguments" should be replace by
a section called "Analysis"



## Liveness Scenarios



# Part V - Additional Discussions



# Old text

Tendermint (as many classic BFT algorithms) have an all-to-all communication pattern (e.g., every validator sends a `precommit` to every other full node). Naive implementations, e.g., maintaining a channel between each of the *N* validators is not scaling to the system sizes of typical Cosmos blockchains (e.g., N = 200 validator nodes + seed nodes + sentry nodes + other full nodes). There is the fundamental necessity to restrict the communication. There is another explicit requirement which is called "deployment flexibility", which means that we do not want to impose a completely-connected network (also for safety concerns).

The design decision is to use an overlay network. Instead of having *N* connections, each node only maintains a relatively small number. In principle, this allows to implement more efficient communication (e.g., gossiping), provided that with this small number of connections per node, the system as a whole stays connected. This overlay network 
is established by the **peer-to-peer system (p2p)**, which is composed of the p2p layers of the participating nodes that locally decide with which peers a node keeps connections.



# References

[[block]] Specification of the block data structure. 

[[RPC]] RPC client for Tendermint

[[fork-detector]] The specification of the light client fork detector.

[[fullnode]] Specification of the full node API

[[ibc-rs]] Rust implementation of IBC modules and relayer.

[[lightclient]] The light client ADR [77d2651 on Dec 27, 2019].

[RPC]: https://docs.tendermint.com/master/rpc/

[block]: https://github.com/tendermint/spec/blob/master/spec/blockchain/blockchain.md

[TMBC-HEADER-link]: #tmbc-header.1
[TMBC-SEQ-link]: #tmbc-seq.1
[TMBC-CorrFull-link]: #tmbc-corr-full.1
[TMBC-Auth-Byz-link]: #tmbc-auth-byz.1
[TMBC-TIME_PARAMS-link]: tmbc-time-params.1
[TMBC-FM-2THIRDS-link]: #tmbc-fm-2thirds.1
[TMBC-VAL-CONTAINS-CORR-link]: tmbc-val-contains-corr.1
[TMBC-VAL-COMMIT-link]: #tmbc-val-commit.1
[TMBC-SOUND-DISTR-POSS-COMMIT-link]: #tmbc-sound-distr-poss-commit.1

[lightclient]: https://github.com/interchainio/tendermint-rs/blob/e2cb9aca0b95430fca2eac154edddc9588038982/docs/architecture/adr-002-lite-client.md
[fork-detector]: https://github.com/informalsystems/tendermint-rs/blob/master/docs/spec/lightclient/detection.md
[fullnode]: https://github.com/tendermint/spec/blob/master/spec/blockchain/fullnode.md

[ibc-rs]:https://github.com/informalsystems/ibc-rs

[FN-LuckyCase-link]: https://github.com/tendermint/spec/blob/master/spec/blockchain/fullnode.md#fn-luckycase

[blockchain-validator-set]: https://github.com/tendermint/spec/blob/master/spec/blockchain/blockchain.md#data-structures
[fullnode-data-structures]: https://github.com/tendermint/spec/blob/master/spec/blockchain/fullnode.md#data-structures

[FN-ManifestFaulty-link]: https://github.com/tendermint/spec/blob/master/spec/blockchain/fullnode.md#fn-manifestfaulty

[arXiv]: https://arxiv.org/abs/1807.04938