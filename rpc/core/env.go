package core

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	abcicli "github.com/cometbft/cometbft/abci/client"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	mempl "github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/state/indexer"
	"github.com/cometbft/cometbft/state/txindex"
	"github.com/cometbft/cometbft/types"
)

const (
	// see README.
	defaultPerPage = 30
	maxPerPage     = 100

	// SubscribeTimeout is the maximum time we wait to subscribe for an event.
	// must be less than the server's write timeout (see rpcserver.DefaultConfig).
	SubscribeTimeout = 5 * time.Second

	// genesisChunkSize is the maximum size, in bytes, of each
	// chunk in the genesis structure for the chunked API.
	genesisChunkSize = 16 * 1024 * 1024 // 16
)

// These interfaces are used by RPC and must be thread safe

type Consensus interface {
	GetState() sm.State
	GetValidators() (int64, []*types.Validator)
	GetLastHeight() int64
	GetRoundStateJSON() ([]byte, error)
	GetRoundStateSimpleJSON() ([]byte, error)
}

type transport interface {
	Listeners() []string
	IsListening() bool
	NodeInfo() p2p.NodeInfo
}

type peers interface {
	AddPersistentPeers(peers []string) error
	AddUnconditionalPeerIDs(peerIDs []string) error
	AddPrivatePeerIDs(peerIDs []string) error
	DialPeersAsync(peers []string) error
	Peers() p2p.IPeerSet
}

// A reactor that transitions from block sync or state sync to consensus mode.
type syncReactor interface {
	WaitSync() bool
}

type mempoolReactor interface {
	syncReactor
	TryAddTx(tx types.Tx, sender p2p.Peer) (*abcicli.ReqRes, error)
}

// Environment contains objects and interfaces used by the RPC. It is expected
// to be setup once during startup.
type Environment struct {
	// external, thread safe interfaces
	ProxyAppQuery   proxy.AppConnQuery
	ProxyAppMempool proxy.AppConnMempool

	// interfaces defined in types and above
	StateStore       sm.Store
	BlockStore       sm.BlockStore
	EvidencePool     sm.EvidencePool
	ConsensusState   Consensus
	ConsensusReactor syncReactor
	MempoolReactor   mempoolReactor
	P2PPeers         peers
	P2PTransport     transport

	// objects
	PubKey       crypto.PubKey
	GenDoc       *types.GenesisDoc // cache the genesis structure
	TxIndexer    txindex.TxIndexer
	BlockIndexer indexer.BlockIndexer
	EventBus     *types.EventBus // thread safe
	Mempool      mempl.Mempool

	Logger log.Logger

	Config cfg.RPCConfig

	// cache of chunked genesis data.
	genChunks []string
}

// InitGenesisChunks checks whether it makes sense to create a cache of chunked
// genesis data. It is called once on Node startup.
// Rules of chunking:
// - if the genesis file's size is <= 16MB, then no chunking. An `Environment`
// object will store a pointer to the genesis in its GenDoc field. Its genChunks
// field will be set to nil. `/genesis` RPC API will return the GenesisDoc itself.
// - if the genesis file's size is > 16MB, then use chunking. An `Environment`
// object will store a slice of base64-encoded chunks in its genChunks field. Its
// GenDoc field will be set to nil. `/genesis` RPC API will redirect users to use
// the `/genesis_chunked` API.
func (env *Environment) InitGenesisChunks() error {
	if env.genChunks != nil || len(env.genChunks) > 0 {
		// we already computed the chunks, return.
		return nil
	}

	if env.GenDoc == nil {
		// chunks not computed yet, but no genesis available.
		// This should not happen.
		return errors.New("pointer to genesis is nil")
	}

	data, err := cmtjson.Marshal(env.GenDoc)
	if err != nil {
		return err
	}

	// If genesis is less than 16MB, then no chunking.
	// Keep a pointer to a GenesisDoc in env.GenDoc.
	if len(data) <= genesisChunkSize {
		env.genChunks = nil
		return nil
	}

	var (
		nChunks = (len(data) + genesisChunkSize - 1) / genesisChunkSize
		chunks  = make([]string, nChunks)
	)
	for i := range nChunks {
		var (
			start = i * genesisChunkSize
			end   = start + genesisChunkSize
		)
		if end > len(data) {
			end = len(data)
		}

		chunk := data[start:end]
		chunks[i] = base64.StdEncoding.EncodeToString(chunk)
	}

	env.genChunks = chunks

	// we store the chunks; don't store a ptr to the genesis anymore.
	env.GenDoc = nil

	return nil
}

func validatePage(pagePtr *int, perPage, totalCount int) (int, error) {
	if perPage < 1 {
		panic(fmt.Sprintf("zero or negative perPage: %d", perPage))
	}

	if pagePtr == nil { // no page parameter
		return 1, nil
	}

	pages := ((totalCount - 1) / perPage) + 1
	if pages == 0 {
		pages = 1 // one page (even if it's empty)
	}
	page := *pagePtr
	if page <= 0 || page > pages {
		return 1, fmt.Errorf("page should be within [1, %d] range, given %d", pages, page)
	}

	return page, nil
}

func (*Environment) validatePerPage(perPagePtr *int) int {
	if perPagePtr == nil { // no per_page parameter
		return defaultPerPage
	}

	perPage := *perPagePtr
	if perPage < 1 {
		return defaultPerPage
	} else if perPage > maxPerPage {
		return maxPerPage
	}
	return perPage
}

func validateSkipCount(page, perPage int) int {
	skipCount := (page - 1) * perPage
	if skipCount < 0 {
		return 0
	}

	return skipCount
}

// latestHeight can be either latest committed or uncommitted (+1) height.
func (env *Environment) getHeight(latestHeight int64, heightPtr *int64) (int64, error) {
	if heightPtr != nil {
		height := *heightPtr
		if height <= 0 {
			return 0, fmt.Errorf("height must be greater than 0, but got %d", height)
		}
		if height > latestHeight {
			return 0, fmt.Errorf("height %d must be less than or equal to the current blockchain height %d",
				height, latestHeight)
		}
		base := env.BlockStore.Base()
		if height < base {
			return 0, fmt.Errorf("height %d is not available, lowest height is %d",
				height, base)
		}
		return height, nil
	}
	return latestHeight, nil
}

func (env *Environment) latestUncommittedHeight() int64 {
	nodeIsSyncing := env.ConsensusReactor.WaitSync()
	if nodeIsSyncing {
		return env.BlockStore.Height()
	}
	return env.BlockStore.Height() + 1
}
