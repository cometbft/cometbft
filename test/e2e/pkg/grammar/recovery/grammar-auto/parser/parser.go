// Package parser is generated by gogll. Do not edit.
package parser

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/recovery/grammar-auto/lexer"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/recovery/grammar-auto/parser/bsr"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/recovery/grammar-auto/parser/slot"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/recovery/grammar-auto/parser/symbols"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/recovery/grammar-auto/token"
)

type parser struct {
	cI int

	R *descriptors
	U *descriptors

	popped   map[poppedNode]bool
	crf      map[clusterNode][]*crfNode
	crfNodes map[crfNode]*crfNode

	lex         *lexer.Lexer
	parseErrors []*Error

	bsrSet *bsr.Set
}

func newParser(l *lexer.Lexer) *parser {
	return &parser{
		cI:     0,
		lex:    l,
		R:      &descriptors{},
		U:      &descriptors{},
		popped: make(map[poppedNode]bool),
		crf: map[clusterNode][]*crfNode{
			{symbols.NT_Start, 0}: {},
		},
		crfNodes:    map[crfNode]*crfNode{},
		bsrSet:      bsr.New(symbols.NT_Start, l),
		parseErrors: nil,
	}
}

// Parse returns the BSR set containing the parse forest.
// If the parse was successfull []*Error is nil
func Parse(l *lexer.Lexer) (*bsr.Set, []*Error) {
	return newParser(l).parse()
}

