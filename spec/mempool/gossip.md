# Mempool gossip

## Protocols

- [Flood](gossip/flood.md) protocol
  - Pros:
    + Optimal latency (given the constrains of the network topology).
    + BFT: it tolerates malicious behaviour.
  - Cons:
    - Bandwidth: exponential redundancy.

- [DOG](gossip/dog.md) protocol
