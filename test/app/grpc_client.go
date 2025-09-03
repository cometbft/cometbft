package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	coregrpc "github.com/cometbft/cometbft/rpc/grpc"
)

var grpcAddr = "tcp://localhost:36656"

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Must enter a transaction to send (hex)")
		os.Exit(1)
	}
	tx := args[1]
	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		fmt.Println("Invalid hex", err)
		os.Exit(1)
	}

	//nolint:staticcheck // SA1019: core_grpc.StartGRPCClient is deprecated: A new gRPC API will be introduced after v0.38.
	clientGRPC := coregrpc.StartGRPCClient(grpcAddr)
	res, err := clientGRPC.BroadcastTx(context.Background(), &coregrpc.RequestBroadcastTx{Tx: txBytes})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bz, err := cmtjson.Marshal(res)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(bz))
}
