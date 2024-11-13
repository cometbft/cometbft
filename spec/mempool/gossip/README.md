# Mempool Gossip

This directory contains specifications of gossip protocols used by the mempool to disseminate
transactions in the network.

## Protocols

- [Flood](flood.qnt). Currently implemented by CometBFT, Flood is a straightforward gossip protocol
  with a focus on rapid transaction propagation.
  - Pros:
    + Latency: nodes forward transactions to their peers as soon as they receive them, resulting in
      the minimum possible latency of decentralised P2P networks.
    + Byzantine Fault Tolerance (BFT): flooding the network with messages ensures malicious actors
      cannot easily prevent transaction dissemination (i.e., censoring), making it resilient to network disruptions
      and attacks.
  - Cons:
    - Bandwidth: the broadcast nature of Flood results in significant redundancy in message
      propagation, leading to exponential increases in bandwidth usage.

## Quint specs

Specifications are written in the [Quint][quint] language. Protocols are mainly described in
comments of the Quint files, where the Quint code can be read as pseudo-code. Quint allows specs to
be executed, tested, and formally verified, but for the moment we use it here just to give structure
to the spec documentation and to type-check the definitions.

The Flood gossip protocol is self-described in its own [flood](flood.qnt) module. It is built on top
of two other modules, which are not strictly needed to understand the protocol:
- [mempool](mempool.qnt) with definitions of common data structures from the mempool, and 
- [p2p](p2p.qnt) with networking definitions, assumptions, and boilerplate.
```mermaid
flowchart TB
    flood --> mempool --> p2p;
```

[quint]: https://quint-lang.org/
