package client

import "fmt"

type ErrInvalidAddress struct {
	Addr   string
	Source error
}

func (e ErrInvalidAddress) Error() string {
	return fmt.Sprintf("invalid address: %s: %v ", e.Addr, e.Source)
}

func (e ErrInvalidAddress) Unwrap() error {
	return e.Source
}

type ErrEncodingParams struct {
	Source error
}

func (e ErrEncodingParams) Error() string {
	return fmt.Sprintf("failed to encode params: %v", e.Source)
}

func (e ErrEncodingParams) Unwrap() error {
	return e.Source
}

type ErrMarshalRequest struct {
	Source error
}

func (e ErrMarshalRequest) Error() string {
	return fmt.Sprintf("failed to marshal request: %v", e.Source)
}

func (e ErrMarshalRequest) Unwrap() error {
	return e.Source
}

type ErrCreateRequest struct {
	Source error
}

func (e ErrCreateRequest) Error() string {
	return fmt.Sprintf("failed to create request: %v", e.Source)
}

func (e ErrCreateRequest) Unwrap() error {
	return e.Source
}

type ErrFailedRequest struct {
	Source error
}

func (e ErrFailedRequest) Error() string {
	return fmt.Sprintf("failed request: %v", e.Source)
}

func (e ErrFailedRequest) Unwrap() error {
	return e.Source
}

type ErrReadResponse struct {
	Source      error
	Description string
}

func (e ErrReadResponse) Error() string {
	if e.Description == "" {
		return fmt.Sprintf("failed to read response: %s : %v", e.Source.Error(), e.Source)
	}

	return fmt.Sprintf("failed to read response: %s : %v", e.Description, e.Source)
}

func (e ErrReadResponse) Unwrap() error {
	return e.Source
}

type ErrUnmarshalResponse struct {
	Source      error
	Description string
}

func (e ErrUnmarshalResponse) Error() string {
	if e.Description == "" {
		return fmt.Sprintf("failed to unmarshal response: %s : %v", e.Source.Error(), e.Source)
	}

	return fmt.Sprintf("failed to unmarshal response: %s : %v", e.Description, e.Source)
}
