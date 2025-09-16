#! /bin/bash
set -u

function toHex() {
    echo -n "$1" | hexdump -ve '1/1 "%.2X"'
}

N=$1
PORT=$2

for ((i=1; i<=N; i++)); do
    # store key value pair
    KEY=$(head -c 10 /dev/urandom | hexdump -ve '1/1 "%.2X"')
    VALUE="$i"
    echo $(toHex "$KEY=$VALUE")
    curl "127.0.0.1:$PORT/broadcast_tx_sync?tx=0x$(toHex "$KEY=$VALUE")"
done
