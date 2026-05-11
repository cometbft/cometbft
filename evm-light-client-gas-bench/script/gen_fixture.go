package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cosmos/gogoproto/proto"
	iavl "github.com/cosmos/iavl"
	iavldb "github.com/cosmos/iavl/db"
	ics23 "github.com/cosmos/ics23/go"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secpEcdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/sha3"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtsecp "github.com/cometbft/cometbft/crypto/secp256k1"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
)

const outDir = "test/fixtures"

var secpNMinusOne = new(big.Int).Sub(secp.Params().N, big.NewInt(1))

type fixture struct {
	Client              clientFixture     `json:"client"`
	Trusted             consensusFixture  `json:"trusted"`
	TrustedValidators   validatorFixture  `json:"trustedValidators"`
	UntrustedValidators validatorFixture  `json:"untrustedValidators"`
	Adjacent            signedHeader      `json:"adjacent"`
	NonAdjacent         signedHeader      `json:"nonAdjacent"`
	NonAdjacentConflict signedHeader      `json:"nonAdjacentConflict"`
	Conflict            signedHeader      `json:"conflict"`
	TimeViolationLower  signedHeader      `json:"timeViolationLower"`
	TimeViolationHigher signedHeader      `json:"timeViolationHigher"`
	Membership          membershipFixture `json:"membership"`
	Iavl                iavlFixture       `json:"iavl"`
	IavlMaxDepth        iavlFixture       `json:"iavlMaxDepth"`
	Canonical           canonicalFixture  `json:"canonical"`
}

type clientFixture struct {
	ChainID               string `json:"chainId"`
	RevisionNumber        uint64 `json:"revisionNumber"`
	LatestHeight          uint64 `json:"latestHeight"`
	TrustingPeriod        uint64 `json:"trustingPeriod"`
	MaxClockDrift         uint64 `json:"maxClockDrift"`
	TrustLevelNumerator   uint64 `json:"trustLevelNumerator"`
	TrustLevelDenominator uint64 `json:"trustLevelDenominator"`
	CurrentTimestamp      uint64 `json:"currentTimestamp"`
	ProcessedHeight       uint64 `json:"processedHeight"`
	ProcessedTime         uint64 `json:"processedTime"`
}

type consensusFixture struct {
	Height             uint64 `json:"height"`
	Timestamp          uint64 `json:"timestamp"`
	AppHash            string `json:"appHash"`
	NextValidatorsHash string `json:"nextValidatorsHash"`
	ProcessedHeight    uint64 `json:"processedHeight"`
	ProcessedTime      uint64 `json:"processedTime"`
}

type validatorFixture struct {
	Validators []string `json:"validators"`
	Powers     []uint64 `json:"powers"`
	Leaves     []string `json:"leaves"`
	Hash       string   `json:"hash"`
}

type signedHeader struct {
	Header          headerFixture `json:"header"`
	Commit          commitFixture `json:"commit"`
	WeakCommit      commitFixture `json:"weakCommit"`
	TrustOnlyCommit commitFixture `json:"trustOnlyCommit"`
}

type headerFixture struct {
	RevisionNumber     uint64 `json:"revisionNumber"`
	Height             uint64 `json:"height"`
	Timestamp          uint64 `json:"timestamp"`
	ValidatorsHash     string `json:"validatorsHash"`
	NextValidatorsHash string `json:"nextValidatorsHash"`
	AppHash            string `json:"appHash"`
	HeaderHash         string `json:"headerHash"`
	CommitBlockHash    string `json:"commitBlockHash"`
	Round              uint32 `json:"round"`
	PartSetTotal       uint32 `json:"partSetTotal"`
	PartSetHash        string `json:"partSetHash"`
}

type commitFixture struct {
	Height        uint64   `json:"height"`
	BlockHash     string   `json:"blockHash"`
	Round         uint32   `json:"round"`
	PartSetTotal  uint32   `json:"partSetTotal"`
	PartSetHash   string   `json:"partSetHash"`
	SignerBitmap  *big.Int `json:"signerBitmap"`
	Signatures    []string `json:"signatures"`
	Timestamps    []uint64 `json:"timestamps"`
	VoteSignBytes []string `json:"voteSignBytes"`
	RequiredPower uint64   `json:"requiredPower"`
	TrustPower    uint64   `json:"trustPower"`
}

type membershipFixture struct {
	Root          string   `json:"root"`
	Key           string   `json:"key"`
	Value         string   `json:"value"`
	Siblings      []string `json:"siblings"`
	SiblingOnLeft []bool   `json:"siblingOnLeft"`
}

// iavlFixture is a focused subset of an IAVL existence proof:
// LeafOp(prefix=0x00, hash=sha256, prehash_value=sha256, length=varint),
// followed by InnerOps with prefix bytes and suffix bytes that get prepended/appended
// to the running hash and then hashed with sha256. This matches the standard ICS23
// IAVL existence proof shape closely enough to estimate verification gas without
// pulling in a full ICS23 protobuf decoder.
type iavlFixture struct {
	Root            string   `json:"root"`
	Key             string   `json:"key"`
	Value           string   `json:"value"`
	LeafPrefix      string   `json:"leafPrefix"`
	InnerPrefixes   []string `json:"innerPrefixes"`
	InnerSuffixes   []string `json:"innerSuffixes"`
	ExistenceProof  string   `json:"existenceProof"`
	CommitmentProof string   `json:"commitmentProof"`
}

