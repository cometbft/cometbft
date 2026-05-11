# CometBFT Signature Verification on EVM Light Clients

This is the high-level narrative report. Exact gas numbers are intentionally deferred to the canonical gas report and generated benchmark data:

- [`evm-light-client-gas-benchmark-report.md`](evm-light-client-gas-benchmark-report.md)
- [`evm-light-client-gas-bench/benchmarks.json`](evm-light-client-gas-bench/benchmarks.json)
- [`evm-light-client-gas-bench/generated/benchmark-tables.md`](evm-light-client-gas-bench/generated/benchmark-tables.md)

Regenerate the data with:

```sh
cd evm-light-client-gas-bench
go run -tags bls12381 ./script
./scripts/update-report-data.sh
forge fmt --check
```

## Core Result

An EVM light client should not try to verify current CometBFT ed25519 commits directly on Ethereum L1. Pure-Solidity ed25519 verification is measured as a real component benchmark, but quorum verification scales linearly and exceeds Ethereum's single-transaction gas cap quickly. The measured ed25519 verifier costs `906,311` gas per signature; a quorum crosses the `16,777,216` gas transaction cap at `N = 27`, and the `N = 50` quorum costs `30,814,574` gas (`1.84x` the cap).

The practical EVM-native baseline is compact `secp256k1eth`: Ethereum-style validator addresses, compact EVM signing bytes, Keccak vote digests, and `ecrecover`. It is the least disruptive design for small and medium validator sets when each signer still signs independently. This is not CometBFT's current `crypto/secp256k1` mode, and it is not the exact canonical protobuf vote-byte path; it requires Ethereum address derivation, an EVM-compatible digest, and recoverable signatures consumable by `ecrecover`. The exact canonical vote-byte reconstruction rows measure the compatibility path separately.

BLS via EIP-2537 is attractive only under a clearly named protocol model. The benchmark now separates:

- BLS-A: supplied aggregate pubkey, supplied aggregate signature, supplied `H(m)`.
- BLS-B: canonical BLS pubkeys and powers checked against the canonical CometBFT BLS validator-set hash during registration, EIP-2537 pubkeys stored under that hash, signer bitmap in calldata.
- BLS-C: validator pubkeys and powers supplied in calldata, aggregate pubkey computed on-chain.
- BLS-D: synthetic multi-message aggregate verification priced as `N + 1` pairing-check pairs, excluding production multi-message calldata.

BLS-A/B/C require one shared precommit message for the committed block: same chain ID, height, round, block ID, and no per-signer timestamp. That differs from CometBFT's per-signer canonical vote bytes with signer-specific timestamps. If that protocol change is not acceptable, BLS-D shows the lower-bound precompile cost of preserving per-signer messages: the EIP-2537 pairing-check formula grows as `37_700 + 32_600 * (N + 1)`, and the benchmark measures N = 10, 50, 100, and 175. Across those anchors, BLS-D is `5.8x` to `7.4x` heavier than compact `secp256k1eth`.

## CometBFT Protocol Reality

BLS-A/B/C are commit-time light-client models, not a drop-in key-type switch. Current CometBFT commits carry one signature per validator slot, current vote gossip exchanges individual votes and bit arrays, and vote extensions remain per-validator data. A production BLS aggregate-commit path would need a new aggregate commit representation with signer bitmaps, same-message precommit signing, and rogue-key mitigation such as proof-of-possession at validator-key registration.

The existing CometBFT BLS implementation is also not production-wired for this model: it is build-tag/cgo gated, default-disabled, and exposes individual signature verification rather than aggregate commit verification.

## Proofs And State

The benchmark now measures raw `cosmos.ics23.v1.ExistenceProof` bytes for IAVL existence proofs. The implemented scope is wire-format IAVL existence-proof known-field decoding plus verification. It skips unknown protobuf fields under proof-size and depth caps, and it does not claim generic ICS23 or non-existence proof support.

Canonical CometBFT vote-byte reconstruction is tested byte-for-byte against generated fixture sign bytes. This lets the reports distinguish:

- compact EVM-native vote bytes,
- prebuilt canonical vote bytes,
- reconstructed canonical vote bytes,
- reconstructed canonical vote bytes plus hash/signature verification.

## Decision Guidance

For validator sets at or below roughly 50 validators, use `secp256k1eth` unless the chain is willing to adopt a BLS same-message precommit protocol and carry the operational complexity of BLS fixtures and validator-set storage.

For validator sets above roughly 100 validators, BLS-B is the most defensible BLS model if the protocol can sign one shared precommit message for the committed block: it avoids trusting relayer-side aggregate pubkey construction while keeping normal verify calldata small, and its registration row checks the canonical validator-set hash before storing EIP-2537 pubkeys. Its one-time storage cost is reported separately.

For chains that must stay byte-canonical with today's CometBFT vote bytes, use the canonical reconstruction rows to price the compatibility cost, and do not cite BLS-A/B/C as drop-in replacements.

For membership proofs, cite the raw IAVL existence-proof rows, not the older pre-parsed IAVL rows, when making production claims about on-chain proof decoding.

When quoting gas budgets, use Ethereum's single-transaction cap as the primary feasibility boundary. A 60M block gas limit is useful context for total block capacity, but it does not make a single over-cap light-client update admissible. EIP-7623 calldata-floor values are token based: zero bytes count as 1 token, non-zero bytes count as 4 tokens, and the floor is 10 gas per token.
