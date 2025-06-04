package kvstore

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/v2/abci/types"
	"github.com/cometbft/cometbft/v2/crypto"
	cryptoenc "github.com/cometbft/cometbft/v2/crypto/encoding"
	"github.com/cometbft/cometbft/v2/libs/log"
	"github.com/cometbft/cometbft/v2/version"
)

var (
	stateKey        = []byte("stateKey")
	kvPairPrefixKey = []byte("kvPairKey:")
)

const (
	ValidatorPrefix        = "val="
	AppVersion      uint64 = 1
	defaultLane     string = "default"
)

var _ types.Application = (*Application)(nil)

// Application is the kvstore state machine. It complies with the abci.Application interface.
// It takes transactions in the form of key=value and saves them in a database. This is
// a somewhat trivial example as there is no real state execution.
type Application struct {
	types.BaseApplication

	state        State
	RetainBlocks int64 // blocks to retain after commit (via CommitResponse.RetainHeight)
	stagedTxs    [][]byte
	logger       log.Logger

	// validator set
	valUpdates         []types.ValidatorUpdate
	valAddrToPubKeyMap map[string]crypto.PubKey

	// If true, the app will generate block events in BeginBlock. Used to test the event indexer
	// Should be false by default to avoid generating too much data.
	genBlockEvents bool

	// Map from lane IDs to their priorities.
	lanePriorities map[string]uint32

	nextBlockDelay time.Duration
}

// NewApplication creates an instance of the kvstore from the provided database,
// with the given lanes and priorities.
func NewApplication(db dbm.DB, lanePriorities map[string]uint32) *Application {
	return &Application{
		logger:             log.NewNopLogger(),
		state:              loadState(db),
		valAddrToPubKeyMap: make(map[string]crypto.PubKey),
		lanePriorities:     lanePriorities,
		nextBlockDelay:     0, // zero by default because kvstore is mostly used for testing
	}
}

// newDB creates a DB engine for persisting the application state.
func newDB(dbDir string) *dbm.PebbleDB {
	name := "kvstore"
	db, err := dbm.NewPebbleDB(name, dbDir)
	if err != nil {
		panic(fmt.Errorf("failed to create persistent app at %s: %w", dbDir, err))
	}
	return db
}

// NewPersistentApplication creates a new application using the pebbledb
// database engine and default lanes.
func NewPersistentApplication(dbDir string) *Application {
	return NewApplication(newDB(dbDir), DefaultLanes())
}

// NewPersistentApplicationWithoutLanes creates a new application using the
// pebbledb database engine and without lanes.
func NewPersistentApplicationWithoutLanes(dbDir string) *Application {
	return NewApplication(newDB(dbDir), nil)
}

// NewInMemoryApplication creates a new application from an in memory database
// that uses default lanes. Nothing will be persisted.
func NewInMemoryApplication() *Application {
	return NewApplication(dbm.NewMemDB(), DefaultLanes())
}

// NewInMemoryApplication creates a new application from an in memory database
// and without lanes. Nothing will be persisted.
func NewInMemoryApplicationWithoutLanes() *Application {
	return NewApplication(dbm.NewMemDB(), nil)
}

// DefaultLanes returns a map from lane names to their priorities. Priority 0 is
// reserved. The higher the value, the higher the priority.
func DefaultLanes() map[string]uint32 {
	return map[string]uint32{
		"val":       9, // for validator updates
		"foo":       7,
		defaultLane: 3,
		"bar":       1,
	}
}

func (app *Application) SetGenBlockEvents() {
	app.genBlockEvents = true
}

// SetNextBlockDelay sets the delay for the next finalized block. Default is 0
// here because kvstore is mostly used for testing. In production, the default
// is 1s, mimicking the default for the deprecated `timeout_commit` parameter.
func (app *Application) SetNextBlockDelay(delay time.Duration) {
	app.nextBlockDelay = delay
}

