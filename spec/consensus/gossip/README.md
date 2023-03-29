# Consensus Reactor

CometBFT is a state machine replication framework and part of the stack used in the Cosmos ecosystem to build distributed applications, directly or through the Cosmos SDK.
CometBFT uses its northbound API, ABCI, to communicate with applications.
South of CometBFT is the OS' network stack.

CometBFT is implemented in a modular way, separating protocol implementations into **reactors**.
Reactors communicate with their counterparts on other nodes using the P2P layer, through what we will call the **P2P-I**.

```
                                                      SDK Apps
                                                   ==================
 Applications                                        Cosmos SDK
======================================ABCI============================ ┐
 [Mempool Reactor] [Evidence Reactor] [Consensus Reactor] [PEX] [...]  |
- - - - - - - - - - - - - - - - P2P-I - - - - - - - - - - - -- - - - - | CometBFT
                                  P2P                                  | Core
====================================================================== ┘
                            Network Stack
```


The Consensus Reactor is further divided into two layers:
- **CONS**, keeps the state and transition functions described in the [Tendermint BFT paper][1] and uses gossiping to communicate with other nodes.
- **GOSSIP**, keeps state and transition functions needed to implement gossiping on top of the 1-to-1 communication facilities provided by the P2P layer.

Exchanges between CONS and GOSSIP happen through the **GOSSIP-I**.

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

The set of documents in this directory aims at specifying the desired behavior and documenting the current implementation of GOSSIP-I, GOSSIP, and P2P-I. More specifically
1. Provide an english specification of
    * what CONS requires from and provides to GOSSIP (GOSSIP-I);
    * what GOSSIP requires from and provides to CONS (GOSSIP-I); and,
    * what GOSSIP requires from and provides to P2P (P2P-I) in order to satisfy CONS' needs.
2. Provide equivalent Quint specifications, used to mechanically check the properties
3. Provide an english description of how the current implementation matches (or not) the specified behavior, that is, a very loose refinement mapping.

# Outline

The specification is divided into multiple documents:
* [gossip.md](./gossip.md)
    - english specification of GOSSIP, GOSSIP-I, and how GOSSIP may use P2P.
    - The text is filled with `<<TAGS>>` that will eventually be used to implement "literate" specification (automatically paste the corresponding snippets from the Quint spec). For now, refer to the `gossip.qnt` directly.
* [gossip.qnt](./gossip.qnt):
    - corresponding specifications in [Quint](https://github.com/informalsystems/quint)
* [gossipold.qnt](./gossipold.qnt): previous version of the Quint spec, before modularizing. Will be eventually removed.
* [implementation.md](./implementation.md): a description of what is currently implemented in CometBFT, in English. Not updated.
* [implementation.qnt](./implementation.qnt): Quint model of current behavior, for model checking of provided properties. To be written.


# Conventions

* MUST, SHOULD, MAY... are used according to RFC2119.
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
> This is a Work In Progress.


> **Warning**
> This table is outdated.

The following table summarized the relationship between requirements and provisions on the GOSSIP-I, if they are formally defined in Quint, and if there is a discussion of how the current implementation of CometBFT matches the provisions.

| Requirement |Quint | Provision | Quint | Match | Implemented |
|----|----|----|----|----|----|
| [REQ-CONS-GOSSIP-BROADCAST.1]     | X | [PROV-GOSSIP-CONS-BROADCAST.1]        | X | X |  |
| [REQ-CONS-GOSSIP-DELIVERY.1]      | X | [PROV-GOSSIP-CONS-DELIVERY.1]         | X | X |  |
| [REQ-CONS-GOSSIP-BROADCAST.2]     |   | [PROV-GOSSIP-CONS-BROADCAST.2]        | P |   |  |
| [REQ-CONS-GOSSIP-DELIVERY.2]      | X | [PROV-GOSSIP-CONS-DELIVERY.2]         | X | X |  |
| [REQ-GOSSIP-CONS-SUPERSESSION.1]  | X | [PROV-CONS-GOSSIP-SUPERSESSION.1]     | X | X |  |
| [REQ-GOSSIP-CONS-SUPERSESSION.2]  | X | [PROV-CONS-GOSSIP-SUPERSESSION.2]     | X |   |  |
|                                   |   | [PROV-CONS-GOSSIP-SUPERSESSION.3]     | X |   |  |
| [REQ-GOSSIP-P2P-CONNECTION.1]     | X |                                       |   |   |  |
| [REQ-GOSSIP-P2P-UNICAST.1]        | X |                                       |   |   |  |
| [REQ-GOSSIP-P2P-UNICAST.2]        | X |                                       |   |   |  |
| [REQ-GOSSIP-P2P-CONCURRENT_CONN]  | X |                                       |   |   |  |
| [REQ-GOSSIP-P2P-IGNORING]         | X |                                       |   |   |  |
