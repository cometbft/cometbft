package errors

import (
	"fmt"
)

type ErrMsgToProto struct {
	Err         error
	MessageName string
}

func (e ErrMsgToProto) Error() string {
	return fmt.Sprintf("%s to proto error: %s", e.MessageName, e.Err.Error())
}

func (e ErrMsgToProto) Unwrap() error {
	return e.Err
}

type ErrMsgFromProto struct {
	Err         error
	MessageName string
}

func (e ErrMsgFromProto) Error() string {
	return fmt.Sprintf("%s msg from proto error: %s", e.MessageName, e.Err.Error())
}

func (e ErrMsgFromProto) Unwrap() error {
	return e.Err
}