func (p *parser) parse() (*bsr.Set, []*Error) {
	var L slot.Label
	m, cU := len(p.lex.Tokens)-1, 0
	p.ntAdd(symbols.NT_Start, 0)
	// p.DumpDescriptors()
	for !p.R.empty() {
		L, cU, p.cI = p.R.remove()

		// fmt.Println()
		// fmt.Printf("L:%s, cI:%d, I[p.cI]:%s, cU:%d\n", L, p.cI, p.lex.Tokens[p.cI], cU)
		// p.DumpDescriptors()

		switch L {
		case slot.Commit0R0: // Commit : ∙commit

			p.bsrSet.Add(slot.Commit0R1, cU, p.cI, p.cI+1)
			p.cI++
			if p.follow(symbols.NT_Commit) {
				p.rtn(symbols.NT_Commit, cU, p.cI)
			} else {
				p.parseError(slot.Commit0R0, p.cI, followSets[symbols.NT_Commit])
			}
		case slot.ConsensusExec0R0: // ConsensusExec : ∙ConsensusHeights

			p.call(slot.ConsensusExec0R1, cU, p.cI)
		case slot.ConsensusExec0R1: // ConsensusExec : ConsensusHeights ∙

			if p.follow(symbols.NT_ConsensusExec) {
				p.rtn(symbols.NT_ConsensusExec, cU, p.cI)
			} else {
				p.parseError(slot.ConsensusExec0R0, p.cI, followSets[symbols.NT_ConsensusExec])
			}
		case slot.ConsensusHeight0R0: // ConsensusHeight : ∙ConsensusRounds FinalizeBlock Commit

			p.call(slot.ConsensusHeight0R1, cU, p.cI)
		case slot.ConsensusHeight0R1: // ConsensusHeight : ConsensusRounds ∙FinalizeBlock Commit

			if !p.testSelect(slot.ConsensusHeight0R1) {
				p.parseError(slot.ConsensusHeight0R1, p.cI, first[slot.ConsensusHeight0R1])
				break
			}

			p.call(slot.ConsensusHeight0R2, cU, p.cI)
		case slot.ConsensusHeight0R2: // ConsensusHeight : ConsensusRounds FinalizeBlock ∙Commit

			if !p.testSelect(slot.ConsensusHeight0R2) {
				p.parseError(slot.ConsensusHeight0R2, p.cI, first[slot.ConsensusHeight0R2])
				break
			}

			p.call(slot.ConsensusHeight0R3, cU, p.cI)
		case slot.ConsensusHeight0R3: // ConsensusHeight : ConsensusRounds FinalizeBlock Commit ∙

			if p.follow(symbols.NT_ConsensusHeight) {
				p.rtn(symbols.NT_ConsensusHeight, cU, p.cI)
			} else {
				p.parseError(slot.ConsensusHeight0R0, p.cI, followSets[symbols.NT_ConsensusHeight])
			}
		case slot.ConsensusHeight1R0: // ConsensusHeight : ∙FinalizeBlock Commit

			p.call(slot.ConsensusHeight1R1, cU, p.cI)
		case slot.ConsensusHeight1R1: // ConsensusHeight : FinalizeBlock ∙Commit

			if !p.testSelect(slot.ConsensusHeight1R1) {
				p.parseError(slot.ConsensusHeight1R1, p.cI, first[slot.ConsensusHeight1R1])
				break
			}

			p.call(slot.ConsensusHeight1R2, cU, p.cI)
		case slot.ConsensusHeight1R2: // ConsensusHeight : FinalizeBlock Commit ∙

			if p.follow(symbols.NT_ConsensusHeight) {
				p.rtn(symbols.NT_ConsensusHeight, cU, p.cI)
			} else {
				p.parseError(slot.ConsensusHeight1R0, p.cI, followSets[symbols.NT_ConsensusHeight])
			}
		case slot.ConsensusHeights0R0: // ConsensusHeights : ∙ConsensusHeight

			p.call(slot.ConsensusHeights0R1, cU, p.cI)
		case slot.ConsensusHeights0R1: // ConsensusHeights : ConsensusHeight ∙

			if p.follow(symbols.NT_ConsensusHeights) {
				p.rtn(symbols.NT_ConsensusHeights, cU, p.cI)
			} else {
				p.parseError(slot.ConsensusHeights0R0, p.cI, followSets[symbols.NT_ConsensusHeights])
			}
		case slot.ConsensusHeights1R0: // ConsensusHeights : ∙ConsensusHeight ConsensusHeights

			p.call(slot.ConsensusHeights1R1, cU, p.cI)
		case slot.ConsensusHeights1R1: // ConsensusHeights : ConsensusHeight ∙ConsensusHeights

			if !p.testSelect(slot.ConsensusHeights1R1) {
				p.parseError(slot.ConsensusHeights1R1, p.cI, first[slot.ConsensusHeights1R1])
				break
			}

			p.call(slot.ConsensusHeights1R2, cU, p.cI)
		case slot.ConsensusHeights1R2: // ConsensusHeights : ConsensusHeight ConsensusHeights ∙

			if p.follow(symbols.NT_ConsensusHeights) {
				p.rtn(symbols.NT_ConsensusHeights, cU, p.cI)
			} else {
				p.parseError(slot.ConsensusHeights1R0, p.cI, followSets[symbols.NT_ConsensusHeights])
			}
		case slot.ConsensusRound0R0: // ConsensusRound : ∙Proposer

			p.call(slot.ConsensusRound0R1, cU, p.cI)
		case slot.ConsensusRound0R1: // ConsensusRound : Proposer ∙

			if p.follow(symbols.NT_ConsensusRound) {
				p.rtn(symbols.NT_ConsensusRound, cU, p.cI)
			} else {
				p.parseError(slot.ConsensusRound0R0, p.cI, followSets[symbols.NT_ConsensusRound])
			}
		case slot.ConsensusRound1R0: // ConsensusRound : ∙NonProposer

			p.call(slot.ConsensusRound1R1, cU, p.cI)
		case slot.ConsensusRound1R1: // ConsensusRound : NonProposer ∙

			if p.follow(symbols.NT_ConsensusRound) {
				p.rtn(symbols.NT_ConsensusRound, cU, p.cI)
			} else {
				p.parseError(slot.ConsensusRound1R0, p.cI, followSets[symbols.NT_ConsensusRound])
			}
		case slot.ConsensusRounds0R0: // ConsensusRounds : ∙ConsensusRound

			p.call(slot.ConsensusRounds0R1, cU, p.cI)
		case slot.ConsensusRounds0R1: // ConsensusRounds : ConsensusRound ∙

			if p.follow(symbols.NT_ConsensusRounds) {
				p.rtn(symbols.NT_ConsensusRounds, cU, p.cI)
			} else {
				p.parseError(slot.ConsensusRounds0R0, p.cI, followSets[symbols.NT_ConsensusRounds])
			}
		case slot.ConsensusRounds1R0: // ConsensusRounds : ∙ConsensusRound ConsensusRounds

			p.call(slot.ConsensusRounds1R1, cU, p.cI)
		case slot.ConsensusRounds1R1: // ConsensusRounds : ConsensusRound ∙ConsensusRounds

			if !p.testSelect(slot.ConsensusRounds1R1) {
				p.parseError(slot.ConsensusRounds1R1, p.cI, first[slot.ConsensusRounds1R1])
				break
			}

			p.call(slot.ConsensusRounds1R2, cU, p.cI)
		case slot.ConsensusRounds1R2: // ConsensusRounds : ConsensusRound ConsensusRounds ∙

			if p.follow(symbols.NT_ConsensusRounds) {
				p.rtn(symbols.NT_ConsensusRounds, cU, p.cI)
			} else {
				p.parseError(slot.ConsensusRounds1R0, p.cI, followSets[symbols.NT_ConsensusRounds])
			}
		case slot.FinalizeBlock0R0: // FinalizeBlock : ∙finalize_block

			p.bsrSet.Add(slot.FinalizeBlock0R1, cU, p.cI, p.cI+1)
			p.cI++
			if p.follow(symbols.NT_FinalizeBlock) {
				p.rtn(symbols.NT_FinalizeBlock, cU, p.cI)
			} else {
				p.parseError(slot.FinalizeBlock0R0, p.cI, followSets[symbols.NT_FinalizeBlock])
			}
		case slot.NonProposer0R0: // NonProposer : ∙ProcessProposal

			p.call(slot.NonProposer0R1, cU, p.cI)
		case slot.NonProposer0R1: // NonProposer : ProcessProposal ∙

			if p.follow(symbols.NT_NonProposer) {
				p.rtn(symbols.NT_NonProposer, cU, p.cI)
			} else {
				p.parseError(slot.NonProposer0R0, p.cI, followSets[symbols.NT_NonProposer])
			}
		case slot.PrepareProposal0R0: // PrepareProposal : ∙prepare_proposal

			p.bsrSet.Add(slot.PrepareProposal0R1, cU, p.cI, p.cI+1)
			p.cI++
			if p.follow(symbols.NT_PrepareProposal) {
				p.rtn(symbols.NT_PrepareProposal, cU, p.cI)
			} else {
				p.parseError(slot.PrepareProposal0R0, p.cI, followSets[symbols.NT_PrepareProposal])
			}
		case slot.ProcessProposal0R0: // ProcessProposal : ∙process_proposal

			p.bsrSet.Add(slot.ProcessProposal0R1, cU, p.cI, p.cI+1)
			p.cI++
			if p.follow(symbols.NT_ProcessProposal) {
				p.rtn(symbols.NT_ProcessProposal, cU, p.cI)
			} else {
				p.parseError(slot.ProcessProposal0R0, p.cI, followSets[symbols.NT_ProcessProposal])
			}
		case slot.Proposer0R0: // Proposer : ∙PrepareProposal

			p.call(slot.Proposer0R1, cU, p.cI)
		case slot.Proposer0R1: // Proposer : PrepareProposal ∙

			if p.follow(symbols.NT_Proposer) {
				p.rtn(symbols.NT_Proposer, cU, p.cI)
			} else {
				p.parseError(slot.Proposer0R0, p.cI, followSets[symbols.NT_Proposer])
			}
		case slot.Proposer1R0: // Proposer : ∙PrepareProposal ProcessProposal

			p.call(slot.Proposer1R1, cU, p.cI)
		case slot.Proposer1R1: // Proposer : PrepareProposal ∙ProcessProposal

			if !p.testSelect(slot.Proposer1R1) {
				p.parseError(slot.Proposer1R1, p.cI, first[slot.Proposer1R1])
				break
			}

			p.call(slot.Proposer1R2, cU, p.cI)
		case slot.Proposer1R2: // Proposer : PrepareProposal ProcessProposal ∙

			if p.follow(symbols.NT_Proposer) {
				p.rtn(symbols.NT_Proposer, cU, p.cI)
			} else {
				p.parseError(slot.Proposer1R0, p.cI, followSets[symbols.NT_Proposer])
			}
		case slot.Recovery0R0: // Recovery : ∙ConsensusExec

			p.call(slot.Recovery0R1, cU, p.cI)
		case slot.Recovery0R1: // Recovery : ConsensusExec ∙

			if p.follow(symbols.NT_Recovery) {
				p.rtn(symbols.NT_Recovery, cU, p.cI)
			} else {
				p.parseError(slot.Recovery0R0, p.cI, followSets[symbols.NT_Recovery])
			}
		case slot.Start0R0: // Start : ∙Recovery

			p.call(slot.Start0R1, cU, p.cI)
		case slot.Start0R1: // Start : Recovery ∙

			if p.follow(symbols.NT_Start) {
				p.rtn(symbols.NT_Start, cU, p.cI)
			} else {
				p.parseError(slot.Start0R0, p.cI, followSets[symbols.NT_Start])
			}

		default:
			panic("This must not happen")
		}
	}
	if !p.bsrSet.Contain(symbols.NT_Start, 0, m) {
		p.sortParseErrors()
		return nil, p.parseErrors
	}
	return p.bsrSet, nil
}

