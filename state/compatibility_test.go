package state_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	abciv1beta1 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta1"
	abciv1beta2 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta2"
	cryptov1 "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	statev1 "github.com/cometbft/cometbft/api/cometbft/state/v1"
	statev1beta2 "github.com/cometbft/cometbft/api/cometbft/state/v1beta2"
	statev1beta3 "github.com/cometbft/cometbft/api/cometbft/state/v1beta3"
	typesv1beta1 "github.com/cometbft/cometbft/api/cometbft/types/v1beta1"
	typesv1beta2 "github.com/cometbft/cometbft/api/cometbft/types/v1beta2"
	sm "github.com/cometbft/cometbft/state"
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

type MultiStore struct {
	sm.Store
	db dbm.DB
	sm.StoreOptions
}

func NewMultiStore(db dbm.DB, options sm.StoreOptions, store sm.Store) *MultiStore {
	return &MultiStore{
		Store:        store,
		db:           db,
		StoreOptions: options,
	}
}

type LegacyStore interface {
	SaveABCIResponses(height int64, abciResponses *statev1beta2.ABCIResponses) error
}

func (multi MultiStore) SaveABCIResponses(height int64, abciResponses *statev1beta2.ABCIResponses) error {
	var dtxs []*abciv1beta2.ResponseDeliverTx
	// strip nil values,
	for _, tx := range abciResponses.DeliverTxs {
		if tx != nil {
			dtxs = append(dtxs, tx)
		}
	}
	abciResponses.DeliverTxs = dtxs

	// If the flag is false then we save the ABCIResponse. This can be used for the /BlockResults
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
	response := &statev1beta2.ABCIResponsesInfo{
		AbciResponses: abciResponses,
		Height:        height,
	}
	bz, err := response.Marshal()
	if err != nil {
		return err
	}

	return multi.db.SetSync(lastABCIResponseKey, bz)
}

func TestSaveLegacyAndLoadFinalizeBlock(t *testing.T) {
	tearDown, stateDB, _, store := setupTestCaseWithStore(t)
	defer tearDown(t)
	options := sm.StoreOptions{
		DiscardABCIResponses: false,
	}

	height := int64(1)
	multiStore := NewMultiStore(stateDB, options, store)

	// try with a complete ABCI Response
	abciResponses := newV1Beta2ABCIResponses()
	require.Equal(t, 1, len(abciResponses.DeliverTxs))
	require.Equal(t, 1, len(abciResponses.BeginBlock.Events))
	require.Equal(t, 1, len(abciResponses.EndBlock.Events))
	err := multiStore.SaveABCIResponses(height, &abciResponses)
	require.NoError(t, err)
	loadedABCIResponses, err := multiStore.LoadFinalizeBlockResponse(height)
	require.NoError(t, err)
	require.Equal(t, 1, len(loadedABCIResponses.TxResults))
	require.Equal(t, abciResponses.DeliverTxs[0].String(), loadedABCIResponses.TxResults[0].String())
	require.Equal(t, 2, len(loadedABCIResponses.Events))
	require.Equal(t, len(abciResponses.BeginBlock.Events)+len(abciResponses.EndBlock.Events), len(loadedABCIResponses.Events))

	// try with an ABCI Response missing fields
	height = int64(2)
	abciResponses = newV1Beta2ABCIResponsesWithNullFields()
	require.Equal(t, 1, len(abciResponses.DeliverTxs))
	require.Equal(t, 1, len(abciResponses.BeginBlock.Events))
	require.Nil(t, abciResponses.EndBlock)
	err = multiStore.SaveABCIResponses(height, &abciResponses)
	require.NoError(t, err)
	loadedABCIResponses, err = multiStore.LoadFinalizeBlockResponse(height)
	require.NoError(t, err)
	require.Equal(t, 1, len(loadedABCIResponses.TxResults))
	require.Equal(t, abciResponses.DeliverTxs[0].String(), loadedABCIResponses.TxResults[0].String())
	require.Equal(t, 1, len(loadedABCIResponses.Events))
	require.Equal(t, len(abciResponses.BeginBlock.Events), len(loadedABCIResponses.Events))
}

