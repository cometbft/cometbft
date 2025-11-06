package priorityqueue

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContainer(t *testing.T) {
	type input struct {
		str      string
		priority uint64
	}
	for _, tt := range []struct {
		name           string
		maxHeap        bool
		input          []input
		expectedOutput []string
	}{
		{
			name:    "max-heap",
			maxHeap: true,
			input: []input{
				{str: "NOT IMPORTANT", priority: 1},
				{str: "MEDIUM", priority: 4},
				{str: "ULTRA", priority: 10},
				{str: "NOT IMPORTANT", priority: 1},
				{str: "SEVERE", priority: 9},
			},
			expectedOutput: []string{
				"ULTRA", "SEVERE", "MEDIUM", "NOT IMPORTANT", "NOT IMPORTANT",
			},
		},
		{
			name: "min-heap",
			input: []input{
				{str: "NOT IMPORTANT", priority: 1},
				{str: "MEDIUM", priority: 4},
				{str: "ULTRA", priority: 10},
				{str: "NOT IMPORTANT", priority: 1},
				{str: "SEVERE", priority: 9},
			},
			expectedOutput: []string{
				"NOT IMPORTANT", "NOT IMPORTANT", "MEDIUM", "SEVERE", "ULTRA",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			c := newContainer(tt.maxHeap)

			// ACT
			for _, item := range tt.input {
				c.PushItem(item.str, item.priority)
			}

			actualOutput := []string{}

			for {
				v, ok := c.PopItem()
				if !ok {
					break
				}

				actualOutput = append(actualOutput, v.(string))
			}

			// ASSERT
			require.Equal(t, tt.expectedOutput, actualOutput)
		})
	}
}
