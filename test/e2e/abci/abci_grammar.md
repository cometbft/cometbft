```
package "github.com/cometbft/cometbft/test/e2e/abci"


Start : CleanStart | Recovery ;

CleanStart : InitChain StateSync ConsensusExec | InitChain ConsensusExec ;
StateSync : StateSyncAttempts SuccessSync ; 
StateSyncAttempts : StateSyncAttempt | StateSyncAttempt StateSyncAttempts ;
StateSyncAttempt : OfferSnapshot ApplyChunks ;
SuccessSync : OfferSnapshot ApplyChunks | OfferSnapshot ; 
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
BeginBlock : "3" ; 
DeliverTx : "4" ;
EndBlock : "5" ;
Commit : "6" ;
OfferSnapshot : "10" ;
ApplyChunk : "11" ; 
PrepareProposal : "13" ; 
ProcessProposal : "14" ;


```