// This test unmarshals a v1beta2.ABCIResponses as a statev1.LegacyABCIResponses
// This logic is important for the LoadFinalizeBlockResponse method in the state sm
// The conversion should not fail because they are compatible.
func TestStateProtoV1Beta2ToV1(t *testing.T) {
	v1beta2ABCIResponses := newV1Beta2ABCIResponses()

	v1b2Resp, err := v1beta2ABCIResponses.Marshal()
	require.NoError(t, err)
	require.NotNil(t, v1b2Resp)

	// un-marshall a v1beta2 ABCI Response as a LegacyABCIResponse
	legacyABCIResponse := new(statev1.LegacyABCIResponses)
	err = legacyABCIResponse.Unmarshal(v1b2Resp)
	require.NoError(t, err)
	require.NotNil(t, legacyABCIResponse.DeliverTxs)
	require.NotNil(t, legacyABCIResponse.BeginBlock)
	require.NotNil(t, legacyABCIResponse.EndBlock)
	require.Equal(t, 1, len(legacyABCIResponse.DeliverTxs))
	require.Equal(t, 1, len(legacyABCIResponse.BeginBlock.Events))
	require.Equal(t, 1, len(legacyABCIResponse.EndBlock.Events))
	require.Equal(t, int64(100000), legacyABCIResponse.EndBlock.ConsensusParamUpdates.Block.MaxBytes)
	require.Equal(t, int64(10000), legacyABCIResponse.EndBlock.ConsensusParamUpdates.Evidence.MaxBytes)
	require.Equal(t, []string{"ed25519"}, legacyABCIResponse.EndBlock.ConsensusParamUpdates.Validator.PubKeyTypes)
	require.Equal(t, uint64(10), legacyABCIResponse.EndBlock.ConsensusParamUpdates.Version.App)
}

// This test unmarshal a v1beta2.ABCIResponses as a v1.FinalizeBlockResponse
// This logic is important for the LoadFinalizeBlockResponse method in the state sm
// The conversion should fail because they are not compatible.
func TestStateV1Beta2ABCIResponsesAsStateV1FinalizeBlockResponse(t *testing.T) {
	v1beta2ABCIResponses := newV1Beta2ABCIResponses()
	data, err := v1beta2ABCIResponses.Marshal()
	require.NoError(t, err)
	require.NotNil(t, data)

	// This cannot work since they have different schemas, a wrong wireType error is generated
	finalizeBlockResponse := new(abci.FinalizeBlockResponse)
	err = finalizeBlockResponse.Unmarshal(data)
	require.Error(t, err)
	require.ErrorContains(t, err, "unexpected EOF")
}

// This test unmarshal a v1beta2.ABCIResponses as a v1.FinalizeBlockResponse
// This logic is important for the LoadFinalizeBlockResponse method in the state sm
// The conversion doesn't fail because no error is return, but they are NOT compatible.
func TestStateV1Beta2ABCIResponsesWithNullAsStateV1FinalizeBlockResponse(t *testing.T) {
	v1beta2ABCIResponsesWithNull := newV1Beta2ABCIResponsesWithNullFields()
	data, err := v1beta2ABCIResponsesWithNull.Marshal()
	require.NoError(t, err)
	require.NotNil(t, data)

	// This should not work since they have different schemas
	// but an error is not returned, so it deserializes an ABCIResponse
	// on top of a FinalizeBlockResponse giving the false impression they are the same
	// but because it doesn't error out, the fields in finalizeBlockResponse will have
	// their zero-value (e.g. nil, 0, "")
	finalizeBlockResponse := new(abci.FinalizeBlockResponse)
	err = finalizeBlockResponse.Unmarshal(data)
	require.NoError(t, err)
	require.Nil(t, finalizeBlockResponse.AppHash)
	require.Nil(t, finalizeBlockResponse.TxResults)
}

// This test unmarshal a v1beta2.ABCIResponses as a statev1beta3.LegacyABCIResponses
// This logic is important for the LoadFinalizeBlockResponse method in the state sm
// The conversion should work because they are compatible.
func TestStateV1Beta2ABCIResponsesAsStateV1Beta3FinalizeBlockResponse(t *testing.T) {
	v1beta2ABCIResponses := newV1Beta2ABCIResponses()

	data, err := v1beta2ABCIResponses.Marshal()
	require.NoError(t, err)
	require.NotNil(t, data)

	// This works because they are equivalent protos and the fields are populated
	legacyABCIResponses := new(statev1beta3.LegacyABCIResponses)
	err = legacyABCIResponses.Unmarshal(data)
	require.NoError(t, err)
	require.NotNil(t, legacyABCIResponses.DeliverTxs)
	require.NotNil(t, legacyABCIResponses.EndBlock)
	require.NotNil(t, legacyABCIResponses.BeginBlock)
}