// Info returns information about the state of the application. This is generally used every time a Tendermint instance
// begins and let's the application know what Tendermint versions it's interacting with. Based from this information,
// Tendermint will ensure it is in sync with the application by potentially replaying the blocks it has. If the
// Application returns a 0 appBlockHeight, Tendermint will call InitChain to initialize the application with consensus related data.
func (app *Application) Info(context.Context, *types.InfoRequest) (*types.InfoResponse, error) {
	// Tendermint expects the application to persist validators, on start-up we need to reload them to memory if they exist
	if len(app.valAddrToPubKeyMap) == 0 && app.state.Height > 0 {
		validators := app.getValidators()
		for _, v := range validators {
			pubkey, err := cryptoenc.PubKeyFromTypeAndBytes(v.PubKeyType, v.PubKeyBytes)
			if err != nil {
				panic(err)
			}
			app.valAddrToPubKeyMap[string(pubkey.Address())] = pubkey
		}
	}

	var defLane string
	if len(app.lanePriorities) != 0 {
		defLane = defaultLane
	}
	return &types.InfoResponse{
		Data:             fmt.Sprintf("{\"size\":%v}", app.state.Size),
		Version:          version.ABCIVersion,
		AppVersion:       AppVersion,
		LastBlockHeight:  app.state.Height,
		LastBlockAppHash: app.state.Hash(),
		LanePriorities:   app.lanePriorities,
		DefaultLane:      defLane,
	}, nil
}

// InitChain takes the genesis validators and stores them in the kvstore. It returns the application hash in the
// case that the application starts prepopulated with values. This method is called whenever a new instance of the application
// starts (i.e. app height = 0).
func (app *Application) InitChain(_ context.Context, req *types.InitChainRequest) (*types.InitChainResponse, error) {
	for _, v := range req.Validators {
		app.updateValidator(v)
	}
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	return &types.InitChainResponse{
		AppHash: appHash,
	}, nil
}

// CheckTx handles inbound transactions or in the case of recheckTx assesses old transaction validity after a state transition.
// As this is called frequently, it's preferably to keep the check as stateless and as quick as possible.
// Here we check that the transaction has the correctly key=value format.
// For the KVStore we check that each transaction has the valid tx format:
// - Contains one and only one `=`
// - `=` is not the first or last byte.
// - if key is `val` that the validator update transaction is also valid.
func (app *Application) CheckTx(_ context.Context, req *types.CheckTxRequest) (*types.CheckTxResponse, error) {
	// If it is a validator update transaction, check that it is correctly formatted
	if isValidatorTx(req.Tx) {
		if _, _, _, err := parseValidatorTx(req.Tx); err != nil {
			return &types.CheckTxResponse{Code: CodeTypeInvalidTxFormat}, nil //nolint:nilerr // error is not nil but it returns nil
		}
	} else if !isValidTx(req.Tx) {
		return &types.CheckTxResponse{Code: CodeTypeInvalidTxFormat}, nil
	}

	if len(app.lanePriorities) == 0 {
		return &types.CheckTxResponse{Code: CodeTypeOK, GasWanted: 1}, nil
	}
	laneID := assignLane(req.Tx)
	return &types.CheckTxResponse{Code: CodeTypeOK, GasWanted: 1, LaneId: laneID}, nil
}

// assignLane deterministically computes a lane for the given tx.
func assignLane(tx []byte) string {
	lane := defaultLane
	if isValidatorTx(tx) {
		return "val" // priority 9
	}
	key, _, err := parseTx(tx)
	if err != nil {
		return lane
	}

	// If the transaction key is an integer (for example, a transaction of the
	// form 2=2), we will assign a lane. Any other type of transaction will go
	// to the default lane.
	keyInt, err := strconv.Atoi(key)
	if err != nil {
		return lane
	}

	// Since a key is usually a numerical value, we assign lanes by computing
	// the key modulo some pre-selected divisors. As a result, some lanes will
	// be assigned less frequently than others, and we will be able to compute
	// in advance the lane assigned to a transaction (useful for testing).
	switch {
	case keyInt%11 == 0:
		return "foo" // priority 7
	case keyInt%3 == 0:
		return "bar" // priority 1
	default:
		return lane // priority 3
	}
}

// parseTx parses a tx in 'key=value' format into a key and value.
func parseTx(tx []byte) (key, value string, err error) {
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid tx format: %q", string(tx))
	}
	if len(parts[0]) == 0 {
		return "", "", errors.New("key cannot be empty")
	}
	return string(parts[0]), string(parts[1]), nil
}

