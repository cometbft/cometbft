package json_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/libs/json"
)

func TestUnmarshalSurroundingWhitespace(t *testing.T) {
	for name, value := range map[string]any{
		"int64":  int64(42),
		"uint64": uint64(42),
		"time":   time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC),
		"bool":   true,
		"bytes":  []byte{1, 2, 3},
	} {
		t.Run(name, func(t *testing.T) {
			encoded, err := json.Marshal(value)
			require.NoError(t, err)
			encoded = append(append([]byte(" \t\r\n"), encoded...), '\n', '\r', '\t', ' ')
			target := reflect.New(reflect.TypeOf(value))
			require.NoError(t, json.Unmarshal(encoded, target.Interface()))
			assert.Equal(t, value, target.Elem().Interface())
		})
	}

	t.Run("null pointer", func(t *testing.T) {
		value := int64(42)
		target := &value
		require.NoError(t, json.Unmarshal([]byte(" \t\r\nnull\n\r\t "), &target))
		assert.Nil(t, target)
	})

	t.Run("null map", func(t *testing.T) {
		target := map[string]int64{"key": 42}
		require.NoError(t, json.Unmarshal([]byte(" \t\r\nnull\n\r\t "), &target))
		assert.Nil(t, target)
	})
}

func TestUnmarshalInvalidWhitespace(t *testing.T) {
	for name, input := range map[string]string{
		"empty":              "",
		"whitespace only":    " \t\r\n",
		"nonbreaking space":  "\u00a0\"42\"",
		"vertical tab":       "\v\"42\"",
		"form feed":          "\"42\"\f",
		"unquoted integer":   " 42 ",
		"multiple documents": " \"42\" \"43\" ",
	} {
		t.Run(name, func(t *testing.T) {
			var target int64
			require.Error(t, json.Unmarshal([]byte(input), &target))
		})
	}
}
