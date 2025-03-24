package kv

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntInSlice(t *testing.T) {
	assert.True(t, slices.Contains([]int{1, 2, 3}, 1))
	assert.False(t, slices.Contains([]int{1, 2, 3}, 4))
	assert.True(t, slices.Contains([]int{0}, 0))
	assert.False(t, slices.Contains([]int{}, 0))
}