type canonicalFixture struct {
	Ed25519    canonicalValidatorSet `json:"ed25519"`
	Secp256k1  canonicalValidatorSet `json:"secp256k1"`
	Bls        canonicalBls          `json:"bls"`
	Vote       canonicalVote         `json:"vote"`
	HeaderHash string                `json:"headerHash"`
}

type canonicalValidatorSet struct {
	PubKeys    []string `json:"pubKeys"`
	Powers     []int64  `json:"powers"`
	Leaves     []string `json:"leaves"`
	Hash       string   `json:"hash"`
	Signatures []string `json:"signatures"`
}

type canonicalBls struct {
	Available      bool     `json:"available"`
	PubKeys        []string `json:"pubKeys"`
	PubKeysEip2537 []string `json:"pubKeysEip2537"`
	Powers         []int64  `json:"powers"`
	Leaves         []string `json:"leaves"`
	Hash           string   `json:"hash"`
	// Signatures holds the individual signer-side signatures (compressed 96-byte G2 points,
	// hex-encoded). Useful for cross-checking against blst-side verification.
	Signatures []string `json:"signatures"`
	// AggregateSigEip2537 is the BLS aggregate signature serialized as a 256-byte
	// EIP-2537 G2 point (uncompressed, with 16-byte zero pads per FP element and
	// the (c0, c1) ordering required by EIP-2537). Empty when blsAvailable is false.
	AggregateSigEip2537 string `json:"aggregateSigEip2537"`
	// AggregatePubKeyEip2537 is the BLS aggregate public key serialized as a 128-byte
	// EIP-2537 G1 point. Empty when blsAvailable is false.
	AggregatePubKeyEip2537 string `json:"aggregatePubKeyEip2537"`
	// HashedMessageEip2537 is the result of hash-to-G2 on the shared message, in
	// EIP-2537 G2 byte format (256 bytes). Used for the `e(-G1, aggSig) * e(aggPk, H(m)) == 1`
	// pairing check on-chain.
	HashedMessageEip2537             string   `json:"hashedMessageEip2537"`
	WrongMessageHashedEip2537        string   `json:"wrongMessageHashedEip2537"`
	MissingSignerAggregateSigEip2537 string   `json:"missingSignerAggregateSigEip2537"`
	SignerBitmap                     *big.Int `json:"signerBitmap"`
	Message                          string   `json:"message"`
}

type canonicalVote struct {
	ChainID       string   `json:"chainId"`
	Height        int64    `json:"height"`
	Round         int32    `json:"round"`
	BlockHash     string   `json:"blockHash"`
	PartSetTotal  uint32   `json:"partSetTotal"`
	PartSetHash   string   `json:"partSetHash"`
	Timestamps    []int64  `json:"timestamps"`
	SignBytes     []string `json:"signBytes"`
	EthSignatures []string `json:"ethSignatures"`
}

type scenario struct {
	name           string
	validatorCount int
	signerCount    int
	mixedPower     bool
	changedSet     bool
}

func main() {
	scenarios := []scenario{
		{name: "update_10_equal", validatorCount: 10, signerCount: 7},
		{name: "misbehaviour_10_equal", validatorCount: 10, signerCount: 7, changedSet: true},
		{name: "update_50_equal", validatorCount: 50, signerCount: 34},
		{name: "update_100_equal", validatorCount: 100, signerCount: 67},
		{name: "update_175_equal", validatorCount: 175, signerCount: 117},
		{name: "update_50_mixed", validatorCount: 50, signerCount: 34, mixedPower: true},
		{name: "misbehaviour_50_equal", validatorCount: 50, signerCount: 34, changedSet: true},
	}

	must(os.MkdirAll(outDir, 0o755))
	for _, s := range scenarios {
		fx := makeFixture(s)
		bz, err := json.MarshalIndent(fx, "", "  ")
		must(err)
		path := filepath.Join(outDir, s.name+".json")
		must(os.WriteFile(path, append(bz, '\n'), 0o644))
		fmt.Println(path)
	}
}