func (p *parser) ntAdd(nt symbols.NT, j int) {
	// fmt.Printf("p.ntAdd(%s, %d)\n", nt, j)
	failed := true
	expected := map[token.Type]string{}
	for _, l := range slot.GetAlternates(nt) {
		if p.testSelect(l) {
			p.dscAdd(l, j, j)
			failed = false
		} else {
			for k, v := range first[l] {
				expected[k] = v
			}
		}
	}
	if failed {
		for _, l := range slot.GetAlternates(nt) {
			p.parseError(l, j, expected)
		}
	}
}

/*** Call Return Forest ***/

type poppedNode struct {
	X    symbols.NT
	k, j int
}

type clusterNode struct {
	X symbols.NT
	k int
}

type crfNode struct {
	L slot.Label
	i int
}

/*
suppose that L is Y ::=αX ·β
if there is no CRF node labelled (L,i)

	create one let u be the CRF node labelled (L,i)

if there is no CRF node labelled (X, j) {

		create a CRF node v labelled (X, j)
		create an edge from v to u
		ntAdd(X, j)
	} else {

		let v be the CRF node labelled (X, j)
		if there is not an edge from v to u {
			create an edge from v to u
			for all ((X, j,h)∈P) {
				dscAdd(L, i, h);
				bsrAdd(L, i, j, h)
			}
		}
	}
*/
func (p *parser) call(L slot.Label, i, j int) {
	// fmt.Printf("p.call(%s,%d,%d)\n", L,i,j)
	u, exist := p.crfNodes[crfNode{L, i}]
	// fmt.Printf("  u exist=%t\n", exist)
	if !exist {
		u = &crfNode{L, i}
		p.crfNodes[*u] = u
	}
	X := L.Symbols()[L.Pos()-1].(symbols.NT)
	ndV := clusterNode{X, j}
	v, exist := p.crf[ndV]
	if !exist {
		// fmt.Println("  v !exist")
		p.crf[ndV] = []*crfNode{u}
		p.ntAdd(X, j)
	} else {
		// fmt.Println("  v exist")
		if !existEdge(v, u) {
			// fmt.Printf("  !existEdge(%v)\n", u)
			p.crf[ndV] = append(v, u)
			// fmt.Printf("|popped|=%d\n", len(popped))
			for pnd := range p.popped {
				if pnd.X == X && pnd.k == j {
					p.dscAdd(L, i, pnd.j)
					p.bsrSet.Add(L, i, j, pnd.j)
				}
			}
		}
	}
}

