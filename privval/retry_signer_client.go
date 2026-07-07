package privval

import (
	"fmt"
	"sync"
	"time"

	"github.com/cometbft/cometbft/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"
)

// RetrySignerClient wraps SignerClient adding retry for each operation (except
// Ping) w/ a timeout.
type RetrySignerClient struct {
	next    *SignerClient
	retries int
	timeout time.Duration

	quitOnce sync.Once
	quit     chan struct{} // closed by Close to abort in-flight retry loops
}

// NewRetrySignerClient returns RetrySignerClient. If +retries+ is 0, the
// client will be retrying each operation indefinitely.
func NewRetrySignerClient(sc *SignerClient, retries int, timeout time.Duration) *RetrySignerClient {
	return &RetrySignerClient{next: sc, retries: retries, timeout: timeout, quit: make(chan struct{})}
}

var _ types.PrivValidator = (*RetrySignerClient)(nil)

// Close aborts any in-flight retry loop and closes the underlying client. Safe
// to call more than once. The abort unblocks node shutdown, which waits on the
// consensus receiveRoutine that signs synchronously.
func (sc *RetrySignerClient) Close() error {
	sc.quitOnce.Do(func() { close(sc.quit) })
	return sc.next.Close()
}

// sleep waits for timeout, returning false if Close was called first.
func (sc *RetrySignerClient) sleep() bool {
	select {
	case <-sc.quit:
		return false
	case <-time.After(sc.timeout):
		return true
	}
}

func (sc *RetrySignerClient) IsConnected() bool {
	return sc.next.IsConnected()
}

func (sc *RetrySignerClient) WaitForConnection(maxWait time.Duration) error {
	return sc.next.WaitForConnection(maxWait)
}

//--------------------------------------------------------
// Implement PrivValidator

func (sc *RetrySignerClient) Ping() error {
	return sc.next.Ping()
}

func (sc *RetrySignerClient) GetPubKey() (crypto.PubKey, error) {
	var (
		pk  crypto.PubKey
		err error
	)
	for i := 0; i < sc.retries || sc.retries == 0; i++ {
		pk, err = sc.next.GetPubKey()
		if err == nil {
			return pk, nil
		}
		// If remote signer errors, we don't retry.
		if _, ok := err.(*RemoteSignerError); ok {
			return nil, err
		}
		if !sc.sleep() {
			return nil, fmt.Errorf("aborted getting pubkey: %w", err)
		}
	}
	return nil, fmt.Errorf("exhausted all attempts to get pubkey: %w", err)
}

func (sc *RetrySignerClient) SignVote(chainID string, vote *cmtproto.Vote) error {
	var err error
	for i := 0; i < sc.retries || sc.retries == 0; i++ {
		err = sc.next.SignVote(chainID, vote)
		if err == nil {
			return nil
		}
		// If remote signer errors, we don't retry.
		if _, ok := err.(*RemoteSignerError); ok {
			return err
		}
		if !sc.sleep() {
			return fmt.Errorf("aborted signing vote: %w", err)
		}
	}
	return fmt.Errorf("exhausted all attempts to sign vote: %w", err)
}

func (sc *RetrySignerClient) SignProposal(chainID string, proposal *cmtproto.Proposal) error {
	var err error
	for i := 0; i < sc.retries || sc.retries == 0; i++ {
		err = sc.next.SignProposal(chainID, proposal)
		if err == nil {
			return nil
		}
		// If remote signer errors, we don't retry.
		if _, ok := err.(*RemoteSignerError); ok {
			return err
		}
		if !sc.sleep() {
			return fmt.Errorf("aborted signing proposal: %w", err)
		}
	}
	return fmt.Errorf("exhausted all attempts to sign proposal: %w", err)
}
