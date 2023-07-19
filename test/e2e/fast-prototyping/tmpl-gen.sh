#!/usr/bin/env bash

DIR=$(dirname "${BASH_SOURCE[0]}")
source ${DIR}/utils.sh

if [ $# -lt 2 ];
then
    echo "usage: gen.sh #nodes propagation_rate [out_degree]"
    exit 1
fi

NODES=$1
RATE=$2
DEGREE=$3

PREAMBLE="\
prometheus = true\n\
load_tx_size_bytes = 100\n\
load_tx_to_send = 100\n\
load_tx_batch_size = 10\n\
pex = false\n\
experimental_gossip_propagation_rate = ${RATE}\n\
experimental_custom_reactors = {CONSENSUS = \"p2p.mock.reactor\", MEMPOOL = \"experimental.reactors.mempool.gossip\"}\
"

ALL_NODES=$(seq -f "%02g" 1 ${NODES})

if [[ ${DEGREE} == "" ]]
then
    DEGREE=$(echo ${ALL_NODES[@]} | wc -w)
fi

echo -e $PREAMBLE
for i in ${ALL_NODES}
do
    peers=($(shuf -e ${ALL_NODES[@]}))
    echo "[node.validator${i}]"
    echo "  persistent_peers = ["$(echo ${peers[@]:0:${DEGREE}} | sed -r 's/([0-9]*)/"validator\1"/g' | sed s/\ /,/g )"]"
done
