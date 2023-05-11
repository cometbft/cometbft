# Architecture

CometBFT is a state machine replication framework and part of the stack used in the Cosmos ecosystem to build distributed applications, directly or through the Cosmos SDK.
CometBFT uses its northbound API, ABCI, to communicate with applications.
South of CometBFT is the OS' network stack.

CometBFT is implemented in a modular way, separating protocol implementations into modules.
Modules implement the **reactor** interface, allowing them to interact with the P2P layer, through which they communicate with their counterparts on other nodes.
Interaction with the P2P layer is done through what we will call the P2P Interface or simply **P2P-I**.

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

The Consensus module is further divided into two layers:

- **CONS**, keeps the state and transition functions described in the [Tendermint BFT paper](https://arxiv.org/abs/1807.04938) and uses gossiping to communicate with other nodes.
- **GOSSIP**, keeps state and transition functions needed to implement the gossiping of proposals and votes of the Tendermint algorithm on top of the 1-to-1 communication facilities provided by the P2P layer.

In the context of this specification, exchanges between CONS and GOSSIP happen through the **GOSSIP-I** interface.

```
...
==========ABCI=========
                         ┐
  |       CONS      |    |
  |.....GOSSIP-I....|    |  Consensus module
  |      GOSSIP     |    |
                         ┘
- - - - -P2P-I- - - - -
          P2P
=======================
     Network Stack
```

In order to perform a state transition, CONS may need to interact with the application or other modules, for example to gather data to compose a proposal or to deliver decisions.
Although important, these interactions are out of scope of this specification.
The focus of this specification are the interactions of CONS with the GOSSIP via GOSSIP-I and of GOSSIP with P2P, via P2P-I.

The reader interested in the interactions between CONS and applications should refer to the [Application Blockchain Interface (ABCI)](../../abci/) specification, which include among other things both what CONS [requires from applications](../../abci/abci%2B%2B_app_requirements.md) and on what CONS [provides to applications](../../abci/abci%2B%2B_tmint_expected_behavior.md).

As for the interactions between CONS and other reactors, these happen through other means that will be covered elsewhere. For now the reader is referred to the code.