func makeFixture(s scenario) fixture {
	chainID := []byte("evm-cometbft-poc-1")
	trustedHeight := uint64(10)
	trustedTimestamp := uint64(1_700_000_000_000_000_000)
	currentTimestamp := uint64(1_700_000_060_000_000_000)

	trusted := makeValidatorSet("trusted", s.validatorCount, s.mixedPower)
	untrusted := trusted
	if s.changedSet {
		untrusted = makeChangedValidatorSet(trusted, s.signerCount, s.mixedPower)
	}

	client := clientFixture{
		ChainID:               "0x" + hex.EncodeToString(chainID),
		RevisionNumber:        1,
		LatestHeight:          trustedHeight,
		TrustingPeriod:        1_209_600_000_000_000,
		MaxClockDrift:         600_000_000_000,
		TrustLevelNumerator:   1,
		TrustLevelDenominator: 3,
		CurrentTimestamp:      currentTimestamp,
		ProcessedHeight:       12345,
		ProcessedTime:         currentTimestamp,
	}

	trustedState := consensusFixture{
		Height:             trustedHeight,
		Timestamp:          trustedTimestamp,
		AppHash:            hexHash("trusted-app", s.name),
		NextValidatorsHash: trusted.Hash,
		ProcessedHeight:    trustedHeight + 1000,
		ProcessedTime:      trustedTimestamp + 1,
	}

	adjacent := makeSignedHeader(chainID, client.RevisionNumber, trustedHeight+1, trustedTimestamp+30_000_000_000, trusted, trusted, s.signerCount, "adjacent")
	nonAdjacent := makeSignedHeader(chainID, client.RevisionNumber, trustedHeight+10, trustedTimestamp+40_000_000_000, untrusted, trusted, s.signerCount, "non-adjacent")
	nonAdjacentConflict := makeSignedHeader(chainID, client.RevisionNumber, trustedHeight+10, trustedTimestamp+41_000_000_000, untrusted, trusted, s.signerCount, "non-adjacent-conflict")
	conflict := makeSignedHeader(chainID, client.RevisionNumber, trustedHeight+1, trustedTimestamp+31_000_000_000, trusted, trusted, s.signerCount, "conflict")
	lower := makeSignedHeader(chainID, client.RevisionNumber, trustedHeight+1, trustedTimestamp+35_000_000_000, trusted, trusted, s.signerCount, "time-lower")
	higher := makeSignedHeader(chainID, client.RevisionNumber, trustedHeight+2, trustedTimestamp+35_000_000_000, trusted, trusted, s.signerCount, "time-higher")

	return fixture{
		Client:              client,
		Trusted:             trustedState,
		TrustedValidators:   trusted,
		UntrustedValidators: untrusted,
		Adjacent:            adjacent,
		NonAdjacent:         nonAdjacent,
		NonAdjacentConflict: nonAdjacentConflict,
		Conflict:            conflict,
		TimeViolationLower:  lower,
		TimeViolationHigher: higher,
		Membership:          makeMembershipFixture(),
		Iavl:                makeIavlFixture(),
		IavlMaxDepth:        makeIavlFixtureAtDepth(16),
		Canonical:           makeCanonicalFixture(s, chainID, client.RevisionNumber, trustedHeight+1, trustedTimestamp+30_000_000_000, trusted, adjacent),
	}
}

func makeValidatorSet(label string, count int, mixed bool) validatorFixture {
	validators := make([]string, count)
	powers := make([]uint64, count)
	leaves := make([]string, count)
	leafBytes := make([][]byte, count)

	for i := 0; i < count; i++ {
		priv := secpPrivKey(label, i)
		pub := priv.PubKey().SerializeUncompressed()
		addr := keccak256(pub[1:])[12:]
		power := votingPower(i, mixed)

		validators[i] = "0x" + hex.EncodeToString(addr)
		powers[i] = power
		leaf := validatorLeaf(addr, power)
		leaves[i] = "0x" + hex.EncodeToString(leaf)
		leafBytes[i] = leaf
	}

	return validatorFixture{
		Validators: validators,
		Powers:     powers,
		Leaves:     leaves,
		Hash:       "0x" + hex.EncodeToString(merkleHash(leafBytes)),
	}
}

func makeChangedValidatorSet(trusted validatorFixture, signerCount int, mixed bool) validatorFixture {
	count := len(trusted.Validators)
	validators := append([]string{}, trusted.Validators...)
	powers := append([]uint64{}, trusted.Powers...)
	leaves := make([]string, count)
	leafBytes := make([][]byte, count)

	for i := signerCount; i < count; i++ {
		priv := secpPrivKey("changed", i)
		pub := priv.PubKey().SerializeUncompressed()
		addr := keccak256(pub[1:])[12:]
		validators[i] = "0x" + hex.EncodeToString(addr)
		powers[i] = votingPower(i, mixed)
	}
	for i := 0; i < count; i++ {
		addr := mustDecodeHex(validators[i])
		leaf := validatorLeaf(addr, powers[i])
		leaves[i] = "0x" + hex.EncodeToString(leaf)
		leafBytes[i] = leaf
	}

	return validatorFixture{
		Validators: validators,
		Powers:     powers,
		Leaves:     leaves,
		Hash:       "0x" + hex.EncodeToString(merkleHash(leafBytes)),
	}
}

func makeSignedHeader(chainID []byte, revision, height, timestamp uint64, validators, nextValidators validatorFixture, signerCount int, label string) signedHeader {
	partSetHash := taggedHash("part-set", label)
	appHash := taggedHash("app", label)
	headerHash := hashHeader(chainID, revision, height, timestamp, mustDecodeHex(validators.Hash), mustDecodeHex(nextValidators.Hash), appHash, 0, 1, partSetHash)
	header := headerFixture{
		RevisionNumber:     revision,
		Height:             height,
		Timestamp:          timestamp,
		ValidatorsHash:     validators.Hash,
		NextValidatorsHash: nextValidators.Hash,
		AppHash:            "0x" + hex.EncodeToString(appHash),
		HeaderHash:         "0x" + hex.EncodeToString(headerHash),
		CommitBlockHash:    "0x" + hex.EncodeToString(headerHash),
		Round:              0,
		PartSetTotal:       1,
		PartSetHash:        "0x" + hex.EncodeToString(partSetHash),
	}

	commit := makeCommit(chainID, height, headerHash, 0, 1, partSetHash, validators, signerCount, timestamp)
	weakSignerCount := len(validators.Validators) / 3
	trustOnlySignerCount := weakSignerCount + 1
	weakCommit := makeCommit(chainID, height, headerHash, 0, 1, partSetHash, validators, weakSignerCount, timestamp)
	trustOnlyCommit := makeCommit(chainID, height, headerHash, 0, 1, partSetHash, validators, trustOnlySignerCount, timestamp)
	return signedHeader{Header: header, Commit: commit, WeakCommit: weakCommit, TrustOnlyCommit: trustOnlyCommit}
}

