package node

import (
	"errors"
	"fmt"
)

var (
	// ErrNonEmptyBlockStore is returned when the blockstore is not empty and the node is trying to initialize non empty state.
	ErrNonEmptyBlockStore = errors.New("blockstore not empty, trying to initialize non empty state")
	// ErrNonEmptyState is returned when the state is not empty and the node is trying to initialize non empty state.
	ErrNonEmptyState = errors.New("state not empty, trying to initialize non empty state")
	// ErrSwitchStateSync is returned when the blocksync reactor does not support switching from state sync.
	ErrSwitchStateSync = errors.New("this blocksync reactor does not support switching from state sync")
	// ErrGenesisHashDecode is returned when the genesis hash provided by the operator cannot be decoded.
	ErrGenesisHashDecode = errors.New("genesis hash provided by operator cannot be decoded")
	// ErrPassedGenesisHashMismatch is returned when the genesis doc hash in the database does not match the passed --genesis_hash value.
	ErrPassedGenesisHashMismatch = errors.New("genesis doc hash in db does not match passed --genesis_hash value")
	// ErrLoadedGenesisDocHashMismatch is returned when the genesis doc hash in the database does not match the loaded genesis doc.
	ErrLoadedGenesisDocHashMismatch = errors.New("genesis doc hash in db does not match loaded genesis doc")
)

// ErrLightClientStateProvider is returned when the node fails to create the blockstore.
type ErrLightClientStateProvider struct {
	Err error
}

func (e ErrLightClientStateProvider) Error() string {
	return fmt.Sprintf("failed to set up light client state provider: %v", e.Err)
}

func (e ErrLightClientStateProvider) Unwrap() error {
	return e.Err
}

// ErrMismatchAppHash is returned when the app hash returned by the light client does not match the provided appHash.
type ErrMismatchAppHash struct {
	Expected, Actual []byte
}

func (e ErrMismatchAppHash) Error() string {
	return fmt.Sprintf("the app hash returned by the light client does not match the provided appHash, expected %X, got %X", e.Expected, e.Actual)
}

// ErrSetSyncHeight is returned when the node fails to set the synced height.
type ErrSetSyncHeight struct {
	Err error
}

func (e ErrSetSyncHeight) Error() string {
	return fmt.Sprintf("failed to set synced height: %v", e.Err)
}

// ErrPrivValidatorSocketClient is returned when the node fails to create private validator socket client.
type ErrPrivValidatorSocketClient struct {
	Err error
}

func (e ErrPrivValidatorSocketClient) Error() string {
	return fmt.Sprintf("error with private validator socket client: %v", e.Err)
}

func (e ErrPrivValidatorSocketClient) Unwrap() error {
	return e.Err
}

// ErrGetPubKey is returned when the node fails to get the public key.
type ErrGetPubKey struct {
	Err error
}

func (e ErrGetPubKey) Error() string {
	return fmt.Sprintf("can't get pubkey: %v", e.Err)
}

func (e ErrGetPubKey) Unwrap() error {
	return e.Err
}

// ErrCreatePruner is returned when the node fails to create the pruner.
type ErrCreatePruner struct {
	Err error
}

func (e ErrCreatePruner) Error() string {
	return fmt.Sprintf("failed to create pruner: %v", e.Err)
}

func (e ErrCreatePruner) Unwrap() error {
	return e.Err
}

// ErrCreateBlockSyncReactor is returned when the node fails to create the blocksync reactor.
type ErrCreateBlockSyncReactor struct {
	Err error
}

func (e ErrCreateBlockSyncReactor) Error() string {
	return fmt.Sprintf("could not create blocksync reactor: %v", e.Err)
}

func (e ErrCreateBlockSyncReactor) Unwrap() error {
	return e.Err
}

// ErrAddPersistentPeers is returned when the node fails to add peers from the persistent_peers field.
type ErrAddPersistentPeers struct {
	Err error
}

func (e ErrAddPersistentPeers) Error() string {
	return fmt.Sprintf("could not add peers from persistent_peers field: %v", e.Err)
}

func (e ErrAddPersistentPeers) Unwrap() error {
	return e.Err
}

// ErrAddUnconditionalPeerIDs is returned when the node fails to add peer ids from the unconditional_peer_ids field.
type ErrAddUnconditionalPeerIDs struct {
	Err error
}

func (e ErrAddUnconditionalPeerIDs) Error() string {
	return fmt.Sprintf("could not add peer ids from unconditional_peer_ids field: %v", e.Err)
}

func (e ErrAddUnconditionalPeerIDs) Unwrap() error {
	return e.Err
}

