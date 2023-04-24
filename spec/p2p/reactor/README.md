# Reactors

Reactor is the generic name for a component that employs the p2p communication layer.

The main source of documentation for a reactor is the [`Reactor`](../../../p2p/base_reactor.go) interface.
In order to use communication services provided by the p2p layer,
a component has to implement the p2p's package `Reactor` interface.

The goal of this document is to complement the existing documentation by
specifying the behaviour of the p2p when interacting with a reactor.
So while the `Reactor` interface defines the methods invoked and determines
what the p2p layer expects from a reactor, this documentation focuses on the
behaviour that a reactor implementation should expect from the p2p layer.

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
start           = registration on-start *peer-management on-stop
registration    = get-channels set-switch

; Refers to a single peer, a reactor should support multiple concurrent peers
peer-management = init-peer peer-start peer-stop
peer-start      = [receive] (peer-connected / start-error)
peer-connected  = add-peer *receive
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
`Switch.AddReactor(name string, reactor Reactor)` method, where `name` can be
an arbitrary string.

The registration should happen before the node, in general, and the p2p layer,
in particular, are started.
In other words, there is no support for registering a reactor on a running node:
reactors must be registered as part of the setup of a node.

```abnf
registration     = get-channels set-switch
```

The p2p layer retrieves from the reactor a list of channels the reactor is
responsible for, using the `GetChannels()` method.
The reactor implementation should thereafter expect the delivery of every
message received by the p2p layer in the informed channels.

The second method `SetSwitch(Switch)` concludes the handshake between the
reactor and the p2p layer.
The `Switch` is the main component of the p2p layer, being responsible for
establishing connections with peers and routing messages.

### Service interface

A reactor should implement the [`Service`](../../../libs/service/service.go) interface,
in particular, a startup `OnStart()` and a shutdown `OnStop()` methods:

```abnf
start           = registration on-start *peer-management on-stop
```

As part of the startup of a node, all registered reactors are started by the p2p layer.
And when the node is shutdown, all registered reactors are stopped by the p2p layer.
Observe that the `Service` interface specification establishes that a service
can be started and stopped only once.
So before being started or once stopped by the p2p layer, the reactor should
not expect any interaction.

### Peer management

```abnf
; Refers to a single peer, a reactor should support multiple concurrent peers
peer-management = init-peer peer-start peer-stop
```

### Add peer

The p2p layer informs all registered reactors when it establishes a connection
with a `Peer`, using the `InitPeer(Peer)` method.

It is up to the reactor to define how to process this event.
The typical behavior is to setup routines that, given some conditions or events,
send messages to the added peer, using the provided `Peer` handler.

Adding a peer to a reactor has two steps.
In the first step, the `Peer` has not yet been started.
This step should be used to initialize state or data related to the new peer,
but not to interact with it:

```abnf
peer-management = init-peer peer-start peer-stop
```

The second step is performed after the peer's send and receive routines are
started without errors.
The updated `Peer` handler provided to this method can then be used to interact with the peer:

```abnf
peer-start      = [receive] (peer-connected / start-error)
peer-connected  = add-peer *receive
```

### Remove Peer

The p2p layer also informs all registered reactors when it disconnects from a `Peer`:

```abnf
peer-stop       = [peer-error] remove-peer
```

When this method is invoked, the peer's send and receive routine were already stopped.
This means that the reactor should not receive any further message from this
peer and should not try sending messages to the removed peer.

### Receiving messages

The main duty of a reactor is to handle incoming messages on the channels it
has registered with the p2p layer.

When a message is received from a connected peer on any of the channels
registered by the reactor, the node will deliver the message to the reactor
invoking the `Receive(Envelope)` method.

```abnf
peer-start      = [receive] (peer-connected / start-error)
peer-connected  = add-peer *receive
```

Notice that _pre-condition_ for receiving a message from a `Peer` is that the
p2p layer has previously invoked `InitPeer(Peer)` for that peer.
While this is not the common case, the reactor should be able to handle
messages from a `Peer` before `AddPeer(Peer)` is invoked.
This happens because starting the peer's send and receive routines
and adding the peer are done in parallel.

The reactor receives a message packed into an `Envelope` with the following content:

- `ChannelID`: the channel the message belongs to
- `Src`: the `Peer` from which the message was received
- `Message`: the message's payload, unmarshalled using protocol buffers

Two important observations regarding the implementation of the `Receive` method:

1. Concurrency: the implementation should consider concurrent invocations of
   the `Receive` method, as messages received from different peers about at the
same time can be delivered to the reactor concurrently.
1. The implementation should be non-blocking, as it is directly invoked
   by the peers' receive routines.
