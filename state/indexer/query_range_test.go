package indexer

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/libs/pubsub/query/syntax"
)

func TestLookForRangesWithHeightMergesConditionsOnSameKey(t *testing.T) {
	const key = "account.number"

	testCases := []struct {
		query     string
		lower     *big.Float
		lowerIncl bool
		upper     *big.Float
		upperIncl bool
	}{
		{query: "account.number > 5", lower: big.NewFloat(5)},
		{query: "account.number >= 5", lower: big.NewFloat(5), lowerIncl: true},

		// Conditions are ANDed, so duplicates must intersect: the strictest
		// bound wins regardless of the order they appear in.
		{query: "account.number > 5 AND account.number > 1", lower: big.NewFloat(5)},
		{query: "account.number > 1 AND account.number > 5", lower: big.NewFloat(5)},
		{query: "account.number >= 5 AND account.number > 5", lower: big.NewFloat(5)},
		{query: "account.number > 5 AND account.number >= 5", lower: big.NewFloat(5)},
		{query: "account.number >= 1 AND account.number >= 5", lower: big.NewFloat(5), lowerIncl: true},

		{query: "account.number < 10 AND account.number < 20", upper: big.NewFloat(10)},
		{query: "account.number < 20 AND account.number < 10", upper: big.NewFloat(10)},
		{query: "account.number <= 10 AND account.number < 10", upper: big.NewFloat(10)},
		{query: "account.number <= 20 AND account.number <= 10", upper: big.NewFloat(10), upperIncl: true},

		{
			query: "account.number > 1 AND account.number <= 9 AND account.number > 4",
			lower: big.NewFloat(4), upper: big.NewFloat(9), upperIncl: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			conditions, err := syntax.Parse(tc.query)
			require.NoError(t, err)

			ranges, _, _ := LookForRangesWithHeight(conditions)
			qr, ok := ranges[key]
			require.True(t, ok)

			requireBound(t, tc.lower, qr.LowerBound)
			require.Equal(t, tc.lowerIncl, qr.IncludeLowerBound)
			requireBound(t, tc.upper, qr.UpperBound)
			require.Equal(t, tc.upperIncl, qr.IncludeUpperBound)
		})
	}
}

func TestLookForRangesWithHeightMergesHeightConditions(t *testing.T) {
	conditions, err := syntax.Parse("tx.height > 5 AND tx.height > 1 AND tx.height < 100")
	require.NoError(t, err)

	_, _, heightRange := LookForRangesWithHeight(conditions)
	requireBound(t, big.NewFloat(5), heightRange.LowerBound)
	require.False(t, heightRange.IncludeLowerBound)
	requireBound(t, big.NewFloat(100), heightRange.UpperBound)
}

func requireBound(t *testing.T, expected *big.Float, actual any) {
	t.Helper()

	if expected == nil {
		require.Nil(t, actual)
		return
	}

	bound, ok := actual.(*big.Float)
	require.True(t, ok, "expected a *big.Float bound, got %T", actual)
	require.Zerof(t, bound.Cmp(expected), "expected bound %v, got %v", expected, bound)
}
