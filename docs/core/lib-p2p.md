# Experimental Lib-P2P Support

CometBFT includes an experimental networking layer based on [go-libp2p](https://libp2p.io/).
It adds a new transport and connection-management layer for peer-to-peer communication,
while keeping the reactor-facing CometBFT API unchanged.

Actors still use the same core p2p concepts (`Switch`, `Peer`, `PeerSet`, `Reactor`, envelopes).
The transport implementation under those abstractions is `lib-p2p` instead of `comet-p2p`.

`lib-p2p` is a widely used networking stack with production-ready peer-to-peer features,
and implementations across many languages and transport protocols (TCP, QUIC, WebSockets, and more).

This feature is implemented alongside `combined_mode`, which blends blocksync with consensus reactors.
That improves network liveliness by letting slower nodes continue ingesting blocks while consensus is running.

You can refer to the implementation in the CometBFT codebase here:

- [`lp2p`](https://github.com/cometbft/cometbft/tree/main/lp2p)
- [`internal/autopool`](https://github.com/cometbft/cometbft/tree/main/internal/autopool)

### Performance and Liveliness

In high-load conditions, legacy `comet-p2p` can become a networking bottleneck for modern blockchain workloads:

- It's more prone to congestion under concurrent message pressure,
- Stream/message handling is less effective at scaling with traffic spikes,
- This limits end-to-end throughput when the rest of the stack is optimized.

The `lib-p2p` integration addresses this with native stream-oriented transport, concurrent receive pipelines,
and autoscaled worker pools per reactor, which helps reduce queue pressure and improve message flow under load.
Beyond raw throughput, this also improves network liveliness by making peer communication and block propagation
more resilient under sustained congestion and sudden load spikes.

In our benchmarks, together with additional performance improvements across the stack, we reached over 2000 TPS,
and `lib-p2p` has been one of the key unblockers enabling that result.

## Differences in transport and peer identification

`lib-p2p` uses its own [peer ID format](https://github.com/libp2p/specs/blob/master/peer-ids/peer-ids.md),
which is different from `comet-p2p`. The two formats are not compatible.

```
# comet-p2p peer ID format
539ffc12a0ac78970dab31ae8cdcfdd4285b2162@10.186.73.3:26656

# lib-p2p peer ID format
{ host = "10.186.73.3:26656", id = "12D3KooWRuTppVZGE7qhanfsHfmzkWUZnnRbTxbgWYvKibij9niy" }
```

`lib-p2p` uses QUIC instead of TCP in this integration, so you need UDP open in your firewall.

## Configuration

Configure `lib-p2p` in the `p2p.libp2p` section of `config.toml`.
All other p2p settings are ignored except `external_address` and `laddr`.

By default, the node listens on UDP port `26656`.

```toml
[p2p.libp2p]

# Enabled set true to use go-libp2p for networking instead of CometBFT's p2p.
enabled = true

# Disables resource manager.
# Warning! This might consume all of the system's resources.
disable_resource_manager = false

# Bootstrap peers to connect to
# format: { host, id, private (opt), persistent (opt), unconditional (opt) }
# dns resolution is also supported  (e.g. "example.com:26656")
bootstrap_peers = [
  { host = "10.186.73.3:26656", id = "12D3KooWRuTppVZGE7qhanfsHfmzkWUZnnRbTxbgWYvKibij9niy", persistent = true },
  { host = "10.186.73.5:26656", id = "12D3KooWHjC8SJFVpAvY3qpM5PPeXSpQZvLxcZb7Tjr1kHMLEFtS", persistent = true },
  { host = "10.186.73.6:26656", id = "12D3KooWJFbLcqdPpNP7E1EXC4tDiPtxEGDM6K7RXpDsdWmVnjSu", persistent = true },
]
```

Each host, similar to `comet-p2p`, supports the following options:
- `persistent` - ensures peer is always (re)connected (even after removal)
- `unconditional` - not affected by the max number of peers limit (see notes on limitations)
- `private` - the peer will not be gossiped to other peers (see notes on limitations)

To validate successful configuration, check the logs for the following message:

```text
Using go-libp2p transport!
```

Using `lib-p2p` with `combined_mode` enabled is recommended:

```toml
[blocksync]

version = "v0"

# Experimental Combined mode (bool):
#
# Run both BLOCKSYNC and CONSENSUS for improved liveness, connectivity, and performance.
combined_mode = true
```

## Implementation details

From the CometBFT actor-model perspective, the API stays the same:
`Reactor`, `Peer`, `PeerSet`, `Switch`, and envelope flow remain compatible,
so existing reactors can run without protocol-level rewrites.

At the connection layer, lib-p2p replaces CometBFT secret connection with lib-p2p native
identity and [secure handshake mechanisms](https://github.com/libp2p/specs/blob/master/tls/tls.md). 
This means peer session establishment, encryption negotiation, and remote identification are handled 
by lib-p2p transport stack.

CometBFT channel traffic is mapped to lib-p2p protocol handlers:

- Each CometBFT channel is exposed under a lib-p2p `protocol.ID` namespace (e.g., `/p2p/cometbft/1.0.0/...`).
- Messages are exchanged over lib-p2p streams bound to those protocol handlers.
- Inbound handling is concurrent, with a priority FIFO queue and worker pool to process
  reactor traffic in parallel under load.

The worker pool is autoscaled by `autopool`:

- It tracks per-message processing durations and computes decisions from throughput EWMA,
  queue pressure, and latency percentile.
- High throughput growth or queue pressure scales workers up; high P90 latency above the
  configured threshold triggers shrink to avoid overload.
- Default limits are 4-32 workers per reactor; mempool uses a wider range (8-512) to
  absorb bursty transaction traffic.
- Priority ordering is preserved before dispatch (`Receive()` pushes by priority, then workers consume in parallel).

## Combined Mode

`combined_mode` runs blocksync and consensus together, instead of treating them
as separate phases. This improves catch-up liveness because a slower node can keep ingesting
verified blocks while consensus is already active.

If enabled, blocksync continuously fetches candidate blocks from peers and
passes them into consensus through an ingestion path:

1. Blocksync takes two consecutive blocks from its pool (fetched from other peers)
2. It builds an ingest candidate for the first block, using commit data from
   the next block.
3. It verifies the candidate against the latest state (light client verification).
4. Then it calls consensus, which applies the block as the next height.

Important behavior in practice:

- Blocksync loads the latest state before verification, and both reactors effectively track the same chain progression.
- If consensus already included a block concurrently, blocksync skips it.
- Ingestion enforces strict invariants and keeps the state machine consistent.


# Comparison and Limitations

> **Important:** The current release does **NOT** include peer exchange (PEX).
> Nodes must use explicit bootstrap peers/static topology; automatic peer exchange will be added in the next release.

| Area                 | `comet-p2p`                        | `lib-p2p`                             |
| -------------------- | ---------------------------------- | ------------------------------------- |
| Transport            | TCP                                | QUIC                                  |
| Peer identity        | Comet peer IDs (`<hex>@host:port`) | lib-p2p peer IDs (`12D3Koo...`)       |
| Connection handshake | Comet secret connection            | lib-p2p identity and secure handshake |
| Peer exchange        | PEX + address book flow            | **No PEX in this release**            |


Note, because peer identification formats differ, you cannot run a mixed `comet-p2p` / `lib-p2p` network.