func makeCommit(chainID []byte, height uint64, blockHash []byte, round uint32, partSetTotal uint32, partSetHash []byte, validators validatorFixture, signerCount int, baseTimestamp uint64) commitFixture {
	signatures := make([]string, signerCount)
	timestamps := make([]uint64, signerCount)
	voteSignBytes := make([]string, signerCount)
	bitmap := new(big.Int)

	for i := 0; i < signerCount; i++ {
		bitmap.SetBit(bitmap, i, 1)
		timestamps[i] = baseTimestamp + uint64(i*1_000_000)
		msg := compactVoteSignBytes(chainID, height, round, blockHash, partSetTotal, partSetHash, timestamps[i])
		priv := secpPrivKeyForAddress(validators.Validators[i])
		signatures[i] = "0x" + hex.EncodeToString(ethSignature(priv, msg))
		voteSignBytes[i] = "0x" + hex.EncodeToString(msg)
	}

	totalPower := totalPower(validators.Powers)
	return commitFixture{
		Height:        height,
		BlockHash:     "0x" + hex.EncodeToString(blockHash),
		Round:         round,
		PartSetTotal:  partSetTotal,
		PartSetHash:   "0x" + hex.EncodeToString(partSetHash),
		SignerBitmap:  bitmap,
		Signatures:    signatures,
		Timestamps:    timestamps,
		VoteSignBytes: voteSignBytes,
		RequiredPower: (totalPower * 2) / 3,
		TrustPower:    totalPower / 3,
	}
}

func secpPrivKeyForAddress(address string) *secp.PrivateKey {
	for _, label := range []string{"trusted", "changed"} {
		for i := 0; i < 256; i++ {
			priv := secpPrivKey(label, i)
			pub := priv.PubKey().SerializeUncompressed()
			addr := "0x" + hex.EncodeToString(keccak256(pub[1:])[12:])
			if addr == address {
				return priv
			}
		}
	}
	panic("private key not found for address")
}

func votingPower(i int, mixed bool) uint64 {
	if !mixed {
		return 1
	}
	if i < 10 {
		return 3
	}
	if i < 25 {
		return 2
	}
	return 1
}

func validatorLeaf(address []byte, power uint64) []byte {
	out := make([]byte, 0, 28)
	out = append(out, address...)
	out = append(out, uint64Bytes(power)...)
	return out
}

func compactVoteSignBytes(chainID []byte, height uint64, round uint32, blockHash []byte, partSetTotal uint32, partSetHash []byte, timestamp uint64) []byte {
	out := make([]byte, 0, len(chainID)+8+4+32+4+32+8)
	out = append(out, chainID...)
	out = append(out, uint64Bytes(height)...)
	out = append(out, uint32Bytes(round)...)
	out = append(out, blockHash...)
	out = append(out, uint32Bytes(partSetTotal)...)
	out = append(out, partSetHash...)
	out = append(out, uint64Bytes(timestamp)...)
	return out
}

func hashHeader(chainID []byte, revision, height, timestamp uint64, validatorsHash, nextValidatorsHash, appHash []byte, round uint32, partSetTotal uint32, partSetHash []byte) []byte {
	out := make([]byte, 0, len(chainID)+8+8+8+32+32+32+4+4+32)
	out = append(out, chainID...)
	out = append(out, uint64Bytes(revision)...)
	out = append(out, uint64Bytes(height)...)
	out = append(out, uint64Bytes(timestamp)...)
	out = append(out, validatorsHash...)
	out = append(out, nextValidatorsHash...)
	out = append(out, appHash...)
	out = append(out, uint32Bytes(round)...)
	out = append(out, uint32Bytes(partSetTotal)...)
	out = append(out, partSetHash...)
	return sha256Bytes(out)
}

func secpPrivKey(label string, i int) *secp.PrivateKey {
	seed := sha256.Sum256([]byte(fmt.Sprintf("%s validator %d", label, i)))
	fe := new(big.Int).SetBytes(seed[:])
	fe.Mod(fe, secpNMinusOne)
	fe.Add(fe, big.NewInt(1))

	privBytes := make([]byte, 32)
	feB := fe.Bytes()
	copy(privBytes[32-len(feB):], feB)
	return secp.PrivKeyFromBytes(privBytes)
}

