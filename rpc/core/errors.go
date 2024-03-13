package core

import (
	"errors"
	"fmt"
)

var (
	ErrNegativeHeight          = errors.New("negative height")
	ErrBlockIndexing           = errors.New("block indexing is disabled")
	ErrTxIndexingDisabled      = errors.New("transaction indexing is disabled")
	ErrNoEvidence              = errors.New("no evidence was provided")
	ErrSlowClient              = errors.New("slow client")
	ErrCometBFTExited          = errors.New("cometBFT exited")
	ErrConfirmationNotReceived = errors.New("broadcast confirmation not received")
	ErrTimedOutWaitingForTx    = errors.New("timed out waiting for tx to be included in a block")
	ErrGenesisRespSize         = errors.New("genesis response is too large, please use the genesis_chunked API instead")
	ErrChunkNotInitialized     = errors.New("genesis chunks are not initialized")
	ErrNoChunks                = errors.New("no chunks")
)

type ErrMaxSubscription struct {
	Max int
}

func (e ErrMaxSubscription) Error() string {
	return fmt.Sprintf("maximum number of subscriptions reached: %d", e.Max)
}

type ErrMaxPerClientSubscription struct {
	Max int
}

func (e ErrMaxPerClientSubscription) Error() string {
	return fmt.Sprintf("maximum number of subscriptions per client reached: %d", e.Max)
}

type ErrQueryLength struct {
	length    int
	maxLength int
}

func (e ErrQueryLength) Error() string {
	return fmt.Sprintf("maximum query length exceeded: length %d, max_length %d", e.length, e.maxLength)
}

type ErrValidation struct {
	Source  error
	ValType string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("%s validation failed: %s", e.ValType, e.Source)
}

func (e ErrValidation) Unwrap() error { return e.Source }

type ErrAddEvidence struct {
	Source error
}

func (e ErrAddEvidence) Error() string {
	return fmt.Sprintf("failed to add evidence: %s", e.Source)
}

func (e ErrAddEvidence) Unwrap() error {
	return e.Source
}

type ErrSubCanceled struct {
	Reason string
}

func (e ErrSubCanceled) Error() string {
	return fmt.Sprintf("subscription canceled: (reason: %s)", e.Reason)
}

type ErrSubFailed struct {
	Source error
}

func (e ErrSubFailed) Error() string {
	return fmt.Sprintf("failed to subscribe: %s", e.Source)
}

func (e ErrSubFailed) Unwrap() error {
	return e.Source
}

type ErrTxBroadcast struct {
	Source error
	Reason string
}

func (e ErrTxBroadcast) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("failed to broadcast tx: %v", e.Source)
	}

	return fmt.Sprintf("failed to broadcast tx: %s: %v", e.Reason, e.Source)
}

func (e ErrTxBroadcast) Unwrap() error {
	return e.Source
}

type ErrServiceConfig struct {
	Source error
}

func (e ErrServiceConfig) Error() string {
	return fmt.Sprintf("service configuration error: %s", e.Source)
}

func (e ErrServiceConfig) Unwrap() error { return e.Source }

type ErrInvalidChunkID struct {
	RequestedID int
	MaxID       int
}

func (e ErrInvalidChunkID) Error() string {
	return fmt.Sprintf("invalid chunk ID: length %d but maximum available is %d", e.RequestedID, e.MaxID)
}

type ErrTxNotFound struct {
	Hash any
}

func (e ErrTxNotFound) Error() string {
	return fmt.Sprintf("tx not found: %X", e.Hash)
}

type ErrInvalidOrderBy struct {
	OrderBy string
}

func (e ErrInvalidOrderBy) Error() string {
	return "invalid order_by: maxLength either `asc` or `desc` or an empty value but got " + e.OrderBy
}

type ErrInvalidNodeType struct {
	PeerID   string
	Expected string
	Actual   string
}

func (e ErrInvalidNodeType) Error() string {
	return fmt.Sprintf("peer %s has an invalid node type: maxLength %s but got %s", e.PeerID, e.Expected, e.Actual)
}
