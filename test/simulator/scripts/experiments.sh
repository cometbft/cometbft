#!/usr/bin/env bash

DIR=$(dirname "${BASH_SOURCE[0]}")
source ${DIR}/utils.sh

echo "nodes;propagation_rate;sent;seen;completion;total_bandwidth;useful_bandwidth;redundancy;overhead;bandwidth"
for r in $(seq 100 100 100);
do
    for i in $(geometric 3 1 1);
    do
	${DIR}/tmpl-gen.sh ${i} ${r} > ${NETDIR}/simple.toml
	${BINDIR}/runner -f ${NETDIR}/simple.toml mempool > ${TMPDIR}/log
	sent=$(grep "txs sent" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	seen=$(grep "txs seen" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	completion=$(grep "completion" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	totalBandwidth=$(grep "total bandwidth" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	usefulBandwidth=$(grep "useful bandwidth" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	overhead=$(grep "overhead" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	redundancy=$(grep "redundancy" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	bandwidth=$(grep "bandwidth graph" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	echo ${i}";"${r}";"${sent}";"${seen}";"${completion}";"${totalBandwidth}";"${usefulBandwidth}";"${overhead}";"${redundancy}";"${bandwidth}
    done
done
