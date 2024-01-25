// Package symbols is generated by gogll. Do not edit.
package symbols

import (
	"bytes"
	"fmt"
)

type Symbol interface {
	isSymbol()
	IsNonTerminal() bool
	String() string
}

func (NT) isSymbol() {}
func (T) isSymbol()  {}

// NT is the type of non-terminals symbols
type NT int

const (
	NT_ApplyChunk NT = iota
<<<<<<< HEAD:test/e2e/pkg/grammar/clean-start/grammar-auto/parser/symbols/symbols.go
	NT_ApplyChunks 
	NT_CleanStart 
	NT_Commit 
	NT_ConsensusExec 
	NT_ConsensusHeight 
	NT_ConsensusHeights 
	NT_ConsensusRound 
	NT_ConsensusRounds 
	NT_FinalizeBlock 
	NT_InitChain 
	NT_NonProposer 
	NT_OfferSnapshot 
	NT_PrepareProposal 
	NT_ProcessProposal 
	NT_Proposer 
	NT_Start 
	NT_StateSync 
	NT_StateSyncAttempt 
	NT_StateSyncAttempts 
	NT_SuccessSync 
=======
	NT_ApplyChunks
	NT_CleanStart
	NT_Commit
	NT_ConsensusExec
	NT_ConsensusHeight
	NT_ConsensusHeights
	NT_ConsensusRound
	NT_ConsensusRounds
	NT_FinalizeBlock
	NT_InitChain
	NT_NonProposer
	NT_OfferSnapshot
	NT_PrepareProposal
	NT_ProcessProposal
	NT_Proposer
	NT_Recovery
	NT_Start
	NT_StateSync
	NT_StateSyncAttempt
	NT_StateSyncAttempts
	NT_SuccessSync
>>>>>>> e9637adbe (feat: add gofumpt (#2049)):test/e2e/pkg/grammar/grammar-auto/parser/symbols/symbols.go
)

// T is the type of terminals symbols
type T int

const (
	T_0 T = iota // apply_snapshot_chunk
	T_1          // commit
	T_2          // finalize_block
	T_3          // init_chain
	T_4          // offer_snapshot
	T_5          // prepare_proposal
	T_6          // process_proposal
)

type Symbols []Symbol

func (ss Symbols) Equal(ss1 Symbols) bool {
	if len(ss) != len(ss1) {
		return false
	}
	for i, s := range ss {
		if s.String() != ss1[i].String() {
			return false
		}
	}
	return true
}

func (ss Symbols) String() string {
	w := new(bytes.Buffer)
	for i, s := range ss {
		if i > 0 {
			fmt.Fprint(w, " ")
		}
		fmt.Fprintf(w, "%s", s)
	}
	return w.String()
}

func (ss Symbols) Strings() []string {
	strs := make([]string, len(ss))
	for i, s := range ss {
		strs[i] = s.String()
	}
	return strs
}

func (NT) IsNonTerminal() bool {
	return true
}

func (T) IsNonTerminal() bool {
	return false
}

func (nt NT) String() string {
	return ntToString[nt]
}

func (t T) String() string {
	return tToString[t]
}

// IsNT returns true iff sym is a non-terminal symbol of the grammar
func IsNT(sym string) bool {
	_, exist := stringNT[sym]
	return exist
}

// ToNT returns the NT value of sym or panics if sym is not a non-terminal of the grammar
func ToNT(sym string) NT {
	nt, exist := stringNT[sym]
	if !exist {
		panic(fmt.Sprintf("No NT: %s", sym))
	}
	return nt
}

<<<<<<< HEAD:test/e2e/pkg/grammar/clean-start/grammar-auto/parser/symbols/symbols.go
var ntToString = []string { 
	"ApplyChunk", /* NT_ApplyChunk */
	"ApplyChunks", /* NT_ApplyChunks */
	"CleanStart", /* NT_CleanStart */
	"Commit", /* NT_Commit */
	"ConsensusExec", /* NT_ConsensusExec */
	"ConsensusHeight", /* NT_ConsensusHeight */
	"ConsensusHeights", /* NT_ConsensusHeights */
	"ConsensusRound", /* NT_ConsensusRound */
	"ConsensusRounds", /* NT_ConsensusRounds */
	"FinalizeBlock", /* NT_FinalizeBlock */
	"InitChain", /* NT_InitChain */
	"NonProposer", /* NT_NonProposer */
	"OfferSnapshot", /* NT_OfferSnapshot */
	"PrepareProposal", /* NT_PrepareProposal */
	"ProcessProposal", /* NT_ProcessProposal */
	"Proposer", /* NT_Proposer */
	"Start", /* NT_Start */
	"StateSync", /* NT_StateSync */
	"StateSyncAttempt", /* NT_StateSyncAttempt */
=======
var ntToString = []string{
	"ApplyChunk",        /* NT_ApplyChunk */
	"ApplyChunks",       /* NT_ApplyChunks */
	"CleanStart",        /* NT_CleanStart */
	"Commit",            /* NT_Commit */
	"ConsensusExec",     /* NT_ConsensusExec */
	"ConsensusHeight",   /* NT_ConsensusHeight */
	"ConsensusHeights",  /* NT_ConsensusHeights */
	"ConsensusRound",    /* NT_ConsensusRound */
	"ConsensusRounds",   /* NT_ConsensusRounds */
	"FinalizeBlock",     /* NT_FinalizeBlock */
	"InitChain",         /* NT_InitChain */
	"NonProposer",       /* NT_NonProposer */
	"OfferSnapshot",     /* NT_OfferSnapshot */
	"PrepareProposal",   /* NT_PrepareProposal */
	"ProcessProposal",   /* NT_ProcessProposal */
	"Proposer",          /* NT_Proposer */
	"Recovery",          /* NT_Recovery */
	"Start",             /* NT_Start */
	"StateSync",         /* NT_StateSync */
	"StateSyncAttempt",  /* NT_StateSyncAttempt */
>>>>>>> e9637adbe (feat: add gofumpt (#2049)):test/e2e/pkg/grammar/grammar-auto/parser/symbols/symbols.go
	"StateSyncAttempts", /* NT_StateSyncAttempts */
	"SuccessSync",       /* NT_SuccessSync */
}

var tToString = []string{
	"apply_snapshot_chunk", /* T_0 */
	"commit",               /* T_1 */
	"finalize_block",       /* T_2 */
	"init_chain",           /* T_3 */
	"offer_snapshot",       /* T_4 */
	"prepare_proposal",     /* T_5 */
	"process_proposal",     /* T_6 */
}

<<<<<<< HEAD:test/e2e/pkg/grammar/clean-start/grammar-auto/parser/symbols/symbols.go
var stringNT = map[string]NT{ 
	"ApplyChunk":NT_ApplyChunk,
	"ApplyChunks":NT_ApplyChunks,
	"CleanStart":NT_CleanStart,
	"Commit":NT_Commit,
	"ConsensusExec":NT_ConsensusExec,
	"ConsensusHeight":NT_ConsensusHeight,
	"ConsensusHeights":NT_ConsensusHeights,
	"ConsensusRound":NT_ConsensusRound,
	"ConsensusRounds":NT_ConsensusRounds,
	"FinalizeBlock":NT_FinalizeBlock,
	"InitChain":NT_InitChain,
	"NonProposer":NT_NonProposer,
	"OfferSnapshot":NT_OfferSnapshot,
	"PrepareProposal":NT_PrepareProposal,
	"ProcessProposal":NT_ProcessProposal,
	"Proposer":NT_Proposer,
	"Start":NT_Start,
	"StateSync":NT_StateSync,
	"StateSyncAttempt":NT_StateSyncAttempt,
	"StateSyncAttempts":NT_StateSyncAttempts,
	"SuccessSync":NT_SuccessSync,
=======
var stringNT = map[string]NT{
	"ApplyChunk":        NT_ApplyChunk,
	"ApplyChunks":       NT_ApplyChunks,
	"CleanStart":        NT_CleanStart,
	"Commit":            NT_Commit,
	"ConsensusExec":     NT_ConsensusExec,
	"ConsensusHeight":   NT_ConsensusHeight,
	"ConsensusHeights":  NT_ConsensusHeights,
	"ConsensusRound":    NT_ConsensusRound,
	"ConsensusRounds":   NT_ConsensusRounds,
	"FinalizeBlock":     NT_FinalizeBlock,
	"InitChain":         NT_InitChain,
	"NonProposer":       NT_NonProposer,
	"OfferSnapshot":     NT_OfferSnapshot,
	"PrepareProposal":   NT_PrepareProposal,
	"ProcessProposal":   NT_ProcessProposal,
	"Proposer":          NT_Proposer,
	"Recovery":          NT_Recovery,
	"Start":             NT_Start,
	"StateSync":         NT_StateSync,
	"StateSyncAttempt":  NT_StateSyncAttempt,
	"StateSyncAttempts": NT_StateSyncAttempts,
	"SuccessSync":       NT_SuccessSync,
>>>>>>> e9637adbe (feat: add gofumpt (#2049)):test/e2e/pkg/grammar/grammar-auto/parser/symbols/symbols.go
}
