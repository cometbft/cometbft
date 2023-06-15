---
order: 1
parent:
  title: P2P
  order: 6
---

# Peer-to-Peer Communication

A CometBFT network is composed of multiple CometBFT instances, hereafter called
`nodes`, that interact by exchanging messages.

CometBFT assumes a partially-connected network model.
This means that a node is not assumed to be directly connected to every other
node in the network.
Instead, each node is directly connected to only a subset of other nodes,
hereafter called its `peers`.

The peer-to-peer (p2p) communication layer is then the component of CometBFT that:

1. establishes connections between nodes in a CometBFT network
2. manages the communication between a node and the connected peers
3. intermediates the exchange of messages between peers in CometBFT protocols


> **Note**
>
> The specification of CometBFT's p2p communication layer is a work in progress,
> tracked by [issue #19](https://github.com/cometbft/cometbft/issues/19).
