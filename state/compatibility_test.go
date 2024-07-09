package state_test

import (
	"fmt"
	"testing"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sm "github.com/cometbft/cometbft/state"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtstate "github.com/cometbft/cometbft/proto/tendermint/state"
	"github.com/stretchr/testify/require"
)

// Compatibility test across different state proto versions

func calcABCIResponsesKey(height int64) []byte {
	return []byte(fmt.Sprintf("abciResponsesKey:%v", height))
}

var lastABCIResponseKey = []byte("lastABCIResponseKey")

var (
	_ sm.Store    = (*MultiStore)(nil)
	_ LegacyStore = (*MultiStore)(nil)
)

// MultiStore represents a state store that implements the Store interface
// and contains additional store and database options.
type MultiStore struct {
	sm.Store
	db dbm.DB
	sm.StoreOptions
}

// NewMultiStore returns a new MultiStore.
// It sets the store, db, and StoreOptions fields of the MultiStore struct.
func NewMultiStore(db dbm.DB, options sm.StoreOptions, store sm.Store) *MultiStore {
	return &MultiStore{
		Store:        store,
		db:           db,
		StoreOptions: options,
	}
}

// LegacyStore represents a legacy data store.
type LegacyStore interface {
	SaveABCIResponses(height int64, abciResponses *cmtstate.LegacyABCIResponses) error
}

// SaveABCIResponses saves the ABCIResponses for a given height in the MultiStore.
// It strips out any nil values from the DeliverTxs field, and saves the ABCIResponses to
// disk if the DiscardABCIResponses flag is set to false. It also saves the last ABCI response
// for crash recovery, overwriting the previously saved response.
func (multi MultiStore) SaveABCIResponses(height int64, abciResponses *cmtstate.LegacyABCIResponses) error {
	var dtxs []*abci.ExecTxResult
	// strip nil values,
	for _, tx := range abciResponses.DeliverTxs {
		if tx != nil {
			dtxs = append(dtxs, tx)
		}
	}
	abciResponses.DeliverTxs = dtxs

	// If the flag is false then we save the ABCIResponse. This can be used for the /block_results
	// query or to reindex an event using the command line.
	if !multi.StoreOptions.DiscardABCIResponses {
		bz, err := abciResponses.Marshal()
		if err != nil {
			return err
		}
		if err := multi.db.Set(calcABCIResponsesKey(height), bz); err != nil {
			return err
		}
	}

	// We always save the last ABCI response for crash recovery.
	// This overwrites the previous saved ABCI Response.
	response := &cmtstate.ABCIResponsesInfo{
		LegacyAbciResponses: abciResponses,
		Height:              height,
	}
	bz, err := response.Marshal()
	if err != nil {
		return err
	}

	return multi.db.SetSync(lastABCIResponseKey, bz)
}

