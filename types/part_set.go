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
	cmtmath "github.com/cometbft/cometbft/libs/math"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/klauspost/reedsolomon"
)

var (
	ErrPartSetUnexpectedIndex = errors.New("error part set unexpected index")
	ErrPartSetInvalidProof    = errors.New("error part set invalid proof")
	ErrPartTooBig             = errors.New("error part size too big")
	ErrPartInvalidSize        = errors.New("error inner part with invalid size")
	ErrPartInvalidEncoding    = errors.New("error part has invalid encoding")
)

// ErrInvalidPart is an error type for invalid parts.
type ErrInvalidPart struct {
	Reason error
}

func (e ErrInvalidPart) Error() string {
	return fmt.Sprintf("invalid part: %v", e.Reason)
}

func (e ErrInvalidPart) Unwrap() error {
	return e.Reason
}

func NumParityParts(total uint32) int {
	return int(total / 2)
}

type Part struct {
	Index    uint32            `json:"index"`
	Bytes    cmtbytes.HexBytes `json:"bytes"`
	Proof    merkle.Proof      `json:"proof"`
	IsParity bool              `json:"is_parity"`
}

type PartEncoding byte

var (
	None        PartEncoding = 0
	ReedSolomon PartEncoding = 1
)

// ValidateBasic performs basic validation.
func (part *Part) ValidateBasic() error {
	if len(part.Bytes) > int(BlockPartSizeBytes) {
		return ErrPartTooBig
	}
	// All parts except the last one should have the same constant size.
	if int64(part.Index) < part.Proof.Total-1 && len(part.Bytes) != int(BlockPartSizeBytes) {
		return ErrPartInvalidSize
	}
	if int64(part.Index) != part.Proof.Index {
		return ErrInvalidPart{Reason: fmt.Errorf("part index %d != proof index %d", part.Index, part.Proof.Index)}
	}
	if err := part.Proof.ValidateBasic(); err != nil {
		return ErrInvalidPart{Reason: fmt.Errorf("wrong Proof: %w", err)}
	}
	return nil
}

// String returns a string representation of Part.
//
// See StringIndented.
func (part *Part) String() string {
	return part.StringIndented("")
}

// StringIndented returns an indented Part.
//
// See merkle.Proof#StringIndented
func (part *Part) StringIndented(indent string) string {
	return fmt.Sprintf(`Part{#%v
%s  Bytes: %X...
%s  Proof: %v
%s}`,
		part.Index,
		indent, cmtbytes.Fingerprint(part.Bytes),
		indent, part.Proof.StringIndented(indent+"  "),
		indent)
}

func (part *Part) ToProto() (*cmtproto.Part, error) {
	if part == nil {
		return nil, errors.New("nil part")
	}
	pb := new(cmtproto.Part)
	proof := part.Proof.ToProto()

	pb.Index = part.Index
	pb.Bytes = part.Bytes
	pb.Proof = *proof

	return pb, nil
}

func PartFromProto(pb *cmtproto.Part) (*Part, error) {
	if pb == nil {
		return nil, errors.New("nil part")
	}

	part := new(Part)
	proof, err := merkle.ProofFromProto(&pb.Proof)
	if err != nil {
		return nil, err
	}
	part.Index = pb.Index
	part.Bytes = pb.Bytes
	part.Proof = *proof

	return part, part.ValidateBasic()
}

//-------------------------------------

type PartSetHeader struct {
	Total  uint32            `json:"total"`
	Hash   cmtbytes.HexBytes `json:"hash"`
	Parity uint32            `json:"parity"`
}

// String returns a string representation of PartSetHeader.
//
// 1. total number of parts
// 2. number of total parts that are for parity
// 2. first 6 bytes of the hash
func (psh PartSetHeader) String() string {
	return fmt.Sprintf("%v:%v:%X", psh.Total, psh.Parity, cmtbytes.Fingerprint(psh.Hash))
}

func (psh PartSetHeader) IsZero() bool {
	return psh.Total == 0 && len(psh.Hash) == 0 && psh.Parity == 0
}

func (psh PartSetHeader) Equals(other PartSetHeader) bool {
	return psh.Total == other.Total && bytes.Equal(psh.Hash, other.Hash) && psh.Parity == other.Parity
}

// ValidateBasic performs basic validation.
func (psh PartSetHeader) ValidateBasic() error {
	// Hash can be empty in case of POLBlockID.PartSetHeader in Proposal.
	if err := ValidateHash(psh.Hash); err != nil {
		return fmt.Errorf("wrong Hash: %w", err)
	}
	return nil
}

