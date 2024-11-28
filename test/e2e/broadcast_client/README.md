# Broadcast Client

Generates a load using a specific broadcast type and compares the latency with the latency of blocks received through ws.

Basic usage:
```sh
make
./build/client  -txBatches 5 --totalConcurrentTx 2 -debug -broadcastType async
python script.py log.txt
```
