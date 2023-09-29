package core

import (
	rpc "github.com/cometbft/cometbft/rpc/jsonrpc/server"
)

// TODO: better system than "unsafe" prefix

type RoutesMap map[string]*rpc.RPCFunc

const v1Prefix = "v1/"

// Routes is a map of available routes.
func (env *Environment) GetRoutes() RoutesMap {
	return RoutesMap{
		// subscribe/unsubscribe are reserved for websocket events.
		// v0
		"subscribe":       rpc.NewWSRPCFunc(env.Subscribe, "query"),
		"unsubscribe":     rpc.NewWSRPCFunc(env.Unsubscribe, "query"),
		"unsubscribe_all": rpc.NewWSRPCFunc(env.UnsubscribeAll, ""),

		// v1
		v1Prefix + "subscribe":       rpc.NewWSRPCFunc(env.Subscribe, "query"),
		v1Prefix + "unsubscribe":     rpc.NewWSRPCFunc(env.Unsubscribe, "query"),
		v1Prefix + "unsubscribe_all": rpc.NewWSRPCFunc(env.UnsubscribeAll, ""),

		// info AP
		// v0
		"health":               rpc.NewRPCFunc(env.Health, ""),
		"status":               rpc.NewRPCFunc(env.Status, ""),
		"net_info":             rpc.NewRPCFunc(env.NetInfo, ""),
		"blockchain":           rpc.NewRPCFunc(env.BlockchainInfo, "minHeight,maxHeight", rpc.Cacheable()),
		"genesis":              rpc.NewRPCFunc(env.Genesis, "", rpc.Cacheable()),
		"genesis_chunked":      rpc.NewRPCFunc(env.GenesisChunked, "chunk", rpc.Cacheable()),
		"block":                rpc.NewRPCFunc(env.Block, "height", rpc.Cacheable("height")),
		"block_by_hash":        rpc.NewRPCFunc(env.BlockByHash, "hash", rpc.Cacheable()),
		"block_results":        rpc.NewRPCFunc(env.BlockResults, "height", rpc.Cacheable("height")),
		"commit":               rpc.NewRPCFunc(env.Commit, "height", rpc.Cacheable("height")),
		"header":               rpc.NewRPCFunc(env.Header, "height", rpc.Cacheable("height")),
		"header_by_hash":       rpc.NewRPCFunc(env.HeaderByHash, "hash", rpc.Cacheable()),
		"check_tx":             rpc.NewRPCFunc(env.CheckTx, "tx"),
		"tx":                   rpc.NewRPCFunc(env.Tx, "hash,prove", rpc.Cacheable()),
		"tx_search":            rpc.NewRPCFunc(env.TxSearch, "query,prove,page,per_page,order_by"),
		"block_search":         rpc.NewRPCFunc(env.BlockSearch, "query,page,per_page,order_by"),
		"validators":           rpc.NewRPCFunc(env.Validators, "height,page,per_page", rpc.Cacheable("height")),
		"dump_consensus_state": rpc.NewRPCFunc(env.DumpConsensusState, ""),
		"consensus_state":      rpc.NewRPCFunc(env.GetConsensusState, ""),
		"consensus_params":     rpc.NewRPCFunc(env.ConsensusParams, "height", rpc.Cacheable("height")),
		"unconfirmed_txs":      rpc.NewRPCFunc(env.UnconfirmedTxs, "limit"),
		"num_unconfirmed_txs":  rpc.NewRPCFunc(env.NumUnconfirmedTxs, ""),

		// v1
		v1Prefix + "health":               rpc.NewRPCFunc(env.Health, ""),
		v1Prefix + "status":               rpc.NewRPCFunc(env.Status, ""),
		v1Prefix + "net_info":             rpc.NewRPCFunc(env.NetInfo, ""),
		v1Prefix + "blockchain":           rpc.NewRPCFunc(env.BlockchainInfo, "minHeight,maxHeight", rpc.Cacheable()),
		v1Prefix + "genesis":              rpc.NewRPCFunc(env.Genesis, "", rpc.Cacheable()),
		v1Prefix + "genesis_chunked":      rpc.NewRPCFunc(env.GenesisChunked, "chunk", rpc.Cacheable()),
		v1Prefix + "block":                rpc.NewRPCFunc(env.Block, "height", rpc.Cacheable("height")),
		v1Prefix + "block_by_hash":        rpc.NewRPCFunc(env.BlockByHash, "hash", rpc.Cacheable()),
		v1Prefix + "block_results":        rpc.NewRPCFunc(env.BlockResults, "height", rpc.Cacheable("height")),
		v1Prefix + "commit":               rpc.NewRPCFunc(env.Commit, "height", rpc.Cacheable("height")),
		v1Prefix + "header":               rpc.NewRPCFunc(env.Header, "height", rpc.Cacheable("height")),
		v1Prefix + "header_by_hash":       rpc.NewRPCFunc(env.HeaderByHash, "hash", rpc.Cacheable()),
		v1Prefix + "check_tx":             rpc.NewRPCFunc(env.CheckTx, "tx"),
		v1Prefix + "tx":                   rpc.NewRPCFunc(env.Tx, "hash,prove", rpc.Cacheable()),
		v1Prefix + "tx_search":            rpc.NewRPCFunc(env.TxSearch, "query,prove,page,per_page,order_by"),
		v1Prefix + "block_search":         rpc.NewRPCFunc(env.BlockSearch, "query,page,per_page,order_by"),
		v1Prefix + "validators":           rpc.NewRPCFunc(env.Validators, "height,page,per_page", rpc.Cacheable("height")),
		v1Prefix + "dump_consensus_state": rpc.NewRPCFunc(env.DumpConsensusState, ""),
		v1Prefix + "consensus_state":      rpc.NewRPCFunc(env.GetConsensusState, ""),
		v1Prefix + "consensus_params":     rpc.NewRPCFunc(env.ConsensusParams, "height", rpc.Cacheable("height")),
		v1Prefix + "unconfirmed_txs":      rpc.NewRPCFunc(env.UnconfirmedTxs, "limit"),
		v1Prefix + "num_unconfirmed_txs":  rpc.NewRPCFunc(env.NumUnconfirmedTxs, ""),

		// tx broadcast API
		// v0
		"broadcast_tx_commit": rpc.NewRPCFunc(env.BroadcastTxCommit, "tx"),
		"broadcast_tx_sync":   rpc.NewRPCFunc(env.BroadcastTxSync, "tx"),
		"broadcast_tx_async":  rpc.NewRPCFunc(env.BroadcastTxAsync, "tx"),

		// v1
		v1Prefix + "broadcast_tx_commit": rpc.NewRPCFunc(env.BroadcastTxCommit, "tx"),
		v1Prefix + "broadcast_tx_sync":   rpc.NewRPCFunc(env.BroadcastTxSync, "tx"),
		v1Prefix + "broadcast_tx_async":  rpc.NewRPCFunc(env.BroadcastTxAsync, "tx"),

		// abci API
		// v0
		"abci_query": rpc.NewRPCFunc(env.ABCIQuery, "path,data,height,prove"),
		"abci_info":  rpc.NewRPCFunc(env.ABCIInfo, "", rpc.Cacheable()),

		// v1
		v1Prefix + "abci_query": rpc.NewRPCFunc(env.ABCIQuery, "path,data,height,prove"),
		v1Prefix + "abci_info":  rpc.NewRPCFunc(env.ABCIInfo, "", rpc.Cacheable()),

		// evidence API
		// v0
		"broadcast_evidence": rpc.NewRPCFunc(env.BroadcastEvidence, "evidence"),

		// v1
		v1Prefix + "broadcast_evidence": rpc.NewRPCFunc(env.BroadcastEvidence, "evidence"),
	}
}

// AddUnsafeRoutes adds unsafe routes.
func (env *Environment) AddUnsafeRoutes(routes RoutesMap) {
	// control API
	// v0
	routes["dial_seeds"] = rpc.NewRPCFunc(env.UnsafeDialSeeds, "seeds")
	routes["dial_peers"] = rpc.NewRPCFunc(env.UnsafeDialPeers, "peers,persistent,unconditional,private")
	routes["unsafe_flush_mempool"] = rpc.NewRPCFunc(env.UnsafeFlushMempool, "")

	// v1
	routes[v1Prefix+"dial_seeds"] = rpc.NewRPCFunc(env.UnsafeDialSeeds, "seeds")
	routes[v1Prefix+"dial_peers"] = rpc.NewRPCFunc(env.UnsafeDialPeers, "peers,persistent,unconditional,private")
	routes[v1Prefix+"unsafe_flush_mempool"] = rpc.NewRPCFunc(env.UnsafeFlushMempool, "")
}