// ToProto converts PartSetHeader to protobuf
func (psh *PartSetHeader) ToProto() cmtproto.PartSetHeader {
	if psh == nil {
		return cmtproto.PartSetHeader{}
	}

	return cmtproto.PartSetHeader{
		Total:  psh.Total,
		Hash:   psh.Hash,
		Parity: psh.Parity,
	}
}

// PartSetHeaderFromProto sets a protobuf PartSetHeader to the given pointer
func PartSetHeaderFromProto(ppsh *cmtproto.PartSetHeader) (*PartSetHeader, error) {
	if ppsh == nil {
		return nil, errors.New("nil PartSetHeader")
	}
	psh := new(PartSetHeader)
	psh.Total = ppsh.Total
	psh.Hash = ppsh.Hash
	psh.Parity = ppsh.Parity

	return psh, psh.ValidateBasic()
}

// ProtoPartSetHeaderIsZero is similar to the IsZero function for
// PartSetHeader, but for the Protobuf representation.
func ProtoPartSetHeaderIsZero(ppsh *cmtproto.PartSetHeader) bool {
	return ppsh.Total == 0 && len(ppsh.Hash) == 0 && ppsh.Parity == 0
}

//-------------------------------------

type PartSet struct {
	total uint32
	hash  []byte

	mtx           cmtsync.Mutex
	parts         []*Part
	partsBitArray *bits.BitArray
	count         uint32
	// a count of the total size (in bytes). Used to ensure that the
	// part set doesn't exceed the maximum block bytes
	byteSize int64

	// parity is the amount of ParityPart's in this PartSet
	parity      uint32
	parityParts []*Part
}

// NewtPartSetFromData returns an immutable, full PartSet from the data bytes.
// The data bytes are split into "partSize" chunks, and merkle tree computed.
// CONTRACT: partSize is greater than zero.
func NewPartSetFromDataWithEncoding(data []byte, partSize uint32, encoding PartEncoding) (*PartSet, error) {
	switch encoding {
	case None:
		return NewPartSetFromData(data, partSize), nil
	case ReedSolomon:
		return NewRSPartSetFromData(data, partSize)
	default:
		return nil, fmt.Errorf("unknown encoding %d", encoding)
	}
}

func NewRSPartSetFromData(data []byte, partSize uint32) (*PartSet, error) {
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
	parts := make([]*Part, dataShards)
	partsBytes := make([][]byte, dataShards)

	parityShards := dataShards / 2
	parityParts := make([]*Part, parityShards)

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

	expectedLen := int(dataShards) + int(parityShards)
	if len(shards) != expectedLen {
		return nil, fmt.Errorf("invalid number of shards after encoding, expected %d but got %d", expectedLen, len(shards))
	}

	for i := uint32(0); i < dataShards; i++ {
		parts[i] = &Part{
			Index: i,
			Bytes: shards[i],
		}
		partsBytes[i] = shards[i]
	}
	root, proofs := merkle.ProofsFromByteSlices(partsBytes)
	for i := uint32(0); i < dataShards; i++ {
		parts[i].Proof = *proofs[i]
	}
	partsBitArray := bits.NewBitArrayFromFn(int(dataShards), func(int) bool { return true })

	for i := dataShards; i < parityShards; i++ {
		parityParts[i] = &Part{
			Index:    i,
			Bytes:    shards[i],
			IsParity: true,
		}
	}

	return &PartSet{
		total:         dataShards,
		hash:          root,
		parts:         parts,
		partsBitArray: partsBitArray,
		count:         dataShards,
		// all shards must have the same length, so this is ok to calc the
		// total size.
		//
		// NOTE: this only counts the size of the data shards, not
		// parity shards.
		byteSize:    int64(len(data)),
		parity:      parityShards,
		parityParts: parityParts,
	}, nil
}

func NewPartSetFromData(data []byte, partSize uint32) *PartSet {
	// divide data into parts of size `partSize`
	total := (uint32(len(data)) + partSize - 1) / partSize
	parts := make([]*Part, total)
	partsBytes := make([][]byte, total)
	for i := uint32(0); i < total; i++ {
		part := &Part{
			Index: i,
			Bytes: data[i*partSize : cmtmath.MinInt(len(data), int((i+1)*partSize))],
		}
		parts[i] = part
		partsBytes[i] = part.Bytes
	}
	// Compute merkle proofs
	root, proofs := merkle.ProofsFromByteSlices(partsBytes)
	for i := uint32(0); i < total; i++ {
		parts[i].Proof = *proofs[i]
	}
	partsBitArray := bits.NewBitArrayFromFn(int(total), func(int) bool { return true })
	return &PartSet{
		total:         total,
		hash:          root,
		parts:         parts,
		partsBitArray: partsBitArray,
		count:         total,
		byteSize:      int64(len(data)),
		parity:        0,
	}
}