func ethSignature(priv *secp.PrivateKey, msg []byte) []byte {
	compact := secpEcdsa.SignCompact(priv, keccak256(msg), false)
	if len(compact) != 65 {
		panic("unexpected compact signature length")
	}
	v := compact[0] - 27
	if v > 1 {
		panic(fmt.Sprintf("unexpected recovery id %d", v))
	}
	sig := append([]byte{}, compact[1:]...)
	sig = append(sig, v)
	return sig
}

func makeMembershipFixture() membershipFixture {
	leaves := make([][]byte, 8)
	for i := range leaves {
		key := []byte(fmt.Sprintf("ibc/key/%02d", i))
		value := []byte(fmt.Sprintf("ibc-value-%02d", i))
		leaves[i] = sha256Tagged(0x00, key, value)
	}

	target := 5
	siblings, siblingOnLeft := merkleProof(leaves, target)
	return membershipFixture{
		Root:          "0x" + hex.EncodeToString(pairMerkleRoot(leaves)),
		Key:           "0x" + hex.EncodeToString([]byte(fmt.Sprintf("ibc/key/%02d", target))),
		Value:         "0x" + hex.EncodeToString([]byte(fmt.Sprintf("ibc-value-%02d", target))),
		Siblings:      bytesListToHex(siblings),
		SiblingOnLeft: siblingOnLeft,
	}
}

// makeIavlFixture constructs an IAVL-shaped existence proof. IAVL leaves are
// sha256(0x00 || varint(keyLen) || key || varint(valueHashLen) || valueHash) where
// valueHash = sha256(value). Inner nodes are sha256(prefix || running_hash || suffix)
// with the prefix encoding height/size/version and the appropriate child hash.
func makeIavlFixture() iavlFixture {
	return makeIavlFixtureAtDepth(8)
}

// makeIavlFixtureAtDepth builds an existence proof with a tree of exactly 2^depth
// leaves so the proof has `depth` inner ops. Depth 16 corresponds to ~65k entries
// and matches the upper end of typical IBC commitment trees.
func makeIavlFixtureAtDepth(depth uint) iavlFixture {
	keys := make([][]byte, 1<<depth)
	values := make([][]byte, len(keys))
	for i := range keys {
		keys[i] = []byte(fmt.Sprintf("ibc/iavl/key/%05d", i))
		values[i] = []byte(fmt.Sprintf("iavl-value-%05d", i))
	}
	target := 99
	if target >= len(keys) {
		target = len(keys) / 3
	}

	tree := iavl.NewMutableTree(iavldb.NewMemDB(), 0, false, iavl.NewNopLogger())
	for i := range keys {
		_, err := tree.Set(keys[i], values[i])
		must(err)
	}
	root, _, err := tree.SaveVersion()
	must(err)
	commitment, err := tree.GetMembershipProof(keys[target])
	must(err)
	ok, err := tree.VerifyMembership(commitment, keys[target])
	must(err)
	if !ok {
		panic("iavl membership proof verification failed")
	}
	existence := commitment.GetExist()
	if existence == nil {
		panic("missing iavl existence proof")
	}
	calculated, err := existence.Calculate()
	must(err)
	if !bytes.Equal(calculated, root) {
		panic("ics23 existence proof root mismatch")
	}
	must(existence.Verify(ics23.IavlSpec, root, keys[target], values[target]))

	leafPrefix := append([]byte{}, existence.Leaf.Prefix...)
	leafPrefix = append(leafPrefix, encodeUvarint(uint64(len(keys[target])))...)
	leafPrefix = append(leafPrefix, keys[target]...)
	leafPrefix = append(leafPrefix, encodeUvarint(32)...)
	innerPrefixes := make([][]byte, len(existence.Path))
	innerSuffixes := make([][]byte, len(existence.Path))
	for i, op := range existence.Path {
		innerPrefixes[i] = op.Prefix
		innerSuffixes[i] = op.Suffix
	}

	existenceBytes, err := proto.Marshal(existence)
	must(err)
	commitmentBytes, err := proto.Marshal(commitment)
	must(err)

	return iavlFixture{
		Root:            "0x" + hex.EncodeToString(root),
		Key:             "0x" + hex.EncodeToString(keys[target]),
		Value:           "0x" + hex.EncodeToString(values[target]),
		LeafPrefix:      "0x" + hex.EncodeToString(leafPrefix),
		InnerPrefixes:   bytesListToHex(innerPrefixes),
		InnerSuffixes:   bytesListToHex(innerSuffixes),
		ExistenceProof:  "0x" + hex.EncodeToString(existenceBytes),
		CommitmentProof: "0x" + hex.EncodeToString(commitmentBytes),
	}
}

// iavlLeafPrefixBytes returns the bytes that, when sha256(prefix || sha256(value))
// is computed, gives the IAVL leaf hash. The prefix encodes 0x00 || varint(height=0) ||
// varint(size=1) || varint(version=1) || varint(keyLen) || key || varint(valueHashLen).
// We pre-hash the value so that the verifier only ever sees a fixed-size 32-byte input.
func iavlLeafPrefixBytes(key, value []byte) []byte {
	out := iavlLeafOpPrefixBytes()
	out = append(out, encodeUvarint(uint64(len(key)))...)
	out = append(out, key...)
	out = append(out, encodeUvarint(32)...)
	return out
}

func iavlLeafOpPrefixBytes() []byte {
	out := []byte{0x00}
	out = append(out, encodeVarint(1)...)
	out = append(out, encodeVarint(1)...)
	return out
}

