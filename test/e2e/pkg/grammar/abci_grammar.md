```
package "github.com/cometbft/cometbft/test/e2e/pkg/grammar"


Start : CleanStart | Recovery ;

CleanStart : InitChain StateSync ConsensusExec | InitChain ConsensusExec ;
StateSync : StateSyncAttempts SuccessSync |  SuccessSync ; 
StateSyncAttempts : StateSyncAttempt | StateSyncAttempt StateSyncAttempts ;
StateSyncAttempt : OfferSnapshot ApplyChunks | OfferSnapshot ;
SuccessSync : OfferSnapshot ApplyChunks ; 
ApplyChunks : ApplyChunk | ApplyChunk ApplyChunks ;  

Recovery :  ConsensusExec ;

ConsensusExec : ConsensusHeights ;
ConsensusHeights : ConsensusHeight | ConsensusHeight ConsensusHeights ;
ConsensusHeight : ConsensusRounds Decide Commit | Decide Commit ;
ConsensusRounds : ConsensusRound | ConsensusRound ConsensusRounds ;
ConsensusRound : Proposer | NonProposer ; 

Proposer : PrepareProposal ProcessProposal ; 
NonProposer: ProcessProposal ;
Decide : BeginBlock DeliverTxs EndBlock | BeginBlock EndBlock ; 
DeliverTxs : DeliverTx | DeliverTx DeliverTxs ; 


InitChain : "1" ;
BeginBlock : "2" ; 
DeliverTx : "3" ;
EndBlock : "4" ;
Commit : "5" ;
OfferSnapshot : "6" ;
ApplyChunk : "7" ; 
PrepareProposal : "8" ; 
ProcessProposal : "9" ;


```

