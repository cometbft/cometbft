package types

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/libs/bits"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/klauspost/reedsolomon"
)

//-------------------------------------

type ECPartSetHeader struct {
	Total  uint32            `json:"total"`
	Parity uint32            `json:"parity"`
	Hash   cmtbytes.HexBytes `json:"hash"`
}

// String returns a string representation of PartSetHeader.
//
// 1. total number of parts
// 2. first 6 bytes of the hash
func (psh ECPartSetHeader) String() string {
	return fmt.Sprintf("%v:%X", psh.Total, cmtbytes.Fingerprint(psh.Hash))
}

func (psh ECPartSetHeader) IsZero() bool {
	return psh.Total == 0 && len(psh.Hash) == 0
}

func (psh ECPartSetHeader) Equals(other ECPartSetHeader) bool {
	return psh.Total == other.Total && bytes.Equal(psh.Hash, other.Hash) && psh.Parity == other.Parity
}

// ValidateBasic performs basic validation.
func (psh ECPartSetHeader) ValidateBasic() error {
	// Hash can be empty in case of POLBlockID.PartSetHeader in Proposal.
	if err := ValidateHash(psh.Hash); err != nil {
		return fmt.Errorf("wrong Hash: %w", err)
	}
	return nil
}

// ToProto converts ECPartSetHeader to protobuf
func (psh *ECPartSetHeader) ToProto() cmtproto.ECPartSetHeader {
	if psh == nil {
		return cmtproto.ECPartSetHeader{}
	}

	return cmtproto.ECPartSetHeader{
		Total:  psh.Total,
		Hash:   psh.Hash,
		Parity: psh.Parity,
	}
}

// ECPartSetHeaderFromProto sets a protobuf PartSetHeader to the given pointer
func ECPartSetHeaderFromProto(ppsh *cmtproto.ECPartSetHeader) (*ECPartSetHeader, error) {
	if ppsh == nil {
		return nil, errors.New("nil PartSetHeader")
	}
	psh := new(ECPartSetHeader)
	psh.Total = ppsh.Total
	psh.Hash = ppsh.Hash
	psh.Parity = ppsh.Parity

	return psh, psh.ValidateBasic()
}

// ProtoECPartSetHeaderIsZero is similar to the IsZero function for
// ECPartSetHeader, but for the Protobuf representation.
func ProtoECPartSetHeaderIsZero(ppsh *cmtproto.ECPartSetHeader) bool {
	return ppsh.Total == 0 && len(ppsh.Hash) == 0 && ppsh.Parity == 0
}

//-------------------------------------

type ECPartSet struct {
	total uint32
	hash  []byte

	parity uint32

	mtx           cmtsync.Mutex
	parts         []*Part
	partsBitArray *bits.BitArray
	count         uint32
	// a count of the total size (in bytes). Used to ensure that the
	// part set doesn't exceed the maximum block bytes
	byteSize int64
}

// NewECPartSetFromData returns an immutable, full ECPartSet from the data bytes.
// The data bytes are split into "partSize" chunks, and merkle tree computed.
// CONTRACT: partSize is greater than zero.
func NewECPartSetFromData(data []byte, partSize uint32) (*ECPartSet, error) {
	// divide data into parts of size `partSize`
	dataShards := (uint32(len(data)) + partSize - 1) / partSize
	if dataShards > 256 {
		// NOTE: Reed Solomon has a restriction where some functionality we need
		// cannot be used with > 256 data shards. If we need more than 256
		// shards, then our shards will start to grow very large due to this
		// limitation (greater than the configured part size), which will break
		// things. This effectively caps our blocks at 16mb.
		panic(fmt.Errorf("cannot have > 256 total shards for a single block, got %d", dataShards))
	}
	parityShards := dataShards / 2
	total := dataShards + parityShards
	parts := make([]*Part, dataShards+parityShards)
	partsBytes := make([][]byte, dataShards+parityShards)

	// TODO: this is pretty arbitrary, need to do more tuning here
	enc, err := reedsolomon.New(int(dataShards), int(parityShards))
	if err != nil {
		return nil, fmt.Errorf("creating rs encoder with %d data shards and %d parity shards: %w", dataShards, parityShards, err)
	}

	shards, err := enc.Split(data)
	if err != nil {
		return nil, fmt.Errorf("splitting %d bytes of data into %d data shards and %d parity shards: %w", len(data), dataShards, parityShards, err)
	}

	if err = enc.Encode(shards); err != nil {
		return nil, fmt.Errorf("encoding parity into data shards: %w", err)
	}

	for i := uint32(0); i < total; i++ {
		parts[i] = &Part{
			Index: i,
			Bytes: shards[i],
		}
		partsBytes[i] = shards[i]
	}
	// Compute merkle proofs
	root, proofs := merkle.ProofsFromByteSlices(partsBytes)
	for i := uint32(0); i < total; i++ {
		parts[i].Proof = *proofs[i]
	}
	partsBitArray := bits.NewBitArrayFromFn(int(total), func(int) bool { return true })
	return &ECPartSet{
		total:         total,
		parity:        parityShards,
		hash:          root,
		parts:         parts,
		partsBitArray: partsBitArray,
		count:         total,
		// all shards must have the same length, so this is ok to calc the
		// total size
		byteSize: int64(len(shards[0]) * int(total)),
	}, nil
}

