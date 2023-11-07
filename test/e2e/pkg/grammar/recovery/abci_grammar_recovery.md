```
package "github.com/cometbft/cometbft/test/e2e/pkg/grammar/recovery/grammar-auto"

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

The part of the original grammar (https://github.com/cometbft/cometbft/blob/main/spec/abci/abci%2B%2B_comet_expected_behavior.md) the grammar above 
refers to is below: 

start               = recovery

recovery            = info consensus-exec

consensus-exec      = (inf)consensus-height
consensus-height    = *consensus-round decide commit
consensus-round     = proposer / non-proposer

proposer            = [prepare-proposal [process-proposal]]
non-proposer        = [process-proposal] 

decide              = %s"<FinalizeBlock>"
commit              = %s"<Commit>"
info                = %s"<Info>"
prepare-proposal    = %s"<PrepareProposal>"
process-proposal    = %s"<ProcessProposal>"
