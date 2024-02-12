// Package slot is generated by gogll. Do not edit.
package slot

import (
	"bytes"
	"fmt"

	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/grammar-auto/parser/symbols"
)

type Label int

const (
	ApplyChunk0R0 Label = iota
	ApplyChunk0R1
	ApplyChunks0R0
	ApplyChunks0R1
	ApplyChunks1R0
	ApplyChunks1R1
	ApplyChunks1R2
	CleanStart0R0
	CleanStart0R1
	CleanStart0R2
	CleanStart1R0
	CleanStart1R1
	CleanStart1R2
	Commit0R0
	Commit0R1
	ConsensusExec0R0
	ConsensusExec0R1
	ConsensusHeight0R0
	ConsensusHeight0R1
	ConsensusHeight0R2
	ConsensusHeight0R3
	ConsensusHeight1R0
	ConsensusHeight1R1
	ConsensusHeight1R2
	ConsensusHeights0R0
	ConsensusHeights0R1
	ConsensusHeights1R0
	ConsensusHeights1R1
	ConsensusHeights1R2
	ConsensusRound0R0
	ConsensusRound0R1
	ConsensusRound1R0
	ConsensusRound1R1
	ConsensusRounds0R0
	ConsensusRounds0R1
	ConsensusRounds1R0
	ConsensusRounds1R1
	ConsensusRounds1R2
	Extend0R0
	Extend0R1
	Extend1R0
	Extend1R1
	Extend1R2
	Extend2R0
	Extend2R1
	Extend2R2
	Extend3R0
	Extend3R1
	Extend3R2
	Extend3R3
	ExtendVote0R0
	ExtendVote0R1
	FinalizeBlock0R0
	FinalizeBlock0R1
	GotVote0R0
	GotVote0R1
	GotVotes0R0
	GotVotes0R1
	GotVotes1R0
	GotVotes1R1
	GotVotes1R2
	InitChain0R0
	InitChain0R1
	NonProposer0R0
	NonProposer0R1
	NonProposer1R0
	NonProposer1R1
	NonProposer2R0
	NonProposer2R1
	NonProposer3R0
	NonProposer3R1
	NonProposer3R2
	NonProposer4R0
	NonProposer4R1
	NonProposer4R2
	NonProposer5R0
	NonProposer5R1
	NonProposer5R2
	NonProposer6R0
	NonProposer6R1
	NonProposer6R2
	NonProposer6R3
	OfferSnapshot0R0
	OfferSnapshot0R1
	PrepareProposal0R0
	PrepareProposal0R1
	ProcessProposal0R0
	ProcessProposal0R1
	Proposer0R0
	Proposer0R1
	Proposer1R0
	Proposer1R1
	Proposer2R0
	Proposer2R1
	Proposer3R0
	Proposer3R1
	Proposer3R2
	Proposer4R0
	Proposer4R1
	Proposer4R2
	Proposer5R0
	Proposer5R1
	Proposer5R2
	Proposer6R0
	Proposer6R1
	Proposer6R2
	Proposer6R3
	ProposerSimple0R0
	ProposerSimple0R1
	ProposerSimple1R0
	ProposerSimple1R1
	ProposerSimple1R2
	Recovery0R0
	Recovery0R1
	Recovery0R2
	Recovery1R0
	Recovery1R1
	Start0R0
	Start0R1
	Start1R0
	Start1R1
	StateSync0R0
	StateSync0R1
	StateSync0R2
	StateSync1R0
	StateSync1R1
	StateSyncAttempt0R0
	StateSyncAttempt0R1
	StateSyncAttempt0R2
	StateSyncAttempt1R0
	StateSyncAttempt1R1
	StateSyncAttempts0R0
	StateSyncAttempts0R1
	StateSyncAttempts1R0
	StateSyncAttempts1R1
	StateSyncAttempts1R2
	SuccessSync0R0
	SuccessSync0R1
	SuccessSync0R2
)

type Slot struct {
	NT      symbols.NT
	Alt     int
	Pos     int
	Symbols symbols.Symbols
	Label   Label
}

type Index struct {
	NT  symbols.NT
	Alt int
	Pos int
}

func GetAlternates(nt symbols.NT) []Label {
	alts, exist := alternates[nt]
	if !exist {
		panic(fmt.Sprintf("Invalid NT %s", nt))
	}
	return alts
}

func GetLabel(nt symbols.NT, alt, pos int) Label {
	l, exist := slotIndex[Index{nt, alt, pos}]
	if exist {
		return l
	}
	panic(fmt.Sprintf("Error: no slot label for NT=%s, alt=%d, pos=%d", nt, alt, pos))
}

func (l Label) EoR() bool {
	return l.Slot().EoR()
}

func (l Label) Head() symbols.NT {
	return l.Slot().NT
}

func (l Label) Index() Index {
	s := l.Slot()
	return Index{s.NT, s.Alt, s.Pos}
}

func (l Label) Alternate() int {
	return l.Slot().Alt
}

func (l Label) Pos() int {
	return l.Slot().Pos
}

func (l Label) Slot() *Slot {
	s, exist := slots[l]
	if !exist {
		panic(fmt.Sprintf("Invalid slot label %d", l))
	}
	return s
}

func (l Label) String() string {
	return l.Slot().String()
}

func (l Label) Symbols() symbols.Symbols {
	return l.Slot().Symbols
}

func (s *Slot) EoR() bool {
	return s.Pos >= len(s.Symbols)
}

func (s *Slot) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%s : ", s.NT)
	for i, sym := range s.Symbols {
		if i == s.Pos {
			fmt.Fprintf(buf, "∙")
		}
		fmt.Fprintf(buf, "%s ", sym)
	}
	if s.Pos >= len(s.Symbols) {
		fmt.Fprintf(buf, "∙")
	}
	return buf.String()
}

