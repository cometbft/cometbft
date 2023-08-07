package errors

import (
	"fmt"
	"reflect"
)

type ErrMsgToProto struct {
	MessageName string
	Err         error
}

func NewErrMsgToProto[T any](msg T, err error) ErrMsgToProto {
	t := reflect.TypeOf(msg)

	// if msg provided is a pointer, get the underlying value
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	messageName := t.Name()
	return ErrMsgToProto{MessageName: messageName, Err: err}
}

func (e ErrMsgToProto) Error() string {
	return fmt.Sprintf("%s to proto error: %s", e.MessageName, e.Err.Error())
}

func (e ErrMsgToProto) Unwrap() error {
	return e.Err
}

type ErrMsgFromProto struct {
	MessageName string
	Err         error
}

func NewErrMsgFromProto[T any](proto T, err error) ErrMsgFromProto {
	t := reflect.TypeOf(proto)

	// if proto provided is a pointer, get the underlying value
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	messageName := t.Name()
	return ErrMsgFromProto{MessageName: messageName, Err: err}
}

func (e ErrMsgFromProto) Error() string {
	return fmt.Sprintf("%s msg from proto error: %s", e.MessageName, e.Err.Error())
}

func (e ErrMsgFromProto) Unwrap() error {
	return e.Err
}
