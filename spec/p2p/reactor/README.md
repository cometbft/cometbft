# Reactors

Reactor is the generic name for a component that interacts with the P2P communication layer.

This document aims to specify the operation of CometBFT reactors.

This is a work in progress, tracked by the issue [#599](https://github.com/cometbft/cometbft/issues/599).

## Interface

To become a reactor, a component has to implement the
[`p2p.Reactor`](../../../p2p/base_reactor.go) interface.

Much of the expected operation of a reactor can be derived from the
documentation of this interface.

## Grammar

The expected sequence of calls to a reactor from the P2P layer will be
described using a [Grammar](./grammar.md).


## Quint

The expected operation of a reactor will be modelled using
[Quint](https://github.com/informalsystems/quint),
an executable specification language.

A reactor is a [`Service`](../../../libs/service/service.go) controlled by the P2P layer.
We modelled a generic service in [quint](./service.qnt).