func iavlLeafHash(key, value []byte) []byte {
	prefix := iavlLeafPrefixBytes(key, value)
	valueHash := sha256Bytes(value)
	return sha256Bytes(append(append([]byte{}, prefix...), valueHash...))
}

func iavlInnerPrefixBytesLeft(height int8, size int) []byte {
	out := []byte{0x01}
	out = append(out, encodeVarint(int64(height))...)
	out = append(out, encodeVarint(int64(size))...)
	out = append(out, encodeVarint(int64(1))...)
	out = append(out, 0x20)
	return out
}

func iavlInnerPrefixBytesRight(height int8, size int, leftHash []byte) []byte {
	out := []byte{0x01}
	out = append(out, encodeVarint(int64(height))...)
	out = append(out, encodeVarint(int64(size))...)
	out = append(out, encodeVarint(int64(1))...)
	out = append(out, 0x20)
	out = append(out, leftHash...)
	out = append(out, 0x20)
	return out
}

func iavlInnerHashOpaque(height int8, size int, left, right []byte) []byte {
	out := []byte{0x01}
	out = append(out, encodeVarint(int64(height))...)
	out = append(out, encodeVarint(int64(size))...)
	out = append(out, encodeVarint(int64(1))...)
	out = append(out, 0x20)
	out = append(out, left...)
	out = append(out, 0x20)
	out = append(out, right...)
	return sha256Bytes(out)
}

func encodeVarint(v int64) []byte {
	uv := uint64(v) << 1
	if v < 0 {
		uv = ^uv
	}
	out := make([]byte, 0, 10)
	for uv >= 0x80 {
		out = append(out, byte(uv)|0x80)
		uv >>= 7
	}
	return append(out, byte(uv))
}

func encodeUvarint(v uint64) []byte {
	out := make([]byte, 0, 10)
	for v >= 0x80 {
		out = append(out, byte(v)|0x80)
		v >>= 7
	}
	return append(out, byte(v))
}

// makeCanonicalFixture produces real CometBFT canonical bytes for the trusted
// validator set: SimpleValidator leaves + ValidatorSet.Hash() in both ed25519
// and secp256k1 PubKey forms, canonical voteSignBytes signed under those keys,
// and a canonical Header.Hash() over the adjacent header. BLS data is added
// separately by gen_fixture_bls.go when the bls12381 build tag is enabled.
func makeCanonicalFixture(s scenario, chainID []byte, revision, height, baseTimestamp uint64, trusted validatorFixture, adjacent signedHeader) canonicalFixture {
	count := len(trusted.Validators)

	ed25519Set := makeCanonicalEd25519Set(count, trusted.Powers)
	secp256k1Set := makeCanonicalSecp256k1Set(count, trusted.Powers)

	timestamps := make([]int64, s.signerCount)
	voteBytes := make([]string, s.signerCount)
	ethSigs := make([]string, s.signerCount)
	ed25519Sigs := make([]string, s.signerCount)
	secp256k1Sigs := make([]string, s.signerCount)
	blockHash := mustDecodeHex(adjacent.Header.HeaderHash)
	partSetHash := mustDecodeHex(adjacent.Header.PartSetHash)

	for i := 0; i < s.signerCount; i++ {
		ts := int64(baseTimestamp) + int64(i)*1_000_000
		timestamps[i] = ts
		signBytes := canonicalVoteSignBytes(string(chainID), int64(height), 0, blockHash, 1, partSetHash, ts)
		voteBytes[i] = "0x" + hex.EncodeToString(signBytes)
		ethSigs[i] = "0x" + hex.EncodeToString(ethSignature(secpPrivKey("trusted", i), signBytes))

		edPriv := canonicalEd25519Key(i)
		edSig, err := edPriv.Sign(signBytes)
		must(err)
		ed25519Sigs[i] = "0x" + hex.EncodeToString(edSig)

		secpPriv := canonicalSecp256k1Key(i)
		secpSig, err := secpPriv.Sign(signBytes)
		must(err)
		secp256k1Sigs[i] = "0x" + hex.EncodeToString(secpSig)
	}

	ed25519Set.Signatures = ed25519Sigs
	secp256k1Set.Signatures = secp256k1Sigs

	headerHash := canonicalHeaderHash(string(chainID), int64(revision), int64(height), int64(baseTimestamp), mustDecodeHex(adjacent.Header.ValidatorsHash), mustDecodeHex(adjacent.Header.NextValidatorsHash), mustDecodeHex(adjacent.Header.AppHash), 0, 1, partSetHash)

	bitmap := new(big.Int)
	for i := 0; i < s.signerCount; i++ {
		bitmap.SetBit(bitmap, i, 1)
	}

	bls := canonicalBls{
		Available:    false,
		PubKeys:      []string{},
		Powers:       []int64{},
		Leaves:       []string{},
		Signatures:   []string{},
		SignerBitmap: bitmap,
	}
	if blsAvailable {
		bls = makeCanonicalBls(count, trusted.Powers, s.signerCount, voteBytesShared(voteBytes, s.signerCount))
	}

	return canonicalFixture{
		Ed25519:   ed25519Set,
		Secp256k1: secp256k1Set,
		Bls:       bls,
		Vote: canonicalVote{
			ChainID:       string(chainID),
			Height:        int64(height),
			Round:         0,
			BlockHash:     "0x" + hex.EncodeToString(blockHash),
			PartSetTotal:  1,
			PartSetHash:   "0x" + hex.EncodeToString(partSetHash),
			Timestamps:    timestamps,
			SignBytes:     voteBytes,
			EthSignatures: ethSigs,
		},
		HeaderHash: "0x" + hex.EncodeToString(headerHash),
	}
}

