#!/bin/bash
# This script is invoked by OSS-Fuzz to run fuzz tests against CometBFT.
# See https://github.com/google/oss-fuzz/blob/master/projects/tendermint/build.sh

<<<<<<< HEAD
compile_go_fuzzer github.com/tendermint/tendermint/test/fuzz/mempool/v0 Fuzz mempool_v0_fuzzer
compile_go_fuzzer github.com/tendermint/tendermint/test/fuzz/mempool/v1 Fuzz mempool_v1_fuzzer
compile_go_fuzzer github.com/tendermint/tendermint/test/fuzz/p2p/addrbook Fuzz p2p_addrbook_fuzzer
compile_go_fuzzer github.com/tendermint/tendermint/test/fuzz/p2p/pex Fuzz p2p_pex_fuzzer
compile_go_fuzzer github.com/tendermint/tendermint/test/fuzz/p2p/secret_connection Fuzz p2p_secret_connection_fuzzer
compile_go_fuzzer github.com/tendermint/tendermint/test/fuzz/rpc/jsonrpc/server Fuzz rpc_jsonrpc_server_fuzzer
=======
export FUZZ_ROOT="github.com/cometbft/cometbft"

build_go_fuzzer() {
	local function="$1"
	local fuzzer="$2"

	go run github.com/orijtech/otils/corpus2ossfuzz@latest -o "$OUT"/"$fuzzer"_seed_corpus.zip -corpus test/fuzz/tests/testdata/fuzz/"$function"
	compile_native_go_fuzzer "$FUZZ_ROOT"/test/fuzz/tests "$function" "$fuzzer"
}

go get github.com/AdamKorcz/go-118-fuzz-build/testing
go get github.com/prometheus/common/expfmt@v0.32.1

build_go_fuzzer FuzzP2PSecretConnection fuzz_p2p_secretconnection

build_go_fuzzer FuzzMempool fuzz_mempool

build_go_fuzzer FuzzRPCJSONRPCServer fuzz_rpc_jsonrpc_server
>>>>>>> 1cb55d49b (Rename Tendermint to CometBFT: further actions (#224))
