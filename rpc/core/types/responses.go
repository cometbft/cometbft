package coretypes

import (
	"encoding/json"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
)

// List of blocks.
type ResultBlockchainInfo struct {
	BlockMetas []*types.BlockMeta `json:"block_metas"`
	LastHeight int64              `json:"last_height"`
}

// Genesis file.
type ResultGenesis struct {
	Genesis *types.GenesisDoc `json:"genesis"`
}

// ResultGenesisChunk is the output format for the chunked/paginated
// interface. These chunks are produced by converting the genesis
// document to JSON and then splitting the resulting payload into
// 16 megabyte blocks and then base64 encoding each block.
type ResultGenesisChunk struct {
	Data        string `json:"data"`
	ChunkNumber int    `json:"chunk"`
	TotalChunks int    `json:"total"`
}

// Single block (with meta).
type ResultBlock struct {
	Block   *types.Block  `json:"block"`
	BlockID types.BlockID `json:"block_id"`
}

// ResultHeader represents the response for a Header RPC Client query.
type ResultHeader struct {
	Header *types.Header `json:"header"`
}

// Commit and Header.
type ResultCommit struct {
	types.SignedHeader `json:"signed_header"`
	CanonicalCommit    bool `json:"canonical"`
}

// ABCI results from a block.
type ResultBlockResults struct {
	ConsensusParamUpdates *cmtproto.ConsensusParams `json:"consensus_param_updates"`
	TxResults             []*abci.ExecTxResult      `json:"txs_results"`
	FinalizeBlockEvents   []abci.Event              `json:"finalize_block_events"`
	ValidatorUpdates      []abci.ValidatorUpdate    `json:"validator_updates"`
	AppHash               []byte                    `json:"app_hash"`
	Height                int64                     `json:"height"`
}

// NewResultCommit is a helper to initialize the ResultCommit with
// the embedded struct.
func NewResultCommit(header *types.Header, commit *types.Commit,
	canonical bool,
) *ResultCommit {
	return &ResultCommit{
		SignedHeader: types.SignedHeader{
			Header: header,
			Commit: commit,
		},
		CanonicalCommit: canonical,
	}
}

// Info about the node's syncing state.
type SyncInfo struct {
	LatestBlockTime     time.Time      `json:"latest_block_time"`
	EarliestBlockTime   time.Time      `json:"earliest_block_time"`
	LatestBlockHash     bytes.HexBytes `json:"latest_block_hash"`
	LatestAppHash       bytes.HexBytes `json:"latest_app_hash"`
	EarliestBlockHash   bytes.HexBytes `json:"earliest_block_hash"`
	EarliestAppHash     bytes.HexBytes `json:"earliest_app_hash"`
	LatestBlockHeight   int64          `json:"latest_block_height"`
	EarliestBlockHeight int64          `json:"earliest_block_height"`
	CatchingUp          bool           `json:"catching_up"`
}

// Info about the node's validator.
type ValidatorInfo struct {
	PubKey      crypto.PubKey  `json:"pub_key"`
	Address     bytes.HexBytes `json:"address"`
	VotingPower int64          `json:"voting_power"`
}

// Node Status.
type ResultStatus struct {
	NodeInfo      p2p.DefaultNodeInfo `json:"node_info"`
	SyncInfo      SyncInfo            `json:"sync_info"`
	ValidatorInfo ValidatorInfo       `json:"validator_info"`
}

// Is TxIndexing enabled.
func (s *ResultStatus) TxIndexEnabled() bool {
	if s == nil {
		return false
	}
	return s.NodeInfo.Other.TxIndex == "on"
}

// Info about peer connections.
type ResultNetInfo struct {
	Listeners []string `json:"listeners"`
	Peers     []Peer   `json:"peers"`
	NPeers    int      `json:"n_peers"`
	Listening bool     `json:"listening"`
}

// Log from dialing seeds.
type ResultDialSeeds struct {
	Log string `json:"log"`
}