func existEdge(nds []*crfNode, nd *crfNode) bool {
	for _, nd1 := range nds {
		if nd1 == nd {
			return true
		}
	}
	return false
}

func (p *parser) rtn(X symbols.NT, k, j int) {
	// fmt.Printf("p.rtn(%s,%d,%d)\n", X,k,j)
	pn := poppedNode{X, k, j}
	if _, exist := p.popped[pn]; !exist {
		p.popped[pn] = true
		for _, nd := range p.crf[clusterNode{X, k}] {
			p.dscAdd(nd.L, nd.i, j)
			p.bsrSet.Add(nd.L, nd.i, k, j)
		}
	}
}

// func CRFString() string {
// 	buf := new(bytes.Buffer)
// 	buf.WriteString("CRF: {")
// 	for cn, nds := range crf{
// 		for _, nd := range nds {
// 			fmt.Fprintf(buf, "%s->%s, ", cn, nd)
// 		}
// 	}
// 	buf.WriteString("}")
// 	return buf.String()
// }

func (cn clusterNode) String() string {
	return fmt.Sprintf("(%s,%d)", cn.X, cn.k)
}

func (n crfNode) String() string {
	return fmt.Sprintf("(%s,%d)", n.L.String(), n.i)
}

// func PoppedString() string {
// 	buf := new(bytes.Buffer)
// 	buf.WriteString("Popped: {")
// 	for p, _ := range popped {
// 		fmt.Fprintf(buf, "(%s,%d,%d) ", p.X, p.k, p.j)
// 	}
// 	buf.WriteString("}")
// 	return buf.String()
// }

