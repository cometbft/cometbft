#! /bin/bash

set -o errexit   # abort on nonzero exitstatus
set -o nounset   # abort on unbound variable
set -o pipefail  # don't hide errors within pipes

function toHex() {
    echo -n "$1" | hexdump -ve '1/1 "%.2X"' | awk '{print "0x" $0}'
}

#####################
# kvstore with curl
#####################
TESTNAME=$1

# store key value pair
KEY="abcd"
VALUE="dcba"
toHex $KEY=$VALUE
curl -s 127.0.0.1:26657/broadcast_tx_commit?tx="$(toHex $KEY=$VALUE)"
echo $?
echo ""

###########################
# test using the abci-cli
###########################

echo "... testing query with abci-cli"

# we should be able to look up the key
RESPONSE=$(abci-cli query \"$KEY\")

if ! grep -q "$VALUE" <<< "$RESPONSE"; then
    echo "Failed to find $VALUE for $KEY. Response:"
    echo "$RESPONSE"
    exit 1
fi

# we should not be able to look up the value
RESPONSE=$(abci-cli query \"$VALUE\")
if grep -q "value: $VALUE" <<< "$RESPONSE"; then
		echo "Found 'value: $VALUE' for $VALUE when we should not have. Response:"
    echo "$RESPONSE"
    exit 1
fi

#############################
# test using the /abci_query
#############################

echo "... testing query with /abci_query 2"

# we should be able to look up the key
RESPONSE=$(curl -s "127.0.0.1:26657/abci_query?path=\"\"&data=$(toHex $KEY)&prove=false")
RESPONSE=$(echo "$RESPONSE" | jq .result.response.log)

if ! grep -q "exists" <<< "$RESPONSE"; then
    echo "Failed to find 'exists' for $KEY. Response:"
    echo "$RESPONSE"
    exit 1
fi

# we should not be able to look up the value
RESPONSE=$(curl -s "127.0.0.1:26657/abci_query?path=\"\"&data=$(toHex $VALUE)&prove=false")
RESPONSE=$(echo "$RESPONSE" | jq .result.response.log)
if ! grep -q "does not exist" <<< "$RESPONSE"; then
		echo "Failed to find 'does not exist' for $VALUE. Response:"
    echo "$RESPONSE"
    exit 1
fi

echo "Passed Test: $TESTNAME"
