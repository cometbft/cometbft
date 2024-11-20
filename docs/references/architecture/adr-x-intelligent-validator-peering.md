
# ADR X: Intelligent Validator Peering

## Changelog

- 11/8: Initial draft

## Status

## Context

The existing CometBFT p2p network is agnostic to the distinction between full nodes and validators. Peers are added via config files and PEX, but validators are equally likely to connect to full nodes as they are to validators via PEX. Thus, validators can be peered suboptimally, requiring messages go through extra hops through full nodes in order for the network to proceed through consensus.

This document proposes a way for validators to peer intelligently with each other, increasing the tendency for validators to be directly peered and thus for consensus messages to travel through the network of validators more quickly. We propose adding a reactor for validator p2p, similar to pex, but specifically for the purpose of getting validators peered with each other.

## Alternative Approaches

- One alternative for the reactor implementation is complete validator peering, where all validators are peered with all other validators. This is not chosen because it can become prohibitively expensive to a large number of peers, when the number of validators is large.


## Decision

## Detailed Design

### Config

The CometBFT config should add some new fields which should be populated by validators:

1.  `val_peer_count_low`, `val_peer_count_high`, `val_peer_count_target`. These are the lower bound, upper bound, and target number of validator peers to maintain, respectively. Note that these have some interaction with `p2p.max_num_inbound_peers` and `p2p.max_num_outbound_peers`, which are already in the config. `p2p.max_num_inbound_peers` and `p2p.max_num_outbound_peers` are used by the switch (rejects inbound peers when above max) and the pex reactor (uses max outbound to determine how many to dial). These new parameters will operate within those bounds, so node operators may need to adjust these parameters to allow for the val peer count bounds to work within the max peer bounds.

2.  `should_join_valp2p` - whether the node should join the valp2p network.

### Peering Handshake

First, we modify the handshake that nodes perform when connecting to each other. Currently, that handshake exchanges a `Default` `NodeInfo`. We will augment `Default` to include an `IsValidator` field, which will be kept for each peer of the node in the peer's `NodeInfo`. In order to determine the `IsValidator` field, the node handshake will be augmented:

1. Nodes will include their `PublicKey` and a `ChallengeBytes` in the `Default` information exchanged during the handshake. Validators must include this, full nodes do not have to.
2. When two nodes perform a handshake and both claim to be validators, they will both sign `"AuthChallenge + ChallengeBytes"` and send the signature to each other.
3. Both nodes will verify the signature against the other node's claimed public key. If the signatures are valid, they will set `IsValidator` to true on the peer.

Note that sentry setups will be broken by this change, as sentry nodes will not be able to participate in the valp2p network.

### Reactor

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

Note that no verification is performed that the Valp2pAddr is actually a validator, so there is a spam attack vector here.

The `AddPeer` behavior is roughly: when I get connected to a new validator, tell my neighbors about this new validator so they can potentially connect as well, if they are a validator.

The `Receive` behavior is roughly: when I hear about a new validator, keep it in my validator address book so I can potentially connect to it later, and let my neighbors know about it so they can do the same.

This should probabilistically achieve a connected subnetwork of validators with a small diameter, although additional thought needs to be put into how to set the val peer bounds and target based on the number of validators in the network.

For `Valp2pRequest`:
1. Send the source peer a `ValP2pResponse` containing all of the validator addresses recorded in memory.

For `Valp2pResponse`:
1. Record the addresses in memory.

Additionally, when this reactor starts on a validator node, it should start a goroutine that periodically:

1. Checks the number of validators in the address book. If it is too little (e.g. less than `val_peer_count_low`), send a `Valp2pRequest` to a random peer.
2. Check the number of validator peers we are currently connected to. If the number < `val_peer_count_low`, start dialing random validators from our memory address book with `switch.DialValidatorWithAddress` until we reach `val_peer_count_target`. If we do not have enough addresses, send `Valp2pRequests` to random peers until we have a sufficiently large validator address book. If the number > `val_peer_count_high`, drop some random validators until we reach `val_peer_count_target`.
3. Periodically check the public keys of connected peers against app state to update the `IsValidator` field of peer node infos.

[ May need to add some `broadcastHasValidatorAddress` type of thing to tell other nodes not to send me a specific validator address, similar to consensus `broadcastHasVote` ]

### Switch

The switch should expose a `DialValidatorWithAddress` function that dials a peer with a given address, and only keeps the peer if it successfully proves itself as a validator.

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