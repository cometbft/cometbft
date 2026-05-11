# PRODUCT.md

## Register

brand

## Users

CometBFT and Cosmos protocol engineers evaluating whether and how to verify a CometBFT light-client commit inside Solidity on Ethereum L1. Fluent in CometBFT internals — canonical-vote bytes, ed25519 signing, validator sets, ICS23/IAVL membership proofs. Not fluent in EVM specifics — precompiles, calldata economics, Ethereum's single-transaction gas cap, EIP-2537 BLS, EIP-7623 calldata-token floor, ecrecover. Reading on a 16-inch laptop in afternoon office light, looking for ammunition for a protocol decision.

## Product Purpose

A single-page explainable that delivers a prescriptive recommendation — compact secp256k1eth — for verifying CometBFT commits in Solidity, with the gas measurements and architectural reasoning behind it. Slide-deck-shaped vertical scroll. Real interactive widgets (N sliders, assumption toggles, gas-price × ETH-price knobs) let the reader poke at the numbers and feel the differences for themselves. Real forge measurements throughout; derived references are computed from the embedded benchmark JSON.

## Brand Personality

Prescriptive. Instrumented. Uncondescending.

## Tone

Confident, doing real explanatory work. The author has measured the thing, sees the answer clearly, and lays it out so the reader can see it too. Never hedges where data is decisive. Never patronizes the reader (a peer protocol engineer). Treats EVM concepts the audience doesn't know with one-line inline definitions, not paragraph-length asides.

## Anti-references

- SaaS marketing pages with hero metric + customer logos + "trusted by" bar
- Crypto-aesthetic landing pages: neon on black, glassmorphism, "the future of [X]"
- Whitepaper-style PDFs ported to web (dense, undesigned, dignified-but-dead)
- Dashboard / calculator UIs that bury the recommendation inside tabs

## Strategic Principles

1. **Recommendation first.** The verdict (compact secp256k1eth) lands on slide one and never gets diluted by "but it depends." Conditionals come after the recommendation, never before.
2. **The chart is the argument.** When data appears, the reader should be able to feel the single-transaction cap, the linearity, the dollar cost. Widgets let them poke; the page reacts.
3. **EVM concepts inline, in one line.** Audience knows CometBFT cold; doesn't know EVM. Tooltip-style definitions, never paragraph-length asides. Define `precompile`, `ecrecover`, `calldata`, `EIP-2537`, `EIP-7623`, and the single-transaction gas cap exactly once each.
4. **Real measurements, visible provenance.** Measured figures trace to a forge run; derived figures are computed from those rows. Methodology gets its own slide, not a footnote. The reader can find the git ref.
5. **Color carries the argument, not decoration.** One warm coral accent on tinted cream. The coral only appears where the argument lands — verdict italics, the recommendation moment, formula tokens, the recommended row in comparisons.

## Content Guardrails

- Treat the `16,777,216` gas single-transaction cap as the primary feasibility boundary for one update transaction; use a 60M block gas limit only as context.
- Describe BLS-D as a synthetic precompile-cost lower bound for multi-message BLS. Do not present helper calldata rows as production calldata.
- Keep compact `secp256k1eth` distinct from exact canonical protobuf vote-byte reconstruction. The recommendation uses compact EVM signing bytes while preserving per-signer vote semantics; canonical reconstruction is a separate compatibility component.
- Describe EIP-7623 as a calldata-token floor: zero bytes count as 1 token, non-zero bytes count as 4 tokens, and the floor is 10 gas per token.
