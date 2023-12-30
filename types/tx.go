package types

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	// <celestia-core>
	"github.com/cometbft/cometbft/pkg/consts"
	proto "github.com/cosmos/gogoproto/proto"
	// </celestia-core>
)

// TxKeySize is the size of the transaction key index
const TxKeySize = sha256.Size

type (
	// Tx is an arbitrary byte array.
	// NOTE: Tx has no types at this level, so when wire encoded it's just length-prefixed.
	// Might we want types here ?
	Tx []byte

	// TxKey is the fixed length array key used as an index.
	TxKey [TxKeySize]byte
)

// Hash computes the TMHASH hash of the wire encoded transaction.
func (tx Tx) Hash() []byte {
	// <celestia-core>
	if indexWrapper, isIndexWrapper := UnmarshalIndexWrapper(tx); isIndexWrapper {
		return tmhash.Sum(indexWrapper.Tx)
	}
	if blobTx, isBlobTx := UnmarshalBlobTx(tx); isBlobTx {
		return tmhash.Sum(blobTx.Tx)
	}
	// </celestia-core>
	return tmhash.Sum(tx)
}

func (tx Tx) Key() TxKey {
	// <celestia-core>
	if blobTx, isBlobTx := UnmarshalBlobTx(tx); isBlobTx {
		return sha256.Sum256(blobTx.Tx)
	}
	if indexWrapper, isIndexWrapper := UnmarshalIndexWrapper(tx); isIndexWrapper {
		return sha256.Sum256(indexWrapper.Tx)
	}
	// </celestia-core>
	return sha256.Sum256(tx)
}

// String returns the hex-encoded transaction as a string.
func (tx Tx) String() string {
	return fmt.Sprintf("Tx{%X}", []byte(tx))
}

// Txs is a slice of Tx.
type Txs []Tx

// Hash returns the Merkle root hash of the transaction hashes.
// i.e. the leaves of the tree are the hashes of the txs.
func (txs Txs) Hash() []byte {
	hl := txs.hashList()
	return merkle.HashFromByteSlices(hl)
}

// Index returns the index of this transaction in the list, or -1 if not found
func (txs Txs) Index(tx Tx) int {
	for i := range txs {
		if bytes.Equal(txs[i], tx) {
			return i
		}
	}
	return -1
}

// IndexByHash returns the index of this transaction hash in the list, or -1 if not found
func (txs Txs) IndexByHash(hash []byte) int {
	for i := range txs {
		if bytes.Equal(txs[i].Hash(), hash) {
			return i
		}
	}
	return -1
}

func (txs Txs) Proof(i int) TxProof {
	hl := txs.hashList()
	root, proofs := merkle.ProofsFromByteSlices(hl)

	return TxProof{
		RootHash: root,
		Data:     txs[i],
		Proof:    *proofs[i],
	}
}

func (txs Txs) hashList() [][]byte {
	hl := make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		hl[i] = txs[i].Hash()
	}
	return hl
}

// Txs is a slice of transactions. Sorting a Txs value orders the transactions
// lexicographically.
func (txs Txs) Len() int      { return len(txs) }
func (txs Txs) Swap(i, j int) { txs[i], txs[j] = txs[j], txs[i] }
func (txs Txs) Less(i, j int) bool {
	return bytes.Compare(txs[i], txs[j]) == -1
}

func ToTxs(txl [][]byte) Txs {
	txs := make([]Tx, 0, len(txl))
	for _, tx := range txl {
		txs = append(txs, tx)
	}
	return txs
}

func (txs Txs) Validate(maxSizeBytes int64) error {
	var size int64
	for _, tx := range txs {
		size += int64(len(tx))
		if size > maxSizeBytes {
			return fmt.Errorf("transaction data size exceeds maximum %d", maxSizeBytes)
		}
	}
	return nil
}

// ToSliceOfBytes converts a Txs to slice of byte slices.
func (txs Txs) ToSliceOfBytes() [][]byte {
	txBzs := make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		txBzs[i] = txs[i]
	}
	return txBzs
}

// TxProof represents a Merkle proof of the presence of a transaction in the Merkle tree.
type TxProof struct {
	RootHash cmtbytes.HexBytes `json:"root_hash"`
	Data     Tx                `json:"data"`
	Proof    merkle.Proof      `json:"proof"`
}

// Leaf returns the hash(tx), which is the leaf in the merkle tree which this proof refers to.
func (tp TxProof) Leaf() []byte {
	return tp.Data.Hash()
}

