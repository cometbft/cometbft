#!/bin/sh

set -euo pipefail

# A basic invocation of the loadtime tool.

./build/load \
    -c 1 -T 600 -r 1000 -s 2048 \
    --broadcast-tx-method sync \
    --endpoints ws://localhost:26657/v1/websocket