// voteBytesShared returns one fixed message used by the BLS path. BLS aggregation
// requires every signer to sign the same message; we use the first signer's
// canonical voteSignBytes so the canonical message is well-defined.
func voteBytesShared(voteBytes []string, signerCount int) []byte {
	if signerCount == 0 {
		return nil
	}
	return mustDecodeHex(voteBytes[0])
}

func makeCanonicalEd25519Set(count int, powers []uint64) canonicalValidatorSet {
	pubKeys := make([]string, count)
	leaves := make([][]byte, count)
	leavesHex := make([]string, count)
	int64Powers := make([]int64, count)

	cmtVals := make([]*cmttypes.Validator, count)
	for i := 0; i < count; i++ {
		priv := canonicalEd25519Key(i)
		pub := priv.PubKey()
		int64Powers[i] = int64(powers[i])

		pubKeys[i] = "0x" + hex.EncodeToString(pub.Bytes())

		v := cmttypes.NewValidator(pub, int64Powers[i])
		cmtVals[i] = v
		leaf := v.Bytes()
		leaves[i] = leaf
		leavesHex[i] = "0x" + hex.EncodeToString(leaf)
	}

	hash := mustValidatorSetHash(cmtVals)
	return canonicalValidatorSet{
		PubKeys: pubKeys,
		Powers:  int64Powers,
		Leaves:  leavesHex,
		Hash:    "0x" + hex.EncodeToString(hash),
	}
}

func makeCanonicalSecp256k1Set(count int, powers []uint64) canonicalValidatorSet {
	pubKeys := make([]string, count)
	leaves := make([][]byte, count)
	leavesHex := make([]string, count)
	int64Powers := make([]int64, count)

	cmtVals := make([]*cmttypes.Validator, count)
	for i := 0; i < count; i++ {
		priv := canonicalSecp256k1Key(i)
		pub := priv.PubKey()
		int64Powers[i] = int64(powers[i])

		pubKeys[i] = "0x" + hex.EncodeToString(pub.Bytes())

		v := cmttypes.NewValidator(pub, int64Powers[i])
		cmtVals[i] = v
		leaf := v.Bytes()
		leaves[i] = leaf
		leavesHex[i] = "0x" + hex.EncodeToString(leaf)
	}

	hash := mustValidatorSetHash(cmtVals)
	return canonicalValidatorSet{
		PubKeys: pubKeys,
		Powers:  int64Powers,
		Leaves:  leavesHex,
		Hash:    "0x" + hex.EncodeToString(hash),
	}
}

func canonicalEd25519Key(i int) cmted25519.PrivKey {
	seed := sha256.Sum256([]byte(fmt.Sprintf("canonical-ed25519 %d", i)))
	return cmted25519.GenPrivKeyFromSecret(seed[:])
}

func canonicalSecp256k1Key(i int) cmtsecp.PrivKey {
	seed := sha256.Sum256([]byte(fmt.Sprintf("canonical-secp256k1 %d", i)))
	return cmtsecp.GenPrivKeySecp256k1(seed[:])
}

func mustValidatorSetHash(vals []*cmttypes.Validator) []byte {
	// types.NewValidatorSet sorts by power; for benchmark stability we want the
	// hash over the validators in *insertion order* as supplied to the EVM
	// contract. We replicate the canonical merkle hash directly from each
	// validator's protobuf bytes to avoid the proposer-priority side-effects in
	// NewValidatorSet.
	bz := make([][]byte, len(vals))
	for i, v := range vals {
		bz[i] = v.Bytes()
	}
	return cmtMerkleHashFromByteSlices(bz)
}

// cmtMerkleHashFromByteSlices replicates crypto/merkle.HashFromByteSlices using
// SHA256 with the standard 0x00 leaf / 0x01 inner prefixes (RFC6962-style).
func cmtMerkleHashFromByteSlices(items [][]byte) []byte {
	if len(items) == 0 {
		return sha256Bytes(nil)
	}
	leafHashes := make([][]byte, len(items))
	for i, item := range items {
		leafHashes[i] = sha256Tagged(0x00, item)
	}
	return cmtMerkleRange(leafHashes, 0, len(leafHashes))
}

func cmtMerkleRange(hashes [][]byte, start, end int) []byte {
	count := end - start
	if count == 1 {
		return hashes[start]
	}
	split := splitPoint(count)
	left := cmtMerkleRange(hashes, start, start+split)
	right := cmtMerkleRange(hashes, start+split, end)
	return sha256Tagged(0x01, left, right)
}

