# EVM Light Client Gas Benchmark Report

Canonical gas data source:

- [`evm-light-client-gas-bench/benchmarks.json`](evm-light-client-gas-bench/benchmarks.json) is the machine-readable manifest and measured-data file.
- [`evm-light-client-gas-bench/generated/benchmark-tables.md`](evm-light-client-gas-bench/generated/benchmark-tables.md) is generated from Forge gas output and calldata events.

Regenerate both:

```sh
cd evm-light-client-gas-bench
go run -tags bls12381 ./script
./scripts/update-report-data.sh
forge fmt --check
```

## Baseline

The current generated run records:

- Forge `1.5.0-stable`
- Solc `0.8.30` for the main benchmark and tests, Solc `0.6.12` for the vendored pure-Solidity ed25519 component
- Go `go1.25.0`
- EVM `prague`, optimizer enabled with 200 runs, `via_ir = true`
- Git commit and dirty status in `benchmarks.json`

`foundry.toml` uses `auto_detect_solc = true` because this benchmark intentionally compiles two source-pinned compiler versions. The pinning is enforced by exact Solidity pragmas (`=0.8.30`, `=0.6.12`) and audited in the generated manifest.

## Current Findings

For a production EVM light-client shape, compact `secp256k1eth` remains the simplest baseline. At 50 validators the compact commit row has median execution gas around `210,755`, with calldata row `commit-compact-50` at `10,244` ABI bytes and `85,100` standard calldata gas, for a combined `295,855` gas. This benchmark model assumes Ethereum-style validator addresses, Keccak vote digests, compact fixed-layout vote fields, and recoverable signatures for `ecrecover`; it is not the same as CometBFT's existing `crypto/secp256k1` behavior and it is not the exact canonical protobuf vote-byte path. The canonical vote-byte reconstruction rows measure that compatibility cost separately.

Ethereum feasibility should be framed against the single-transaction gas cap, not only the block gas limit. EIP-7825 caps one transaction at `16,777,216` gas. The measured pure-Solidity ed25519 verifier costs `906,311` gas per signature; quorum gas crosses the single-transaction cap at `N = 27`, and at `N = 50` the 34-signer quorum costs `30,814,574` gas (`1.84x` the cap). A 60M block gas limit is useful capacity context, but it is not the acceptance boundary for one light-client update transaction.

BLS must be read as four separate protocol models:

| Model | What It Answers | Primary Rows |
|---|---|---|
| BLS-A | Smallest real aggregate verify when aggregate pubkey, aggregate signature, and `H(m)` are supplied | `verifyBlsAggregate`, `bls-a-aggregate-verify-supplied-aggregate-pubkey` |
| BLS-B | Registration checks canonical BLS pubkeys/powers against the canonical CometBFT BLS validator-set hash, then stores EIP-2537 pubkeys/powers under that hash; bitmap in calldata | `storeBlsValidatorSet`, `verifyBlsAggregateStoredValidatorSet`, `bls-b-stored-validator-set-bitmap-50` |
| BLS-C | Validator pubkeys and powers supplied in calldata, aggregate pubkey computed on-chain | `verifyBlsAggregateCalldataValidatorSet`, `bls-c-calldata-validator-set-50` |
| BLS-D | Synthetic per-signer multi-message aggregate verification lower bound under EIP-2537; production calldata not modeled | `benchBlsMultiMessagePairing`, generated BLS-D table |

BLS-A/B/C assume one shared precommit message for the committed block: same chain ID, height, round, block ID, and no per-signer timestamp. That is a protocol change from CometBFT's per-signer canonical vote bytes with signer-specific timestamps. BLS-D prices the alternative as a synthetic precompile-cost bound: `N + 1` pairing-check pairs for `N` signer messages. Production multi-message BLS calldata is intentionally not included. The measured BLS-D rows match the EIP-2537 formula `37_700 + 32_600 * (N + 1)` closely:

| Signers | Pairings | Measured Gas | Formula Gas |
|---:|---:|---:|---:|
| 10 | 11 | 401,618 | 396,300 |
| 50 | 51 | 1,714,446 | 1,700,300 |
| 100 | 101 | 3,356,747 | 3,330,300 |
| 175 | 176 | 5,822,835 | 5,775,300 |

