package blocksync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMovingAverageEmpty(t *testing.T) {
	ma := NewMovingAverage(5)
	_, ok := ma.Avg()
	require.False(t, ok, "empty moving average should return false")
	require.Equal(t, 0, ma.Len())
}

func TestMovingAverageSingleValue(t *testing.T) {
	ma := NewMovingAverage(5)
	avg := ma.Add(10 * time.Second)
	require.Equal(t, 10*time.Second, avg)
	require.Equal(t, 1, ma.Len())

	avg, ok := ma.Avg()
	require.True(t, ok)
	require.Equal(t, 10*time.Second, avg)
}

func TestMovingAverageMultipleValues(t *testing.T) {
	ma := NewMovingAverage(3)
	ma.Add(1 * time.Second)
	ma.Add(2 * time.Second)
	avg := ma.Add(3 * time.Second)
	require.Equal(t, 2*time.Second, avg) // (1+2+3)/3 = 2
	require.Equal(t, 3, ma.Len())
}

func TestMovingAverageRingBuffer(t *testing.T) {
	ma := NewMovingAverage(3)
	ma.Add(1 * time.Second)
	ma.Add(2 * time.Second)
	ma.Add(3 * time.Second) // buffer full: [1,2,3]
	require.Equal(t, 3, ma.Len())

	// Add 4th value — evicts 1.
	avgs := ma.Add(10 * time.Second)      // buffer: [10,2,3]
	require.Equal(t, 5*time.Second, avgs) // (10+2+3)/3 = 5

	// Add 5th value — evicts 2.
	avg := ma.Add(20 * time.Second)       // buffer: [10,20,3]
	require.Equal(t, 11*time.Second, avg) // (10+20+3)/3 = 11
}

func TestMovingAverageReset(t *testing.T) {
	ma := NewMovingAverage(3)
	ma.Add(1 * time.Second)
	ma.Add(2 * time.Second)
	require.Equal(t, 2, ma.Len())

	ma.Reset()
	_, ok := ma.Avg()
	require.False(t, ok, "after reset, avg should return false")
	require.Equal(t, 0, ma.Len())

	avg := ma.Add(5 * time.Second)
	require.Equal(t, 5*time.Second, avg)
}
