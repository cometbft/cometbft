#!/usr/bin/env bash

if [ $# -ne 2 ]; then
    echo "usage: gen.sh #nodes propagation_rate"
    exit 1
fi

NODES=$1
RATE=$2
PREAMBLE="\
prometheus = true\n\
load_tx_size_bytes = 100\n\
load_tx_to_send = 1000\n\
load_tx_batch_size = 100\n\
experimental_gossip_propagation_rate = ${RATE}\n\
experimental_custom_reactors = {CONSENSUS = \"p2p.mock.reactor\", MEMPOOL = \"experimental.reactors.mempool.gossip\"}\
"

echo -e $PREAMBLE
for i in $(seq -f "%02g" 1 ${NODES})
do
    echo "[node.validator${i}]"
    echo "  seeds = [\"validator01\",\"validator02\"]"
done
