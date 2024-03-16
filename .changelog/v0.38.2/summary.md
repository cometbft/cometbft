*November 27, 2023*

This release provides the **nop** mempool for applications that want to build their own mempool.
Using this mempool effectively disables all mempool functionality in CometBFT, including transaction dissemination and the `broadcast_tx_*` endpoints.

Also fixes a small bug in the mempool for an experimental feature.