var slots = map[Label]*Slot{
	ApplyChunk0R0: {
		symbols.NT_ApplyChunk, 0, 0,
		symbols.Symbols{
			symbols.T_0,
		},
		ApplyChunk0R0,
	},
	ApplyChunk0R1: {
		symbols.NT_ApplyChunk, 0, 1,
		symbols.Symbols{
			symbols.T_0,
		},
		ApplyChunk0R1,
	},
	ApplyChunks0R0: {
		symbols.NT_ApplyChunks, 0, 0,
		symbols.Symbols{
			symbols.NT_ApplyChunk,
		},
		ApplyChunks0R0,
	},
	ApplyChunks0R1: {
		symbols.NT_ApplyChunks, 0, 1,
		symbols.Symbols{
			symbols.NT_ApplyChunk,
		},
		ApplyChunks0R1,
	},
	ApplyChunks1R0: {
		symbols.NT_ApplyChunks, 1, 0,
		symbols.Symbols{
			symbols.NT_ApplyChunk,
			symbols.NT_ApplyChunks,
		},
		ApplyChunks1R0,
	},
	ApplyChunks1R1: {
		symbols.NT_ApplyChunks, 1, 1,
		symbols.Symbols{
			symbols.NT_ApplyChunk,
			symbols.NT_ApplyChunks,
		},
		ApplyChunks1R1,
	},
	ApplyChunks1R2: {
		symbols.NT_ApplyChunks, 1, 2,
		symbols.Symbols{
			symbols.NT_ApplyChunk,
			symbols.NT_ApplyChunks,
		},
		ApplyChunks1R2,
	},
	CleanStart0R0: {
		symbols.NT_CleanStart, 0, 0,
		symbols.Symbols{
			symbols.NT_InitChain,
			symbols.NT_ConsensusExec,
		},
		CleanStart0R0,
	},
	CleanStart0R1: {
		symbols.NT_CleanStart, 0, 1,
		symbols.Symbols{
			symbols.NT_InitChain,
			symbols.NT_ConsensusExec,
		},
		CleanStart0R1,
	},
	CleanStart0R2: {
		symbols.NT_CleanStart, 0, 2,
		symbols.Symbols{
			symbols.NT_InitChain,
			symbols.NT_ConsensusExec,
		},
		CleanStart0R2,
	},
	CleanStart1R0: {
		symbols.NT_CleanStart, 1, 0,
		symbols.Symbols{
			symbols.NT_StateSync,
			symbols.NT_ConsensusExec,
		},
		CleanStart1R0,
	},
	CleanStart1R1: {
		symbols.NT_CleanStart, 1, 1,
		symbols.Symbols{
			symbols.NT_StateSync,
			symbols.NT_ConsensusExec,
		},
		CleanStart1R1,
	},
	CleanStart1R2: {
		symbols.NT_CleanStart, 1, 2,
		symbols.Symbols{
			symbols.NT_StateSync,
			symbols.NT_ConsensusExec,
		},
		CleanStart1R2,
	},
	Commit0R0: {
		symbols.NT_Commit, 0, 0,
		symbols.Symbols{
			symbols.T_1,
		},
		Commit0R0,
	},
	Commit0R1: {
		symbols.NT_Commit, 0, 1,
		symbols.Symbols{
			symbols.T_1,
		},
		Commit0R1,
	},
	ConsensusExec0R0: {
		symbols.NT_ConsensusExec, 0, 0,
		symbols.Symbols{
			symbols.NT_ConsensusHeights,
		},
		ConsensusExec0R0,
	},
	ConsensusExec0R1: {
		symbols.NT_ConsensusExec, 0, 1,
		symbols.Symbols{
			symbols.NT_ConsensusHeights,
		},
		ConsensusExec0R1,
	},
	ConsensusHeight0R0: {
		symbols.NT_ConsensusHeight, 0, 0,
		symbols.Symbols{
			symbols.NT_ConsensusRounds,
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight0R0,
	},
	ConsensusHeight0R1: {
		symbols.NT_ConsensusHeight, 0, 1,
		symbols.Symbols{
			symbols.NT_ConsensusRounds,
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight0R1,
	},
	ConsensusHeight0R2: {
		symbols.NT_ConsensusHeight, 0, 2,
		symbols.Symbols{
			symbols.NT_ConsensusRounds,
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight0R2,
	},
	ConsensusHeight0R3: {
		symbols.NT_ConsensusHeight, 0, 3,
		symbols.Symbols{
			symbols.NT_ConsensusRounds,
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight0R3,
	},
	ConsensusHeight1R0: {
		symbols.NT_ConsensusHeight, 1, 0,
		symbols.Symbols{
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight1R0,
	},
	ConsensusHeight1R1: {
		symbols.NT_ConsensusHeight, 1, 1,
		symbols.Symbols{
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight1R1,
	},
	ConsensusHeight1R2: {
		symbols.NT_ConsensusHeight, 1, 2,
		symbols.Symbols{
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight1R2,
	},
	ConsensusHeights0R0: {
		symbols.NT_ConsensusHeights, 0, 0,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
		},
		ConsensusHeights0R0,
	},
	ConsensusHeights0R1: {
		symbols.NT_ConsensusHeights, 0, 1,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
		},
		ConsensusHeights0R1,
	},
	ConsensusHeights1R0: {
		symbols.NT_ConsensusHeights, 1, 0,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
			symbols.NT_ConsensusHeights,
		},
		ConsensusHeights1R0,
	},
	ConsensusHeights1R1: {
		symbols.NT_ConsensusHeights, 1, 1,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
			symbols.NT_ConsensusHeights,
		},
		ConsensusHeights1R1,
	},
	ConsensusHeights1R2: {
		symbols.NT_ConsensusHeights, 1, 2,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
			symbols.NT_ConsensusHeights,
		},
		ConsensusHeights1R2,
	},
	ConsensusRound0R0: {
		symbols.NT_ConsensusRound, 0, 0,
		symbols.Symbols{
			symbols.NT_Proposer,
		},
		ConsensusRound0R0,
	},
	ConsensusRound0R1: {
		symbols.NT_ConsensusRound, 0, 1,
		symbols.Symbols{
			symbols.NT_Proposer,
		},
		ConsensusRound0R1,
	},
	ConsensusRound1R0: {
		symbols.NT_ConsensusRound, 1, 0,
		symbols.Symbols{
			symbols.NT_NonProposer,
		},
		ConsensusRound1R0,
	},
	ConsensusRound1R1: {
		symbols.NT_ConsensusRound, 1, 1,
		symbols.Symbols{
			symbols.NT_NonProposer,
		},
		ConsensusRound1R1,
	},
	ConsensusRounds0R0: {
		symbols.NT_ConsensusRounds, 0, 0,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
		},
		ConsensusRounds0R0,
	},
	ConsensusRounds0R1: {
		symbols.NT_ConsensusRounds, 0, 1,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
		},
		ConsensusRounds0R1,
	},
	ConsensusRounds1R0: {
		symbols.NT_ConsensusRounds, 1, 0,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
			symbols.NT_ConsensusRounds,
		},
		ConsensusRounds1R0,
	},
	ConsensusRounds1R1: {
		symbols.NT_ConsensusRounds, 1, 1,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
			symbols.NT_ConsensusRounds,
		},
		ConsensusRounds1R1,
	},
	ConsensusRounds1R2: {
		symbols.NT_ConsensusRounds, 1, 2,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
			symbols.NT_ConsensusRounds,
		},
		ConsensusRounds1R2,
	},
	Extend0R0: {
		symbols.NT_Extend, 0, 0,
		symbols.Symbols{
			symbols.NT_ExtendVote,
		},
		Extend0R0,
	},
	Extend0R1: {
		symbols.NT_Extend, 0, 1,
		symbols.Symbols{
			symbols.NT_ExtendVote,
		},
		Extend0R1,
	},
	Extend1R0: {
		symbols.NT_Extend, 1, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ExtendVote,
		},
		Extend1R0,
	},
	Extend1R1: {
		symbols.NT_Extend, 1, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ExtendVote,
		},
		Extend1R1,
	},
	Extend1R2: {
		symbols.NT_Extend, 1, 2,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ExtendVote,
		},
		Extend1R2,
	},
	Extend2R0: {
		symbols.NT_Extend, 2, 0,
		symbols.Symbols{
			symbols.NT_ExtendVote,
			symbols.NT_GotVotes,
		},
		Extend2R0,
	},
	Extend2R1: {
		symbols.NT_Extend, 2, 1,
		symbols.Symbols{
			symbols.NT_ExtendVote,
			symbols.NT_GotVotes,
		},
		Extend2R1,
	},
	Extend2R2: {
		symbols.NT_Extend, 2, 2,
		symbols.Symbols{
			symbols.NT_ExtendVote,
			symbols.NT_GotVotes,
		},
		Extend2R2,
	},
	Extend3R0: {
		symbols.NT_Extend, 3, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ExtendVote,
			symbols.NT_GotVotes,
		},
		Extend3R0,
	},
	Extend3R1: {
		symbols.NT_Extend, 3, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ExtendVote,
			symbols.NT_GotVotes,
		},
		Extend3R1,
	},
	Extend3R2: {
		symbols.NT_Extend, 3, 2,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ExtendVote,
			symbols.NT_GotVotes,
		},
		Extend3R2,
	},
	Extend3R3: {
		symbols.NT_Extend, 3, 3,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ExtendVote,
			symbols.NT_GotVotes,
		},
		Extend3R3,
	},
	ExtendVote0R0: {
		symbols.NT_ExtendVote, 0, 0,
		symbols.Symbols{
			symbols.T_2,
		},
		ExtendVote0R0,
	},
	ExtendVote0R1: {
		symbols.NT_ExtendVote, 0, 1,
		symbols.Symbols{
			symbols.T_2,
		},
		ExtendVote0R1,
	},
	FinalizeBlock0R0: {
		symbols.NT_FinalizeBlock, 0, 0,
		symbols.Symbols{
			symbols.T_3,
		},
		FinalizeBlock0R0,
	},
	FinalizeBlock0R1: {
		symbols.NT_FinalizeBlock, 0, 1,
		symbols.Symbols{
			symbols.T_3,
		},
		FinalizeBlock0R1,
	},
	GotVote0R0: {
		symbols.NT_GotVote, 0, 0,
		symbols.Symbols{
			symbols.T_8,
		},
		GotVote0R0,
	},
	GotVote0R1: {
		symbols.NT_GotVote, 0, 1,
		symbols.Symbols{
			symbols.T_8,
		},
		GotVote0R1,
	},
	GotVotes0R0: {
		symbols.NT_GotVotes, 0, 0,
		symbols.Symbols{
			symbols.NT_GotVote,
		},
		GotVotes0R0,
	},
	GotVotes0R1: {
		symbols.NT_GotVotes, 0, 1,
		symbols.Symbols{
			symbols.NT_GotVote,
		},
		GotVotes0R1,
	},
	GotVotes1R0: {
		symbols.NT_GotVotes, 1, 0,
		symbols.Symbols{
			symbols.NT_GotVote,
			symbols.NT_GotVotes,
		},
		GotVotes1R0,
	},
	GotVotes1R1: {
		symbols.NT_GotVotes, 1, 1,
		symbols.Symbols{
			symbols.NT_GotVote,
			symbols.NT_GotVotes,
		},
		GotVotes1R1,
	},
	GotVotes1R2: {
		symbols.NT_GotVotes, 1, 2,
		symbols.Symbols{
			symbols.NT_GotVote,
			symbols.NT_GotVotes,
		},
		GotVotes1R2,
	},
	InitChain0R0: {
		symbols.NT_InitChain, 0, 0,
		symbols.Symbols{
			symbols.T_4,
		},
		InitChain0R0,
	},
	InitChain0R1: {
		symbols.NT_InitChain, 0, 1,
		symbols.Symbols{
			symbols.T_4,
		},
		InitChain0R1,
	},
	NonProposer0R0: {
		symbols.NT_NonProposer, 0, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
		},
		NonProposer0R0,
	},
	NonProposer0R1: {
		symbols.NT_NonProposer, 0, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
		},
		NonProposer0R1,
	},
	NonProposer1R0: {
		symbols.NT_NonProposer, 1, 0,
		symbols.Symbols{
			symbols.NT_ProcessProposal,
		},
		NonProposer1R0,
	},
	NonProposer1R1: {
		symbols.NT_NonProposer, 1, 1,
		symbols.Symbols{
			symbols.NT_ProcessProposal,
		},
		NonProposer1R1,
	},
	NonProposer2R0: {
		symbols.NT_NonProposer, 2, 0,
		symbols.Symbols{
			symbols.NT_Extend,
		},
		NonProposer2R0,
	},
	NonProposer2R1: {
		symbols.NT_NonProposer, 2, 1,
		symbols.Symbols{
			symbols.NT_Extend,
		},
		NonProposer2R1,
	},
	NonProposer3R0: {
		symbols.NT_NonProposer, 3, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProcessProposal,
		},
		NonProposer3R0,
	},
	NonProposer3R1: {
		symbols.NT_NonProposer, 3, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProcessProposal,
		},
		NonProposer3R1,
	},
	NonProposer3R2: {
		symbols.NT_NonProposer, 3, 2,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProcessProposal,
		},
		NonProposer3R2,
	},
	NonProposer4R0: {
		symbols.NT_NonProposer, 4, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_Extend,
		},
		NonProposer4R0,
	},
	NonProposer4R1: {
		symbols.NT_NonProposer, 4, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_Extend,
		},
		NonProposer4R1,
	},
	NonProposer4R2: {
		symbols.NT_NonProposer, 4, 2,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_Extend,
		},
		NonProposer4R2,
	},
	NonProposer5R0: {
		symbols.NT_NonProposer, 5, 0,
		symbols.Symbols{
			symbols.NT_ProcessProposal,
			symbols.NT_Extend,
		},
		NonProposer5R0,
	},
	NonProposer5R1: {
		symbols.NT_NonProposer, 5, 1,
		symbols.Symbols{
			symbols.NT_ProcessProposal,
			symbols.NT_Extend,
		},
		NonProposer5R1,
	},
	NonProposer5R2: {
		symbols.NT_NonProposer, 5, 2,
		symbols.Symbols{
			symbols.NT_ProcessProposal,
			symbols.NT_Extend,
		},
		NonProposer5R2,
	},
	NonProposer6R0: {
		symbols.NT_NonProposer, 6, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProcessProposal,
			symbols.NT_Extend,
		},
		NonProposer6R0,
	},
	NonProposer6R1: {
		symbols.NT_NonProposer, 6, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProcessProposal,
			symbols.NT_Extend,
		},
		NonProposer6R1,
	},
	NonProposer6R2: {
		symbols.NT_NonProposer, 6, 2,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProcessProposal,
			symbols.NT_Extend,
		},
		NonProposer6R2,
	},
	NonProposer6R3: {
		symbols.NT_NonProposer, 6, 3,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProcessProposal,
			symbols.NT_Extend,
		},
		NonProposer6R3,
	},
	OfferSnapshot0R0: {
		symbols.NT_OfferSnapshot, 0, 0,
		symbols.Symbols{
			symbols.T_5,
		},
		OfferSnapshot0R0,
	},
	OfferSnapshot0R1: {
		symbols.NT_OfferSnapshot, 0, 1,
		symbols.Symbols{
			symbols.T_5,
		},
		OfferSnapshot0R1,
	},
	PrepareProposal0R0: {
		symbols.NT_PrepareProposal, 0, 0,
		symbols.Symbols{
			symbols.T_6,
		},
		PrepareProposal0R0,
	},
	PrepareProposal0R1: {
		symbols.NT_PrepareProposal, 0, 1,
		symbols.Symbols{
			symbols.T_6,
		},
		PrepareProposal0R1,
	},
	ProcessProposal0R0: {
		symbols.NT_ProcessProposal, 0, 0,
		symbols.Symbols{
			symbols.T_7,
		},
		ProcessProposal0R0,
	},
	ProcessProposal0R1: {
		symbols.NT_ProcessProposal, 0, 1,
		symbols.Symbols{
			symbols.T_7,
		},
		ProcessProposal0R1,
	},
	Proposer0R0: {
		symbols.NT_Proposer, 0, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
		},
		Proposer0R0,
	},
	Proposer0R1: {
		symbols.NT_Proposer, 0, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
		},
		Proposer0R1,
	},
	Proposer1R0: {
		symbols.NT_Proposer, 1, 0,
		symbols.Symbols{
			symbols.NT_ProposerSimple,
		},
		Proposer1R0,
	},
	Proposer1R1: {
		symbols.NT_Proposer, 1, 1,
		symbols.Symbols{
			symbols.NT_ProposerSimple,
		},
		Proposer1R1,
	},
	Proposer2R0: {
		symbols.NT_Proposer, 2, 0,
		symbols.Symbols{
			symbols.NT_Extend,
		},
		Proposer2R0,
	},
	Proposer2R1: {
		symbols.NT_Proposer, 2, 1,
		symbols.Symbols{
			symbols.NT_Extend,
		},
		Proposer2R1,
	},
	Proposer3R0: {
		symbols.NT_Proposer, 3, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProposerSimple,
		},
		Proposer3R0,
	},
	Proposer3R1: {
		symbols.NT_Proposer, 3, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProposerSimple,
		},
		Proposer3R1,
	},
	Proposer3R2: {
		symbols.NT_Proposer, 3, 2,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProposerSimple,
		},
		Proposer3R2,
	},
	Proposer4R0: {
		symbols.NT_Proposer, 4, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_Extend,
		},
		Proposer4R0,
	},
	Proposer4R1: {
		symbols.NT_Proposer, 4, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_Extend,
		},
		Proposer4R1,
	},
	Proposer4R2: {
		symbols.NT_Proposer, 4, 2,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_Extend,
		},
		Proposer4R2,
	},
	Proposer5R0: {
		symbols.NT_Proposer, 5, 0,
		symbols.Symbols{
			symbols.NT_ProposerSimple,
			symbols.NT_Extend,
		},
		Proposer5R0,
	},
	Proposer5R1: {
		symbols.NT_Proposer, 5, 1,
		symbols.Symbols{
			symbols.NT_ProposerSimple,
			symbols.NT_Extend,
		},
		Proposer5R1,
	},
	Proposer5R2: {
		symbols.NT_Proposer, 5, 2,
		symbols.Symbols{
			symbols.NT_ProposerSimple,
			symbols.NT_Extend,
		},
		Proposer5R2,
	},
	Proposer6R0: {
		symbols.NT_Proposer, 6, 0,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProposerSimple,
			symbols.NT_Extend,
		},
		Proposer6R0,
	},
	Proposer6R1: {
		symbols.NT_Proposer, 6, 1,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProposerSimple,
			symbols.NT_Extend,
		},
		Proposer6R1,
	},
	Proposer6R2: {
		symbols.NT_Proposer, 6, 2,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProposerSimple,
			symbols.NT_Extend,
		},
		Proposer6R2,
	},
	Proposer6R3: {
		symbols.NT_Proposer, 6, 3,
		symbols.Symbols{
			symbols.NT_GotVotes,
			symbols.NT_ProposerSimple,
			symbols.NT_Extend,
		},
		Proposer6R3,
	},
	ProposerSimple0R0: {
		symbols.NT_ProposerSimple, 0, 0,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
		},
		ProposerSimple0R0,
	},
	ProposerSimple0R1: {
		symbols.NT_ProposerSimple, 0, 1,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
		},
		ProposerSimple0R1,
	},
	ProposerSimple1R0: {
		symbols.NT_ProposerSimple, 1, 0,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
			symbols.NT_ProcessProposal,
		},
		ProposerSimple1R0,
	},
	ProposerSimple1R1: {
		symbols.NT_ProposerSimple, 1, 1,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
			symbols.NT_ProcessProposal,
		},
		ProposerSimple1R1,
	},
	ProposerSimple1R2: {
		symbols.NT_ProposerSimple, 1, 2,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
			symbols.NT_ProcessProposal,
		},
		ProposerSimple1R2,
	},
	Recovery0R0: {
		symbols.NT_Recovery, 0, 0,
		symbols.Symbols{
			symbols.NT_InitChain,
			symbols.NT_ConsensusExec,
		},
		Recovery0R0,
	},
	Recovery0R1: {
		symbols.NT_Recovery, 0, 1,
		symbols.Symbols{
			symbols.NT_InitChain,
			symbols.NT_ConsensusExec,
		},
		Recovery0R1,
	},
	Recovery0R2: {
		symbols.NT_Recovery, 0, 2,
		symbols.Symbols{
			symbols.NT_InitChain,
			symbols.NT_ConsensusExec,
		},
		Recovery0R2,
	},
	Recovery1R0: {
		symbols.NT_Recovery, 1, 0,
		symbols.Symbols{
			symbols.NT_ConsensusExec,
		},
		Recovery1R0,
	},
	Recovery1R1: {
		symbols.NT_Recovery, 1, 1,
		symbols.Symbols{
			symbols.NT_ConsensusExec,
		},
		Recovery1R1,
	},
	Start0R0: {
		symbols.NT_Start, 0, 0,
		symbols.Symbols{
			symbols.NT_CleanStart,
		},
		Start0R0,
	},
	Start0R1: {
		symbols.NT_Start, 0, 1,
		symbols.Symbols{
			symbols.NT_CleanStart,
		},
		Start0R1,
	},
	Start1R0: {
		symbols.NT_Start, 1, 0,
		symbols.Symbols{
			symbols.NT_Recovery,
		},
		Start1R0,
	},
	Start1R1: {
		symbols.NT_Start, 1, 1,
		symbols.Symbols{
			symbols.NT_Recovery,
		},
		Start1R1,
	},
	StateSync0R0: {
		symbols.NT_StateSync, 0, 0,
		symbols.Symbols{
			symbols.NT_StateSyncAttempts,
			symbols.NT_SuccessSync,
		},
		StateSync0R0,
	},
	StateSync0R1: {
		symbols.NT_StateSync, 0, 1,
		symbols.Symbols{
			symbols.NT_StateSyncAttempts,
			symbols.NT_SuccessSync,
		},
		StateSync0R1,
	},
	StateSync0R2: {
		symbols.NT_StateSync, 0, 2,
		symbols.Symbols{
			symbols.NT_StateSyncAttempts,
			symbols.NT_SuccessSync,
		},
		StateSync0R2,
	},
	StateSync1R0: {
		symbols.NT_StateSync, 1, 0,
		symbols.Symbols{
			symbols.NT_SuccessSync,
		},
		StateSync1R0,
	},
	StateSync1R1: {
		symbols.NT_StateSync, 1, 1,
		symbols.Symbols{
			symbols.NT_SuccessSync,
		},
		StateSync1R1,
	},
	StateSyncAttempt0R0: {
		symbols.NT_StateSyncAttempt, 0, 0,
		symbols.Symbols{
			symbols.NT_OfferSnapshot,
			symbols.NT_ApplyChunks,
		},
		StateSyncAttempt0R0,
	},
	StateSyncAttempt0R1: {
		symbols.NT_StateSyncAttempt, 0, 1,
		symbols.Symbols{
			symbols.NT_OfferSnapshot,
			symbols.NT_ApplyChunks,
		},
		StateSyncAttempt0R1,
	},
	StateSyncAttempt0R2: {
		symbols.NT_StateSyncAttempt, 0, 2,
		symbols.Symbols{
			symbols.NT_OfferSnapshot,
			symbols.NT_ApplyChunks,
		},
		StateSyncAttempt0R2,
	},
	StateSyncAttempt1R0: {
		symbols.NT_StateSyncAttempt, 1, 0,
		symbols.Symbols{
			symbols.NT_OfferSnapshot,
		},
		StateSyncAttempt1R0,
	},
	StateSyncAttempt1R1: {
		symbols.NT_StateSyncAttempt, 1, 1,
		symbols.Symbols{
			symbols.NT_OfferSnapshot,
		},
		StateSyncAttempt1R1,
	},
	StateSyncAttempts0R0: {
		symbols.NT_StateSyncAttempts, 0, 0,
		symbols.Symbols{
			symbols.NT_StateSyncAttempt,
		},
		StateSyncAttempts0R0,
	},
	StateSyncAttempts0R1: {
		symbols.NT_StateSyncAttempts, 0, 1,
		symbols.Symbols{
			symbols.NT_StateSyncAttempt,
		},
		StateSyncAttempts0R1,
	},
	StateSyncAttempts1R0: {
		symbols.NT_StateSyncAttempts, 1, 0,
		symbols.Symbols{
			symbols.NT_StateSyncAttempt,
			symbols.NT_StateSyncAttempts,
		},
		StateSyncAttempts1R0,
	},
	StateSyncAttempts1R1: {
		symbols.NT_StateSyncAttempts, 1, 1,
		symbols.Symbols{
			symbols.NT_StateSyncAttempt,
			symbols.NT_StateSyncAttempts,
		},
		StateSyncAttempts1R1,
	},
	StateSyncAttempts1R2: {
		symbols.NT_StateSyncAttempts, 1, 2,
		symbols.Symbols{
			symbols.NT_StateSyncAttempt,
			symbols.NT_StateSyncAttempts,
		},
		StateSyncAttempts1R2,
	},
	SuccessSync0R0: {
		symbols.NT_SuccessSync, 0, 0,
		symbols.Symbols{
			symbols.NT_OfferSnapshot,
			symbols.NT_ApplyChunks,
		},
		SuccessSync0R0,
	},
	SuccessSync0R1: {
		symbols.NT_SuccessSync, 0, 1,
		symbols.Symbols{
			symbols.NT_OfferSnapshot,
			symbols.NT_ApplyChunks,
		},
		SuccessSync0R1,
	},
	SuccessSync0R2: {
		symbols.NT_SuccessSync, 0, 2,
		symbols.Symbols{
			symbols.NT_OfferSnapshot,
			symbols.NT_ApplyChunks,
		},
		SuccessSync0R2,
	},
}

