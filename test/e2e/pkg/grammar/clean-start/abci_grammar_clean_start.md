```
package "github.com/cometbft/cometbft/test/e2e/pkg/grammar/clean-start/grammar-auto"

Start : CleanStart ;

CleanStart : InitChain StateSync ConsensusExec | InitChain ConsensusExec ;
StateSync : StateSyncAttempts SuccessSync |  SuccessSync ; 
StateSyncAttempts : StateSyncAttempt | StateSyncAttempt StateSyncAttempts ;
StateSyncAttempt : OfferSnapshot ApplyChunks | OfferSnapshot ;
SuccessSync : OfferSnapshot ApplyChunks ; 
ApplyChunks : ApplyChunk | ApplyChunk ApplyChunks ;  

ConsensusExec : ConsensusHeights ;
ConsensusHeights : ConsensusHeight | ConsensusHeight ConsensusHeights ;
ConsensusHeight : ConsensusRounds FinalizeBlock Commit | FinalizeBlock Commit ;
ConsensusRounds : ConsensusRound | ConsensusRound ConsensusRounds ;
ConsensusRound : Proposer | NonProposer ; 

Proposer : PrepareProposal | PrepareProposal ProcessProposal ; 
NonProposer: ProcessProposal ;

InitChain : "init_chain" ;
FinalizeBlock : "finalize_block" ; 
Commit : "commit" ;
OfferSnapshot : "offer_snapshot" ;
ApplyChunk : "apply_snapshot_chunk" ; 
PrepareProposal : "prepare_proposal" ; 
ProcessProposal : "process_proposal" ;

```

The part of the original grammar (https://github.com/cometbft/cometbft/blob/main/spec/abci/abci%2B%2B_comet_expected_behavior.md) the grammar above 
refers to is below: 

start               = clean-start

clean-start         = init-chain [state-sync] consensus-exec
state-sync          = *state-sync-attempt success-sync info
state-sync-attempt  = offer-snapshot *apply-chunk
success-sync        = offer-snapshot 1*apply-chunk

consensus-exec      = (inf)consensus-height
consensus-height    = *consensus-round decide commit
consensus-round     = proposer / non-proposer

proposer            = [prepare-proposal [process-proposal]]
non-proposer        = [process-proposal] 

init-chain          = %s"<InitChain>"
decide              = %s"<FinalizeBlock>"
commit              = %s"<Commit>"
offer-snapshot      = %s"<OfferSnapshot>"
apply-chunk         = %s"<ApplySnapshotChunk>"
info                = %s"<Info>"
prepare-proposal    = %s"<PrepareProposal>"
process-proposal    = %s"<ProcessProposal>"





