package api_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	v1beta1abci "github.com/cometbft/cometbft/api/cometbft/abci/v1beta1"
	v1beta2abci "github.com/cometbft/cometbft/api/cometbft/abci/v1beta2"
	v1 "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	cmtstate "github.com/cometbft/cometbft/api/cometbft/state/v1"
	v1beta2state "github.com/cometbft/cometbft/api/cometbft/state/v1beta2"
	v1beta1types "github.com/cometbft/cometbft/api/cometbft/types/v1beta1"
	v1beta2types "github.com/cometbft/cometbft/api/cometbft/types/v1beta2"
)

// The 'v1beta2' is the proto level used by CometBFT v0.37 release (check /proto/README.md for details)
// This test creates an ABCIResponse message at the v0.37 level and tries to convert to a 'LegacyABCIResponses'
// that is used in the state store logic to retrieve messages stored using a previous version of ABCI Responses
// The test checks if fields in the original message are present in the converted legacy message.
func TestLoadLegacyResponseFromV1Beta2(t *testing.T) {
	v1beta2ABCIResponses := v1beta2state.ABCIResponses{
		DeliverTxs: []*v1beta2abci.ResponseDeliverTx{
			{
				Code: abci.CodeTypeOK,
				Data: []byte("result tx data"),
				Log:  "tx committed successfully",
				Info: "tx processing info",
				Events: []v1beta2abci.Event{{
					Type: "deliver_tx_event",
					Attributes: []v1beta2abci.EventAttribute{{
						Key:   "key",
						Value: "value",
					}},
				}},
			},
		},
		EndBlock: &v1beta2abci.ResponseEndBlock{
			ValidatorUpdates: []v1beta1abci.ValidatorUpdate{{
				PubKey: v1.PublicKey{Sum: &v1.PublicKey_Ed25519{Ed25519: make([]byte, 1)}},
				Power:  int64(10),
			}},
			ConsensusParamUpdates: &v1beta2types.ConsensusParams{
				Block: &v1beta2types.BlockParams{
					MaxBytes: int64(100000),
					MaxGas:   int64(10000),
				},
				Evidence: &v1beta1types.EvidenceParams{
					MaxAgeNumBlocks: int64(10),
					MaxAgeDuration:  time.Duration(1000),
					MaxBytes:        int64(10000),
				},
				Validator: &v1beta1types.ValidatorParams{
					PubKeyTypes: []string{"ed25519"},
				},
				Version: &v1beta1types.VersionParams{
					App: uint64(1),
				},
			},
			Events: []v1beta2abci.Event{
				{
					Type: "end_block_event",
					Attributes: []v1beta2abci.EventAttribute{{
						Key:   "key",
						Value: "value",
					}},
				},
			},
		},
		BeginBlock: &v1beta2abci.ResponseBeginBlock{
			Events: []v1beta2abci.Event{{
				Type: "begin_block_event",
				Attributes: []v1beta2abci.EventAttribute{{
					Key:   "key",
					Value: "value",
				}},
			}},
		},
	}

	v1b2Resp, err := v1beta2ABCIResponses.Marshal()
	require.NoError(t, err)
	require.NotNil(t, v1b2Resp)

	// un-marshall a v1beta2 ABCI Response as a LegacyABCIResponse
	legacyABCIResponse := new(cmtstate.LegacyABCIResponses)
	err = legacyABCIResponse.Unmarshal(v1b2Resp)
	require.NoError(t, err)
	require.Equal(t, 1, len(legacyABCIResponse.DeliverTxs))
	require.Equal(t, 1, len(legacyABCIResponse.BeginBlock.Events))
	require.Equal(t, 1, len(legacyABCIResponse.EndBlock.Events))
	require.Equal(t, int64(100000), legacyABCIResponse.EndBlock.ConsensusParamUpdates.Block.MaxBytes)
	require.Equal(t, int64(10000), legacyABCIResponse.EndBlock.ConsensusParamUpdates.Evidence.MaxBytes)
	require.Equal(t, []string{"ed25519"}, legacyABCIResponse.EndBlock.ConsensusParamUpdates.Validator.PubKeyTypes)
}

// This test uses a binary file for a v1beta2 ABCI Responses and parses it into a
// LegacyABCIResponses. The conversion should succeed since they should be compatible.
func Test_V1Beta2_ABCIResponses_To_LegacyABCIResponses(t *testing.T) {
	data, err := readBytesFromFile("./test_files/v1beta2_ABCI_Responses.bin")
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}
	legacyABCIResponse := new(cmtstate.LegacyABCIResponses)
	err = legacyABCIResponse.Unmarshal(data)
	require.NoError(t, err)
}

// This test uses a binary file for a v1beta2 ABCI Responses and parses it into a
// FinalizeBlockResponse. The conversion should not succeed since the schema is not compatible.
func Test_V1Beta2_ABCIResponses_To_FinalizeBlockResponse(t *testing.T) {
	data, err := readBytesFromFile("./test_files/v1beta2_ABCI_Responses.bin")
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	finalizeBlockResponse := new(abci.FinalizeBlockResponse)
	err = finalizeBlockResponse.Unmarshal(data)
	require.Error(t, err)
}

func readBytesFromFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}
