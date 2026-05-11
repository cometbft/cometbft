# EVM Light Client Gas Bench

This Foundry project benchmarks EVM Tendermint/CometBFT light-client design choices with explicit models, generated fixtures, Go/reference parity, execution gas, and exact ABI calldata accounting.

The canonical report is [`../evm-light-client-gas-benchmark-report.md`](../evm-light-client-gas-benchmark-report.md). The generated source of truth is [`benchmarks.json`](benchmarks.json), with Markdown tables in [`generated/benchmark-tables.md`](generated/benchmark-tables.md).

## Reproduce

```sh
go run -tags bls12381 ./script
./scripts/update-report-data.sh
forge fmt --check
```

The definition-of-done command sequence is:

```sh
go run -tags bls12381 ./script
forge test --summary
forge test --gas-report --summary
forge test --match-test testCalldataGasEstimates -vvvv
forge fmt --check
```

Use `-tags bls12381` to generate real BLS fixtures with `blst`: signer bitmap, signer pubkeys, aggregate pubkey, aggregate signature, shared message, supplied `H(m)`, wrong-message fixture, and missing-signer fixture. The generator verifies the BLS fixture before writing it.

## Baseline Locking

`benchmarks.json` records Forge, Solc, Go, git commit/status, EVM version, optimizer settings, `via_ir`, fixture hashes, benchmark models, assumptions, expected outputs, execution rows, and calldata rows.

Compiler pinning is source-level exact pinning:

- `EvmLightClientBench.sol` and tests use `pragma solidity =0.8.30`.
- The vendored ed25519 wrapper uses `pragma solidity =0.6.12`.
- `foundry.toml` uses `auto_detect_solc = true` because both pinned compilers are required in one project.

## Benchmark Models

The main implemented groups are:

- `secp256k1eth` compact commit verification: EVM-native address derivation, compact fixed-layout vote fields, Keccak digests, recoverable signatures, and `ecrecover`. This preserves per-signer vote semantics but is not the exact canonical protobuf vote-byte path.
- Canonical vote bytes: reconstruct-only, reconstruct+hash, reconstruct+secp256k1 verify, and prebuilt canonical vote bytes.
- BLS-A: off-chain aggregate pubkey, off-chain aggregate signature, supplied `H(m)`.
- BLS-B: canonical BLS pubkeys and powers checked against the canonical CometBFT BLS validator-set hash during registration, EIP-2537 pubkeys stored under that hash, signer bitmap in calldata, normal verify row excluding one-time storage.
- BLS-C: BLS validator pubkeys and powers supplied in calldata, aggregate pubkey computed on-chain.
- BLS-D: synthetic multi-message aggregate verification priced as `N + 1` EIP-2537 pairing-check pairs for `N = 10, 50, 100, 175`; the helper calldata row encodes only `signerCount` and is not production multi-message calldata.
- ICS23/IAVL: raw `cosmos.ics23.v1.ExistenceProof` bytes decoded on-chain and verified with the supported IAVL existence-proof known fields.
- ed25519: measured pure-Solidity component using the vendored `chengwenxi/Ed25519` verifier.

BLS-A/B/C assume a single shared precommit message for the committed block: same chain ID, height, round, block ID, and no per-signer timestamp. That is a protocol change from per-signer timestamped CometBFT vote sign bytes. A production BLS aggregate-commit path also needs signer bitmaps and rogue-key mitigation such as proof-of-possession. BLS-D prices the non-changed multi-message pairing-check shape and is intentionally separate; do not combine it with aggregate calldata rows.

## Budget References

Use Ethereum's single-transaction gas cap as the primary feasibility boundary for one light-client update transaction. EIP-7825 caps a transaction at `16,777,216` gas. In the current benchmark data:

- Pure-Solidity ed25519 costs `906,311` gas per signature.
- ed25519 quorum verification crosses the single-transaction cap at `N = 27`.
- At `N = 50`, the ed25519 quorum is `30,814,574` gas, or `1.84x` the cap.
- At `N = 50`, compact `secp256k1eth` is `295,855` gas including standard calldata.

A 60M block gas limit is useful capacity context, but it is not the acceptance boundary for one transaction.

## ICS23 Scope

Implemented claim: wire-format IAVL existence-proof known-field decode and verify.

Not claimed: generic ICS23, non-existence proofs, or outer `CommitmentProof` wrapper decoding. The fixture generator emits the wrapper when cheap, but the decision row defaults to raw `ExistenceProof` bytes because that is the cost needed for proof verification.

Unknown protobuf fields are skipped under proof-size and proof-depth caps. Unsupported known hash ops, length ops, proof specs, malformed prefixes, wrong key/value, tampering, and oversized proofs reject explicitly.

## Generated Data

Run:

```sh
./scripts/update-report-data.sh
```

The script runs Forge, captures gas reports and calldata events, benchmarks BLS-D per signer count, and writes:

- `benchmarks.json`
- `generated/benchmark-tables.md`
- raw command logs in `generated/*.txt`

Use `benchmarks.json` for any combined execution + calldata table. It is the guardrail against combining mismatched models.

EIP-7623 floor values in the calldata table are token based: zero bytes count as 1 token, non-zero bytes count as 4 tokens, and the floor is 10 gas per token.
