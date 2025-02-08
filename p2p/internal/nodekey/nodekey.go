package nodekey

import (
	"bytes"
	"fmt"
	"os"

	mh "github.com/multiformats/go-multihash"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtos "github.com/cometbft/cometbft/internal/os"
	cmtjson "github.com/cometbft/cometbft/libs/json"
)

// ID is a hex-encoded crypto.Address.
type ID string

// IDByteLength is the length of a crypto.Address. Currently only 20.
// TODO: support other length addresses ?
const IDByteLength = 32

// ------------------------------------------------------------------------------
// Persistent peer ID
// TODO: encrypt on disk

// NodeKey is the persistent peer key.
// It contains the nodes private key for authentication.
type NodeKey struct {
	PrivKey crypto.PrivKey `json:"priv_key"` // our priv key
}

// ID returns the peer's canonical ID - the hash of its public key.
func (nk *NodeKey) ID() ID {
	return PubKeyToID(nk.PubKey())
}

// PubKey returns the peer's PubKey.
func (nk *NodeKey) PubKey() crypto.PubKey {
	return nk.PrivKey.PubKey()
}

// PubKeyToID returns the ID corresponding to the given PubKey.
// It's the hex-encoding of the pubKey.Address().
func PubKeyToID(pubKey crypto.PubKey) ID {
	var alg uint64 = mh.SHA2_256
	hash, _ := mh.Sum(pubKey.Bytes(), alg, -1)
	return ID(hash.B58String())
}

func DecodeID(s string) (ID, error) {
	// base58 encoded sha256 or identity multihash
	m, err := mh.FromB58String(s)
	if err != nil {
		return "", fmt.Errorf("failed to parse peer ID: %s", err)
	}
	return ID(m), nil
}

// LoadOrGen attempts to load the NodeKey from the given filePath. If
// the file does not exist, it generates and saves a new NodeKey.
func LoadOrGen(filePath string) (*NodeKey, error) {
	if cmtos.FileExists(filePath) {
		nodeKey, err := Load(filePath)
		if err != nil {
			return nil, err
		}
		return nodeKey, nil
	}

	privKey := ed25519.GenPrivKey()
	nodeKey := &NodeKey{
		PrivKey: privKey,
	}

	if err := nodeKey.SaveAs(filePath); err != nil {
		return nil, err
	}

	return nodeKey, nil
}

// Load loads NodeKey located in filePath.
func Load(filePath string) (*NodeKey, error) {
	jsonBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	nodeKey := new(NodeKey)
	err = cmtjson.Unmarshal(jsonBytes, nodeKey)
	if err != nil {
		return nil, err
	}
	return nodeKey, nil
}

// SaveAs persists the NodeKey to filePath.
func (nk *NodeKey) SaveAs(filePath string) error {
	jsonBytes, err := cmtjson.Marshal(nk)
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, jsonBytes, 0o600)
	if err != nil {
		return err
	}
	return nil
}

// ------------------------------------------------------------------------------

// MakePoWTarget returns the big-endian encoding of 2^(targetBits - difficulty) - 1.
// It can be used as a Proof of Work target.
// NOTE: targetBits must be a multiple of 8 and difficulty must be less than targetBits.
func MakePoWTarget(difficulty, targetBits uint) []byte {
	if targetBits%8 != 0 {
		panic(fmt.Sprintf("targetBits (%d) not a multiple of 8", targetBits))
	}
	if difficulty >= targetBits {
		panic(fmt.Sprintf("difficulty (%d) >= targetBits (%d)", difficulty, targetBits))
	}
	targetBytes := targetBits / 8
	zeroPrefixLen := (int(difficulty) / 8)
	prefix := bytes.Repeat([]byte{0}, zeroPrefixLen)
	mod := (difficulty % 8)
	if mod > 0 {
		nonZeroPrefix := byte(1<<(8-mod) - 1)
		prefix = append(prefix, nonZeroPrefix)
	}
	tailLen := int(targetBytes) - len(prefix)
	return append(prefix, bytes.Repeat([]byte{0xFF}, tailLen)...)
}
