#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

mkdir -p generated

echo "== forge build --force =="
forge build --force 2>&1 | tee generated/forge-build.txt

echo "== forge test --summary =="
forge test --summary | tee generated/forge-summary.txt

echo "== forge test --gas-report --summary =="
forge test --gas-report --summary | tee generated/gas-report.txt

echo "== forge test --match-test testCalldataGasEstimates -vvvv =="
forge test --match-test testCalldataGasEstimates -vvvv | tee generated/calldata-report.txt

for test in \
  testIcs23IavlExistenceProofDecodeOnlyGas \
  testIcs23IavlExistenceProofVerifyGas
do
  echo "== forge test --gas-report --summary --match-test ${test} =="
  forge test --gas-report --summary --match-test "${test}" | tee "generated/${test}.txt"
done

for n in 10 50 100 175; do
  echo "== forge test --gas-report --summary --match-test testBenchBlsMultiMessagePairing${n}Gas =="
  forge test --gas-report --summary --match-test "testBenchBlsMultiMessagePairing${n}Gas" \
    | tee "generated/bls-d-${n}.txt"
done

go run ./scripts/extract_gas_report.go