// Tx must have a format like key:value or key=value. That is:
// - it must have one and only one ":" or "="
// - It must not begin or end with these special characters.
func isValidTx(tx []byte) bool {
	if bytes.Count(tx, []byte(":")) == 1 && bytes.Count(tx, []byte("=")) == 0 {
		if !bytes.HasPrefix(tx, []byte(":")) && !bytes.HasSuffix(tx, []byte(":")) {
			return true
		}
	} else if bytes.Count(tx, []byte("=")) == 1 && bytes.Count(tx, []byte(":")) == 0 {
		if !bytes.HasPrefix(tx, []byte("=")) && !bytes.HasSuffix(tx, []byte("=")) {
			return true
		}
	}
	return false
}

// PrepareProposal is called when the node is a proposer. CometBFT stages a set of transactions to the application. As the
// KVStore has two accepted formats, `:` and `=`, we modify all instances of `:` with `=` to make it consistent. Note: this is
// quite a trivial example of transaction modification.
// NOTE: we assume that CometBFT will never provide more transactions than can fit in a block.
func (app *Application) PrepareProposal(ctx context.Context, req *types.PrepareProposalRequest) (*types.PrepareProposalResponse, error) {
	return &types.PrepareProposalResponse{Txs: app.formatTxs(ctx, req.Txs)}, nil
}

// formatTxs validates and excludes invalid transactions
// also substitutes all the transactions with x:y to x=y.
func (app *Application) formatTxs(ctx context.Context, blockData [][]byte) [][]byte {
	txs := make([][]byte, 0, len(blockData))
	for _, tx := range blockData {
		resp, err := app.CheckTx(ctx, &types.CheckTxRequest{Tx: tx, Type: types.CHECK_TX_TYPE_CHECK})
		if err != nil {
			panic(fmt.Sprintln("formatTxs: CheckTx call had an unrecoverable error", err))
		}
		if resp.Code == CodeTypeOK {
			txs = append(txs, bytes.Replace(tx, []byte(":"), []byte("="), 1))
		}
	}
	return txs
}

// ProcessProposal is called whenever a node receives a complete proposal. It allows the application to validate the proposal.
// Only validators who can vote will have this method called. For the KVstore we reuse CheckTx.
func (app *Application) ProcessProposal(ctx context.Context, req *types.ProcessProposalRequest) (*types.ProcessProposalResponse, error) {
	for _, tx := range req.Txs {
		// As CheckTx is a full validity check we can simply reuse this
		resp, err := app.CheckTx(ctx, &types.CheckTxRequest{Tx: tx, Type: types.CHECK_TX_TYPE_CHECK})
		if err != nil {
			panic(fmt.Sprintln("ProcessProposal: CheckTx call had an unrecoverable error", err))
		}
		if resp.Code != CodeTypeOK {
			return &types.ProcessProposalResponse{Status: types.PROCESS_PROPOSAL_STATUS_REJECT}, nil
		}
	}
	return &types.ProcessProposalResponse{Status: types.PROCESS_PROPOSAL_STATUS_ACCEPT}, nil
}

