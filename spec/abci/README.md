---
order: 1
parent:
  title: ABCI
  order: 2
---

# ABCI

ABCI stands for "**A**pplication **B**lock**c**hain **I**nterface".
ABCI is the interface between CometBFT (a state-machine replication engine)
and your application (the actual state machine). It consists of a set of
_methods_, each with a corresponding `Request` and `Response`message type. 
To perform state-machine replication, CometBFT calls the ABCI methods on the 
ABCI application by sending the `Request*` messages and receiving the `Response*` messages in return.

<<<<<<< HEAD
All ABCI messages and methods are defined in [protocol buffers](https://github.com/cometbft/cometbft/blob/v0.34.x/proto/abci/types.proto). 
=======
ABCI++ is a major evolution of ABCI (**A**pplication **B**lock**c**hain **I**nterface).
Like its predecessor, ABCI++ is the interface between CometBFT (a state-machine
replication engine) and the actual state machine being replicated (i.e., the Application).
The API consists of a set of _methods_, each with a corresponding `Request` and `Response`
message type.

The methods are always initiated by CometBFT. The Application implements its logic
for handling all ABCI++ methods.
Thus, CometBFT always sends the `Request*` messages and receives the `Response*` messages
in return.

All ABCI++ messages and methods are defined in [protocol buffers](https://github.com/cometbft/cometbft/blob/main/proto/tendermint/abci/types.proto).
>>>>>>> 28baba3ed (Docs fixes (#368))
This allows CometBFT to run with applications written in many programming languages.

This specification is split as follows:

<<<<<<< HEAD
- [Methods and Types](./abci.md) - complete details on all ABCI methods and
  message types
- [Applications](./apps.md) - how to manage ABCI application state and other
  details about building ABCI applications
- [Client and Server](./client-server.md) - for those looking to implement their
  own ABCI application servers
=======
- [Overview and basic concepts](./abci++_basic_concepts.md) - interface's overview and concepts
  needed to understand other parts of this specification.
- [Methods](./abci++_methods.md) - complete details on all ABCI++ methods
  and message types.
- [Requirements for the Application](./abci++_app_requirements.md) - formal requirements
  on the Application's logic to ensure CometBFT properties such as liveness. These requirements define what
  CometBFT expects from the Application; second part on managing ABCI application state and related topics.
- [CometBFT's expected behavior](./abci++_comet_expected_behavior.md) - specification of
  how the different ABCI++ methods may be called by CometBFT. This explains what the Application
  is to expect from CometBFT.
- [Example scenarios](./abci++_example_scenarios.md) - specific scenarios showing why the Application needs to account
for any CometBFT's behaviour prescribed by the specification.
- [Client and Server](./abci++_client_server.md) - for those looking to implement their
  own ABCI application servers.
>>>>>>> 28baba3ed (Docs fixes (#368))
