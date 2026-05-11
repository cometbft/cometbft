# Generated Benchmark Tables

Generated at `2026-05-07T11:30:45Z`.

## Toolchain

| Field | Value |
|---|---|
| Forge | `forge Version: 1.5.0-stable` |
| Solc | `0.8.30, 0.6.12` |
| Go | `go version go1.25.0 darwin/arm64` |
| Git commit | `62a15e75371982c4cfe42298bcbaf24439bb758d` |
| Git status | `M ../.gitignore;  M ../go.mod;  M ../go.sum; ?? ../.deepsec/; ?? ./; ?? ../evm-light-client-gas-benchmark-report.md; ?? ../evm-tendermint-light-client-poc-plan.md; ?? ../signature-light-client-evm-report.md` |
| EVM / optimizer / via IR | `prague / true (200 runs) / true` |

## BLS-D Multi-Message Pairing

| Signers N | Pairings | Measured gas | Formula gas | Overhead |
|---:|---:|---:|---:|---:|
| 10 | 11 | 401618 | 396300 | +5318 |
| 50 | 51 | 1714446 | 1700300 | +14146 |
| 100 | 101 | 3356747 | 3330300 | +26447 |
| 175 | 176 | 5822835 | 5775300 | +47535 |

Formula: `37_700 + 32_600 * (N + 1)` for EIP-2537 PAIRING_CHECK with `N` message/signature pairs plus the aggregate-signature pair.

These rows are synthetic precompile-cost lower bounds for the multi-message shape. The generated calldata rows named `bls-d-synthetic-helper-*` encode only the benchmark helper's `signerCount` and must not be used as production multi-message BLS calldata.

## ICS23 Happy-Path Execution Gas

| Function | Min | Avg | Median | Max | Calls |
|---|---:|---:|---:|---:|---:|
| `decodeIcs23IavlExistenceProof` | 64303 | 64303 | 64303 | 64303 | 1 |
| `verifyIcs23IavlExistenceProof` | 75215 | 75215 | 75215 | 75215 | 1 |

## Selected Execution Gas

Aggregate Forge rows include all matching calls in the full test suite. For raw ICS23 happy-path-only decode/verify gas, use `benchmarks.json.ics23_happy_path_gas`.

| Function | Min | Avg | Median | Max | Calls |
|---|---:|---:|---:|---:|---:|
| `benchBlsAggregateApprox` | 162127 | 325114 | 304519 | 529293 | 4 |
| `benchBlsMultiMessagePairing` | 401618 | 2823911 | 2535596 | 5822835 | 4 |
| `decodeIcs23IavlExistenceProof` | 864 | 26620 | 22601 | 64303 | 5 |
| `hashCanonicalVoteSignBytes` | 11642 | 11642 | 11642 | 11642 | 1 |
| `reconstructCanonicalVoteSignBytes` | 12140 | 15026 | 15145 | 15145 | 34 |
| `storeBlsValidatorSet` | 513660 | 5416357 | 6396897 | 6396897 | 6 |
| `verifyBlsAggregate` | 106536 | 118622 | 106536 | 203230 | 8 |
| `verifyBlsAggregateCalldataValidatorSet` | 1681 | 97348 | 97348 | 193016 | 2 |
| `verifyBlsAggregateStoredValidatorSet` | 5334 | 394601 | 588235 | 590235 | 3 |
| `verifyCanonicalVoteSecp256k1` | 15760 | 15760 | 15760 | 15760 | 1 |
| `verifyCommitCompact` | 46526 | 287837 | 210755 | 716478 | 7 |
| `verifyCommitPrebuiltVoteBytes` | 190571 | 190571 | 190571 | 190571 | 1 |
| `verifyIavlExistenceProof` | 11883 | 17436 | 21138 | 21138 | 5 |
| `verifyIcs23IavlExistenceProof(bytes32,bytes,bytes,bytes)` | 65704 | 69028 | 66166 | 75215 | 3 |
| `verifyMembershipProof` | 4739 | 4739 | 4739 | 4739 | 1 |

## Calldata Rows