// FinalizeBlock executes the block against the application state. It punishes validators who equivocated and
// updates validators according to transactions in a block. The rest of the transactions are regular key value
// updates and are cached in memory and will be persisted once Commit is called.
// ConsensusParams are never changed.
func (app *Application) FinalizeBlock(_ context.Context, req *types.FinalizeBlockRequest) (*types.FinalizeBlockResponse, error) {
	// reset valset changes
	app.valUpdates = make([]types.ValidatorUpdate, 0)
	app.stagedTxs = make([][]byte, 0)

	// Punish validators who committed equivocation.
	for _, ev := range req.Misbehavior {
		if ev.Type == types.MISBEHAVIOR_TYPE_DUPLICATE_VOTE {
			addr := string(ev.Validator.Address)
			//nolint:revive // this is a false positive from early-return
			if pubKey, ok := app.valAddrToPubKeyMap[addr]; ok {
				app.valUpdates = append(app.valUpdates, types.ValidatorUpdate{
					Power:       ev.Validator.Power - 1,
					PubKeyType:  pubKey.Type(),
					PubKeyBytes: pubKey.Bytes(),
				})
				app.logger.Info("Decreased val power by 1 because of the equivocation",
					"val", addr)
			} else {
				panic(fmt.Errorf("wanted to punish val %q but can't find it", addr))
			}
		}
	}

	respTxs := make([]*types.ExecTxResult, len(req.Txs))
	for i, tx := range req.Txs {
		if isValidatorTx(tx) {
			keyType, pubKey, power, err := parseValidatorTx(tx)
			if err != nil {
				panic(err)
			}
			app.valUpdates = append(app.valUpdates, types.ValidatorUpdate{Power: power, PubKeyType: keyType, PubKeyBytes: pubKey})
		} else {
			app.stagedTxs = append(app.stagedTxs, tx)
		}

		var key, value string
		parts := bytes.Split(tx, []byte("="))
		if len(parts) == 2 {
			key, value = string(parts[0]), string(parts[1])
		} else {
			key, value = string(tx), string(tx)
		}
		respTxs[i] = &types.ExecTxResult{
			Code: CodeTypeOK,
			// With every transaction we can emit a series of events. To make it simple, we just emit the same events.
			Events: []types.Event{
				{
					Type: "app",
					Attributes: []types.EventAttribute{
						{Key: "creator", Value: "Cosmoshi Netowoko", Index: true},
						{Key: "key", Value: key, Index: true},
						{Key: "index_key", Value: "index is working", Index: true},
						{Key: "noindex_key", Value: "index is working", Index: false},
					},
				},
				{
					Type: "app",
					Attributes: []types.EventAttribute{
						{Key: "creator", Value: "Cosmoshi", Index: true},
						{Key: "key", Value: value, Index: true},
						{Key: "index_key", Value: "index is working", Index: true},
						{Key: "noindex_key", Value: "index is working", Index: false},
					},
				},
			},
		}
		app.state.Size++
	}

	app.state.Height = req.Height

	response := &types.FinalizeBlockResponse{
		TxResults:        respTxs,
		ValidatorUpdates: app.valUpdates,
		AppHash:          app.state.Hash(),
		NextBlockDelay:   app.nextBlockDelay,
	}

	if !app.genBlockEvents {
		return response, nil
	}
	if app.state.Height%2 == 0 {
		response.Events = []types.Event{
			{
				Type: "begin_event",
				Attributes: []types.EventAttribute{
					{
						Key:   "foo",
						Value: "100",
						Index: true,
					},
					{
						Key:   "bar",
						Value: "200",
						Index: true,
					},
				},
			},
			{
				Type: "begin_event",
				Attributes: []types.EventAttribute{
					{
						Key:   "foo",
						Value: "200",
						Index: true,
					},
					{
						Key:   "bar",
						Value: "300",
						Index: true,
					},
				},
			},
		}
	} else {
		response.Events = []types.Event{
			{
				Type: "begin_event",
				Attributes: []types.EventAttribute{
					{
						Key:   "foo",
						Value: "400",
						Index: true,
					},
					{
						Key:   "bar",
						Value: "300",
						Index: true,
					},
				},
			},
		}
	}
	return response, nil
}

// Commit is called after FinalizeBlock and after Tendermint state which includes the updates to
// AppHash, ConsensusParams and ValidatorSet has occurred.
// The KVStore persists the validator updates and the new key values.
func (app *Application) Commit(context.Context, *types.CommitRequest) (*types.CommitResponse, error) {
	// apply the validator updates to state (note this is really the validator set at h + 2)
	for _, valUpdate := range app.valUpdates {
		app.updateValidator(valUpdate)
	}

	// persist all the staged txs in the kvstore
	for _, tx := range app.stagedTxs {
		parts := bytes.Split(tx, []byte("="))
		if len(parts) != 2 {
			panic(fmt.Sprintf("unexpected tx format. Expected 2 got %d: %s", len(parts), parts))
		}
		key, value := string(parts[0]), string(parts[1])
		err := app.state.db.Set(prefixKey([]byte(key)), []byte(value))
		if err != nil {
			panic(err)
		}
	}

	// persist the state (i.e. size and height)
	saveState(app.state)

	resp := &types.CommitResponse{}
	if app.RetainBlocks > 0 && app.state.Height >= app.RetainBlocks {
		resp.RetainHeight = app.state.Height - app.RetainBlocks + 1
	}
	return resp, nil
}