// TestLegacySaveAndLoadFinalizeBlock tests saving and loading of ABCIResponses
// using the multiStore. It verifies that the loaded ABCIResponses match the
// original ones and that missing fields are correctly handled.
// This test is important for the LoadFinalizeBlockResponse method in the state store.
func TestLegacySaveAndLoadFinalizeBlock(t *testing.T) {
	tearDown, stateDB, _, store := setupTestCaseWithStore(t)
	defer tearDown(t)
	options := sm.StoreOptions{
		DiscardABCIResponses: false,
	}

	height := int64(1)
	multiStore := NewMultiStore(stateDB, options, store)

	// try with a complete ABCI Response
	legacyABCIResponses := newLegacyABCIResponses()
	err := multiStore.SaveABCIResponses(height, &legacyABCIResponses)
	require.NoError(t, err)
	require.Equal(t, 1, len(legacyABCIResponses.DeliverTxs))
	require.Equal(t, 1, len(legacyABCIResponses.BeginBlock.Events))
	require.Equal(t, 1, len(legacyABCIResponses.EndBlock.Events))

	responseFinalizeBlock, err := multiStore.LoadFinalizeBlockResponse(height)
	require.NoError(t, err)

	// Test for not nil
	require.NotNil(t, responseFinalizeBlock.TxResults)
	require.NotNil(t, responseFinalizeBlock.Events)
	require.NotNil(t, responseFinalizeBlock.ValidatorUpdates)
	require.NotNil(t, responseFinalizeBlock.ConsensusParamUpdates)
	require.Nil(t, responseFinalizeBlock.AppHash)

	// Test for equality
	require.Equal(t, 1, len(responseFinalizeBlock.TxResults))
	require.Equal(t, len(legacyABCIResponses.DeliverTxs), len(responseFinalizeBlock.TxResults))
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].Code, responseFinalizeBlock.TxResults[0].Code)
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].Data, responseFinalizeBlock.TxResults[0].Data)
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].Log, responseFinalizeBlock.TxResults[0].Log)
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].GasWanted, responseFinalizeBlock.TxResults[0].GasWanted)
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].GasUsed, responseFinalizeBlock.TxResults[0].GasUsed)
	require.Equal(t, len(legacyABCIResponses.DeliverTxs[0].Events), len(responseFinalizeBlock.TxResults[0].Events))
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].Events[0].Type, responseFinalizeBlock.TxResults[0].Events[0].Type)
	require.Equal(t, len(legacyABCIResponses.DeliverTxs[0].Events[0].Attributes), len(responseFinalizeBlock.TxResults[0].Events[0].Attributes))
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].Events[0].Attributes[0].Key, responseFinalizeBlock.TxResults[0].Events[0].Attributes[0].Key)
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].Events[0].Attributes[0].Value, responseFinalizeBlock.TxResults[0].Events[0].Attributes[0].Value)
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].Codespace, responseFinalizeBlock.TxResults[0].Codespace)

	require.Equal(t, 2, len(responseFinalizeBlock.Events))
	require.Equal(t, len(legacyABCIResponses.BeginBlock.Events)+len(legacyABCIResponses.EndBlock.Events), len(responseFinalizeBlock.Events))

	require.Equal(t, legacyABCIResponses.BeginBlock.Events[0].Type, responseFinalizeBlock.Events[0].Type)
	require.Equal(t, len(legacyABCIResponses.BeginBlock.Events[0].Attributes)+1, len(responseFinalizeBlock.Events[0].Attributes)) // +1 for inject 'mode' attribute
	require.Equal(t, legacyABCIResponses.BeginBlock.Events[0].Attributes[0].Key, responseFinalizeBlock.Events[0].Attributes[0].Key)
	require.Equal(t, legacyABCIResponses.BeginBlock.Events[0].Attributes[0].Value, responseFinalizeBlock.Events[0].Attributes[0].Value)

	require.Equal(t, legacyABCIResponses.EndBlock.ConsensusParamUpdates.Block.MaxBytes, responseFinalizeBlock.ConsensusParamUpdates.Block.MaxBytes)
	require.Equal(t, legacyABCIResponses.EndBlock.ConsensusParamUpdates.Block.MaxGas, responseFinalizeBlock.ConsensusParamUpdates.Block.MaxGas)
	require.Equal(t, legacyABCIResponses.EndBlock.ConsensusParamUpdates.Evidence.MaxAgeNumBlocks, responseFinalizeBlock.ConsensusParamUpdates.Evidence.MaxAgeNumBlocks)
	require.Equal(t, legacyABCIResponses.EndBlock.ConsensusParamUpdates.Evidence.MaxAgeDuration, responseFinalizeBlock.ConsensusParamUpdates.Evidence.MaxAgeDuration)
	require.Equal(t, legacyABCIResponses.EndBlock.ConsensusParamUpdates.Evidence.MaxBytes, responseFinalizeBlock.ConsensusParamUpdates.Evidence.MaxBytes)
	require.Equal(t, legacyABCIResponses.EndBlock.ConsensusParamUpdates.Validator.PubKeyTypes, responseFinalizeBlock.ConsensusParamUpdates.Validator.PubKeyTypes)
	require.Equal(t, legacyABCIResponses.EndBlock.ConsensusParamUpdates.Version.App, responseFinalizeBlock.ConsensusParamUpdates.Version.App)

	require.Nil(t, responseFinalizeBlock.ConsensusParamUpdates.Abci)
	require.Nil(t, responseFinalizeBlock.AppHash)

	require.Equal(t, len(legacyABCIResponses.EndBlock.ValidatorUpdates), len(responseFinalizeBlock.ValidatorUpdates))
	require.Equal(t, legacyABCIResponses.EndBlock.ValidatorUpdates[0].Power, responseFinalizeBlock.ValidatorUpdates[0].Power)

	// skip until an equivalency test is possible
	require.Equal(t, legacyABCIResponses.EndBlock.ValidatorUpdates[0].PubKey.GetEd25519(), responseFinalizeBlock.ValidatorUpdates[0].PubKey.GetEd25519())

	// try with an ABCI Response missing fields
	height = int64(2)
	legacyABCIResponses = newLegacyABCIResponsesWithNullFields()
	require.Equal(t, 1, len(legacyABCIResponses.DeliverTxs))
	require.Equal(t, 1, len(legacyABCIResponses.BeginBlock.Events))
	require.Nil(t, legacyABCIResponses.EndBlock)
	err = multiStore.SaveABCIResponses(height, &legacyABCIResponses)
	require.NoError(t, err)
	responseFinalizeBlock, err = multiStore.LoadFinalizeBlockResponse(height)
	require.NoError(t, err)

	require.Equal(t, len(legacyABCIResponses.DeliverTxs), len(responseFinalizeBlock.TxResults))
	require.Equal(t, legacyABCIResponses.DeliverTxs[0].String(), responseFinalizeBlock.TxResults[0].String())
	require.Equal(t, len(legacyABCIResponses.BeginBlock.Events), len(responseFinalizeBlock.Events))
}

