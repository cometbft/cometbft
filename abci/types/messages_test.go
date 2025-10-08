package types

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	version "github.com/cometbft/cometbft/proto/tendermint/version"
)

func TestMarshalJSON(t *testing.T) {
	b, err := json.Marshal(&ExecTxResult{Code: 1})
	assert.NoError(t, err)
	// include empty fields.
	assert.True(t, strings.Contains(string(b), "code"))
	r1 := ResponseCheckTx{
		Code:      1,
		Data:      []byte("hello"),
		GasWanted: 43,
		Events: []Event{
			{
				Type: "testEvent",
				Attributes: []EventAttribute{
					{Key: "pho", Value: "bo"},
				},
			},
		},
	}
	b, err = json.Marshal(&r1)
	assert.Nil(t, err)

	var r2 ResponseCheckTx
	err = json.Unmarshal(b, &r2)
	assert.Nil(t, err)
	assert.Equal(t, r1, r2)
}

func TestWriteReadMessageSimple(t *testing.T) {
	cases := []proto.Message{
		&RequestEcho{
			Message: "Hello",
		},
	}

	for _, c := range cases {
		buf := new(bytes.Buffer)
		err := WriteMessage(c, buf)
		assert.Nil(t, err)

		msg := new(RequestEcho)
		err = ReadMessage(buf, msg)
		assert.Nil(t, err)

		assert.True(t, proto.Equal(c, msg))
	}
}

func TestWriteReadMessage(t *testing.T) {
	cases := []proto.Message{
		&cmtproto.Header{
			Height:  4,
			ChainID: "test",
		},
		&cmtproto.Header{
			Version: version.Consensus{Block: 11, App: 22},
			ChainID: "chain-A",
			Height:  42,
			Time:    time.Unix(1_700_000_000, 0).UTC(),
			LastBlockId: cmtproto.BlockID{
				Hash: []byte{0x01, 0x02, 0x03},
				PartSetHeader: cmtproto.PartSetHeader{
					Total: 123,
					Hash:  []byte{0xaa, 0xbb, 0xcc},
				},
			},
			LastCommitHash:     []byte{0x10},
			DataHash:           []byte{0x20},
			ValidatorsHash:     []byte{0x30},
			NextValidatorsHash: []byte{0x40},
			ConsensusHash:      []byte{0x50},
			AppHash:            []byte{0x60},
			LastResultsHash:    []byte{0x70},
			EvidenceHash:       []byte{0x80},
			ProposerAddress:    []byte{0x90},
		},
		&cmtproto.Header{
			Version: version.Consensus{Block: 0, App: 0},
			ChainID: "chain-B",
			Height:  1,
			Time:    time.Unix(0, 0).UTC(),
			LastBlockId: cmtproto.BlockID{
				Hash: []byte{},
				PartSetHeader: cmtproto.PartSetHeader{
					Total: 0,
					Hash:  nil,
				},
			},
		},
	}

	for _, c := range cases {
		buf := new(bytes.Buffer)
		err := WriteMessage(c, buf)
		assert.Nil(t, err)

		msg := new(cmtproto.Header)
		err = ReadMessage(buf, msg)
		assert.Nil(t, err)

		assert.True(t, proto.Equal(c, msg))
	}
}

func TestWriteReadMessage2(t *testing.T) {
	phrase := "hello-world"
	cases := []proto.Message{
		&ResponseCheckTx{
			Data:      []byte(phrase),
			Log:       phrase,
			GasWanted: 10,
			Events: []Event{
				{
					Type: "testEvent",
					Attributes: []EventAttribute{
						{Key: "abc", Value: "def"},
					},
				},
			},
		},
		&ResponseCheckTx{
			Code:      1,
			Data:      []byte("transaction data"),
			Log:       "check tx log",
			Info:      "additional info",
			GasWanted: 1000,
			GasUsed:   800,
			Codespace: "test-codespace",
			Events: []Event{
				{
					Type: "transfer",
					Attributes: []EventAttribute{
						{Key: "sender", Value: "alice", Index: true},
						{Key: "receiver", Value: "bob", Index: false},
					},
				},
				{
					Type: "fee",
					Attributes: []EventAttribute{
						{Key: "amount", Value: "100", Index: true},
					},
				},
			},
		},
		&ResponseCheckTx{
			Code:      0,
			Data:      nil,
			Log:       "",
			Info:      "",
			GasWanted: 0,
			GasUsed:   0,
			Codespace: "",
			Events:    nil,
		},
		&ResponseCheckTx{
			Code:      42,
			Data:      []byte{0x01, 0x02, 0x03, 0x04},
			Log:       "error occurred",
			Info:      "detailed error info",
			GasWanted: 5000,
			GasUsed:   4500,
			Codespace: "error-codespace",
			Events: []Event{
				{
					Type: "error",
					Attributes: []EventAttribute{
						{Key: "error_code", Value: "42", Index: true},
						{Key: "error_message", Value: "validation failed", Index: false},
					},
				},
			},
		},
	}

	for _, c := range cases {
		buf := new(bytes.Buffer)
		err := WriteMessage(c, buf)
		assert.Nil(t, err)

		msg := new(ResponseCheckTx)
		err = ReadMessage(buf, msg)
		assert.Nil(t, err)

		assert.True(t, proto.Equal(c, msg))
	}
}