// Validate verifies the proof. It returns nil if the RootHash matches the dataHash argument,
// and if the proof is internally consistent. Otherwise, it returns a sensible error.
func (tp TxProof) Validate(dataHash []byte) error {
	if !bytes.Equal(dataHash, tp.RootHash) {
		return errors.New("proof matches different data hash")
	}
	if tp.Proof.Index < 0 {
		return errors.New("proof index cannot be negative")
	}
	if tp.Proof.Total <= 0 {
		return errors.New("proof total must be positive")
	}
	valid := tp.Proof.Verify(tp.RootHash, tp.Leaf())
	if valid != nil {
		return errors.New("proof is not internally consistent")
	}
	return nil
}

func (tp TxProof) ToProto() cmtproto.TxProof {

	pbProof := tp.Proof.ToProto()

	pbtp := cmtproto.TxProof{
		RootHash: tp.RootHash,
		Data:     tp.Data,
		Proof:    pbProof,
	}

	return pbtp
}
func TxProofFromProto(pb cmtproto.TxProof) (TxProof, error) {

	pbProof, err := merkle.ProofFromProto(pb.Proof)
	if err != nil {
		return TxProof{}, err
	}

	pbtp := TxProof{
		RootHash: pb.RootHash,
		Data:     pb.Data,
		Proof:    *pbProof,
	}

	return pbtp, nil
}

// ComputeProtoSizeForTxs wraps the transactions in cmtproto.Data{} and calculates the size.
// https://developers.google.com/protocol-buffers/docs/encoding
func ComputeProtoSizeForTxs(txs []Tx) int64 {
	data := Data{Txs: txs}
	pdData := data.ToProto()
	return int64(pdData.Size())
}

// <celestia-core>
// UnmarshalIndexWrapper attempts to unmarshal the provided transaction into an
// IndexWrapper transaction. It returns true if the provided transaction is an
// IndexWrapper transaction. An IndexWrapper transaction is a transaction that contains
// a MsgPayForBlob that has been wrapped with a share index.
//
// NOTE: protobuf sometimes does not throw an error if the transaction passed is
// not a tmproto.IndexWrapper, since the protobuf definition for MsgPayForBlob is
// kept in the app, we cannot perform further checks without creating an import
// cycle.
func UnmarshalIndexWrapper(tx Tx) (indexWrapper cmtproto.IndexWrapper, isIndexWrapper bool) {
	// attempt to unmarshal into an IndexWrapper transaction
	err := proto.Unmarshal(tx, &indexWrapper)
	if err != nil {
		return indexWrapper, false
	}
	if indexWrapper.TypeId != consts.ProtoIndexWrapperTypeID {
		return indexWrapper, false
	}
	return indexWrapper, true
}

// MarshalIndexWrapper creates a wrapped Tx that includes the original transaction
// and the share index of the start of its blob.
//
// NOTE: must be unwrapped to be a viable sdk.Tx
func MarshalIndexWrapper(tx Tx, shareIndexes ...uint32) (Tx, error) {
	wTx := cmtproto.IndexWrapper{
		Tx:           tx,
		ShareIndexes: shareIndexes,
		TypeId:       consts.ProtoIndexWrapperTypeID,
	}
	return proto.Marshal(&wTx)
}

// UnmarshalBlobTx attempts to unmarshal a transaction into blob transaction. If an
// error is thrown, false is returned.
func UnmarshalBlobTx(tx Tx) (bTx cmtproto.BlobTx, isBlob bool) {
	err := bTx.Unmarshal(tx)
	if err != nil {
		return cmtproto.BlobTx{}, false
	}
	// perform some quick basic checks to prevent false positives
	if bTx.TypeId != consts.ProtoBlobTxTypeID {
		return bTx, false
	}
	if len(bTx.Blobs) == 0 {
		return bTx, false
	}
	for _, b := range bTx.Blobs {
		if len(b.NamespaceId) != consts.NamespaceIDSize {
			return bTx, false
		}
	}
	return bTx, true
}

// MarshalBlobTx creates a BlobTx using a normal transaction and some number of
// blobs.
//
// NOTE: Any checks on the blobs or the transaction must be performed in the
// application
func MarshalBlobTx(tx []byte, blobs ...*cmtproto.Blob) (Tx, error) {
	bTx := cmtproto.BlobTx{
		Tx:     tx,
		Blobs:  blobs,
		TypeId: consts.ProtoBlobTxTypeID,
	}
	return bTx.Marshal()
}

// </celestia-core>
