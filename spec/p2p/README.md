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
Instead, each node is directly connected to a subset of other nodes,
hereafter called its `peers`.

The peer-to-peer (p2p) communication layer is the component of CometBFT responsible for:

1. establishing connections between nodes in a CometBFT network
2. managing the communication between a node and the connected peers
3. intermediating the exchange of messages between peers in CometBFT protocols


> **Note**
>
> The specification of CometBFT's p2p communication layer is a work in progress,
> tracked by [issue #19](https://github.com/cometbft/cometbft/issues/19).
