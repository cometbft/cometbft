#!/usr/bin/env bash

set -u

INV=${1:-notFullChain}
[[ ! -z "${INV}" ]] && echo "Generating trace for ${INV}..." || (echo "INV is empty" && exit 1)

TRACES_DIR="traces/$INV"
mkdir -p "$TRACES_DIR"

function nextFilename() {
    local dir=$1
    local name=$2
    local ext=$3
    i=1
    while [[ -e "$dir/$name-$i.$ext" || -L "$dir/$name-$i.$ext" ]] ; do
        let i++
    done
    name=$name-$i
    echo $name.$ext
}

FILE_NAME=$(nextFilename $TRACES_DIR "${INV}_trace" "itf.json")
TRACE_PATH="$TRACES_DIR/$FILE_NAME"
echo "trace path: $TRACE_PATH"

time quint run \
    --verbosity 5 \
    --max-steps=100 \
    --max-samples=3 \
    --invariant "$INV" \
    --out-itf "$TRACE_PATH" \
    mempoolv0.qnt