// ErrCreateAddrBook is returned when the node fails to create the address book.
type ErrCreateAddrBook struct {
	Err error
}

func (e ErrCreateAddrBook) Error() string {
	return fmt.Sprintf("could not create addrbook: %v", e.Err)
}

func (e ErrCreateAddrBook) Unwrap() error {
	return e.Err
}

// ErrDialPeers is returned when the node fails to dial peers from the persistent_peers field.
type ErrDialPeers struct {
	Err error
}

func (e ErrDialPeers) Error() string {
	return fmt.Sprintf("could not dial peers from persistent_peers field: %v", e.Err)
}

func (e ErrDialPeers) Unwrap() error {
	return e.Err
}

// ErrHandshake is returned when CometBFT fails to complete the handshake with the ABCI app.
type ErrHandshake struct {
	Err error
}

func (e ErrHandshake) Error() string {
	return fmt.Sprintf("could not complete handshake with the app: %v", e.Err)
}

func (e ErrHandshake) Unwrap() error {
	return e.Err
}

// ErrStartStateSync is returned when the node fails to start the statesync.
type ErrStartStateSync struct {
	Err error
}

func (e ErrStartStateSync) Error() string {
	return fmt.Sprintf("failed to start statesync: %v", e.Err)
}

func (e ErrStartStateSync) Unwrap() error {
	return e.Err
}

// ErrStartBlockSync is returned when the node fails to start the blocksync.
type ErrStartBlockSync struct {
	Err error
}

func (e ErrStartBlockSync) Error() string {
	return fmt.Sprintf("failed to start blocksync: %v", e.Err)
}

func (e ErrStartBlockSync) Unwrap() error {
	return e.Err
}

// ErrStartPruning is returned when the node fails to start background pruning routine.
type ErrStartPruning struct {
	Err error
}

func (e ErrStartPruning) Error() string {
	return fmt.Sprintf("failed to start background pruning routine: %v", e.Err)
}

func (e ErrStartPruning) Unwrap() error {
	return e.Err
}

// ErrLoadOrGenNodeKey is returned when the node fails to load or generate the node key.
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

// ErrRetrieveGenesisDocHash is returned when the node fails to retrieve the genesis doc hash from the database.
type ErrRetrieveGenesisDocHash struct {
	Err error
}

func (e ErrRetrieveGenesisDocHash) Error() string {
	return fmt.Sprintf("error retrieving genesis doc hash: %v", e.Err)
}

func (e ErrRetrieveGenesisDocHash) Unwrap() error {
	return e.Err
}

// ErrGenesisDoc is returned when the node fails to load the genesis doc.
type ErrGenesisDoc struct {
	Err error
}

func (e ErrGenesisDoc) Error() string {
	return fmt.Sprintf("error in genesis doc: %v", e.Err)
}

func (e ErrGenesisDoc) Unwrap() error {
	return e.Err
}

// ErrSaveGenesisDocHash is returned when the node fails to save the genesis doc hash to the database.
type ErrSaveGenesisDocHash struct {
	Err error
}

func (e ErrSaveGenesisDocHash) Error() string {
	return fmt.Sprintf("failed to save genesis doc hash to db: %v", e.Err)
}

func (e ErrSaveGenesisDocHash) Unwrap() error {
	return e.Err
}

// ErrorReadingGenesisDoc is returned when the node fails to read the genesis doc file.
type ErrorReadingGenesisDoc struct {
	Err error
}

func (e ErrorReadingGenesisDoc) Error() string {
	return fmt.Sprintf("could not read GenesisDoc file: %v", e.Err)
}

func (e ErrorReadingGenesisDoc) Unwrap() error {
	return e.Err
}

// ErrorLoadOrGenNodeKey is returned when the node fails to load or generate node key.
type ErrorLoadOrGenNodeKey struct {
	Err         error
	NodeKeyFile string
}

func (e ErrorLoadOrGenNodeKey) Error() string {
	return fmt.Sprintf("failed to load or generate node key %s: %v", e.NodeKeyFile, e.Err)
}

func (e ErrorLoadOrGenNodeKey) Unwrap() error {
	return e.Err
}

// ErrorLoadOrGenFilePV is returned when the node fails to load or generate priv validator file.
type ErrorLoadOrGenFilePV struct {
	Err       error
	KeyFile   string
	StateFile string
}

func (e ErrorLoadOrGenFilePV) Error() string {
	return fmt.Sprintf("failed to load or generate privval file; "+
		"key file %s, state file %s: %v", e.KeyFile, e.StateFile, e.Err)
}

func (e ErrorLoadOrGenFilePV) Unwrap() error {
	return e.Err
}
