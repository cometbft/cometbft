```
package "github.com/cometbft/cometbft/test/e2e/pkg/grammar/clean-start"


Start : CleanStart ;

CleanStart : InitChain StateSync ConsensusExec | InitChain ConsensusExec ;
StateSync : StateSyncAttempts SuccessSync |  SuccessSync ; 
StateSyncAttempts : StateSyncAttempt | StateSyncAttempt StateSyncAttempts ;
StateSyncAttempt : OfferSnapshot ApplyChunks | OfferSnapshot ;
SuccessSync : OfferSnapshot ApplyChunks ; 
ApplyChunks : ApplyChunk | ApplyChunk ApplyChunks ;  

ConsensusExec : ConsensusHeights ;
ConsensusHeights : ConsensusHeight | ConsensusHeight ConsensusHeights ;
ConsensusHeight : ConsensusRounds Decide Commit | Decide Commit ;
ConsensusRounds : ConsensusRound | ConsensusRound ConsensusRounds ;
ConsensusRound : Proposer | NonProposer ; 

Proposer : PrepareProposal | PrepareProposal ProcessProposal ; 
NonProposer: ProcessProposal ;

InitChain : "init_chain" ;
Decide : "finalize_block" ; 
Commit : "commit" ;
OfferSnapshot : "offer_snapshot" ;
ApplyChunk : "apply_snapshot_chunk" ; 
PrepareProposal : "prepare_proposal" ; 
ProcessProposal : "process_proposal" ;


```

