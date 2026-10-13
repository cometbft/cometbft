# RFC 108: Lib-P2P Peer Exchange

Status: **DRAFT**

## Changelog

- 2026-09-18: First draft (@swift1337)

## Abstract

Comet [v0.39.0](https://github.com/cometbft/cometbft/releases/tag/v0.39.0) introduces a new networking layer built on 
top of Lib-P2P. The initial release had dynamic peer exchange (PEX) out of scope, supporting only a static list of
bootstrapped peers. This RFC aims to close the gap and bring feature parity with the original CometBFT PEX by introducing 
similar behavior of discovering and connecting to peers.

## Background

A Comet node runs **one** of two networking stacks, selected by `p2p.libp2p.enabled` (default off):

- **Comet P2P**: multiplexed TCP, `p2p.Switch`, `p2p/pex` reactor, `addrbook.json`.
- **Lib-P2P**: QUIC host in `lp2p/`. Today PEX is NOT implemented. Connectivity is done via bootstrap peers.

Classic PEX is documented in the [Comet PEX protocol](https://github.com/cometbft/cometbft/blob/main/spec/p2p/implementation/pex-protocol.md)
and [address book](https://github.com/cometbft/cometbft/blob/main/spec/p2p/implementation/addressbook.md).

Classic PEX cannot be reused as-is because Comet P2P and libp2p identify and
address peers differently. Comet uses a 20-byte `p2p.ID` and `id@ip:port`
addresses. Libp2p uses `peer.ID` values derived from public keys, native
multiaddresses, and its own
[Identify protocol](https://github.com/libp2p/specs/blob/master/identify/README.md)
to exchange identity metadata, addresses, and supported protocols. Although `lp2p` exposes a libp2p peer ID
through the `p2p.ID` type for compatibility with shared interfaces, Comet and libp2p IDs are not interchangeable.

Libp2p also provides several discovery mechanisms, but none directly replaces Comet PEX:

- [mDNS](https://github.com/libp2p/specs/blob/master/discovery/mdns.md) is
  decentralized but limited to the local network.
- [Rendezvous](https://github.com/libp2p/specs/blob/master/rendezvous/README.md)
  depends on dedicated rendezvous servers.
- [Kademlia DHT](https://github.com/libp2p/specs/blob/master/kad-dht/README.md)
  is a general-purpose distributed routing system rather than an
  operator-managed peer book.
- [GossipSub](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md)
  can exchange peers within a pubsub mesh, but Comet's libp2p reactors do not
  currently use GossipSub.

The libp2p implementation already has a [peerstore](https://pkg.go.dev/github.com/libp2p/go-libp2p/core/peerstore) 
and connection-management primitives keyed by `peer.ID` and multiaddress. 
**The new PEX address book should build on those primitives**, adding only Comet-specific policy such as privacy,
persistence, diversity, retry history, and bans. It should not reimplement the classic `NetAddress`-based stack.

## Proposed implementation

This section covers configuration, the address book format, the PEX protocol and reactor, the
interface changes this requires, peer tracking, and pex/addressbook rules:

### Config

Add PEX options to config.toml:

```toml
[p2p.libp2p.pex]
enabled = true
addr_book_file = "path/to/addressbook.json"
```

The following options stay in the top-level `[p2p]`:

- `max_num_inbound_peers`
- `max_num_outbound_peers`

> Note: there should be a WARN log if [p2p.pex] is enabled (old one) at the same time as lib-p2p PEX.
> The former should be ignored.

**Private peers**

Bootstrap peers under `[p2p.libp2p.bootstrap_peers]` support `private = true` which is a stub.
This feature should be implemented. In this case the node should NOT gossip this peer to the rest of the network.

### Addressbook

Define the following address book format:

```jsonc
{
    // indicates lib-p2p addressbook.
    // parsing should fail for CometPEX addressbook
    "version": "2",
    "addrs": [
        {
            // lib-p2p identity
            "id": "12D3KooWLetY2eapfEAujHqYBhz6jP9vyw7dZu9YwGpqZ9FHw8dU",
            "host": "1.2.3.4:26656",
            "attempts": 0,
            "last_attempt": "2026-01-01T17:00:00Z",
            "last_success": "2026-01-01T17:00:00Z",
            "last_ban_time": "2026-01-01T17:00:00Z",
            // who told us about this peer?
            "origin": {
                "id": "12D3KooWLetY2eapfEAujHqYBhz6jP9vyw7dZu9YwGpqZ9FHw8dU",
                "host": "1.2.3.4:26656"
            }
        }
    ]
}
```

- Note that bootstrap peers or private peers don't go into the addressbook.
- Addressbook should be flushed onto the disk every 2 minutes (similarly to current PEX).

### Protocol and Reactor

PEX would live under `/p2p/cometbft/1.0.0/channel/0x00` Lib-p2p protocol ID similarly to
existing `PexChannel = byte(0x00)`.

For simplicity, Comet's `pex.proto` and `types.proto` can be reused.

```proto
message PexRequest {}

message PexAddrs {
  repeated NetAddress addrs = 1 [(gogoproto.nullable) = false];
}

message Message {
  oneof sum {
    PexRequest pex_request = 1;
    PexAddrs pex_addrs = 2;
  }
}

message NetAddress {
  string id = 1 [(gogoproto.customname) = "ID"];
  string ip = 2 [(gogoproto.customname) = "IP"];
  uint32 port = 3;
}
```

Note that ID would be a libp2p identity that should be validated as it differs
from original Comet's peer ID.

Reactor and addressbook implementation would live under `lp2p/pex`:

```
lp2p/
└── pex/
    ├── reactor.go
    ├── addressbook.go
    └── ...
```

The reactor and addressbook should be wired similarly to other reactors in the codebase
(under the `"PEX"` name)

### Addressbook and Peer Exchange Rules

The current Comet PEX address book uses Bitcoin-inspired bucketing: peers are grouped by network and split into 
"new" and "old" buckets to limit address-book poisoning and preserve network diversity. The libp2p implementation will 
not reproduce this structure because libp2p models peers as `peer.ID` values with one or more multi-addresses
and already stores them in its peerstore. Instead, `lp2p/pex` will apply Comet-specific admission, diversity,
retry, and eviction policies on top of the peerstore.

#### Loop

- One discovery loop every 30s.
- Each tick: `need = max_outbound - (outbound + dialing)`. If `need <= 0`, skip.
- Pick random book entries to dial
- Build a per-tick plan of at most `need` unique targets to dial or ask. Dedup by `peer.ID`. 
  Skip already connected / dialing identities.

#### Peers

- Track inbound and outbound separately.
- Persistent peers are immortal: never evicted, never banned, always retried. Eviction of others MUST NOT touch them.
- Non-persistent peers get a reconnect cooldown after disconnect.
- PEX policy and the libp2p resource manager both apply. Hitting either limit is enough to refuse.

#### Dial

- One host IP per `peer.ID`. Extra advertised IPs for the same ID are ignored.
- Dial timeout `3s`. Failed dials use exponential backoff plus jitter.
- Validate `peer.ID` and address before store, dial, or gossip.

#### PEX stream

- At most one outstanding request per peer.
- Accept at most one inbound PEX request per peer per `5s`.
- Cap concurrent PEX streams globally. This should be a low-rate channel;
- Response cap: at most `128` addresses, and never more than the requester's remaining need.
- Encoded message hard cap: size of a `PexAddrs` with `128` max-size valid records. Larger frames are protocol abuse.
- Ban of a peer/connection also bans its IP for `X` hours (libp2p gater / resource manager). Persistent peers are exempt.

#### Gossip

- Never gossip: self, private, bootstrap, banned, expired, unsupported transports, non-dialable inbound-only sources.
- A sender MAY advertise a wrong IP for an ID; signed advertisements are supported by lib-p2p (`CertifiedAddrBook`)
  and can be added later.

### Necessary interface changes

Current `Switcher` and `Peer` interfaces carry too many comet-p2p-specific methods which should be internalized and dropped from public usage

https://github.com/cometbft/cometbft/blob/3ce165de1c764a825103dae7ca5a3dd15aeab695/p2p/switcher.go#L29-L49

```go
type PeerManager interface {
    // only these 3 methods are useful for generic networking implementation
    // other methods can be internalized to p2p/pex
	Peers() IPeerSet
	StopPeerForError(peer Peer, reason any)
	MarkPeerAsGood(peer Peer)

    // ...
}
```

https://github.com/cometbft/cometbft/blob/3ce165de1c764a825103dae7ca5a3dd15aeab695/p2p/peer.go#L23-L48

```go
type Peer interface {
    // these methods are also comet-p2p specific
    // and can be internalized for comet-p2p only
	Status() cmtconn.ConnectionStatus
	SetRemovalFailed()
	GetRemovalFailed() bool
	FlushStop()

    // ...
}
```

rpc/core/net's `Unsafe*` functionality can be copied to `unsafe_rpc_v2` that exposes similar methods for the lib-p2p implementation.

#### Inbound vs Outbound peers

The current Lib-P2P implementation doesn't distinguish inbound/outbound peer tracking. 
Libp2p connections are bidirectional, but we need to keep track of that. Current `lp2p.(Peer{}).IsOutbound bool` 
returns true, which is a stub.

Outbound peers are those we connected to explicitly; inbound ones are those that asked our node for the connection.

## Feature Completeness

- Implementation: config, addressbook v2, `lp2p/pex` reactor, etc.
- Unit and E2E tests
- Metrics reflecting addressbook
- Update documentation

## Discussion

@swift1337: "Addressbook and Peer Exchange Rules" section should pass a thorough review to reach the alignment
on desirable logic, keeping security and chain liveness in mind. Proposed implementation doesn't use Comet's bucketing
logic. We can also explore porting it as is.

### References

- [CometBFT PEX protocol](https://github.com/cometbft/cometbft/blob/main/spec/p2p/implementation/pex-protocol.md)
- [CometBFT address book](https://github.com/cometbft/cometbft/blob/main/spec/p2p/implementation/addressbook.md)
- [ADR 073: Adopt LibP2P](https://github.com/cometbft/cometbft/blob/main/docs/references/architecture/tendermint-core/adr-073-libp2p.md)
- [Tendermint PEX messages (historical)](https://github.com/tendermint/spec/blob/master/spec/p2p/messages/pex.md)
- [libp2p mDNS](https://github.com/libp2p/specs/blob/master/discovery/mdns.md)
- [go-libp2p mDNS](https://github.com/libp2p/go-libp2p/tree/master/p2p/discovery/mdns)
- [libp2p rendezvous](https://github.com/libp2p/specs/blob/master/rendezvous/README.md)
- [libp2p Kademlia DHT](https://github.com/libp2p/specs/blob/master/kad-dht/README.md)
