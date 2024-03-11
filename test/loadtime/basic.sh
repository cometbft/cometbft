#!/bin/sh

set -euo pipefail

# A basic invocation of the loadtime tool.

#./build/load \
#    -c 1 -T 600 -r 400 -s 1024 \
#    --broadcast-tx-method sync \
#    --endpoints ws://localhost:5701/v1/websocket

#sleep 10

#./build/load \
#    -c 1 -T 600 -r 1000 -s 1024 \
#    --broadcast-tx-method sync \
#    --endpoints ws://localhost:5704/v1/websocket

#sleep 10

#echo "Sending to 5703"
#./build/load \
#    -c 1 -T 3600 -r 1000 -s 1024 \
#    --broadcast-tx-method sync \
#    --endpoints ws://localhost:5707/v1/websocket


sleep 10 

echo "Sending to 5704"
./build/load \
    -c 1 -T 10 -r 500 -s 1024 \
    --broadcast-tx-method sync \
    --endpoints ws://localhost:26657/v1/websocket

echo "Sending to 5705"
sleep 10

./build/load \
    -c 1 -T 600 -r 400 -s 1024 \
    --broadcast-tx-method sync \
    --endpoints ws://localhost:5713/v1/websocket

sleep 10
echo "Sending to 5706"

./build/load \
    -c 1 -T 600 -r 400 -s 1024 \
    --broadcast-tx-method sync \
    --endpoints ws://localhost:5716/v1/websocket

echo "Sending to 5706"

./build/load \
    -c 1 -T 600 -r 400 -s 1024 \
    --broadcast-tx-method sync \
    --endpoints ws://localhost:5719/v1/websocket
