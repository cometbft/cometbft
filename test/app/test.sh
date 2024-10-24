#! /bin/bash

set -o errexit   # abort on nonzero exitstatus
set -o nounset   # abort on unbound variable
set -o pipefail  # don't hide errors within pipes

#- kvstore over socket, curl

export CMTHOME=$HOME/.cometbft

function kvstore_over_socket(){
    rm -rf "$CMTHOME"
    cometbft init
    echo "Starting kvstore_over_socket"
    abci-cli kvstore > /dev/null &
    pid_kvstore=$!
    cometbft node > cometbft.log &
    pid_cometbft=$!
    sleep 5

    echo "running test"
    bash test/app/kvstore_test.sh "KVStore over Socket"

    kill -9 $pid_kvstore $pid_cometbft
}

# start cometbft first
function kvstore_over_socket_reorder(){
    rm -rf "$CMTHOME"
    cometbft init
    echo "Starting kvstore_over_socket_reorder (ie. start cometbft first)"
    cometbft node > cometbft.log &
    pid_cometbft=$!
    sleep 2
    abci-cli kvstore > /dev/null &
    pid_kvstore=$!
    sleep 5

    echo "running test"
    bash test/app/kvstore_test.sh "KVStore over Socket"

    kill -9 $pid_kvstore $pid_cometbft
}

kvstore_over_socket
kvstore_over_socket_reorder
