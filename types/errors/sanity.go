package errors

import "fmt"

type (
	// ErrNegativeField is returned every time some field which should be non-negative turns out negative
	ErrNegativeField struct {
		Field string
	}

	// ErrRequiredField is returned every time a required field is not provided
	ErrRequiredField struct {
		Field string
	}
)

func (e ErrNegativeField) Error() string {
	return fmt.Sprintf("%s can't be negative", e.Field)
}

func (e ErrRequiredField) Error() string {
	return fmt.Sprintf("%s is required", e.Field)
}
