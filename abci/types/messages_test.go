package types

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
)

func TestMarshalJSON(t *testing.T) {
	b, err := json.Marshal(&ExecTxResult{Code: 1})
	assert.NoError(t, err)
	// include empty fields.
	assert.True(t, strings.Contains(string(b), "code"))
	r1 := CheckTxResponse{
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

	var r2 CheckTxResponse
	err = json.Unmarshal(b, &r2)
	assert.Nil(t, err)
	assert.Equal(t, r1, r2)
}

func TestWriteReadMessageSimple(t *testing.T) {
	cases := []proto.Message{
		&EchoRequest{
			Message: "Hello",
		},
	}

	for _, c := range cases {
		buf := new(bytes.Buffer)
		err := WriteMessage(c, buf)
		assert.Nil(t, err)

		msg := new(EchoRequest)
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
		// TODO: add the rest
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
		&CheckTxResponse{
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
		// TODO: add the rest
	}

	for _, c := range cases {
		buf := new(bytes.Buffer)
		err := WriteMessage(c, buf)
		assert.Nil(t, err)

		msg := new(CheckTxResponse)
		err = ReadMessage(buf, msg)
		assert.Nil(t, err)

		assert.True(t, proto.Equal(c, msg))
	}
}
