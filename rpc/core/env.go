package core

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	abcicli "github.com/cometbft/cometbft/abci/client"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	mempl "github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/p2p"
	ni "github.com/cometbft/cometbft/p2p/nodeinfo"
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
	genesisChunkSize = 2 * 1024 * 1024 // 2 MB

	_chunksDir = "genesis-chunks"
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
	NodeInfo() ni.NodeInfo
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

// Environment contains the objects and interfaces used to serve the RPC APIs.
// A Node creates an object of this type at startup.
// An Environment should not be created directly, and it is recommended that
// only one instance of Environment be created at runtime.
// For this reason, callers should create an Environment object using
// the ConfigureRPC() method of the Node type, because the Environment object it
// returns is a singleton.
// Note: The Environment type was exported in the initial RPC API design; therefore,
// unexporting it now could potentially break users.
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

	GenesisFilePath string // the genesis file's full path on disk

	// genesisChunk is a map of chunk ID to its full path on disk.
	// If the genesis file is smaller than genesisChunkSize, then this map will be
	// nil, because there will be no chunks on disk.
	// This map is convenient for the `/genesis_chunked` API to quickly find a chunk
	// by its ID, instead of having to reconstruct its path each time, which would
	// involve multiple string operations.
	genesisChunks map[int]string
}

// InitGenesisChunks checks whether it makes sense to create a cache of chunked
// genesis data. It is called on Node startup and should be called only once.
// Rules of chunking:
//   - if the genesis file's size is <= genesisChunkSize, then no chunking.
//     An `Environment` object will store a pointer to the genesis in its GenDoc
//     field. Its genChunks field will be set to nil. `/genesis` RPC API will return
//     the GenesisDoc itself.
//   - if the genesis file's size is > genesisChunkSize, then use chunking. An
//     `Environment` object will store a slice of base64-encoded chunks in its
//     genChunks field. Its GenDoc field will be set to nil. `/genesis` RPC API will
//     redirect users to use the `/genesis_chunked` API.
func (env *Environment) InitGenesisChunks() error {
	if len(env.genesisChunks) > 0 {
		// we already computed the chunks, return.
		return nil
	}

	gFilePath := env.GenesisFilePath
	if len(gFilePath) == 0 {
		// chunks not computed yet, but no genesis available.
		// This should not happen.
		return errors.New("the genesis file path on disk is missing")
	}

	gFileSize, err := fileSize(gFilePath)
	if err != nil {
		return fmt.Errorf("estimating genesis file size: %s", err)
	}

	if gFileSize <= genesisChunkSize {
		// no chunking required
		return nil
	}

	// chunking required
	var (
		nChunks       = (gFileSize + genesisChunkSize - 1) / genesisChunkSize
		chunkIDToPath = make(map[int]string, nChunks)
	)
	// we'll create the chunks while reading the file as a stream rather than loading
	// it into memory.
	gFile, err := os.Open(gFilePath)
	if err != nil {
		return fmt.Errorf("opening the genesis file at %s: %s", gFilePath, err)
	}
	defer gFile.Close()

	var (
		buf     = make([]byte, genesisChunkSize)
		chunkID = 0

		gFileDir   = filepath.Dir(gFilePath)
		gChunksDir = filepath.Join(gFileDir, _chunksDir)
	)

	if err := os.Mkdir(gChunksDir, 0o755); err != nil {
		formatStr := "creating genesis chunks directory at %s: %s"
		return fmt.Errorf(formatStr, gChunksDir, err)
	}

	for {
		// Read the file in chunks of size genesisChunkSize
		n, err := gFile.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("reading genesis file at %s: %s", gFilePath, err)
		}

		var (
			chunk     = buf[:n]
			chunkName = "chunk_" + strconv.Itoa(chunkID) + ".part"
			chunkPath = filepath.Join(gChunksDir, chunkName)
		)
		if err := os.WriteFile(chunkPath, chunk, 0o600); err != nil {
			return fmt.Errorf("creating genesis chunk at %s: %s", chunkPath, err)
		}

		chunkIDToPath[chunkID] = chunkPath
		chunkID++
	}

	env.genesisChunks = chunkIDToPath

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

// deleteGenesisChunks deletes the directory storing the genesis file chunks on disk
// if it exists. If the directory does not exist, the function is a no-op.
// The chunks' directory is a sub-directory of the `config/` directory of the
// running node.
// We call the function:
// - when creating the genesis file chunks, to make sure
// - when a Node shuts down, to clean up the file system.
func (env *Environment) deleteGenesisChunks() error {
	gFileDir := filepath.Dir(env.GenesisFilePath)
	chunksDir := filepath.Join(gFileDir, _chunksDir)

	if _, err := os.Stat(chunksDir); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		formatStr := "accessing genesis file chunks directory at %q: %s"
		return fmt.Errorf(formatStr, chunksDir, err)
	}

	// Directory exists, delete it
	if err := os.RemoveAll(chunksDir); err != nil {
		formatStr := "deleting pre-existing genesis chunks at %s: %s"
		return fmt.Errorf(formatStr, chunksDir, err)
	}

	return nil
}

// fileSize returns the size of the file at the given path.
func fileSize(fPath string) (int64, error) {
	// we use os.Stat here instead of os.ReadFile, because we don't want to load
	// the entire file into memory just to compute its size from the resulting
	// []byte slice.
	fInfo, err := os.Stat(fPath)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, fmt.Errorf("the file is unavailable at %s", fPath)
	} else if err != nil {
		return 0, fmt.Errorf("accessing file at %s: %s", fPath, err)
	}
	return fInfo.Size(), nil
}

// mkChunksDir creates a new directory to store the genesis file's chunks.
// gFilePath is the genesis file's full path on disk, and mkChunksDir creates a new
// directory as a sub-directory of the genesis file's directory.
// It returns the new directory's full path or an empty string if there is an
// error.
func mkChunksDir(gFilePath string, dirName string) (string, error) {
	gFileDir := filepath.Dir(gFilePath)
	dirPath := filepath.Join(gFileDir, dirName)

	if err := os.Mkdir(dirPath, 0o755); err != nil {
		return "", fmt.Errorf("creating directory at %s: %s", dirPath, err)
	}
	return dirPath, nil
}