// Generate a Legacy ABCIResponses with data for all fields.
func newLegacyABCIResponses() cmtstate.LegacyABCIResponses {
	eventAttr := abci.EventAttribute{
		Key:   "key",
		Value: "value",
	}

	deliverTxEvent := abci.Event{
		Type:       "deliver_tx_event",
		Attributes: []abci.EventAttribute{eventAttr},
	}

	endBlockEvent := abci.Event{
		Type:       "end_block_event",
		Attributes: []abci.EventAttribute{eventAttr},
	}

	beginBlockEvent := abci.Event{
		Type:       "begin_block_event",
		Attributes: []abci.EventAttribute{eventAttr},
	}

	responseDeliverTx := abci.ExecTxResult{
		Code:   abci.CodeTypeOK,
		Events: []abci.Event{deliverTxEvent},
	}

	validatorUpdates := []abci.ValidatorUpdate{{
		PubKey: cmtcrypto.PublicKey{Sum: &cmtcrypto.PublicKey_Ed25519{Ed25519: make([]byte, 1)}},
		Power:  int64(10),
	}}

	consensusParams := &cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxBytes: int64(100000),
			MaxGas:   int64(10000),
		},
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeNumBlocks: int64(10),
			MaxAgeDuration:  time.Duration(1000),
			MaxBytes:        int64(10000),
		},
		Validator: &cmtproto.ValidatorParams{
			PubKeyTypes: []string{"ed25519"},
		},
		Version: &cmtproto.VersionParams{
			App: uint64(10),
		},
	}

	// Legacy ABCI Responses
	legacyABCIResponses := cmtstate.LegacyABCIResponses{
		DeliverTxs: []*abci.ExecTxResult{
			&responseDeliverTx,
		},
		EndBlock: &cmtstate.ResponseEndBlock{
			Events:                []abci.Event{endBlockEvent},
			ConsensusParamUpdates: consensusParams,
			ValidatorUpdates:      validatorUpdates,
		},
		BeginBlock: &cmtstate.ResponseBeginBlock{
			Events: []abci.Event{beginBlockEvent},
		},
	}
	return legacyABCIResponses
}

// Generate a Legacy ABCIResponses with null data for some fields.
func newLegacyABCIResponsesWithNullFields() cmtstate.LegacyABCIResponses {
	eventAttr := abci.EventAttribute{
		Key:   "key",
		Value: "value",
	}

	deliverTxEvent := abci.Event{
		Type:       "deliver_tx_event",
		Attributes: []abci.EventAttribute{eventAttr},
	}

	beginBlockEvent := abci.Event{
		Type:       "begin_block_event",
		Attributes: []abci.EventAttribute{eventAttr},
	}

	responseDeliverTx := abci.ExecTxResult{
		Code:   abci.CodeTypeOK,
		Events: []abci.Event{deliverTxEvent},
	}

	// Legacy ABCI Responses
	legacyABCIResponses := cmtstate.LegacyABCIResponses{
		DeliverTxs: []*abci.ExecTxResult{
			&responseDeliverTx,
		},
		BeginBlock: &cmtstate.ResponseBeginBlock{
			Events: []abci.Event{beginBlockEvent},
		},
	}
	return legacyABCIResponses
}