This shows why multi-message BLS is the wrong EVM-light-client shape for CometBFT-style vote bytes: it keeps the linear signer loop and is `5.8x` to `7.4x` heavier than compact `secp256k1eth` across the measured anchors. BLS only becomes compelling after the protocol changes to a same-message aggregate model or introduces a proof system that avoids `N` pairing-check pairs.

The BLS-D calldata rows in the generated calldata table are named `bls-d-synthetic-helper-*`. They encode only the benchmark helper's `signerCount`; they are not production multi-message BLS calldata rows and must not be combined with production totals. When quoting BLS-D, use execution gas as the synthetic lower bound and state that production calldata remains unmodeled.

The BLS-A/B/C rows are commit-time light-client models, not vote-gossip models. Current CometBFT consensus gossip and vote-set handling still operate on individual votes, and vote extensions remain per-validator data. A production BLS aggregate-commit protocol would also need a signer bitmap, aggregate commit encoding, and rogue-key mitigation such as proof-of-possession for BLS validator keys.

ICS23 is now benchmarked as raw `cosmos.ics23.v1.ExistenceProof` bytes. The implemented scope is wire-format IAVL existence-proof known-field decoding plus verification, not generic ICS23 and not non-existence proofs. Unknown protobuf fields are skipped under proof-size and depth caps; known hash, prehash, length, leaf-prefix, and inner-op markers are validated explicitly. The current depth-8 happy path costs `75,215` execution gas plus `8,640` standard calldata gas; depth 16 keeps the same verify execution gas and raises standard calldata gas to `14,616`. The current rows are:

- `decodeIcs23IavlExistenceProof`: decode-only component
- `verifyIcs23IavlExistenceProof(bytes32,bytes,bytes,bytes)`: raw existence proof verification against supplied expected key/value
- `ics23-iavl-existence-verify-depth8`: exact ABI calldata for the raw existence-proof verify call
- `ics23-iavl-existence-verify-depth16`: maximum-depth row currently generated

Negative tests reject wrong key, wrong value, tampered proof, unsupported hash op, unsupported leaf prefix, unsupported length op, and oversized proof before expensive hashing. The aggregate Forge gas rows include these negative calls; `benchmarks.json.ics23_happy_path_gas` contains the isolated happy-path decode and verify component rows.

Canonical vote-byte reconstruction is byte-parity checked against fixture-generated CometBFT canonical sign bytes:

```text
keccak256(solidityReconstructed) == keccak256(fixture.canonical.vote.signBytes[i])
```

The component rows distinguish reconstruct-only, reconstruct+hash, reconstruct+secp256k1 verify, and the prebuilt vote-byte baseline.

## Decision Matrix

For validator sets at or below roughly 50 validators, `secp256k1eth` compact verification is the least disruptive production option. It requires an EVM-friendly signing mode and compact EVM signing bytes, but it preserves per-signer vote semantics and avoids BLS same-message precommit changes, BLS proof-of-possession rollout, aggregate commit encoding, and BLS pubkey storage.

For 100+ validators, BLS-A or BLS-B becomes attractive only if the protocol can sign one shared precommit message for the committed block. BLS-A has the smallest execution and calldata surface but trusts off-chain aggregate-pubkey construction. BLS-B removes that ambiguity by checking the canonical CometBFT BLS validator-set hash during registration and storing EIP-2537 pubkeys and powers under that hash; the normal verify row excludes one-time storage cost, and `storeBlsValidatorSet` is reported separately.

Staying byte-canonical costs extra but is bounded for the component implemented here. Canonical vote-byte reconstruction is now measured directly, and canonical validator-set proto reconstruction remains a separate component benchmark.

Wire-format IAVL existence-proof known-field protobuf decoding is no longer an estimate. The raw proof path is implemented and measured; absence proofs remain out of scope unless production requires them.

All combined totals should be computed from matching execution and calldata rows in `benchmarks.json`. EIP-7623 floor values in the calldata table are token based: zero bytes count as 1 token, non-zero bytes count as 4 tokens, and the floor is 10 gas per token. Do not combine synthetic BLS execution rows with aggregate-only calldata unless the row explicitly models that ABI call.
