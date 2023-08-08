```
package "github.com/cometbft/cometbft/test/e2e/pkg/grammar/recovery"

Start : Recovery ; 

Recovery :  ConsensusExec ;

ConsensusExec : ConsensusHeights ;
ConsensusHeights : ConsensusHeight | ConsensusHeight ConsensusHeights ;
ConsensusHeight : ConsensusRounds FinalizeBlock Commit | FinalizeBlock Commit ;
ConsensusRounds : ConsensusRound | ConsensusRound ConsensusRounds ;
ConsensusRound : Proposer | NonProposer ; 

Proposer : PrepareProposal | PrepareProposal ProcessProposal ; 
NonProposer: ProcessProposal ;


FinalizeBlock : "finalize_block" ; 
Commit : "commit" ;
PrepareProposal : "prepare_proposal" ; 
ProcessProposal : "process_proposal" ;


```

