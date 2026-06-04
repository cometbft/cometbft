# cometkms

`cometkms` is an external remote signer for CometBFT validators. It lives as a
nested Go module (`github.com/cometbft/cometbft/kms`) inside the CometBFT
repository and is conceptually similar to [tmkms](https://github.com/iqlusioninc/tmkms),
but implemented entirely in Go and building directly on top of the CometBFT
libraries. It dials *out* to one or more validator nodes, authenticates each
connection with either CometBFT's SecretConnection protocol or the libp2p Noise
transport (selected per-validator via the address scheme), and serves Ed25519
consensus signing requests (votes, proposals, and vote extensions) with
mandatory per-chain double-sign protection.

---

## Status

| Scope | Status |
|---|---|
| Ed25519 consensus signing (votes, proposals, vote extensions) | **Supported** |
| `softsign` key backend (file-based, in-memory Ed25519) | **Supported** |
| `pkcs11` key backend (HSM / token, Ed25519) | **Supported** |
| `cometp2p` transport (TCP + SecretConnection) | **Supported** |
| Multi-chain, multi-validator support | **Supported** |
| Double-sign protection (reuses CometBFT FilePV state machine) | **Supported** |
| Dial-out + automatic exponential-backoff reconnect | **Supported** |
| AWS KMS backend | Planned |
| libp2p transport (Noise) | **Supported**  |
| Account / raw-bytes / ECDSA signing | Planned |
| ML-DSA / eth_secp256k1 key types | Planned |

---

## How it works

- **Dial-out model.** `cometkms` dials out to the validator's privval listener
  (`priv_validator_laddr` in `config.toml`) rather than listening itself. This
  removes the need to expose any port on the KMS host.
- **SecretConnection authentication.** Each connection is authenticated and
  encrypted using CometBFT's SecretConnection protocol. `cometkms` uses a
  dedicated Ed25519 *identity key* (distinct from the consensus signing key) to
  authenticate itself to the validator.
- **Request serving.** Once connected, the KMS handles `PubKey`, `SignVote`,
  `SignProposal`, and `Ping` requests using `privval.DefaultValidationRequestHandler`.
- **Double-sign protection.** Signing is delegated to CometBFT's `FilePV` state
  machine, which persists the last-signed height/round/step to a per-chain state
  file and refuses to sign any regression. This survives process restarts.
- **Automatic reconnect.** When a connection drops (validator restart, network
  hiccup), `cometkms` reconnects with capped exponential backoff (200 ms initial,
  10 s ceiling) without any manual intervention.

---

## Build

Requires Go 1.25 or newer. Build from the repository root via the top-level
Makefile (the binary is written to `build/cometkms`):

```sh
make kms-build          # build/cometkms
make kms-install        # install to GOBIN
```

---

## Quick start

### 1. Initialise the home directory

```sh
cometkms init --home ~/.cometkms
```

This creates:

- `~/.cometkms/cometkms.toml` — a stub configuration file.
- `~/.cometkms/identity.json` — a fresh Ed25519 identity key for the
  SecretConnection.

### 2. Edit the config

Open `~/.cometkms/cometkms.toml` and fill in the real values (see the
[example config](#example-config) below).

### 3. Place the consensus key file

Copy or symlink the `priv_validator_key.json` for each chain to the path you
set as `key_file` in `[[providers.softsign]]`.

### 4. Configure the validator

In the validator's `config.toml` enable the remote signer listener:

```toml
priv_validator_laddr = "tcp://0.0.0.0:26659"
```

The address must be reachable from the host running `cometkms`, and it must
match the `addr` you set in `[[validator]]`.

### 5. Start the KMS

```sh
cometkms start -c ~/.cometkms/cometkms.toml --home ~/.cometkms
```

`cometkms` will dial the validator and begin serving signing requests. It logs
each connection and any signing errors to stdout.

---

## Configuration reference

### `[[chain]]`

Declares one blockchain. You need exactly one `[[chain]]` block per chain you
want to sign for.

| Field | Type | Required | Description |
|---|---|---|---|
| `id` | string | yes | The chain-id string (e.g. `cosmoshub-4`). |
| `state_file` | string | no | Path to the double-sign state file. Defaults to `<home>/state/<id>.json`. Relative paths are resolved against `--home`. |

### `[[validator]]`

Declares one outbound connection to a validator node. A single chain can have
multiple `[[validator]]` blocks (e.g. primary + backup nodes).

| Field | Type | Required | Description |
|---|---|---|---|
| `chain_id` | string | yes | Must match a declared `[[chain]].id`. |
| `addr` | string | yes | Address of the validator's privval listener. Use `tcp://host:port` for the standard SecretConnection transport, or `noise://<validator-peer-id>@host:port` for the libp2p Noise transport (see [libp2p Noise transport](#libp2p-noise-transport)). |
| `identity_key` | string | yes | Path to the Ed25519 identity key file used to authenticate the SecretConnection. Relative paths are resolved against `--home`. Use the file generated by `cometkms init`. |
| `reconnect` | bool | no | Whether to reconnect automatically after a dropped connection. Defaults to `true`. |

### `[[providers.softsign]]`

Binds a file-based Ed25519 private key to one or more chains.

| Field | Type | Required | Description |
|---|---|---|---|
| `chain_ids` | list of strings | yes | Chain IDs this key is used to sign for. Each chain may only have one softsign provider. |
| `key_file` | string | yes | Path to the key file. Accepts either a CometBFT `priv_validator_key.json` (typed JSON with a `"priv_key"` field) or a file containing the raw base64-encoded 64-byte Ed25519 private key. |

### `[[providers.pkcs11]]`

Binds an Ed25519 key stored on a PKCS#11 token or HSM to one or more chains. The
private key never leaves the token: signing is performed on-device via `CKM_EDDSA`.

| Field | Type | Required | Description |
|---|---|---|---|
| `chain_ids` | list of strings | yes | Chain IDs this key is used to sign for. Each chain may only have one backend (softsign *or* pkcs11). |
| `module` | string | yes | Path to the PKCS#11 module shared library (e.g. `/usr/lib/softhsm/libsofthsm2.so`). Relative paths are resolved against `--home`. |
| `token_label` | string | one of token_label/slot | `CKA_LABEL` of the token to use. |
| `slot` | integer | one of token_label/slot | Slot number of the token to use. Mutually exclusive with `token_label`. |
| `key_label` | string | at least one of key_label/key_id | `CKA_LABEL` of the key object. |
| `key_id` | string (hex) | at least one of key_label/key_id | Hex-encoded `CKA_ID` of the key object. |
| `pin` | string | exactly one PIN source | User PIN, inline. |
| `pin_env` | string | exactly one PIN source | Name of an environment variable holding the user PIN. Preferred over inline. |
| `pin_file` | string | exactly one PIN source | Path to a file containing the user PIN (trailing whitespace trimmed). Relative paths resolved against `--home`. |
| `algorithm` | string | no | Key algorithm. Defaults to `ed25519` (the only supported value today). |

Provision the key with your HSM tooling before starting `cometkms`; the KMS only
*uses* an existing key, it does not generate or import keys. The key must be an
Ed25519 (`CKK_EC_EDWARDS`) signing key. Example using `pkcs11-tool` with SoftHSM2:

```sh
softhsm2-util --init-token --free --label comet --pin 1234 --so-pin 4321
pkcs11-tool --module /usr/lib/softhsm/libsofthsm2.so --login --pin 1234 \
  --keypairgen --key-type EC:edwards25519 --label validator --id 01
```

Example provider block (PIN supplied via environment, keeping it out of the
config file):

```toml
[[providers.pkcs11]]
chain_ids   = ["cosmoshub-4"]
module      = "/usr/lib/softhsm/libsofthsm2.so"
token_label = "comet"
key_label   = "validator"
key_id      = "01"
pin_env     = "COMETKMS_PIN"
# algorithm defaults to "ed25519"
```

---

## Example config

```toml
# ~/.cometkms/cometkms.toml

[[chain]]
id = "cosmoshub-4"
# state_file defaults to <home>/state/cosmoshub-4.json when omitted

[[validator]]
chain_id    = "cosmoshub-4"
addr        = "tcp://10.0.0.1:26659"
identity_key = "identity.json"   # relative to --home

[[providers.softsign]]
chain_ids = ["cosmoshub-4"]
key_file  = "/secrets/priv_validator_key.json"
```

Multi-chain example:

```toml
[[chain]]
id = "cosmoshub-4"

[[chain]]
id = "osmosis-1"

[[validator]]
chain_id     = "cosmoshub-4"
addr         = "tcp://10.0.0.1:26659"
identity_key = "identity.json"

[[validator]]
chain_id     = "osmosis-1"
addr         = "tcp://10.0.0.2:26659"
identity_key = "identity.json"

[[providers.softsign]]
chain_ids = ["cosmoshub-4"]
key_file  = "/secrets/cosmoshub_priv_validator_key.json"

[[providers.softsign]]
chain_ids = ["osmosis-1"]
key_file  = "/secrets/osmosis_priv_validator_key.json"
```

---

## libp2p Noise transport

The libp2p Noise transport is an alternative to the default SecretConnection
(`tcp://`) channel. Both sides use the same TCP port and listener — the address
scheme (`noise://` vs `tcp://`) is what selects which handshake is performed.
No libp2p switch, host, or gossip network is involved; it is a direct TCP
connection secured by the [Noise_XX handshake](https://noiseprotocol.org/).

The key difference from SecretConnection is **pinned peer IDs on both sides**.
SecretConnection uses an ephemeral, unpinned key on the validator's listener,
which means the KMS cannot verify it is talking to the right validator. With the
Noise transport, each side asserts a stable libp2p identity derived from its
existing keys (the validator's node key, the KMS's `identity.json`), and each
side refuses any connection from an unexpected peer.

### Obtaining the two peer IDs

**KMS peer ID** (give this to the validator operator so they can pin it in
`priv_validator_laddr`):

```sh
cometkms peer-id --home ~/.cometkms
```

This reads `identity.json` from `<home>` and prints the corresponding libp2p
peer ID.

**Validator peer ID** (give this to the KMS operator so they can pin it in
`[[validator]].addr`):

```sh
cometbft show-libp2p-id --home ~/.cometbft
```

This prints the libp2p peer ID derived from the validator's node key. The
equivalent flag form is `cometbft show-node-id --libp2p`.

### KMS configuration

Set `addr` to a `noise://` URI that embeds the validator's peer ID:

```toml
[[validator]]
chain_id     = "cosmoshub-4"
addr         = "noise://12D3KooW...validatorPeerID...@10.0.0.1:26659"
identity_key = "identity.json"   # reused as the KMS's libp2p identity
```

The `identity_key` field serves double duty: it authenticates the
SecretConnection channel when `tcp://` is used, and it provides the KMS's
libp2p identity (peer ID) when `noise://` is used. No additional key file is
needed.

### Validator configuration

Set `priv_validator_laddr` in the validator's `config.toml` to a `noise://`
URI that embeds the KMS peer ID:

```toml
priv_validator_laddr = "noise://12D3KooW...kmsPeerID...@0.0.0.0:26659"
```

The validator uses its **node key** (`node_key.json`) as its Noise identity.
Any incoming connection whose authenticated peer ID does not match the pinned
KMS peer ID is rejected before any signing request is served.

### Mutual pinning requirement

Both sides **must** configure the other's peer ID:

- The KMS encodes the validator's peer ID in the `noise://` address it dials —
  the handshake fails immediately if the remote key does not match.
- The validator encodes the KMS's peer ID in `priv_validator_laddr` — any
  connection from a different peer is dropped.

There is no way to disable peer-ID pinning when using `noise://`; it is
enforced unconditionally.

### Key types

Ed25519 and secp256k1 keys are supported. Consensus keys and node keys in
CometBFT are Ed25519, so no extra setup is needed.


## Security notes

- **softsign is NOT for production custody.** The private key is loaded from
  disk and held in process memory in plaintext for the lifetime of the process.
  Use the `pkcs11` backend (or a future AWS KMS backend) for production
  environments where the key must never leave secure hardware.
- **The identity key is not the consensus signing key.** `identity.json`
  authenticates the SecretConnection channel; it does not sign consensus
  messages and does not need to be protected to the same degree as the
  `priv_validator_key.json`.
- **Double-sign protection is per state file.** The state file records the
  highest height/round/step that has been signed. `cometkms` refuses to sign
  any message that would regress this high-water mark. Never run two `cometkms`
  instances against the same validator with different (or missing) state files —
  doing so removes the double-sign protection.
- **Validator listener exposure.** The validator's `priv_validator_laddr` binds
  a TCP port. Ensure it is not reachable from untrusted networks (use a firewall
  or a private VLAN between the validator and the KMS host).

---

## Testing

From the repository root:

```sh
make kms-test           # go test ./... -count=1
make kms-test-race      # with the race detector
```

---

## Repository layout

```
kms/
├── cmd/
│   └── cometkms/          # Binary entrypoint; CLI subcommands (version, init, start, peer-id)
│       ├── main.go
│       └── main_test.go
└── internal/
    ├── version/           # Version string; overridable at link time via -ldflags
    ├── config/            # TOML config types (config.go) and validation (validate.go)
    ├── identity/          # Identity key load/generate (wraps CometBFT p2p.NodeKey)
    ├── backend/           # backend.Signer interface
    │   ├── softsign/      # File-based Ed25519 backend
    │   └── pkcs11/        # PKCS#11 / HSM Ed25519 backend (+ pkcs11test helpers)
    ├── signer/            # ChainSigner: double-sign protection + PrivValidator impl
    │   ├── chain_signer.go
    │   └── privkey_adapter.go
    ├── manager/           # Manager: supervised dial-out connections with backoff
    │   ├── manager.go
    │   └── dialer.go
    └── app/               # Wiring: Build() assembles Manager from a validated Config
```
