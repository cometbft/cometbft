# CONS and GOSSIP

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

The Consensus Reactor is further divided into two layers:

- **CONS**, keeps the state and transition functions described in the [Tendermint BFT paper](https://arxiv.org/abs/1807.04938) and uses gossiping to communicate with other nodes.
- **GOSSIP**, keeps state and transition functions needed to implement the gossiping of proposals and votes of the Tendermint algorithm on top of the 1-to-1 communication facilities provided by the P2P layer.

In the context of this specification, exchanges between CONS and GOSSIP happen through the **GOSSIP-I** interface.

```
...
==========ABCI=========
                         ┐
  |       CONS      |    |
  |.....GOSSIP-I....|    |  Consensus Reactor
  |      GOSSIP     |    |
                         ┘
- - - - -P2P-I- - - - -
          P2P
=======================
     Network Stack
```
