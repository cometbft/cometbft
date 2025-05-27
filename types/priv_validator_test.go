package types

import (
	"bytes"
	"github.com/cometbft/cometbft/proto/tendermint/privval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRawDataSignBytes(t *testing.T) {
	const (
		testPrefix   = RawDataSignBytesPrefix
		testChainID  = "test-chain"
		testUniqueID = "unique-id-123"
	)

	testCases := []struct {
		name        string
		chainID     string
		uniqueID    string
		rawBytes    []byte
		expectError bool
	}{
		{
			name:        "success with normal inputs",
			chainID:     testChainID,
			uniqueID:    testUniqueID,
			rawBytes:    []byte("test data"),
			expectError: false,
		},
		{
			name:        "error with empty chain ID",
			chainID:     "",
			uniqueID:    testUniqueID,
			rawBytes:    []byte("test data"),
			expectError: true,
		},
		{
			name:        "error with empty unique ID",
			chainID:     testChainID,
			uniqueID:    "",
			rawBytes:    []byte("test data"),
			expectError: true,
		},
		{
			name:        "error with empty raw bytes",
			chainID:     testChainID,
			uniqueID:    testUniqueID,
			rawBytes:    []byte{},
			expectError: true,
		},
		{
			name:        "error with nil raw bytes",
			chainID:     testChainID,
			uniqueID:    testUniqueID,
			rawBytes:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function under test
			signBytes, err := RawDataSignBytes(tc.chainID, tc.uniqueID, tc.rawBytes)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Create expected sign bytes manually for comparison
			expectedReq := &privval.SignRawBytesRequest{
				ChainId:  tc.chainID,
				UniqueId: tc.uniqueID,
				RawBytes: tc.rawBytes,
			}
			expectedProtoBytes, err := expectedReq.Marshal()
			require.NoError(t, err)

			expectedSignBytes := append([]byte(testPrefix), expectedProtoBytes...)

			// Verify the result has the expected prefix
			prefixLen := len(testPrefix)
			assert.True(t, bytes.Equal([]byte(testPrefix), signBytes[:prefixLen]),
				"sign bytes should start with the expected prefix")

			// Verify the entire sign bytes match the expected result
			assert.Equal(t, expectedSignBytes, signBytes,
				"sign bytes should match the expected format")

			// Additional verification: unmarshal the protobuf part and check fields
			protoBytes := signBytes[prefixLen:]
			unmarshaledReq := &privval.SignRawBytesRequest{}
			err = unmarshaledReq.Unmarshal(protoBytes)
			require.NoError(t, err, "should be able to unmarshal the protobuf bytes")

			assert.Equal(t, tc.chainID, unmarshaledReq.ChainId)
			assert.Equal(t, tc.uniqueID, unmarshaledReq.UniqueId)
			assert.Equal(t, tc.rawBytes, unmarshaledReq.RawBytes)
		})
	}
}