// Log from dialing peers.
type ResultDialPeers struct {
	Log string `json:"log"`
}

// A peer.
type Peer struct {
	NodeInfo         p2p.DefaultNodeInfo  `json:"node_info"`
	RemoteIP         string               `json:"remote_ip"`
	ConnectionStatus p2p.ConnectionStatus `json:"connection_status"`
	IsOutbound       bool                 `json:"is_outbound"`
}

// Validators for a height.
type ResultValidators struct {
	Validators  []*types.Validator `json:"validators"`
	BlockHeight int64              `json:"block_height"`
	Count       int                `json:"count"`
	Total       int                `json:"total"`
}

// ConsensusParams for given height.
type ResultConsensusParams struct {
	ConsensusParams types.ConsensusParams `json:"consensus_params"`
	BlockHeight     int64                 `json:"block_height"`
}

// Info about the consensus state.
// UNSTABLE.
type ResultDumpConsensusState struct {
	RoundState json.RawMessage `json:"round_state"`
	Peers      []PeerStateInfo `json:"peers"`
}

// UNSTABLE.
type PeerStateInfo struct {
	NodeAddress string          `json:"node_address"`
	PeerState   json.RawMessage `json:"peer_state"`
}

// UNSTABLE.
type ResultConsensusState struct {
	RoundState json.RawMessage `json:"round_state"`
}

// CheckTx result.
type ResultBroadcastTx struct {
	Log       string         `json:"log"`
	Codespace string         `json:"codespace"`
	Data      bytes.HexBytes `json:"data"`
	Hash      bytes.HexBytes `json:"hash"`
	Code      uint32         `json:"code"`
}

// CheckTx and ExecTx results.
type ResultBroadcastTxCommit struct {
	CheckTx  abci.CheckTxResponse `json:"check_tx"`
	TxResult abci.ExecTxResult    `json:"tx_result"`
	Hash     bytes.HexBytes       `json:"hash"`
	Height   int64                `json:"height"`
}

// ResultCheckTx wraps abci.CheckTxResponse.
type ResultCheckTx struct {
	abci.CheckTxResponse
}

// Result of querying for a tx.
type ResultTx struct {
	TxResult abci.ExecTxResult `json:"tx_result"`
	Proof    types.TxProof     `json:"proof,omitempty"`
	Hash     bytes.HexBytes    `json:"hash"`
	Tx       types.Tx          `json:"tx"`
	Height   int64             `json:"height"`
	Index    uint32            `json:"index"`
}

// Result of searching for txs.
type ResultTxSearch struct {
	Txs        []*ResultTx `json:"txs"`
	TotalCount int         `json:"total_count"`
}

// ResultBlockSearch defines the RPC response type for a block search by events.
type ResultBlockSearch struct {
	Blocks     []*ResultBlock `json:"blocks"`
	TotalCount int            `json:"total_count"`
}

// List of mempool txs.
type ResultUnconfirmedTxs struct {
	Txs        []types.Tx `json:"txs"`
	Count      int        `json:"n_txs"`
	Total      int        `json:"total"`
	TotalBytes int64      `json:"total_bytes"`
}

// Info abci msg.
type ResultABCIInfo struct {
	Response abci.InfoResponse `json:"response"`
}

// Query abci msg.
type ResultABCIQuery struct {
	Response abci.QueryResponse `json:"response"`
}

// Result of broadcasting evidence.
type ResultBroadcastEvidence struct {
	Hash []byte `json:"hash"`
}

// empty results.
type (
	ResultUnsafeFlushMempool struct{}
	ResultUnsafeProfile      struct{}
	ResultSubscribe          struct{}
	ResultUnsubscribe        struct{}
	ResultHealth             struct{}
)

// Event data from a subscription.
type ResultEvent struct {
	Data   types.TMEventData   `json:"data"`
	Events map[string][]string `json:"events"`
	Query  string              `json:"query"`
}
