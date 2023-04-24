# Reactors

Reactor is the generic name for a component that employs the p2p communication layer.

The main source of documentation for a reactor is the p2p's
[`Reactor`](../../../p2p/base_reactor.go) interface.
A component has to implement the `Reactor` interface in order to use the
communication services provided by the p2p layer.

The goal of this document is to complement the existing documentation by
specifying the behaviour of the p2p when interacting with a reactor.
So while the `Reactor` interface determines what the p2p layer expects from a
reactor, this documentation focuses on the behaviour that a reactor
implementation should expect from the p2p layer.

> This is a work in progress, tracked by [issue #599](https://github.com/cometbft/cometbft/issues/599).


## Contents

The operation of the p2p layer when interacting with a reactor is modelled in
[Quint](https://github.com/informalsystems/quint), an executable specification language.

The Quint specification is available on the companion [`reactor.qnt`](./reactor.qnt) file.
The `Reactor` interface methods are the base for the specification, while
the behaviour of the p2p layer described in the remaining of this document is
modeled through state transitions.


## Operation

The following _grammar_ is a simplified representation of the expected sequence of calls
from the p2p layer to a reactor:


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

The grammar is written in case-sensitive Augmented Backus–Naur form (ABNF,
specified in [IETF rfc7405](https://datatracker.ietf.org/doc/html/rfc7405)).
It is inspired on the grammar produced to specify the interaction of CometBFT
with an ABCI++ application, available [here](../../abci/abci%2B%2B_comet_expected_behavior.md).

### Registration

To become a reactor, a component has first to implement the `Reactor` interface,
then to register the implementation with the p2p layer, using the
`Switch.AddReactor(name string, reactor Reactor)` method.

The registration should happen before the node, in general, and the p2p layer,
in particular, are started.
In other words, there is no support for registering a reactor on a running node:
reactors must be registered as part of the setup of a node.

```abnf
add-reactor     = get-channels set-switch
```

As part of the registration procedure, the p2p layer retrieves from the reactor
a list of channels the reactor is responsible for, using the `GetChannels()` method.
The reactor implementation should thereafter expect that every message received
by the p2p layer in the informed channels will be delivered to the reactor.

The second method `SetSwitch(Switch)` concludes the registration of a reactor,
which can be seen as a handshake between the reactor and the p2p layer.
The `Switch` is the main component of the p2p layer, being responsible for
establishing connections with peers and routing messages.

### Service interface

A reactor should implement the [`Service`](../../../libs/service/service.go) interface.
As such, it should implement a startup `OnStart()` and a shutdown `OnStop()` methods.

```abnf
start           = add-reactor on-start *peer-connection on-stop
```

As part of the startup of a node, all registered reactors are started by the p2p layer.
When the node is shutdown, all registered reactors are stopped by the p2p layer.
Observe that the `Service` interface specification establishes that a service
can be started and stopped only once.