/*** descriptors ***/

type descriptors struct {
	set []*descriptor
}

func (ds *descriptors) contain(d *descriptor) bool {
	for _, d1 := range ds.set {
		if d1 == d {
			return true
		}
	}
	return false
}

func (ds *descriptors) empty() bool {
	return len(ds.set) == 0
}

func (ds *descriptors) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("{")
	for i, d := range ds.set {
		if i > 0 {
			buf.WriteString("; ")
		}
		fmt.Fprintf(buf, "%s", d)
	}
	buf.WriteString("}")
	return buf.String()
}

type descriptor struct {
	L slot.Label
	k int
	i int
}

func (d *descriptor) String() string {
	return fmt.Sprintf("%s,%d,%d", d.L, d.k, d.i)
}

func (p *parser) dscAdd(L slot.Label, k, i int) {
	// fmt.Printf("p.dscAdd(%s,%d,%d)\n", L, k, i)
	d := &descriptor{L, k, i}
	if !p.U.contain(d) {
		p.R.set = append(p.R.set, d)
		p.U.set = append(p.U.set, d)
	}
}

func (ds *descriptors) remove() (L slot.Label, k, i int) {
	d := ds.set[len(ds.set)-1]
	ds.set = ds.set[:len(ds.set)-1]
	// fmt.Printf("remove: %s,%d,%d\n", d.L, d.k, d.i)
	return d.L, d.k, d.i
}

func (p *parser) DumpDescriptors() {
	p.DumpR()
	p.DumpU()
}

func (p *parser) DumpR() {
	fmt.Println("R:")
	for _, d := range p.R.set {
		fmt.Printf(" %s\n", d)
	}
}

func (p *parser) DumpU() {
	fmt.Println("U:")
	for _, d := range p.U.set {
		fmt.Printf(" %s\n", d)
	}
}

/*** TestSelect ***/

func (p *parser) follow(nt symbols.NT) bool {
	_, exist := followSets[nt][p.lex.Tokens[p.cI].Type()]
	return exist
}

func (p *parser) testSelect(l slot.Label) bool {
	_, exist := first[l][p.lex.Tokens[p.cI].Type()]
	// fmt.Printf("testSelect(%s) = %t\n", l, exist)
	return exist
}

