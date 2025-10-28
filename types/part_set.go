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
	return int(total)
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

	// parity parts have no proofs
	if !part.IsParity {
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

	pb.Index = part.Index
	pb.Bytes = part.Bytes
	pb.IsParity = part.IsParity
	if !pb.IsParity {
		proof := part.Proof.ToProto()
		pb.Proof = *proof
	}

	return pb, nil
}

func PartFromProto(pb *cmtproto.Part) (*Part, error) {
	if pb == nil {
		return nil, errors.New("nil part")
	}

	part := new(Part)
	part.Index = pb.Index
	part.Bytes = pb.Bytes
	part.IsParity = pb.IsParity

	if !part.IsParity {
		proof, err := merkle.ProofFromProto(&pb.Proof)
		if err != nil {
			return nil, err
		}
		part.Proof = *proof
	}

	return part, part.ValidateBasic()
}

//-------------------------------------

type PartSetHeader struct {
	Total    uint32            `json:"total"`
	Hash     cmtbytes.HexBytes `json:"hash"`
	ByteSize uint64            `json:"byte_size,omitempty"`
}

// String returns a string representation of PartSetHeader.
//
// 1. total number of parts
// 2. number of total parts that are for parity
// 2. first 6 bytes of the hash
func (psh PartSetHeader) String() string {
	return fmt.Sprintf("%v:%X", psh.Total, cmtbytes.Fingerprint(psh.Hash))
}

func (psh PartSetHeader) IsZero() bool {
	return psh.Total == 0 && len(psh.Hash) == 0 && psh.ByteSize == 0
}

func (psh PartSetHeader) Equals(other PartSetHeader) bool {
	return psh.Total == other.Total && bytes.Equal(psh.Hash, other.Hash) && psh.ByteSize == other.ByteSize
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

	ppsh := cmtproto.PartSetHeader{
		Total: psh.Total,
		Hash:  psh.Hash,
	}
	if psh.ByteSize != 0 {
		ppsh.XByteSize = &cmtproto.PartSetHeader_ByteSize{ByteSize: psh.ByteSize}
	}

	return ppsh
}

// PartSetHeaderFromProto sets a protobuf PartSetHeader to the given pointer
func PartSetHeaderFromProto(ppsh *cmtproto.PartSetHeader) (*PartSetHeader, error) {
	if ppsh == nil {
		return nil, errors.New("nil PartSetHeader")
	}
	psh := new(PartSetHeader)
	psh.Total = ppsh.Total
	psh.Hash = ppsh.Hash
	psh.ByteSize = ppsh.GetByteSize()

	return psh, psh.ValidateBasic()
}

// ProtoPartSetHeaderIsZero is similar to the IsZero function for
// PartSetHeader, but for the Protobuf representation.
func ProtoPartSetHeaderIsZero(ppsh *cmtproto.PartSetHeader) bool {
	return ppsh.Total == 0 && len(ppsh.Hash) == 0 && ppsh.GetByteSize() == 0
}

//-------------------------------------

type PartSet struct {
	total uint32
	hash  []byte

	mtx           cmtsync.Mutex
	parts         []*Part
	partsBitArray *bits.BitArray
	count         uint32

	// a count of the total size (in bytes). Used to ensure that the part set
	// doesn't exceed the maximum block bytes. When building a PartSet part by
	// part (via AddPart), byteSize will grow.
	byteSize int64

	// The total byte size of all parts. As reported by the header or from
	// constructing via data. This value is static.
	totalByteSize int64

	// parity is the amount of ParityPart's in this PartSet
	parity         uint32
	parityCount    uint32
	parityBitArray *bits.BitArray
	parityParts    []*Part
}

// NewPartSetFromData returns an immutable, full PartSet from the data bytes.
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

