#!/usr/bin/env bash

DIR=$(dirname "${BASH_SOURCE[0]}")
TMPDIR=/tmp

info() {
    local message=$1
    echo >& 2 "["$(date +%s:%N)"] ${message}"
}

log() {
    if [[ $(config verbose) -eq 1 ]]
    then
	local message=$1
	echo >& 2 "["$(date +%s:%N)"] ${message}"
    fi
}

function geometric {
    if [ $# -ne 3 ]; then
	echo "usage: genometric #val #mult #count"
	exit -1
    fi
    
    ((val = $1))
    ((mult = $2))
    ((count = $3))

    while [[ ${count} -gt 0 ]] ; do
        echo ${val}
        ((val *= mult))
        ((count -= 1))
    done
}
