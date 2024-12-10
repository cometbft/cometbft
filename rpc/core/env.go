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
	TxIndexer    txindex.TxIndexer
	BlockIndexer indexer.BlockIndexer
	EventBus     *types.EventBus // thread safe
	Mempool      mempl.Mempool

	Logger log.Logger

	Config cfg.RPCConfig

	GenesisFilePath string // the genesis file's full path on disk

	// genesisChunk is a map of chunk ID to its full path on disk.
	// If the genesis file is smaller than genesisChunkSize, then this map will be
	// nil, because there will be no chunks on disk.
	// This map is convenient for the `/genesis_chunked` API to quickly find a chunk
	// by its ID, instead of having to reconstruct its path each time, which would
	// involve multiple string operations.
	genesisChunksFiles map[int]string
}

// InitGenesisChunks checks whether it makes sense to split the genesis file into
// small chunks to be stored on disk.
// It is called on Node startup and should be called only once.
// Rules of chunking:
//   - if the genesis file's size is <= genesisChunkSize, this function returns
//     without doing anything. The `/genesis` RPC API endpoint will fetch the genesis
//     file from disk to serve requests.
//   - if the genesis file's size is > genesisChunkSize, then use chunking. The
//     function splits the genesis file into chunks of genesisChunkSize and stores
//     each chunk on disk.  The `/genesis_chunked` RPC API endpoint will fetch the
//     genesis file chunks from disk to serve requests.
func (env *Environment) InitGenesisChunks() error {
	if len(env.genesisChunksFiles) > 0 {
		// we already computed the chunks, return.
		return nil
	}

	gFilePath := env.GenesisFilePath
	if len(gFilePath) == 0 {
		// chunks not computed yet, but no genesis available.
		// This should not happen.
		return errors.New("missing genesis file path on disk")
	}

	gFileSize, err := fileSize(gFilePath)
	if err != nil {
		return fmt.Errorf("gauging genesis file size: %w", err)
	}

	if gFileSize <= genesisChunkSize {
		// no chunking required
		return nil
	}

	gChunksDir, err := mkChunksDir(gFilePath, _chunksDir)
	if err != nil {
		return fmt.Errorf("preparing chunks directory: %w", err)
	}

	// chunking required
	chunkIDToPath, err := writeChunks(gFilePath, gChunksDir, genesisChunkSize)
	if err != nil {
		return fmt.Errorf("chunking large genesis file: %w", err)
	}

	env.genesisChunksFiles = chunkIDToPath

	return nil
}

// Cleanup deletes the directory storing the genesis file chunks on disk
// if it exists. If the directory does not exist, the function is a no-op.
// The chunks' directory is a sub-directory of the `config/` directory of the
// running node (i.e., where the genesis.json file is stored).
// We call the function:
//   - before creating new genesis file chunks, to make sure we start with a clean
//     directory.
//   - when a Node shuts down, to clean up the file system.
func (env *Environment) Cleanup() error {
	gFileDir := filepath.Dir(env.GenesisFilePath)
	chunksDir := filepath.Join(gFileDir, _chunksDir)

	if err := os.RemoveAll(chunksDir); err != nil {
		return fmt.Errorf("deleting genesis file chunks' folder: %w", err)
	}

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

// fileSize returns the size of the file at the given path.
func fileSize(fPath string) (int, error) {
	// we use os.Stat here instead of os.ReadFile, because we don't want to load
	// the entire file into memory just to compute its size from the resulting
	// []byte slice.
	fInfo, err := os.Stat(fPath)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, fmt.Errorf("the file is unavailable at %s", fPath)
	} else if err != nil {
		return 0, fmt.Errorf("accessing file: %w", err)
	}
	return int(fInfo.Size()), nil
}

// mkChunksDir creates a new directory to store the genesis file's chunks.
// gFilePath is the genesis file's full path on disk.
// dirName is the name of the directory to be created, not it's path on disk.
// mkChunksDir will create a directory named 'dirName' as a sub-directory of the
// genesis file's directory (gFileDir).
// It returns the new directory's full path or an empty string if there is an
// error.
func mkChunksDir(gFilePath string, dirName string) (string, error) {
	var (
		gFileDir = filepath.Dir(gFilePath)
		dirPath  = filepath.Join(gFileDir, dirName)
	)
	if _, err := os.Stat(dirPath); err == nil {
		// directory already exists; this might happen it the node crashed and
		// could not do cleanup. Delete it to start from scratch.
		if err := os.RemoveAll(dirPath); err != nil {
			return "", fmt.Errorf("deleting existing chunks directory: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("accessing directory: %w", err)
	}

	if err := os.Mkdir(dirPath, 0o700); err != nil {
		return "", fmt.Errorf("creating chunks directory: %s", err)
	}

	return dirPath, nil
}

// writeChunk writes a chunk of the genesis file to disk, saving it to dir.
// Each chunk file name's format will be: chunk_[chunkID].part, e.g., chunk_42.part.
func writeChunk(chunk []byte, dir string, chunkID int) (string, error) {
	var (
		chunkName = "chunk_" + strconv.Itoa(chunkID) + ".part"
		chunkPath = filepath.Join(dir, chunkName)
	)
	if err := os.WriteFile(chunkPath, chunk, 0o600); err != nil {
		return "", fmt.Errorf("writing chunk to disk: %w", err)
	}

	return chunkPath, nil
}

// writeChunks reads the genesis file in chunks of size chunkSize, and writes them
// to disk.
// gFilePath is the genesis file's full path on disk.
// gChunksDir is the directory where the chunks will be stored on disk.
// chunkSize is the size of a chunk, that is, writeChunks will read the genesis file
// in chunks of size chunkSize.
// It returns a map where the keys are the chunk IDs, and the values are the chunks'
// path on disk. E.g.,:
// map[0] = $HOME/.cometbft/config/genesis-chunks/chunk_0.part
// map[1] = $HOME/.cometbft/config/genesis-chunks/chunk_1.part
// and so on for all chunks.
// The map will be useful for the `/genesis_chunked` RPC endpoint to quickly find
// a chunk on disk given its ID.
func writeChunks(
	gFilePath, gChunksDir string,
	chunkSize int,
) (map[int]string, error) {
	gFile, err := os.Open(gFilePath)
	if err != nil {
		return nil, fmt.Errorf("opening genesis file: %s", err)
	}
	defer gFile.Close()

	var (
		buf           = make([]byte, chunkSize)
		chunkIDToPath = make(map[int]string)
	)
	for chunkID := 0; ; chunkID++ {
		n, err := gFile.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			formatStr := "chunk %d: reading genesis file: %w"
			return nil, fmt.Errorf(formatStr, chunkID, err)
		}

		chunkPath, err := writeChunk(buf[:n], gChunksDir, chunkID)
		if err != nil {
			return nil, fmt.Errorf("chunk %d: %w", chunkID, err)
		}

		chunkIDToPath[chunkID] = chunkPath
	}

	return chunkIDToPath, nil
}