// NewECPartSetFromHeader returns an empty PartSet ready to be populated.
func NewECPartSetFromHeader(header ECPartSetHeader) *ECPartSet {
	return &ECPartSet{
		total:         header.Total,
		hash:          header.Hash,
		parity:        header.Parity,
		parts:         make([]*Part, header.Total),
		partsBitArray: bits.NewBitArray(int(header.Total)),
		count:         0,
		byteSize:      0,
	}
}

func (ps *ECPartSet) Header() ECPartSetHeader {
	if ps == nil {
		return ECPartSetHeader{}
	}
	return ECPartSetHeader{
		Total:  ps.total,
		Parity: ps.parity,
		Hash:   ps.hash,
	}
}

func (ps *ECPartSet) HasHeader(header ECPartSetHeader) bool {
	if ps == nil {
		return false
	}
	return ps.Header().Equals(header)
}

func (ps *ECPartSet) BitArray() *bits.BitArray {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.partsBitArray.Copy()
}

func (ps *ECPartSet) Hash() []byte {
	if ps == nil {
		return merkle.HashFromByteSlices(nil)
	}
	return ps.hash
}

func (ps *ECPartSet) HashesTo(hash []byte) bool {
	if ps == nil {
		return false
	}
	return bytes.Equal(ps.hash, hash)
}

func (ps *ECPartSet) Count() uint32 {
	if ps == nil {
		return 0
	}
	return ps.count
}

func (ps *ECPartSet) ByteSize() int64 {
	if ps == nil {
		return 0
	}
	return ps.byteSize
}

func (ps *ECPartSet) Total() uint32 {
	if ps == nil {
		return 0
	}
	return ps.total
}

func (ps *ECPartSet) Parity() uint32 {
	if ps == nil {
		return 0
	}
	return ps.parity
}

// CONTRACT: part is validated using ValidateBasic.
// NOTE: part may be a parity part or data part, this does not matter.
func (ps *ECPartSet) AddPart(part *Part) (bool, error) {
	// TODO: remove this? would be preferable if this only returned (false, nil)
	// when it's a duplicate block part
	if ps == nil {
		return false, nil
	}

	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	// Invalid part index
	if part.Index >= ps.total {
		return false, ErrPartSetUnexpectedIndex
	}

	// If part already exists, return false.
	if ps.parts[part.Index] != nil {
		return false, nil
	}

	// The proof should be compatible with the number of parts.
	if part.Proof.Total != int64(ps.total) {
		return false, ErrPartSetInvalidProof
	}

	// Check hash proof
	if part.Proof.Verify(ps.Hash(), part.Bytes) != nil {
		return false, ErrPartSetInvalidProof
	}

	// Add part
	ps.parts[part.Index] = part
	ps.partsBitArray.SetIndex(int(part.Index), true)
	ps.count++
	ps.byteSize += int64(len(part.Bytes))
	return true, nil
}

func (ps *ECPartSet) GetPart(index int) *Part {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.parts[index]
}

func (ps *ECPartSet) IsComplete() bool {
	return ps.count == ps.total
}

// GetReader gets an io.Reader that will read part bytes from the PartSet.
//
// NOTE: this does not read parity data.
func (ps *ECPartSet) GetReader() io.Reader {
	if !ps.IsComplete() {
		panic("Cannot GetReader() on incomplete PartSet")
	}
	lastDataShardIdx := ps.total - ps.parity
	return NewPartSetReader(ps.parts[0:lastDataShardIdx])
}

// StringShort returns a short version of String.
//
// (Count of Total)
func (ps *ECPartSet) StringShort() string {
	if ps == nil {
		return "nil-PartSet"
	}
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return fmt.Sprintf("(%v of %v, %d parity)", ps.Count(), ps.Total(), ps.Parity())
}

func (ps *ECPartSet) MarshalJSON() ([]byte, error) {
	if ps == nil {
		return []byte("{}"), nil
	}

	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	return cmtjson.Marshal(struct {
		CountTotal    string         `json:"count/total"`
		PartsBitArray *bits.BitArray `json:"parts_bit_array"`
	}{
		fmt.Sprintf("%d/%d", ps.Count(), ps.Total()),
		ps.partsBitArray,
	})
}
