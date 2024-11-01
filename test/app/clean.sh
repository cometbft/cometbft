#! /bin/bash

set -o nounset   # abort on unbound variable
set -o pipefail  # don't hide errors within pipes

killall cometbft
killall abci-cli
rm -rf ~/.cometbft
