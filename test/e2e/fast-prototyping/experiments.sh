#!/usr/bin/env bash

DIR=$(dirname "${BASH_SOURCE[0]}")
BINDIR=${DIR}/../build
TMPDIR=/tmp

DOCKER_VERSION=$(docker version | grep Version | head -n 1  | awk '{print $2}' | sed s,\\.,,g);
if [[ ${DOCKER_VERSION} == "" ]] || [ ${DOCKER_VERSION} -lt 2400 ]; then
	echo "Docker >= 24.0 required";
	exit 1
fi

if [ $# -lt 1 ]; then
    echo "usage: experiments output.csv [out_degree] [-none|solo|all]"
    echo "where -none = (default) no validator, consensus reactor is mocked everywhere."
    echo "      -solo = validator01 has full power, the rest are full nodes."
    echo "      -all  = all the nodes are validating with the same power."
    exit 1
fi

FILE=$1
DEGREE=$2
MODE=$3
PROP_RATE=(50 60 70 80 90 100)
NUM_NODES=(8 16 32 64)

TMPL="custom"

# this generate a csv file
# the dimensions of each entry in this file are below (nodes and propagation_rate are the x and y axis when rendering)
echo "nodes;propagation_rate;submitted;added;sent;completion;total_mempool_bandwidth;useful_mempool_bandwidth;overhead;redundancy;degree;cpu_load;latency;bandwidth" > ${FILE}
for r in "${PROP_RATE[@]}"
do
    for i in "${NUM_NODES[@]}"
    do
	${DIR}/tmpl-gen.sh ${i} ${r} ${DEGREE} ${MODE}> ${TMPDIR}/${TMPL}.toml
	${BINDIR}/runner -f ${TMPDIR}/${TMPL}.toml custom > ${TMPDIR}/log
	submitted=$(grep "txs submitted" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	added=$(grep "txs added" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	sent=$(grep "txs sent" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	completion=$(grep "completion" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	totalBandwidth=$(grep "total mempool bandwidth" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	usefulBandwidth=$(grep "useful mempool bandwidth" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	overhead=$(grep "overhead" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	redundancy=$(grep "redundancy" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	degree=$(grep "degree" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	cpu=$(grep "cpu load" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	latency=$(grep "latency" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	bandwidth=$(grep "bandwidth graph" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	echo ${i}";"${r}";"${submitted}";"${added}";"${sent}";"${completion}";"${totalBandwidth}";"${usefulBandwidth}";"${overhead}";"${redundancy}";"${degree}";"${cpu}";"${latency}";"${bandwidth} >> ${FILE}
	${BINDIR}/runner -f ${NETDIR}/${TMPL}.toml cleanup >> /dev/null
	sleep 1
    done
done

cat ${FILE}
