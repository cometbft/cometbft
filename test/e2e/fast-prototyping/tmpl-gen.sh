#!/usr/bin/env bash

DIR=$(dirname "${BASH_SOURCE[0]}")
source ${DIR}/utils.sh

if [ $# -lt 2 ];
then
    echo "usage: gen.sh #nodes propagation_rate [out_degree] [-none|solo|all]"
    echo "where -none = (default) no validator, consensus reactor is mocked everywhere."
    echo "      -solo = validator01 has full power, the rest are full nodes."
    echo "      -all  = all the nodes are validating with the same power"
    exit 1
fi

NODES=$1
RATE=$2
DEGREE=$3
MODE=$4

VALIDATORS=""
CUSTOM_CONSENSUS_REACTOR="p2p.mock.reactor"
CUSTOM_MEMPOOL_REACTOR="experimental.reactors.mempool.gossip"
INITIAL_HEIGHT=0

SEND_ONCE="false"

if [[ "${MODE}" == "-solo" ]] || [[ "${MODE}" == "-all" ]];
then
	  INITIAL_HEIGHT=1
    CUSTOM_CONSENSUS_REACTOR=""
fi

PREAMBLE="\
initial_height = ${INITIAL_HEIGHT}\n\
prometheus = true\n\
load_tx_size_bytes = 100\n\
load_tx_to_send = 100\n\
load_tx_batch_size = 10\n\
pex = false\n\
experimental_gossip_propagation_rate = ${RATE}\n\
experimental_gossip_send_once = ${SEND_ONCE}\n\
experimental_custom_reactors = {CONSENSUS = \"${CUSTOM_CONSENSUS_REACTOR}\", MEMPOOL = \"${CUSTOM_MEMPOOL_REACTOR}\"}
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
    if [[ "${MODE}" == "-solo" ]] && [[ "${i}" != "01" ]];
    then
        echo "  mode = \"full\""
    fi
done