EIP-7623 floor gas is token based: zero bytes count as 1 token, non-zero bytes count as 4 tokens, and the floor is 10 gas per token.

| Row | Bytes | Zero | Nonzero | Standard gas | EIP-7623 floor | Blob bytes |
|---|---:|---:|---:|---:|---:|---:|
| `commit-compact-10` | 2500 | 1682 | 818 | 19816 | 49540 | 2592 |
| `commit-compact-50` | 10244 | 6567 | 3677 | 85100 | 212750 | 10592 |
| `commit-compact-100` | 19780 | 12573 | 7207 | 165604 | 414010 | 20448 |
| `commit-compact-175` | 34180 | 21652 | 12528 | 287056 | 717640 | 35296 |
| `commit-prebuilt-vote-bytes-50` | 15684 | 9053 | 6631 | 142308 | 355770 | 16192 |
| `adjacent-update-calldata-10` | 3652 | 2410 | 1242 | 29512 | 73780 | 3776 |
| `adjacent-update-calldata-50` | 13956 | 9019 | 4937 | 115068 | 287670 | 14432 |
| `adjacent-update-mixed-power-50` | 13956 | 9013 | 4943 | 115140 | 287850 | 14432 |
| `non-adjacent-update-changed-valset-50` | 13956 | 9016 | 4940 | 115104 | 287760 | 14432 |
| `misbehaviour-10` | 7332 | 4859 | 2473 | 59004 | 147510 | 7584 |
| `changed-valset-misbehaviour-10` | 7332 | 4858 | 2474 | 59016 | 147540 | 7584 |
| `changed-valset-misbehaviour-50` | 27940 | 18072 | 9868 | 230176 | 575440 | 28864 |
| `adjacent-update-stored-valset-50` | 7332 | 4494 | 2838 | 63384 | 158460 | 7584 |
| `membership-proof-baseline` | 548 | 382 | 166 | 4184 | 10460 | 576 |
| `iavl-existence-proof-depth8` | 2020 | 1597 | 423 | 13156 | 32890 | 2112 |
| `iavl-existence-proof-depth16` | 3812 | 3037 | 775 | 24548 | 61370 | 3936 |
| `ics23-iavl-existence-decode-depth8` | 484 | 73 | 411 | 6868 | 17170 | 512 |
| `ics23-iavl-existence-verify-depth8` | 708 | 224 | 484 | 8640 | 21600 | 736 |
| `ics23-iavl-existence-verify-depth16` | 1092 | 238 | 854 | 14616 | 36540 | 1152 |
| `canonical-vote-reconstruct` | 292 | 195 | 97 | 2332 | 5830 | 320 |
| `canonical-vote-reconstruct-hash` | 292 | 195 | 97 | 2332 | 5830 | 320 |
| `canonical-vote-reconstruct-secp256k1` | 484 | 298 | 186 | 4168 | 10420 | 512 |
| `canonical-validator-set-ed25519-50` | 3332 | 1679 | 1653 | 33164 | 82910 | 3456 |
| `canonical-validator-set-secp256k1-50` | 8132 | 6289 | 1843 | 54644 | 136610 | 8416 |
| `bls-a-aggregate-verify-supplied-aggregate-pubkey` | 836 | 348 | 488 | 9200 | 23000 | 864 |
| `bls-b-store-validator-set-50` | 19428 | 9482 | 9946 | 197064 | 492660 | 20064 |
| `bls-b-stored-validator-set-bitmap-50` | 740 | 311 | 429 | 8108 | 20270 | 768 |
| `bls-c-calldata-validator-set-50` | 12036 | 6660 | 5376 | 112656 | 281640 | 12448 |
| `bls-d-synthetic-helper-pairing-10` | 36 | 31 | 5 | 204 | 510 | 64 |
| `bls-d-synthetic-helper-pairing-50` | 36 | 31 | 5 | 204 | 510 | 64 |
| `bls-d-synthetic-helper-pairing-100` | 36 | 31 | 5 | 204 | 510 | 64 |
| `bls-d-synthetic-helper-pairing-175` | 36 | 31 | 5 | 204 | 510 | 64 |