var first = []map[token.Type]string{
	// Commit : ∙commit
	{
		token.T_0: "commit",
	},
	// Commit : commit ∙
	{
		token.EOF: "$",
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusExec : ∙ConsensusHeights
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusExec : ConsensusHeights ∙
	{
		token.EOF: "$",
	},
	// ConsensusHeight : ∙ConsensusRounds FinalizeBlock Commit
	{
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusHeight : ConsensusRounds ∙FinalizeBlock Commit
	{
		token.T_1: "finalize_block",
	},
	// ConsensusHeight : ConsensusRounds FinalizeBlock ∙Commit
	{
		token.T_0: "commit",
	},
	// ConsensusHeight : ConsensusRounds FinalizeBlock Commit ∙
	{
		token.EOF: "$",
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusHeight : ∙FinalizeBlock Commit
	{
		token.T_1: "finalize_block",
	},
	// ConsensusHeight : FinalizeBlock ∙Commit
	{
		token.T_0: "commit",
	},
	// ConsensusHeight : FinalizeBlock Commit ∙
	{
		token.EOF: "$",
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusHeights : ∙ConsensusHeight
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusHeights : ConsensusHeight ∙
	{
		token.EOF: "$",
	},
	// ConsensusHeights : ∙ConsensusHeight ConsensusHeights
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusHeights : ConsensusHeight ∙ConsensusHeights
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusHeights : ConsensusHeight ConsensusHeights ∙
	{
		token.EOF: "$",
	},
	// ConsensusRound : ∙Proposer
	{
		token.T_2: "prepare_proposal",
	},
	// ConsensusRound : Proposer ∙
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusRound : ∙NonProposer
	{
		token.T_3: "process_proposal",
	},
	// ConsensusRound : NonProposer ∙
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusRounds : ∙ConsensusRound
	{
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusRounds : ConsensusRound ∙
	{
		token.T_1: "finalize_block",
	},
	// ConsensusRounds : ∙ConsensusRound ConsensusRounds
	{
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusRounds : ConsensusRound ∙ConsensusRounds
	{
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusRounds : ConsensusRound ConsensusRounds ∙
	{
		token.T_1: "finalize_block",
	},
	// FinalizeBlock : ∙finalize_block
	{
		token.T_1: "finalize_block",
	},
	// FinalizeBlock : finalize_block ∙
	{
		token.T_0: "commit",
	},
	// NonProposer : ∙ProcessProposal
	{
		token.T_3: "process_proposal",
	},
	// NonProposer : ProcessProposal ∙
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// PrepareProposal : ∙prepare_proposal
	{
		token.T_2: "prepare_proposal",
	},
	// PrepareProposal : prepare_proposal ∙
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ProcessProposal : ∙process_proposal
	{
		token.T_3: "process_proposal",
	},
	// ProcessProposal : process_proposal ∙
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// Proposer : ∙PrepareProposal
	{
		token.T_2: "prepare_proposal",
	},
	// Proposer : PrepareProposal ∙
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// Proposer : ∙PrepareProposal ProcessProposal
	{
		token.T_2: "prepare_proposal",
	},
	// Proposer : PrepareProposal ∙ProcessProposal
	{
		token.T_3: "process_proposal",
	},
	// Proposer : PrepareProposal ProcessProposal ∙
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// Recovery : ∙ConsensusExec
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// Recovery : ConsensusExec ∙
	{
		token.EOF: "$",
	},
	// Start : ∙Recovery
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// Start : Recovery ∙
	{
		token.EOF: "$",
	},
}

var followSets = []map[token.Type]string{
	// Commit
	{
		token.EOF: "$",
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusExec
	{
		token.EOF: "$",
	},
	// ConsensusHeight
	{
		token.EOF: "$",
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusHeights
	{
		token.EOF: "$",
	},
	// ConsensusRound
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ConsensusRounds
	{
		token.T_1: "finalize_block",
	},
	// FinalizeBlock
	{
		token.T_0: "commit",
	},
	// NonProposer
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// PrepareProposal
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// ProcessProposal
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// Proposer
	{
		token.T_1: "finalize_block",
		token.T_2: "prepare_proposal",
		token.T_3: "process_proposal",
	},
	// Recovery
	{
		token.EOF: "$",
	},
	// Start
	{
		token.EOF: "$",
	},
}

/*** Errors ***/

/*
Error is returned by Parse at every point at which the parser fails to parse
a grammar production. For non-LL-1 grammars there will be an error for each
alternate attempted by the parser.

The errors are sorted in descending order of input position (index of token in
the stream of tokens).

Normally the error of interest is the one that has parsed the largest number of
tokens.
*/
type Error struct {
	// Index of token that caused the error.
	cI int

	// Grammar slot at which the error occured.
	Slot slot.Label

	// The token at which the error occurred.
	Token *token.Token

	// The line and column in the input text at which the error occurred
	Line, Column int

	// The tokens expected at the point where the error occurred
	Expected map[token.Type]string
}

func (pe *Error) String() string {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "Parse Error: %s I[%d]=%s at line %d col %d\n",
		pe.Slot, pe.cI, pe.Token, pe.Line, pe.Column)
	exp := []string{}
	for _, e := range pe.Expected {
		exp = append(exp, e)
	}
	fmt.Fprintf(w, "Expected one of: [%s]", strings.Join(exp, ","))
	return w.String()
}

func (p *parser) parseError(slot slot.Label, i int, expected map[token.Type]string) {
	pe := &Error{cI: i, Slot: slot, Token: p.lex.Tokens[i], Expected: expected}
	p.parseErrors = append(p.parseErrors, pe)
}

func (p *parser) sortParseErrors() {
	sort.Slice(p.parseErrors,
		func(i, j int) bool {
			return p.parseErrors[j].Token.Lext() < p.parseErrors[i].Token.Lext()
		})
	for _, pe := range p.parseErrors {
		pe.Line, pe.Column = p.lex.GetLineColumn(pe.Token.Lext())
	}
}
