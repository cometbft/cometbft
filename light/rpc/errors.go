package rpc

import (
	"errors"
	"fmt"
	"regexp"

	cmtbytes "github.com/cometbft/cometbft/v2/libs/bytes"
)

var (
	ErrNegOrZeroHeight = errors.New("negative or zero height")
	ErrNoProofOps      = errors.New("no proof ops")
	ErrNilKeyPathFn    = errors.New("please configure Client with KeyPathFn option")
)

type ErrMissingStoreName struct {
	Path string
	Rex  *regexp.Regexp
}

func (e ErrMissingStoreName) Error() string {
	return fmt.Sprintf("can't find store name in %s using %s", e.Path, e.Rex)
}

type ErrResponseCode struct {
	Code uint32
}

func (e ErrResponseCode) Error() string {
	return fmt.Sprintf("err response code: %v", e.Code)
}

type ErrPageRange struct {
	Pages int
	Page  int
}

func (e ErrPageRange) Error() string {
	return fmt.Sprintf("page should be within [1, %d] range, given %d", e.Pages, e.Page)
}

type ErrNilBlockMeta struct {
	Index int
}

func (e ErrNilBlockMeta) Error() string {
	return fmt.Sprintf("nil block meta %d", e.Index)
}

type ErrParamHashMismatch struct {
	ConsensusParamsHash []byte
	ConsensusHash       cmtbytes.HexBytes
}

func (e ErrParamHashMismatch) Error() string {
	return fmt.Sprintf("params hash %X does not match trusted hash %X", e.ConsensusParamsHash, e.ConsensusHash)
}

type ErrLastResultMismatch struct {
	ResultHash     []byte
	LastResultHash cmtbytes.HexBytes
}

func (e ErrLastResultMismatch) Error() string {
	return fmt.Sprintf("last results %X does not match with trusted last results %X", e.ResultHash, e.LastResultHash)
}

type ErrPrimaryHeaderMismatch struct {
	PrimaryHeaderHash cmtbytes.HexBytes
	TrustedHeaderHash cmtbytes.HexBytes
}

func (e ErrPrimaryHeaderMismatch) Error() string {
	return fmt.Sprintf("primary header hash does not match trusted header hash. (%X != %X)", e.PrimaryHeaderHash, e.TrustedHeaderHash)
}

type ErrBlockHeaderMismatch struct {
	BlockHeader   cmtbytes.HexBytes
	TrustedHeader cmtbytes.HexBytes
}

func (e ErrBlockHeaderMismatch) Error() string {
	return fmt.Sprintf("block header %X does not match with trusted header %X", e.BlockHeader, e.TrustedHeader)
}

type ErrBlockMetaHeaderMismatch struct {
	BlockMetaHeader cmtbytes.HexBytes
	TrustedHeader   cmtbytes.HexBytes
}

func (e ErrBlockMetaHeaderMismatch) Error() string {
	return fmt.Sprintf("block meta header %X does not match with trusted header %X", e.BlockMetaHeader, e.TrustedHeader)
}

type ErrBlockIDMismatch struct {
	BlockID cmtbytes.HexBytes
	Block   cmtbytes.HexBytes
}

func (e ErrBlockIDMismatch) Error() string {
	return fmt.Sprintf("blockID %X does not match with block %X", e.BlockID, e.Block)
}

type ErrBuildMerkleKeyPath struct {
	Err error
}

func (e ErrBuildMerkleKeyPath) Error() string {
	return fmt.Sprintf("can't build merkle key path: %v", e.Err)
}

func (e ErrBuildMerkleKeyPath) Unwrap() error {
	return e.Err
}

type ErrVerifyValueProof struct {
	Err error
}

func (e ErrVerifyValueProof) Error() string {
	return fmt.Sprintf("verify value proof: %v", e.Err)
}

func (e ErrVerifyValueProof) Unwrap() error {
	return e.Err
}

type ErrVerifyAbsenceProof struct {
	Err error
}

func (e ErrVerifyAbsenceProof) Error() string {
	return fmt.Sprintf("verify absence proof: %v", e.Err)
}

func (e ErrVerifyAbsenceProof) Unwrap() error {
	return e.Err
}

type ErrGetLatestHeight struct {
	Err error
}

func (e ErrGetLatestHeight) Error() string {
	return fmt.Sprintf("can't get latest height: %v", e.Err)
}

func (e ErrGetLatestHeight) Unwrap() error {
	return e.Err
}

type ErrInvalidBlockMeta struct {
	I   int
	Err error
}

func (e ErrInvalidBlockMeta) Error() string {
	return fmt.Sprintf("invalid block meta %d: %v", e.I, e.Err)
}

func (e ErrInvalidBlockMeta) Unwrap() error {
	return e.Err
}

type ErrTrustedHeader struct {
	Height int64
	Err    error
}

func (e ErrTrustedHeader) Error() string {
	return fmt.Sprintf("trusted header %d: %v", e.Height, e.Err)
}

func (e ErrTrustedHeader) Unwrap() error {
	return e.Err
}

type ErrUpdateClient struct {
	Height int64
	Err    error
}

func (e ErrUpdateClient) Error() string {
	return fmt.Sprintf("failed to update light client to %d: %v", e.Height, e.Err)
}

func (e ErrUpdateClient) Unwrap() error {
	return e.Err
}