// This test unmarshal a v1beta2.ABCIResponses as a statev1beta3.LegacyABCIResponses
// This logic is important for the LoadFinalizeBlockResponse method in the state sm
// The conversion should work because they are compatible even if fields to be converted are null.
func TestStateV1Beta2ABCIResponsesWithNullAsStateV1Beta3FinalizeBlockResponse(t *testing.T) {
	v1beta2ABCIResponsesWithNull := newV1Beta2ABCIResponsesWithNullFields()
	data, err := v1beta2ABCIResponsesWithNull.Marshal()
	require.NoError(t, err)
	require.NotNil(t, data)

	// This works because they are equivalent protos and the fields are populated
	// even if a field is null, it will be converted properly
	legacyResponseWithNull := new(statev1beta3.LegacyABCIResponses)
	err = legacyResponseWithNull.Unmarshal(data)
	require.NoError(t, err)
	require.NotNil(t, legacyResponseWithNull.DeliverTxs)
	require.Nil(t, legacyResponseWithNull.EndBlock)
	require.NotNil(t, legacyResponseWithNull.BeginBlock)
}

// Generate a v1beta2 ABCIResponses with data for all fields.
func newV1Beta2ABCIResponses() statev1beta2.ABCIResponses {
	eventAttr := abciv1beta2.EventAttribute{
		Key:   "key",
		Value: "value",
	}

	deliverTxEvent := abciv1beta2.Event{
		Type:       "deliver_tx_event",
		Attributes: []abciv1beta2.EventAttribute{eventAttr},
	}

	responseDeliverTx := abciv1beta2.ResponseDeliverTx{
		Code:   abci.CodeTypeOK,
		Data:   []byte("result tx data"),
		Log:    "tx committed successfully",
		Info:   "tx processing info",
		Events: []abciv1beta2.Event{deliverTxEvent},
	}

	validatorUpdates := []abciv1beta1.ValidatorUpdate{{
		PubKey: cryptov1.PublicKey{Sum: &cryptov1.PublicKey_Ed25519{Ed25519: make([]byte, 1)}},
		Power:  int64(10),
	}}

	consensusParams := &typesv1beta2.ConsensusParams{
		Block: &typesv1beta2.BlockParams{
			MaxBytes: int64(100000),
			MaxGas:   int64(10000),
		},
		Evidence: &typesv1beta1.EvidenceParams{
			MaxAgeNumBlocks: int64(10),
			MaxAgeDuration:  time.Duration(1000),
			MaxBytes:        int64(10000),
		},
		Validator: &typesv1beta1.ValidatorParams{
			PubKeyTypes: []string{"ed25519"},
		},
		Version: &typesv1beta1.VersionParams{
			App: uint64(10),
		},
	}

	endBlockEvent := abciv1beta2.Event{
		Type:       "end_block_event",
		Attributes: []abciv1beta2.EventAttribute{eventAttr},
	}

	beginBlockEvent := abciv1beta2.Event{
		Type:       "begin_block_event",
		Attributes: []abciv1beta2.EventAttribute{eventAttr},
	}

	// v1beta2 ABCI Responses
	v1beta2ABCIResponses := statev1beta2.ABCIResponses{
		BeginBlock: &abciv1beta2.ResponseBeginBlock{
			Events: []abciv1beta2.Event{beginBlockEvent},
		},
		DeliverTxs: []*abciv1beta2.ResponseDeliverTx{
			&responseDeliverTx,
		},
		EndBlock: &abciv1beta2.ResponseEndBlock{
			ValidatorUpdates:      validatorUpdates,
			ConsensusParamUpdates: consensusParams,
			Events:                []abciv1beta2.Event{endBlockEvent},
		},
	}
	return v1beta2ABCIResponses
}

// Generate a v1beta2 ABCIResponses with fields missing data (nil).
func newV1Beta2ABCIResponsesWithNullFields() statev1beta2.ABCIResponses {
	eventAttr := abciv1beta2.EventAttribute{
		Key:   "key",
		Value: "value",
	}

	deliverTxEvent := abciv1beta2.Event{
		Type:       "deliver_tx_event",
		Attributes: []abciv1beta2.EventAttribute{eventAttr},
	}

	responseDeliverTx := abciv1beta2.ResponseDeliverTx{
		Code:   abci.CodeTypeOK,
		Events: []abciv1beta2.Event{deliverTxEvent},
	}

	beginBlockEvent := abciv1beta2.Event{
		Type:       "begin_block_event",
		Attributes: []abciv1beta2.EventAttribute{eventAttr},
	}

	// v1beta2 ABCI Responses
	v1beta2ABCIResponses := statev1beta2.ABCIResponses{
		BeginBlock: &abciv1beta2.ResponseBeginBlock{
			Events: []abciv1beta2.Event{beginBlockEvent},
		},
		DeliverTxs: []*abciv1beta2.ResponseDeliverTx{
			&responseDeliverTx,
		},
	}
	return v1beta2ABCIResponses
}
