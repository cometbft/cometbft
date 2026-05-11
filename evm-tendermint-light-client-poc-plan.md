# EVM Tendermint Light Client POC Implementation Status

Date: 2026-05-08

The EVM Tendermint light-client POC has been implemented and then updated to address the semantic review findings.

## Implemented

- Deterministic Go fixtures for:
  - 10 validators / 7 signers
  - 10-validator same-height and changed-validator-set misbehaviour
  - 50 validators / 34 signers
  - 100 validators / 67 signers
  - 175 validators / 117 signers
  - mixed-power 50-validator case
  - changed-validator-set update and misbehaviour cases
- `uint64` voting power throughout Solidity hashing and threshold accounting.
- EVM-native validator-set hashing with CometBFT/RFC6962 tree shape.
- Prebuilt validator-leaf hashing component benchmark.
- Compact commit verification without unused prebuilt vote bytes in calldata.
- Prebuilt vote sign-byte verification as a component-only benchmark.
- Adjacent update verification and consensus-state storage.
- Non-adjacent update verification using one signed commit, first verified positionally against the untrusted validator set at `>2/3`, then counted against the trusted validator set at trust level.
- Stored validator-set update path.
- Same-height, BFT-time-violation, and changed-validator-set misbehaviour verification.
- Duplicate consensus-state update rejection.
- Duplicate validator-address rejection in update and misbehaviour paths.
- Full compact-commit signature consumption before threshold result.
- Pruning benchmarks.
- Simple SHA256 membership proof baseline.
- `signerBitmap` validator-index limit guarded at 256 validators.
- Canonical vote-byte reconstruction, reconstruction+hash, and reconstruction+secp256k1 component benchmarks.
- BLS-A/B/C/D benchmark models, including stored BLS validator-set registration, shared-precommit assumptions for BLS-A/B/C, and measured multi-message `N + 1` pairing-check rows for current CometBFT-style per-signer vote bytes. BLS-D is a synthetic precompile-cost lower bound; production multi-message calldata is not modeled.
- Raw `cosmos.ics23.v1.ExistenceProof` IAVL known-field decode and verify component benchmarks.
- Machine-readable `benchmarks.json` manifest with toolchain, git status, fixture hashes, execution gas, calldata rows, benchmark assumptions, and generated Markdown tables.
- Derived EVM budget references for docs/explainers: ed25519 quorum crosses Ethereum's `16,777,216` gas single-transaction cap at `N = 27`; the `N = 50` ed25519 quorum costs `30,814,574` gas (`1.84x` the cap); compact `secp256k1eth` at `N = 50` is `295,855` gas.

## Primary Files

- `evm-light-client-gas-bench/src/EvmLightClientBench.sol`
- `evm-light-client-gas-bench/test/EvmLightClientBench.t.sol`
- `evm-light-client-gas-bench/script/gen_fixture.go`
- `evm-light-client-gas-bench/test/fixtures/*.json`
- `evm-light-client-gas-bench/README.md`
- `evm-light-client-gas-benchmark-report.md`

## Verification

```sh
cd evm-light-client-gas-bench
go run -tags bls12381 ./script
forge test --summary
forge test --gas-report --summary
forge test --match-test testCalldataGasEstimates -vvvv
forge fmt --check
```

Latest result:

```text
103 passed; 0 failed
forge fmt --check passes
```

## Remaining Non-Goals

These are intentionally not implemented in this phase:

- full production CometBFT protobuf handling beyond the measured `SimpleValidator` reconstruction component
- generic ICS23 support beyond the implemented IAVL existence-proof known-field decoder
- ICS23 non-existence proof verification
- production IBC client APIs
- full current-CometBFT ed25519 commit verification across a signer set; one real pure-Solidity signature verification is measured
- production multi-message BLS calldata; BLS-D rows price only the required EIP-2537 pairing-check shape
- validator sets larger than 256 with the current bitmap representation
- adopting a production BLS same-message precommit protocol redesign, including aggregate commit encoding, signer bitmaps, proof-of-possession, and vote-extension/evidence handling

## Constraints

<anti-slop-guidelines>
  <implementation>
  - Don't consider backwards compatibility or add any fallbacks or migrations (there are no users or deployments to worry about)
  - Don't overengineer solutions, or add _unecessary_ abstraction or indirection where a simpler, cleaner and more clear solution would do
  - Prefer elegant and readable code over comments when the code can speak for itself
    - Use clear naming and structure so intent is obvious — don't paper over confusing code with comments
  - Don't add defensive guards or null checks for states that are unlikely to happen
  - Don't add unnecessary type casts, aliases, or redefinitions — respect the existing source of truth
  - Don't leave placeholder logic, debug leftovers, or deprecated paths behind
  - Keep tests focused on meaningful behavior; skip low-value tests and prefer direct coverage over mocks.
  - Trust internal code and framework guarantees; only test our own code.
  </implementation>
  <design>
  - Avoid shallow abstractions — if an interface is as complex as its implementation, inline it
  - Respect module boundaries — don't reach past a module's public API or dump unrelated code into "common"/shared modules
  - Keep dependency direction clean — domain/core logic shouldn't depend on UI, transport, persistence, or framework details
  - Avoid change amplification — if a small behavior change needs coordinated edits across many files, switches, or registries, the structure is wrong
  - Translate errors at module boundaries so infrastructure errors (DB, HTTP, filesystem) don't leak raw through higher-level APIs
  - Give mutable state a clear owner; don't create shared state that anyone can reach into
  - Avoid temporal coupling — don't require implicit invocation order or objects that are invalid until some init/setup method runs
  </design>
</anti-slop-guidelines>
