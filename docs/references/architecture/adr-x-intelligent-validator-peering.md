
# ADR X: Intelligent Validator Peering

## Changelog

- 11/8: Initial draft

## Status

## Context

The existing CometBFT p2p network is agnostic to the distinction between full nodes and validators. Peers are added via config files and PEX, but validators are equally likely to connect to full nodes as they are to validators via PEX. Thus, validators can be peered suboptimally, requiring messages go through extra hops through full nodes in order for the network to proceed through consensus.

This document proposes a way for validators to peer intelligently with each other, increasing the tendency for validators to be directly peered and thus for consensus messages to travel through the network of validators more quickly. We propose adding a reactor for validator p2p, similar to pex, but specifically for the purpose of getting validators peered with each other.

Note that in this document we will refer to validators as any of the core nodes operated by validator entities in the network. They can be sentry nodes or any non-signing nodes, but should be on the hot-path for consensus messages (i.e. they must participate in the network in order for some consensus messages to be propagated).

## Alternative Approaches

- One alternative for the reactor implementation is complete validator peering, where all validators are peered with all other validators. This is not chosen because it can become prohibitively expensive to a large number of peers, when the number of validators is large.


## Decision

## Detailed Design

The CometBFT config should add some new fields which should be populated by validators:

1.  `val_peer_count_low`, `val_peer_count_high`, `val_peer_count_target`. These are the lower bound, upper bound, and target number of validator peers to maintain, respectively. Note that these have some interaction with `p2p.max_num_inbound_peers` and `p2p.max_num_outbound_peers`, which are already in the config. In order to avoid confusion, we will assume that `p2p.max_num_inbound_peers` and `p2p.max_num_outbound_peers` are the maximum number of peers that the node will ever try to maintain, and that the validator peer count bounds are a subset of that. Node operators may need to adjust these parameters to allow for the val peer count bounds to work within the max peer bounds.

2.  `should_join_valp2p` - whether the node should join the valp2p network.

First, we modify the handshake that nodes perform when connecting to each other. Currently, that handshake exchanges `NodeInfo`. We will augment `NodeInfo` to include an `isValidator` field, which will be kept for each peer of the node in the peer's `NodeInfo`. In order to determine the `isValidator` field, nodes should
perform a Diffie-Hellman key exchange on connection, and use this along with state to determine if the peer is a validator.
Note that sentry setups will be broken by this change, as sentry nodes will not be able to participate in the valp2p network.

We will add a new reactor for validator peering, which we will refer to here as valp2p. All nodes in the network should run this new reactor.

This reactor handles the following message types:
-  `Valp2pRequest` - a message requesting the list of validator addresses.
-  `Valp2pResponse` - a message containing a list of validator addresses.
-  `Valp2pAddr` - a message containing a validator address.

The valp2p reactor overrides the `AddPeer` and `Receive` reactor methods.

`AddPeer`: this is called by the switch on every reactor when a peer is added. For valp2p, this should:
1. Check if the peer is validator, if it is not, return.
2. If the peer is validator, record its net address in memory. Then, broadcast its net address to all peers with a `Valp2pAddr` message.

`Receive`: this processes the messages received by this reactor.
For `Valp2pAddr`:
1. Parse and validate the message. The message should contain a valid net address.
2. Record the net address in memory, then broadcast the message to all peers.

For `Valp2pRequest`:
1. Send the source peer a `ValP2pResponse` containing all of the validator addresses recorded in memory.

For `Valp2pResponse`:
1. Record the addresses in memory.

This reactor should also create a `DialValidatorWithAddress` function that uses the switch to dial a peer with a given address, and only keeps the peer if it successfully proves itself as a validator.

Additionally, when this reactor starts on a validator node, it should start a goroutine that periodically:

1. Checks the number of validator peers we are currently connected to.
2. If the number < `val_peer_count_low`, start dialing random validators from our memory address book with `DialValidatorWithAddress` until we reach `val_peer_count_target`. If we do not have enough addresses, send `Valp2pRequests` to random peers until we have a sufficiently large validator address book.
3. If the number > `val_peer_count_high`, drop some random validators until we reach `val_peer_count_target`.

It should also start another goroutine that periodically checks state to update the `isValidator` field of peer node infos.

The `AddPeer` behavior is roughly: when I get connected to a new validator, tell my neighbors about this new validator so they can potentially connect as well, if they are a validator.

The `Receive` behavior is roughly: when I hear about a new validator, keep it in my validator address book so I can potentially connect to it later, and let my neighbors know about it so they can do the same.

[ May need to add some `broadcastHasValidatorAddress` type of thing to tell other nodes not to send me a specific validator address, similar to consensus `broadcastHasVote` ]

This should probabilistically achieve a connected subnetwork of validators with a small diameter, although additional thought needs to be put into how to set the val peer bounds and target based on the number of validators in the network.

## Consequences

> This section describes the consequences, after applying the decision. All
> consequences should be summarized here, not just the "positive" ones.

### Positive
- The validator subnetwork diameter will be smaller, allowing messages to propagate between validators with less latency

### Negative
- There is additional network bandwidth used by the new reactor

### Neutral

## References

- [issue: Create a separate p2p network for validator consensus](https://github.com/skip-mev/cometbft/issues/5)
- [libp2p gossipsub v1.0 spec](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.0.md)