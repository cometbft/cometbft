package lp2p

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	t.Run("TransientErrorFromAny", func(t *testing.T) {
		innerErr := errors.New("connection reset")
		transientErr := &ErrorTransient{Err: innerErr}

		for _, tt := range []struct {
			name    string
			input   any
			wantOK  bool
			wantErr error
		}{
			{
				name:   "non-error value returns false",
				input:  "not an error",
				wantOK: false,
			},
			{
				name:   "regular error returns false",
				input:  errors.New("regular error"),
				wantOK: false,
			},
			{
				name:    "ErrorTransient returns true",
				input:   transientErr,
				wantOK:  true,
				wantErr: innerErr,
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				// ACT
				got, ok := TransientErrorFromAny(tt.input)

				// ASSERT
				require.Equal(t, tt.wantOK, ok)
				if !tt.wantOK {
					assert.Nil(t, got)
					return
				}

				require.NotNil(t, got)
				assert.ErrorIs(t, got.Err, tt.wantErr)
			})
		}
	})
}