// NewRSPartSetFromData returns an immutable, full PartSet from the data bytes.
// The data bytes are split into "partSize" chunks, and merkle tree computed.
// The PartSet will be ReedSolomon encoded, meaning there will be parity parts
// computed, so the PartSet can be recomputed from the parity parts, even if
// some Part's are missing.
//
// CONTRACT: partSize is greater than zero.
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

	parityShards := NumParityParts(dataShards)
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

	for i := dataShards; i < dataShards+uint32(parityShards); i++ {
		parityParts[i-dataShards] = &Part{
			Index:    i - dataShards,
			Bytes:    shards[i],
			IsParity: true,
		}
	}
	parityBitArray := bits.NewBitArrayFromFn(int(parityShards), func(int) bool { return true })

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
		byteSize:       int64(len(data)),
		totalByteSize:  int64(len(data)),
		parity:         uint32(parityShards),
		parityBitArray: parityBitArray,
		parityParts:    parityParts,
		parityCount:    uint32(parityShards),
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
		totalByteSize: int64(len(data)),
		parity:        0,
	}
}

// NewPartSetFromHeader returns an empty PartSet ready to be populated.
func NewPartSetFromHeader(header PartSetHeader) *PartSet {
	parity := NumParityParts(header.Total)
	return &PartSet{
		total:          header.Total,
		hash:           header.Hash,
		parts:          make([]*Part, header.Total),
		partsBitArray:  bits.NewBitArray(parity),
		parityParts:    make([]*Part, parity),
		parityBitArray: bits.NewBitArray(parity),
		parity:         uint32(parity),
		count:          0,
		byteSize:       0,
		totalByteSize:  int64(header.ByteSize),
		parityCount:    0,
	}
}

func (ps *PartSet) Header() PartSetHeader {
	if ps == nil {
		return PartSetHeader{}
	}
	return PartSetHeader{
		Total:    ps.total,
		Hash:     ps.hash,
		ByteSize: uint64(ps.totalByteSize),
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

func (ps *PartSet) ParityBitArray() *bits.BitArray {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.parityBitArray.Copy()
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

func (ps *PartSet) ParityCount() uint32 {
	if ps == nil {
		return 0
	}
	return ps.parityCount
}

func (ps *PartSet) TryReconstruct() (bool, error) {
	shards := ps.ToShards()

	enc, err := reedsolomon.New(int(ps.total), int(ps.parity))
	if err != nil {
		return false, fmt.Errorf("creating reedsolomon encoder with %d data shards and %d parity shards: %w", ps.total, ps.parity, err)
	}

	err = enc.ReconstructData(shards)
	if err == reedsolomon.ErrTooFewShards {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reconstructing data: %w", err)
	}

	for i := 0; i < int(ps.total); i++ {
		if ps.parts[i] != nil {
			continue
		}

		ps.parts[i] = &Part{
			Index: uint32(i),
			Bytes: shards[i],
		}
	}
	ps.count = ps.total
	return true, nil
}

func (ps *PartSet) ToShards() [][]byte {
	totalShards := ps.total + uint32(ps.parity)
	shards := make([][]byte, totalShards)

	for i := 0; i < int(ps.total); i++ {
		part := ps.parts[i]
		if part == nil {
			continue
		}
		shards[part.Index] = part.Bytes
	}

	for i := 0; i < int(ps.parity); i++ {
		parityPart := ps.parityParts[i]
		if parityPart == nil {
			continue
		}
		shards[parityPart.Index+ps.total] = parityPart.Bytes
	}

	return shards
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
	ps.parityBitArray.SetIndex(int(part.Index), true)
	ps.parityCount++
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
	return NewPartSetReader(ps.parts, ps.totalByteSize)
}

type PartSetReader struct {
	i         int
	parts     []*Part
	reader    *bytes.Reader
	MaxBytes  int64
	bytesRead int64
}

func NewPartSetReader(parts []*Part, totalSize int64) *PartSetReader {
	return &PartSetReader{
		i:         0,
		parts:     parts,
		reader:    bytes.NewReader(parts[0].Bytes),
		MaxBytes:  totalSize,
		bytesRead: 0,
	}
}

func (psr *PartSetReader) Read(p []byte) (n int, err error) {
	// Check if we've already read the maximum allowed bytes
	if psr.bytesRead >= psr.MaxBytes {
		return 0, io.EOF
	}

	// Limit the read to not exceed MaxBytes
	remainingBytes := psr.MaxBytes - psr.bytesRead
	if int64(len(p)) > remainingBytes {
		p = p[:remainingBytes]
	}

	readerLen := psr.reader.Len()
	if readerLen >= len(p) {
		n, err = psr.reader.Read(p)
		psr.bytesRead += int64(n)
		return n, err
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
