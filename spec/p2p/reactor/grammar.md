# Grammar

The following is a simplified representation of the expected sequence of calls
from the P2P layer to a reactor:


```abnf
start           = add-reactor on-start *peer-connection on-stop
add-reactor     = get-channels set-switch

; Refers to a single peer, reactor should support multiple concurrent peers
peer-connection = init-peer peer-start
peer-start      = [receive] (peer-connected / start-error)
peer-connected  = add-peer *receive peer-stop
peer-stop       = [peer-error] remove-peer

; Service interface
on-start        = %s"<OnStart>"
on-stop         = %s"<OnStop>"
; Reactor interface
get-channels    = %s"<GetChannels>"
set-switch      = %s"<SetSwitch>"
init-peer       = %s"<InitPeer>"
add-peer        = %s"<AddPeer>"
remove-peer     = %s"<RemovePeer>"
receive         = %s"<Receive>"

; Errors, for reference
start-error     = %s"Error starting peer"
peer-error      = %s"Stopping peer for error"
```

This is a grammar written in case-sensitive Augmented Backus–Naur form (ABNF,
specified in [IETF rfc7405](https://datatracker.ietf.org/doc/html/rfc7405)).
It is inspired on the [grammar](../../abci/abci%2B%2B_comet_expected_behavior.md)
written to specify the interaction of CometBFT with an ABCI++ application.
