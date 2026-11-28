package json_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/libs/json"
)

func TestUnmarshalStringKeyTypes(t *testing.T) {
	type key string
	type namedMap map[key]int64

	for name, value := range map[string]any{
		"string": map[string]int64{"key": 7},
		"defined string": map[key]int64{
			"key": 7, "quoted\"key": 8, "κλειδί": 9,
		},
		"defined map": namedMap{"key": 7},
		"nested":      map[key]map[key]int64{"outer": {"inner": 7}},
		"empty":       map[key]int64{},
	} {
		t.Run(name, func(t *testing.T) {
			encoded, err := json.Marshal(value)
			require.NoError(t, err)
			target := reflect.New(reflect.TypeOf(value))
			require.NotPanics(t, func() {
				err = json.Unmarshal(encoded, target.Interface())
			})
			require.NoError(t, err)
			assert.Equal(t, value, target.Elem().Interface())
		})
	}

	t.Run("invalid value", func(t *testing.T) {
		var target map[key]int64
		require.Error(t, json.Unmarshal([]byte(`{"key":7}`), &target))
	})
}