var slotIndex = map[Index]Label{
	{symbols.NT_ApplyChunk, 0, 0}:        ApplyChunk0R0,
	{symbols.NT_ApplyChunk, 0, 1}:        ApplyChunk0R1,
	{symbols.NT_ApplyChunks, 0, 0}:       ApplyChunks0R0,
	{symbols.NT_ApplyChunks, 0, 1}:       ApplyChunks0R1,
	{symbols.NT_ApplyChunks, 1, 0}:       ApplyChunks1R0,
	{symbols.NT_ApplyChunks, 1, 1}:       ApplyChunks1R1,
	{symbols.NT_ApplyChunks, 1, 2}:       ApplyChunks1R2,
	{symbols.NT_CleanStart, 0, 0}:        CleanStart0R0,
	{symbols.NT_CleanStart, 0, 1}:        CleanStart0R1,
	{symbols.NT_CleanStart, 0, 2}:        CleanStart0R2,
	{symbols.NT_CleanStart, 1, 0}:        CleanStart1R0,
	{symbols.NT_CleanStart, 1, 1}:        CleanStart1R1,
	{symbols.NT_CleanStart, 1, 2}:        CleanStart1R2,
	{symbols.NT_Commit, 0, 0}:            Commit0R0,
	{symbols.NT_Commit, 0, 1}:            Commit0R1,
	{symbols.NT_ConsensusExec, 0, 0}:     ConsensusExec0R0,
	{symbols.NT_ConsensusExec, 0, 1}:     ConsensusExec0R1,
	{symbols.NT_ConsensusHeight, 0, 0}:   ConsensusHeight0R0,
	{symbols.NT_ConsensusHeight, 0, 1}:   ConsensusHeight0R1,
	{symbols.NT_ConsensusHeight, 0, 2}:   ConsensusHeight0R2,
	{symbols.NT_ConsensusHeight, 0, 3}:   ConsensusHeight0R3,
	{symbols.NT_ConsensusHeight, 1, 0}:   ConsensusHeight1R0,
	{symbols.NT_ConsensusHeight, 1, 1}:   ConsensusHeight1R1,
	{symbols.NT_ConsensusHeight, 1, 2}:   ConsensusHeight1R2,
	{symbols.NT_ConsensusHeights, 0, 0}:  ConsensusHeights0R0,
	{symbols.NT_ConsensusHeights, 0, 1}:  ConsensusHeights0R1,
	{symbols.NT_ConsensusHeights, 1, 0}:  ConsensusHeights1R0,
	{symbols.NT_ConsensusHeights, 1, 1}:  ConsensusHeights1R1,
	{symbols.NT_ConsensusHeights, 1, 2}:  ConsensusHeights1R2,
	{symbols.NT_ConsensusRound, 0, 0}:    ConsensusRound0R0,
	{symbols.NT_ConsensusRound, 0, 1}:    ConsensusRound0R1,
	{symbols.NT_ConsensusRound, 1, 0}:    ConsensusRound1R0,
	{symbols.NT_ConsensusRound, 1, 1}:    ConsensusRound1R1,
	{symbols.NT_ConsensusRounds, 0, 0}:   ConsensusRounds0R0,
	{symbols.NT_ConsensusRounds, 0, 1}:   ConsensusRounds0R1,
	{symbols.NT_ConsensusRounds, 1, 0}:   ConsensusRounds1R0,
	{symbols.NT_ConsensusRounds, 1, 1}:   ConsensusRounds1R1,
	{symbols.NT_ConsensusRounds, 1, 2}:   ConsensusRounds1R2,
	{symbols.NT_Extend, 0, 0}:            Extend0R0,
	{symbols.NT_Extend, 0, 1}:            Extend0R1,
	{symbols.NT_Extend, 1, 0}:            Extend1R0,
	{symbols.NT_Extend, 1, 1}:            Extend1R1,
	{symbols.NT_Extend, 1, 2}:            Extend1R2,
	{symbols.NT_Extend, 2, 0}:            Extend2R0,
	{symbols.NT_Extend, 2, 1}:            Extend2R1,
	{symbols.NT_Extend, 2, 2}:            Extend2R2,
	{symbols.NT_Extend, 3, 0}:            Extend3R0,
	{symbols.NT_Extend, 3, 1}:            Extend3R1,
	{symbols.NT_Extend, 3, 2}:            Extend3R2,
	{symbols.NT_Extend, 3, 3}:            Extend3R3,
	{symbols.NT_ExtendVote, 0, 0}:        ExtendVote0R0,
	{symbols.NT_ExtendVote, 0, 1}:        ExtendVote0R1,
	{symbols.NT_FinalizeBlock, 0, 0}:     FinalizeBlock0R0,
	{symbols.NT_FinalizeBlock, 0, 1}:     FinalizeBlock0R1,
	{symbols.NT_GotVote, 0, 0}:           GotVote0R0,
	{symbols.NT_GotVote, 0, 1}:           GotVote0R1,
	{symbols.NT_GotVotes, 0, 0}:          GotVotes0R0,
	{symbols.NT_GotVotes, 0, 1}:          GotVotes0R1,
	{symbols.NT_GotVotes, 1, 0}:          GotVotes1R0,
	{symbols.NT_GotVotes, 1, 1}:          GotVotes1R1,
	{symbols.NT_GotVotes, 1, 2}:          GotVotes1R2,
	{symbols.NT_InitChain, 0, 0}:         InitChain0R0,
	{symbols.NT_InitChain, 0, 1}:         InitChain0R1,
	{symbols.NT_NonProposer, 0, 0}:       NonProposer0R0,
	{symbols.NT_NonProposer, 0, 1}:       NonProposer0R1,
	{symbols.NT_NonProposer, 1, 0}:       NonProposer1R0,
	{symbols.NT_NonProposer, 1, 1}:       NonProposer1R1,
	{symbols.NT_NonProposer, 2, 0}:       NonProposer2R0,
	{symbols.NT_NonProposer, 2, 1}:       NonProposer2R1,
	{symbols.NT_NonProposer, 3, 0}:       NonProposer3R0,
	{symbols.NT_NonProposer, 3, 1}:       NonProposer3R1,
	{symbols.NT_NonProposer, 3, 2}:       NonProposer3R2,
	{symbols.NT_NonProposer, 4, 0}:       NonProposer4R0,
	{symbols.NT_NonProposer, 4, 1}:       NonProposer4R1,
	{symbols.NT_NonProposer, 4, 2}:       NonProposer4R2,
	{symbols.NT_NonProposer, 5, 0}:       NonProposer5R0,
	{symbols.NT_NonProposer, 5, 1}:       NonProposer5R1,
	{symbols.NT_NonProposer, 5, 2}:       NonProposer5R2,
	{symbols.NT_NonProposer, 6, 0}:       NonProposer6R0,
	{symbols.NT_NonProposer, 6, 1}:       NonProposer6R1,
	{symbols.NT_NonProposer, 6, 2}:       NonProposer6R2,
	{symbols.NT_NonProposer, 6, 3}:       NonProposer6R3,
	{symbols.NT_OfferSnapshot, 0, 0}:     OfferSnapshot0R0,
	{symbols.NT_OfferSnapshot, 0, 1}:     OfferSnapshot0R1,
	{symbols.NT_PrepareProposal, 0, 0}:   PrepareProposal0R0,
	{symbols.NT_PrepareProposal, 0, 1}:   PrepareProposal0R1,
	{symbols.NT_ProcessProposal, 0, 0}:   ProcessProposal0R0,
	{symbols.NT_ProcessProposal, 0, 1}:   ProcessProposal0R1,
	{symbols.NT_Proposer, 0, 0}:          Proposer0R0,
	{symbols.NT_Proposer, 0, 1}:          Proposer0R1,
	{symbols.NT_Proposer, 1, 0}:          Proposer1R0,
	{symbols.NT_Proposer, 1, 1}:          Proposer1R1,
	{symbols.NT_Proposer, 2, 0}:          Proposer2R0,
	{symbols.NT_Proposer, 2, 1}:          Proposer2R1,
	{symbols.NT_Proposer, 3, 0}:          Proposer3R0,
	{symbols.NT_Proposer, 3, 1}:          Proposer3R1,
	{symbols.NT_Proposer, 3, 2}:          Proposer3R2,
	{symbols.NT_Proposer, 4, 0}:          Proposer4R0,
	{symbols.NT_Proposer, 4, 1}:          Proposer4R1,
	{symbols.NT_Proposer, 4, 2}:          Proposer4R2,
	{symbols.NT_Proposer, 5, 0}:          Proposer5R0,
	{symbols.NT_Proposer, 5, 1}:          Proposer5R1,
	{symbols.NT_Proposer, 5, 2}:          Proposer5R2,
	{symbols.NT_Proposer, 6, 0}:          Proposer6R0,
	{symbols.NT_Proposer, 6, 1}:          Proposer6R1,
	{symbols.NT_Proposer, 6, 2}:          Proposer6R2,
	{symbols.NT_Proposer, 6, 3}:          Proposer6R3,
	{symbols.NT_ProposerSimple, 0, 0}:    ProposerSimple0R0,
	{symbols.NT_ProposerSimple, 0, 1}:    ProposerSimple0R1,
	{symbols.NT_ProposerSimple, 1, 0}:    ProposerSimple1R0,
	{symbols.NT_ProposerSimple, 1, 1}:    ProposerSimple1R1,
	{symbols.NT_ProposerSimple, 1, 2}:    ProposerSimple1R2,
	{symbols.NT_Recovery, 0, 0}:          Recovery0R0,
	{symbols.NT_Recovery, 0, 1}:          Recovery0R1,
	{symbols.NT_Recovery, 0, 2}:          Recovery0R2,
	{symbols.NT_Recovery, 1, 0}:          Recovery1R0,
	{symbols.NT_Recovery, 1, 1}:          Recovery1R1,
	{symbols.NT_Start, 0, 0}:             Start0R0,
	{symbols.NT_Start, 0, 1}:             Start0R1,
	{symbols.NT_Start, 1, 0}:             Start1R0,
	{symbols.NT_Start, 1, 1}:             Start1R1,
	{symbols.NT_StateSync, 0, 0}:         StateSync0R0,
	{symbols.NT_StateSync, 0, 1}:         StateSync0R1,
	{symbols.NT_StateSync, 0, 2}:         StateSync0R2,
	{symbols.NT_StateSync, 1, 0}:         StateSync1R0,
	{symbols.NT_StateSync, 1, 1}:         StateSync1R1,
	{symbols.NT_StateSyncAttempt, 0, 0}:  StateSyncAttempt0R0,
	{symbols.NT_StateSyncAttempt, 0, 1}:  StateSyncAttempt0R1,
	{symbols.NT_StateSyncAttempt, 0, 2}:  StateSyncAttempt0R2,
	{symbols.NT_StateSyncAttempt, 1, 0}:  StateSyncAttempt1R0,
	{symbols.NT_StateSyncAttempt, 1, 1}:  StateSyncAttempt1R1,
	{symbols.NT_StateSyncAttempts, 0, 0}: StateSyncAttempts0R0,
	{symbols.NT_StateSyncAttempts, 0, 1}: StateSyncAttempts0R1,
	{symbols.NT_StateSyncAttempts, 1, 0}: StateSyncAttempts1R0,
	{symbols.NT_StateSyncAttempts, 1, 1}: StateSyncAttempts1R1,
	{symbols.NT_StateSyncAttempts, 1, 2}: StateSyncAttempts1R2,
	{symbols.NT_SuccessSync, 0, 0}:       SuccessSync0R0,
	{symbols.NT_SuccessSync, 0, 1}:       SuccessSync0R1,
	{symbols.NT_SuccessSync, 0, 2}:       SuccessSync0R2,
}

