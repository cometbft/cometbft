package node

import (
	"errors"
	"fmt"
)

var (
	// ErrNonEmptyBlockStore is returned when the blockstore is not empty and the node is trying to initialize non empty state
	ErrNonEmptyBlockStore = errors.New("blockstore not empty, trying to initialize non empty state")
	// ErrNonEmptyState is returned when the state is not empty and the node is trying to initialize non empty state
	ErrNonEmptyState = errors.New("state not empty, trying to initialize non empty state")
	// ErrSwitchStateSync is returned when the blocksync reactor does not support switching from state sync
	ErrSwitchStateSync = errors.New("this blocksync reactor does not support switching from state sync")
	// ErrGenesisHashDecode is returned when the genesis hash provided by the operator cannot be decoded
	ErrGenesisHashDecode = errors.New("genesis hash provided by operator cannot be decoded")
	// ErrPassedGenesisHashMismatch is returned when the genesis doc hash in the database does not match the passed --genesis_hash value
	ErrPassedGenesisHashMismatch = errors.New("genesis doc hash in db does not match passed --genesis_hash value")
	// ErrLoadedGenesisDocHashMismatch is returned when the genesis doc hash in the database does not match the loaded genesis doc
	ErrLoadedGenesisDocHashMismatch = errors.New("genesis doc hash in db does not match loaded genesis doc")
)

type ErrLightClientStateProvider struct {
	Err error
}

func (e ErrLightClientStateProvider) Error() string {
	return fmt.Sprintf("failed to set up light client state provider: %v", e.Err)
}

func (e ErrLightClientStateProvider) Unwrap() error {
	return e.Err
}

type ErrMismatchAppHash struct {
	Expected, Actual []byte
}

func (e ErrMismatchAppHash) Error() string {
	return fmt.Sprintf("the app hash returned by the light client does not match the provided appHash, expected %X, got %X", e.Expected, e.Actual)
}

type ErrSetSyncHeight struct {
	Err error
}

func (e ErrSetSyncHeight) Error() string {
	return fmt.Sprintf("failed to set synced height: %v", e.Err)
}

type ErrPrivValidatorSocketClient struct {
	Err error
}

func (e ErrPrivValidatorSocketClient) Error() string {
	return fmt.Sprintf("error with private validator socket client: %v", e.Err)
}

func (e ErrPrivValidatorSocketClient) Unwrap() error {
	return e.Err
}

type ErrGetPubKey struct {
	Err error
}

func (e ErrGetPubKey) Error() string {
	return fmt.Sprintf("can't get pubkey: %v", e.Err)
}

func (e ErrGetPubKey) Unwrap() error {
	return e.Err
}

type ErrCreatePruner struct {
	Err error
}

func (e ErrCreatePruner) Error() string {
	return fmt.Sprintf("failed to create pruner: %v", e.Err)
}

func (e ErrCreatePruner) Unwrap() error {
	return e.Err
}

// fmt.Errorf("could not create blocksync reactor: %w", err)

type ErrCreateBlockSyncReactor struct {
	Err error
}

func (e ErrCreateBlockSyncReactor) Error() string {
	return fmt.Sprintf("could not create blocksync reactor: %v", e.Err)
}

func (e ErrCreateBlockSyncReactor) Unwrap() error {
	return e.Err
}

type ErrAddPersistentPeers struct {
	Err error
}

func (e ErrAddPersistentPeers) Error() string {
	return fmt.Sprintf("could not add peers from persistent_peers field: %v", e.Err)
}

func (e ErrAddPersistentPeers) Unwrap() error {
	return e.Err
}

// fmt.Errorf("could not add peer ids from unconditional_peer_ids field: %w", err)

type ErrAddUnconditionalPeerIDs struct {
	Err error
}

func (e ErrAddUnconditionalPeerIDs) Error() string {
	return fmt.Sprintf("could not add peer ids from unconditional_peer_ids field: %v", e.Err)
}

func (e ErrAddUnconditionalPeerIDs) Unwrap() error {
	return e.Err
}

// fmt.Errorf("could not create addrbook: %w", err)

type ErrCreateAddrBook struct {
	Err error
}

func (e ErrCreateAddrBook) Error() string {
	return fmt.Sprintf("could not create addrbook: %v", e.Err)
}

func (e ErrCreateAddrBook) Unwrap() error {
	return e.Err
}

// fmt.Errorf("could not dial peers from persistent_peers field: %w", err)

type ErrDialPeers struct {
	Err error
}

func (e ErrDialPeers) Error() string {
	return fmt.Sprintf("could not dial peers from persistent_peers field: %v", e.Err)
}

func (e ErrDialPeers) Unwrap() error {
	return e.Err
}

// ErrStartStateSync is returned when the node fails to start state sync
type ErrStartStateSync struct {
	Err error
}

func (e ErrStartStateSync) Error() string {
	return fmt.Sprintf("failed to start state sync: %v", e.Err)
}

func (e ErrStartStateSync) Unwrap() error {
	return e.Err
}

// ErrStartPruning is returned when the node fails to start background pruning routine
type ErrStartPruning struct {
	Err error
}

func (e ErrStartPruning) Error() string {
	return fmt.Sprintf("failed to start background pruning routine: %v", e.Err)
}

func (e ErrStartPruning) Unwrap() error {
	return e.Err
}

// fmt.Errorf("failed to load or gen node key %s: %w", config.NodeKeyFile(), err)

type ErrLoadOrGenNodeKey struct {
	Err         error
	NodeKeyFile string
}

func (e ErrLoadOrGenNodeKey) Error() string {
	return fmt.Sprintf("failed to load or gen node key %s: %v", e.NodeKeyFile, e.Err)
}

func (e ErrLoadOrGenNodeKey) Unwrap() error {
	return e.Err
}

// fmt.Errorf("error retrieving genesis doc hash: %w", err)

type ErrRetrieveGenesisDocHash struct {
	Err error
}

func (e ErrRetrieveGenesisDocHash) Error() string {
	return fmt.Sprintf("error retrieving genesis doc hash: %v", e.Err)
}

func (e ErrRetrieveGenesisDocHash) Unwrap() error {
	return e.Err
}

// fmt.Errorf("error in genesis doc: %w", err)

type ErrGenesisDoc struct {
	Err error
}

func (e ErrGenesisDoc) Error() string {
	return fmt.Sprintf("error in genesis doc: %v", e.Err)
}

func (e ErrGenesisDoc) Unwrap() error {
	return e.Err
}

type ErrSaveGenesisDocHash struct {
	Err error
}

func (e ErrSaveGenesisDocHash) Error() string {
	return fmt.Sprintf("failed to save genesis doc hash to db: %v", e.Err)
}

func (e ErrSaveGenesisDocHash) Unwrap() error {
	return e.Err
}
