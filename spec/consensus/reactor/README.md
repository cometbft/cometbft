# Consensus Reactor

CometBFT is a state machine replication framework and part of the stack used in the Cosmos ecosystem to build distributed applications, such as the Cosmos SDK.
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
- - - - - - - - - - - - - - - - P2P-I - - - - - - - - - - - -- - - - - | Tendermint
                                  P2P                                  | Core
====================================================================== ┘
                            Network Stack
```


This set of documents focuses on the interactions between the P2P layer and the Consensus Reactor, which is further divided into two layers.
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

The overall goal here is to specify the following:
1. Provide english specification of
    * what CONS requires from and provides to GOSSIP (GOSSIP-I);
    * what GOSSIP requires from and provides to CONS (GOSSIP-I); and,
    * what GOSSIP requires from and provides to P2P (P2P-I) in order to satisfy CONS' needs.
2. Provide equivalent Quint specifications, used to mechanically check the properties
3. Provide an english description of how the current implementation matches (or not) the specified behavior (a very loose refinement mapping)


# Outline

> **TODO**: Provide an outline. Can we use Jekyll?


The specification is divided in multiple documents
* [reactor.md] (./reactor.md): specification in English
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
> Permalinks to excerpts of the Quint specification are provided throughout this document for convenience, but may be outdated.

The following table summarizes the relationship between requirements and provisions on the GOSSIP-I, if they are formally defined in Quint, and if there is a discussion of how the current implementation of CometBFT matches the provisions.

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


## TODO

This is a high level TODO list.
Smaller items are spread throughout the document.

- Complete the QNT specs
- Update permalink references to QNT
- Consider splitting the QNT spec?
    - Common vocabulary
    - CONS
    - GOSSIP

