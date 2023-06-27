#!/usr/bin/env bash

if [ $# -ne 2 ]; then
    echo "usage: gen.sh #nodes propagation_rate"
    exit -1
fi

NODES=$1
RATE=$2
PREAMBLE="initial_height=0\nprometheus = true\nload_tx_size_bytes = 100\nload_tx_batch_size = 100\npropagation_ratio=${RATE}"

echo -e $PREAMBLE
for i in $(seq -f "%02g" 1 ${NODES})
do
    echo "[node.validator${i}]"
    echo "  seeds = [\"validator01\",\"validator02\"]"
done
