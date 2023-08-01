package types

import fmt "fmt"

type ErrReadingMessage struct {
	Err error
}

func (e ErrReadingMessage) Error() string {
	return fmt.Sprintf("Error reading message %e", e.Err)
}

type ErrWritingMessage struct {
	Err error
}

func (e ErrWritingMessage) Error() string {
	return fmt.Sprintf("Error writing message %s", e.Err.Error())
}
