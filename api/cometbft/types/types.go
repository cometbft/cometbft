package types

import (
	v1 "github.com/cometbft/cometbft/api/cometbft/types/v1"
	v3 "github.com/cometbft/cometbft/api/cometbft/types/v3"
)

type BlockID = v1.BlockID
type BlockMeta = v1.BlockMeta
type Commit = v1.Commit
type CommitSig = v1.CommitSig
type ExtendedCommit = v3.ExtendedCommit
type ExtendedCommitSig = v3.ExtendedCommitSig
type Header = v1.Header
type LightBlock = v1.LightBlock
type Part = v1.Part
type PartSetHeader = v1.PartSetHeader
type Proposal = v1.Proposal
type SignedHeader = v1.SignedHeader
type TxProof = v1.TxProof
type Validator = v1.Validator
type Vote = v3.Vote