var alternates = map[symbols.NT][]Label{
	symbols.NT_Start:             {Start0R0, Start1R0},
	symbols.NT_CleanStart:        {CleanStart0R0, CleanStart1R0},
	symbols.NT_StateSync:         {StateSync0R0, StateSync1R0},
	symbols.NT_StateSyncAttempts: {StateSyncAttempts0R0, StateSyncAttempts1R0},
	symbols.NT_StateSyncAttempt:  {StateSyncAttempt0R0, StateSyncAttempt1R0},
	symbols.NT_SuccessSync:       {SuccessSync0R0},
	symbols.NT_ApplyChunks:       {ApplyChunks0R0, ApplyChunks1R0},
	symbols.NT_Recovery:          {Recovery0R0, Recovery1R0},
	symbols.NT_ConsensusExec:     {ConsensusExec0R0},
	symbols.NT_ConsensusHeights:  {ConsensusHeights0R0, ConsensusHeights1R0},
	symbols.NT_ConsensusHeight:   {ConsensusHeight0R0, ConsensusHeight1R0},
	symbols.NT_ConsensusRounds:   {ConsensusRounds0R0, ConsensusRounds1R0},
	symbols.NT_ConsensusRound:    {ConsensusRound0R0, ConsensusRound1R0},
	symbols.NT_Proposer:          {Proposer0R0, Proposer1R0, Proposer2R0, Proposer3R0, Proposer4R0, Proposer5R0, Proposer6R0},
	symbols.NT_ProposerSimple:    {ProposerSimple0R0, ProposerSimple1R0},
	symbols.NT_NonProposer:       {NonProposer0R0, NonProposer1R0, NonProposer2R0, NonProposer3R0, NonProposer4R0, NonProposer5R0, NonProposer6R0},
	symbols.NT_Extend:            {Extend0R0, Extend1R0, Extend2R0, Extend3R0},
	symbols.NT_GotVotes:          {GotVotes0R0, GotVotes1R0},
	symbols.NT_InitChain:         {InitChain0R0},
	symbols.NT_FinalizeBlock:     {FinalizeBlock0R0},
	symbols.NT_Commit:            {Commit0R0},
	symbols.NT_OfferSnapshot:     {OfferSnapshot0R0},
	symbols.NT_ApplyChunk:        {ApplyChunk0R0},
	symbols.NT_PrepareProposal:   {PrepareProposal0R0},
	symbols.NT_ProcessProposal:   {ProcessProposal0R0},
	symbols.NT_ExtendVote:        {ExtendVote0R0},
	symbols.NT_GotVote:           {GotVote0R0},
}
