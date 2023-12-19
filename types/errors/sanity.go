package errors

import "fmt"

type (
	// ErrNegativeField is returned every time some field which should be non-negative turns out negative.
	ErrNegativeField struct {
		Field string
	}

	// ErrRequiredField is returned every time a required field is not provided.
	ErrRequiredField struct {
		Field string
	}

	// ErrInvalidField is returned every time a value does not pass a validity check.
	ErrInvalidField struct {
		Field  string
		Reason string
	}

	// ErrWrongField is returned every time a value does not pass a validaty check, accompanied with error.
	ErrWrongField struct {
		Field string
		Err   error
	}
)

func (e ErrNegativeField) Error() string {
	return fmt.Sprintf("%s can't be negative", e.Field)
}

func (e ErrRequiredField) Error() string {
	return fmt.Sprintf("%s is required", e.Field)
}

func (e ErrInvalidField) Error() string {
	return fmt.Sprintf("invalid field %s %s", e.Field, e.Reason)
}

func (e ErrWrongField) Error() string {
	return fmt.Sprintf("wrong %s: %v", e.Field, e.Err)
}
