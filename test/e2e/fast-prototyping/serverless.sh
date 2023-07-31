#!/bin/bash

DIR=$(dirname "${BASH_SOURCE[0]}")
source ${DIR}/utils.sh

for r in $(cat ${DIR}/render/results.csv | awk -F';' '{if (NR!=1) {print $3}}' );
do
    gsutil cp ${DIR}/render/${r} gs://render-experiments/${r}
    gcloud storage objects update gs://render-experiments/${r} --add-acl-grant=entity=AllUsers,role=READER
done

sed -r 's,;([^;]+\.csv),;https://storage\.googleapis\.com/render-experiments/\1,g' ${DIR}/render/results.csv > ${TMPDIR}/results.csv
gsutil cp ${TMPDIR}/results.csv gs://render-experiments/results.csv
gcloud storage objects update gs://render-experiments/results.csv --add-acl-grant=entity=AllUsers,role=READER

sed s,results\.csv,https://storage\.googleapis\.com/render-experiments/results\.csv, ${DIR}/render/experiments.js > ${TMPDIR}/experiment.js
gsutil cp ${TMPDIR}/experiment.js gs://render-experiments/experiments.js
gcloud storage objects update gs://render-experiments/experiments.js --add-acl-grant=entity=AllUsers,role=READER

