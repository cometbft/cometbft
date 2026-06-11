# crypto

crypto is the cryptographic package adapted for CometBFT's uses

## Importing it

To get the interfaces,
`import "github.com/cometbft/cometbft/crypto"`

For any specific algorithm, use its specific module e.g.
`import "github.com/cometbft/cometbft/crypto/ed25519"`

## Validator key types

CometBFT includes validator key implementations for Ed25519, secp256k1,
BLS12-381, ML-DSA-65, and `secp256k1eth`.

The `secp256k1eth` key type stores 33-byte compressed secp256k1 public keys,
derives 20-byte Ethereum addresses as
`Keccak256(uncompressedPubKey[1:])[12:]`, and signs with exact 65-byte
recoverable `[R || S || V]` signatures over legacy Keccak-256 message hashes.
CometBFT uses canonical `V` values `0` or `1`.

## Binary encoding

For Binary encoding, please refer to the [CometBFT encoding specification](https://github.com/cometbft/cometbft/blob/v0.38.x/spec/core/encoding.md).

## JSON Encoding

JSON encoding is done using CometBFT's internal json encoder. For more information on JSON encoding, please refer to [CometBFT JSON encoding](https://github.com/cometbft/cometbft/blob/v0.38.x/libs/json/doc.go)

```go
Example JSON encodings:

ed25519.PrivKey     - {"type":"tendermint/PrivKeyEd25519","value":"EVkqJO/jIXp3rkASXfh9YnyToYXRXhBr6g9cQVxPFnQBP/5povV4HTjvsy530kybxKHwEi85iU8YL0qQhSYVoQ=="}
ed25519.PubKey      - {"type":"tendermint/PubKeyEd25519","value":"AT/+aaL1eB0477Mud9JMm8Sh8BIvOYlPGC9KkIUmFaE="}
crypto.PrivKeySecp256k1   - {"type":"tendermint/PrivKeySecp256k1","value":"zx4Pnh67N+g2V+5vZbQzEyRerX9c4ccNZOVzM9RvJ0Y="}
crypto.PubKeySecp256k1    - {"type":"tendermint/PubKeySecp256k1","value":"A8lPKJXcNl5VHt1FK8a244K9EJuS4WX1hFBnwisi0IJx"}
```