// NewPartSetFromHeader returns an empty PartSet ready to be populated.
func NewPartSetFromHeader(header PartSetHeader) *PartSet {
	return &PartSet{
		total:         header.Total,
		hash:          header.Hash,
		parts:         make([]*Part, header.Total),
		partsBitArray: bits.NewBitArray(int(header.Total)),
		count:         0,
		byteSize:      0,
	}
}

func (ps *PartSet) Header() PartSetHeader {
	if ps == nil {
		return PartSetHeader{}
	}
	return PartSetHeader{
		Total:  ps.total,
		Parity: ps.parity,
		Hash:   ps.hash,
	}
}

func (ps *PartSet) HasHeader(header PartSetHeader) bool {
	if ps == nil {
		return false
	}
	return ps.Header().Equals(header)
}

func (ps *PartSet) BitArray() *bits.BitArray {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.partsBitArray.Copy()
}

func (ps *PartSet) Hash() []byte {
	if ps == nil {
		return merkle.HashFromByteSlices(nil)
	}
	return ps.hash
}

func (ps *PartSet) HashesTo(hash []byte) bool {
	if ps == nil {
		return false
	}
	return bytes.Equal(ps.hash, hash)
}

func (ps *PartSet) Count() uint32 {
	if ps == nil {
		return 0
	}
	return ps.count
}

func (ps *PartSet) ByteSize() int64 {
	if ps == nil {
		return 0
	}
	return ps.byteSize
}

func (ps *PartSet) Total() uint32 {
	if ps == nil {
		return 0
	}
	return ps.total
}

func (ps *PartSet) Parity() uint32 {
	if ps == nil {
		return 0
	}
	return ps.parity
}

// CONTRACT: part is validated using ValidateBasic.
func (ps *PartSet) AddPart(part *Part) (bool, error) {
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
	// parity parts do not count towards total byteSize
	ps.byteSize += int64(len(part.Bytes))

	return true, nil
}

func (ps *PartSet) AddParityPart(part *Part) (bool, error) {
	if ps == nil {
		return false, nil
	}

	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	// Invalid part index
	if part.Index >= ps.parity {
		return false, ErrPartSetUnexpectedIndex
	}

	// If part already exists, return false.
	if ps.parityParts[part.Index] != nil {
		return false, nil
	}

	// Add part
	ps.parityParts[part.Index] = part
	return true, nil
}

func (ps *PartSet) GetPart(index int) *Part {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.parts[index]
}

func (ps *PartSet) GetParityPart(index int) *Part {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.parityParts[index]
}

func (ps *PartSet) IsComplete() bool {
	return ps.count == ps.total
}

func (ps *PartSet) GetReader() io.Reader {
	if !ps.IsComplete() {
		panic("Cannot GetReader() on incomplete PartSet")
	}
	return NewPartSetReader(ps.parts)
}

type PartSetReader struct {
	i      int
	parts  []*Part
	reader *bytes.Reader
}

func NewPartSetReader(parts []*Part) *PartSetReader {
	return &PartSetReader{
		i:      0,
		parts:  parts,
		reader: bytes.NewReader(parts[0].Bytes),
	}
}

func (psr *PartSetReader) Read(p []byte) (n int, err error) {
	readerLen := psr.reader.Len()
	if readerLen >= len(p) {
		return psr.reader.Read(p)
	} else if readerLen > 0 {
		n1, err := psr.Read(p[:readerLen])
		if err != nil {
			return n1, err
		}
		n2, err := psr.Read(p[readerLen:])
		return n1 + n2, err
	}

	psr.i++
	if psr.i >= len(psr.parts) {
		return 0, io.EOF
	}
	psr.reader = bytes.NewReader(psr.parts[psr.i].Bytes)
	return psr.Read(p)
}

// StringShort returns a short version of String.
//
// (Count of Total)
func (ps *PartSet) StringShort() string {
	if ps == nil {
		return "nil-PartSet"
	}
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return fmt.Sprintf("(%v of %v with %d parity)", ps.Count(), ps.Total(), ps.Parity())
}

func (ps *PartSet) MarshalJSON() ([]byte, error) {
	if ps == nil {
		return []byte("{}"), nil
	}

	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	return cmtjson.Marshal(struct {
		CountTotal    string         `json:"count/total/parity"`
		PartsBitArray *bits.BitArray `json:"parts_bit_array"`
	}{
		fmt.Sprintf("%d/%d/%d", ps.Count(), ps.Total(), ps.Parity()),
		ps.partsBitArray,
	})
}