// canonicalVoteSignBytes returns CometBFT's canonical Vote sign bytes by
// constructing a cmtproto.Vote and running it through the public
// cmttypes.VoteSignBytes function. Signing this exact byte string is what
// cometbft validators do today.
func canonicalVoteSignBytes(chainID string, height int64, round int32, blockHash []byte, partSetTotal uint32, partSetHash []byte, unixNanos int64) []byte {
	vote := cmtproto.Vote{
		Type:   cmtproto.PrecommitType,
		Height: height,
		Round:  round,
		BlockID: cmtproto.BlockID{
			Hash: blockHash,
			PartSetHeader: cmtproto.PartSetHeader{
				Total: partSetTotal,
				Hash:  partSetHash,
			},
		},
		Timestamp: time.Unix(unixNanos/1_000_000_000, unixNanos%1_000_000_000).UTC(),
	}
	return cmttypes.VoteSignBytes(chainID, &vote)
}

// canonicalHeaderHash computes a real CometBFT Header.Hash() over a Header that
// matches the benchmark's reduced 10-field shape; the eight unmodelled fields
// are zero-valued so the canonical hash is comparable across runs.
func canonicalHeaderHash(chainID string, revision, height, unixNanos int64, validatorsHash, nextValidatorsHash, appHash []byte, round int32, partSetTotal uint32, partSetHash []byte) []byte {
	_ = round
	h := cmttypes.Header{
		ChainID:            chainID,
		Height:             height,
		Time:               time.Unix(unixNanos/1_000_000_000, unixNanos%1_000_000_000).UTC(),
		ValidatorsHash:     validatorsHash,
		NextValidatorsHash: nextValidatorsHash,
		AppHash:            appHash,
	}
	h.LastBlockID = cmttypes.BlockID{
		Hash: make([]byte, 32),
		PartSetHeader: cmttypes.PartSetHeader{
			Total: partSetTotal,
			Hash:  partSetHash,
		},
	}
	_ = revision
	return h.Hash()
}

func merkleHash(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		return sha256Bytes(nil)
	}
	hashes := make([][]byte, len(leaves))
	for i, leaf := range leaves {
		hashes[i] = sha256Tagged(0x00, leaf)
	}
	return merkleHashRange(hashes, 0, len(hashes))
}

func merkleHashRange(hashes [][]byte, start, end int) []byte {
	count := end - start
	if count == 1 {
		return hashes[start]
	}
	split := splitPoint(count)
	left := merkleHashRange(hashes, start, start+split)
	right := merkleHashRange(hashes, start+split, end)
	return sha256Tagged(0x01, left, right)
}

func splitPoint(count int) int {
	split := 1
	for split*2 < count {
		split *= 2
	}
	return split
}

func pairMerkleRoot(leaves [][]byte) []byte {
	level := append([][]byte{}, leaves...)
	for len(level) > 1 {
		next := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 == len(level) {
				next = append(next, level[i])
			} else {
				next = append(next, sha256Tagged(0x01, level[i], level[i+1]))
			}
		}
		level = next
	}
	return level[0]
}

func merkleProof(leaves [][]byte, target int) ([][]byte, []bool) {
	level := append([][]byte{}, leaves...)
	index := target
	var siblings [][]byte
	var siblingOnLeft []bool
	for len(level) > 1 {
		if index%2 == 0 {
			if index+1 < len(level) {
				siblings = append(siblings, level[index+1])
				siblingOnLeft = append(siblingOnLeft, false)
			}
		} else {
			siblings = append(siblings, level[index-1])
			siblingOnLeft = append(siblingOnLeft, true)
		}

		next := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 == len(level) {
				next = append(next, level[i])
			} else {
				next = append(next, sha256Tagged(0x01, level[i], level[i+1]))
			}
		}
		index /= 2
		level = next
	}
	return siblings, siblingOnLeft
}

func sha256Tagged(tag byte, parts ...[]byte) []byte {
	h := sha256.New()
	_, _ = h.Write([]byte{tag})
	for _, part := range parts {
		_, _ = h.Write(part)
	}
	return h.Sum(nil)
}

func taggedHash(parts ...string) []byte {
	return sha256Bytes([]byte(fmt.Sprint(parts)))
}

func hexHash(parts ...string) string {
	return "0x" + hex.EncodeToString(taggedHash(parts...))
}

func sha256Bytes(bz []byte) []byte {
	sum := sha256.Sum256(bz)
	return sum[:]
}

func bytesListToHex(items [][]byte) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = "0x" + hex.EncodeToString(item)
	}
	return out
}

func mustDecodeHex(s string) []byte {
	s = strings.TrimPrefix(s, "0x")
	bz, err := hex.DecodeString(s)
	must(err)
	return bz
}

func uint64Bytes(v uint64) []byte {
	out := make([]byte, 8)
	binary.BigEndian.PutUint64(out, v)
	return out
}

func uint32Bytes(v uint32) []byte {
	out := make([]byte, 4)
	binary.BigEndian.PutUint32(out, v)
	return out
}

func keccak256(bz []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	_, _ = h.Write(bz)
	return h.Sum(nil)
}

func totalPower(powers []uint64) uint64 {
	var total uint64
	for _, power := range powers {
		total += power
	}
	return total
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// Not used directly today, but kept to keep import surface stable when adding
// future BLS scenarios that need to sort signers deterministically.
var _ = sort.Sort