// Query returns an associated value or nil if missing.
func (app *Application) Query(_ context.Context, reqQuery *types.QueryRequest) (*types.QueryResponse, error) {
	resQuery := &types.QueryResponse{}

	if reqQuery.Path == "/val" {
		key := []byte(ValidatorPrefix + string(reqQuery.Data))
		value, err := app.state.db.Get(key)
		if err != nil {
			panic(err)
		}

		return &types.QueryResponse{
			Key:   reqQuery.Data,
			Value: value,
		}, nil
	}

	if reqQuery.Prove {
		value, err := app.state.db.Get(prefixKey(reqQuery.Data))
		if err != nil {
			panic(err)
		}

		if value == nil {
			resQuery.Log = "does not exist"
		} else {
			resQuery.Log = "exists"
		}
		resQuery.Index = -1 // TODO make Proof return index
		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		resQuery.Height = app.state.Height

		return resQuery, nil
	}

	resQuery.Key = reqQuery.Data
	value, err := app.state.db.Get(prefixKey(reqQuery.Data))
	if err != nil {
		panic(err)
	}
	if value == nil {
		resQuery.Log = "does not exist"
	} else {
		resQuery.Log = "exists"
	}
	resQuery.Value = value
	resQuery.Height = app.state.Height

	return resQuery, nil
}

func (app *Application) Close() error {
	return app.state.db.Close()
}

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorPrefix)
}

func parseValidatorTx(tx []byte) (string, []byte, int64, error) {
	tx = tx[len(ValidatorPrefix):]

	//  get the pubkey and power
	typePubKeyAndPower := strings.Split(string(tx), "!")
	if len(typePubKeyAndPower) != 3 {
		return "", nil, 0, fmt.Errorf("expected 'pubkeytype!pubkey!power'. Got %v", typePubKeyAndPower)
	}
	keyType, pubkeyS, powerS := typePubKeyAndPower[0], typePubKeyAndPower[1], typePubKeyAndPower[2]

	// decode the pubkey
	pubkey, err := base64.StdEncoding.DecodeString(pubkeyS)
	if err != nil {
		return "", nil, 0, fmt.Errorf("pubkey (%s) is invalid base64", pubkeyS)
	}

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return "", nil, 0, fmt.Errorf("power (%s) is not an int", powerS)
	}

	if power < 0 {
		return "", nil, 0, fmt.Errorf("power can not be less than 0, got %d", power)
	}

	return keyType, pubkey, power, nil
}

// add, update, or remove a validator.
func (app *Application) updateValidator(v types.ValidatorUpdate) {
	pubkey, err := cryptoenc.PubKeyFromTypeAndBytes(v.PubKeyType, v.PubKeyBytes)
	if err != nil {
		panic(err)
	}
	key := []byte(ValidatorPrefix + string(pubkey.Bytes()))

	if v.Power == 0 {
		// remove validator
		hasKey, err := app.state.db.Has(key)
		if err != nil {
			panic(err)
		}
		if !hasKey {
			pubStr := base64.StdEncoding.EncodeToString(pubkey.Bytes())
			app.logger.Info("tried to remove non existent validator. Skipping...", "pubKey", pubStr)
		}
		if err = app.state.db.Delete(key); err != nil {
			panic(err)
		}
		delete(app.valAddrToPubKeyMap, string(pubkey.Address()))
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			panic(err)
		}
		if err = app.state.db.Set(key, value.Bytes()); err != nil {
			panic(err)
		}
		app.valAddrToPubKeyMap[string(pubkey.Address())] = pubkey
	}
}

func (app *Application) getValidators() (validators []types.ValidatorUpdate) {
	itr, err := app.state.db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	for ; itr.Valid(); itr.Next() {
		if isValidatorTx(itr.Key()) {
			validator := new(types.ValidatorUpdate)
			err := types.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
			if err != nil {
				panic(err)
			}
			validators = append(validators, *validator)
		}
	}
	if err = itr.Error(); err != nil {
		panic(err)
	}
	return validators
}

// -----------------------------

type State struct {
	db dbm.DB
	// Size is essentially the amount of transactions that have been processes.
	// This is used for the appHash
	Size   int64 `json:"size"`
	Height int64 `json:"height"`
}

func loadState(db dbm.DB) State {
	var state State
	state.db = db
	stateBytes, err := db.Get(stateKey)
	if err != nil {
		panic(err)
	}
	if len(stateBytes) == 0 {
		return state
	}
	err = json.Unmarshal(stateBytes, &state)
	if err != nil {
		panic(err)
	}
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	err = state.db.Set(stateKey, stateBytes)
	if err != nil {
		panic(err)
	}
}

// Hash returns the hash of the application state. This is computed
// as the size or number of transactions processed within the state. Note that this isn't
// a strong guarantee of state machine replication because states could
// have different kv values but still have the same size.
// This function is used as the "AppHash".
func (s State) Hash() []byte {
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, s.Size)
	return appHash
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}
