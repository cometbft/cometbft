#!/usr/bin/env bash

DIR=$(dirname "${BASH_SOURCE[0]}")
BINDIR=${DIR}/../build/

source ${DIR}/utils.sh

echo "nodes,propagation rate,sent,seen,completion,total bandwidth,useful bandwidth,overhead"
for r in $(seq 5 5 40);
do
    for i in $(geometric 2 2 5);
    do
	${DIR}/gen.sh ${i} ${r} > ${DIR}/simple.toml
	${BINDIR}/runner -f ${DIR}/simple.toml mempool > ${TMPDIR}/log
	sent=$(grep "txs sent" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	seen=$(grep "txs seen" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	completion=$(grep "completion" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	totalBandwidth=$(grep "total bandwidth" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	usefulBandwidth=$(grep "useful bandwidth" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	overhead=$(grep "overhead" ${TMPDIR}/log | awk -F= '{print $2}' | sed -r 's/\s+//g')
	echo ${i}","${r}","${sent}","${seen}","${completion}","${totalBandwidth}","${usefulBandwidth}","${overhead}
    done
done
