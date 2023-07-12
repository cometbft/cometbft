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
load_tx_to_send = 100\n\
load_tx_batch_size = 10\n\
pex = false\n\
experimental_gossip_propagation_rate = ${RATE}\n\
experimental_custom_reactors = {CONSENSUS = \"p2p.mock.reactor\", MEMPOOL = \"experimental.reactors.mempool.gossip\"}\
"

# let's form a clique

ALL_NODES=$(seq -f "%02g" 1 ${NODES})
echo -e $PREAMBLE
for i in ${ALL_NODES}
do
    echo "[node.validator${i}]"    
    echo "  persistent_peers = ["$(echo ${ALL_NODES} | sed -r 's/([0-9]*)/"validator\1"/g' | sed s/\ /,/g )"]"
done
