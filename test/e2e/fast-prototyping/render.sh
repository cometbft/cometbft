#!/usr/bin/env bash

DIR=$(dirname "${BASH_SOURCE[0]}")
source ${DIR}/utils.sh

trap "pkill -KILL -P $$; exit 255" SIGINT SIGTERM

python3 -m http.server --directory ${DIR}/render 8080 &
echo "Point your browser to http://localhost:8080/experiments.html"

wait

