package main

import "fmt"

type ErrHexDecoding struct {
	err error
}

func (e ErrHexDecoding) Error() string {
	return fmt.Sprintf("Error while hex decoding %s", e.err.Error())
}

type ErrInvalidString struct {
	str     string
	comment string
}

func (e ErrInvalidString) Error() string {
	return fmt.Sprintf("Invalid string: \"%s\". %s", e.str, e.comment)
}
