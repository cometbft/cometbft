package client

import (
	"context"
	"time"

	"github.com/cometbft/cometbft/v2/types"
)

const WaitThreshold = 10

// Waiter is informed of current height, decided whether to quit early.
type Waiter func(delta int64) (abort error)

// DefaultWaitStrategy is the standard backoff algorithm,
// but you can plug in another one.
func DefaultWaitStrategy(delta int64) (abort error) {
	if delta > WaitThreshold {
		return ErrWaitThreshold{Got: delta, Expected: WaitThreshold}
	} else if delta > 0 {
		// estimate of wait time....
		// wait half a second for the next block (in progress)
		// plus one second for every full block
		delay := time.Duration(delta-1)*time.Second + 500*time.Millisecond
		time.Sleep(delay)
	}
	return nil
}

// Wait for height will poll status at reasonable intervals until
// the block at the given height is available.
//
// If waiter is nil, we use DefaultWaitStrategy, but you can also
// provide your own implementation.
func WaitForHeight(c StatusClient, h int64, waiter Waiter) error {
	if waiter == nil {
		waiter = DefaultWaitStrategy
	}
	delta := int64(1)
	for delta > 0 {
		s, err := c.Status(context.Background())
		if err != nil {
			return err
		}
		// delta might be negative (if h is less than LatestBlockHeight
		// but this should not cause an error when calling the waiter with
		// a negative value
		delta = h - s.SyncInfo.LatestBlockHeight
		// wait for the time, or abort early
		if err := waiter(delta); err != nil {
			return err
		}
	}

	return nil
}

// WaitForOneEvent subscribes to a websocket event for the given
// event time and returns upon receiving it one time, or
// when the timeout duration has expired.
//
// This handles subscribing and unsubscribing under the hood.
func WaitForOneEvent(c EventsClient, evtType string, timeout time.Duration) (types.TMEventData, error) {
	const subscriber = "helpers"
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// register for the next event of this type
	eventCh, err := c.Subscribe(ctx, subscriber, types.QueryForEvent(evtType).String())
	if err != nil {
		return nil, ErrSubscribe{Source: err}
	}
	// make sure to unregister after the test is over
	defer func() {
		if deferErr := c.UnsubscribeAll(ctx, subscriber); deferErr != nil {
			panic(deferErr)
		}
	}()

	select {
	case event := <-eventCh:
		return event.Data, nil
	case <-ctx.Done():
		return nil, ErrEventTimeout
	}
}
