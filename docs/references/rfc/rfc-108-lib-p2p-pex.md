# RFC 108: Lib-P2P Peer Exchange

## Changelog

- 2026-09-18: First draft (@swift1337)

## Abstract

Comet [v0.39.0](https://github.com/cometbft/cometbft/releases/tag/v0.39.0) introduces a new networking layer built on 
top of Lib-P2P. Initial release had dynamic peer exchange (PEX) out of scope, supporting only static list of
bootstrapped peers. This RFC aims to close the gap and bring feature parity with original CometBFT pex by introducing 
similar behavior of discovering and connecting to peers.

## Background

`TODO`

1. Two networking modes
2. Link to PEX rfc
3. Brief logic on PEX (config, addressbook, *buckets*, ...)
4. Peer types (regular, seed, private, ...)
5. Why we can't reuse comet PEX (lib-p2p identities, addressbook, etc...)

> Any context or orientation needed for a reader to understand and participate
> in the substance of the Discussion. If necessary, this section may include
> links to other documentation or sources rather than restating existing
> material, but should provide enough detail that the reader can tell what they
> need to read to be up-to-date.

### References

`TODO`

> Links to external materials needed to follow the discussion may be added here.
>
> In addition, if the discussion in a request for comments leads to any design
> decisions, it may be helpful to add links to the ADR documents here after the
> discussion has settled.

## Discussion

> This section contains the core of the discussion.
>
> There is no fixed format for this section, but ideally changes to this
> section should be updated before merging to reflect any discussion that took
> place on the PR that made those changes.


## Proposed changes

`TODO`

1. Mention lib-p2p support mdns, Kademlia, randevouz, ...

### Config

1. update config.toml
   1. options for lp2p.pex (durations, etc... max peers)
2. define addressbook.json
3. define changes for simplifying address manager interface
   1. TODO list
4. define pex protocol and reactor
5. define protobufs
6. define rules for peer logic, DDOS protections, ensure parity with comet-p2p pex
