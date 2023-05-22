package server

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test__nonJSONStringToArg(t *testing.T) {
	s := "unquoted-string"
	v, ok, err := _nonJSONStringToArg(reflect.TypeOf(""), s)
	require.NoError(t, err)
	require.True(t, ok)

	assert.Equal(t, v.String(), s)
}
