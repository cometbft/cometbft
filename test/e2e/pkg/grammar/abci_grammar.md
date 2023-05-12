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


InitChain : "<InitChain>" ;
BeginBlock : "<BeginBlock>" ; 
DeliverTx : "<DeliverTx>" ;
EndBlock : "<EndBlock>" ;
Commit : "<Commit>" ;
OfferSnapshot : "<OfferSnapshot>" ;
ApplyChunk : "<ApplyChunk>" ; 
PrepareProposal : "<PrepareProposal>" ; 
ProcessProposal : "<ProcessProposal>" ;


```

