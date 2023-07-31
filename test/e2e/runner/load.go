package main

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/loadtime/payload"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/google/uuid"
)

const workerPoolSize = 16

// Load generates transactions against the network until the given context is
// canceled.
func Load(ctx context.Context, testnet *e2e.Testnet) (int, error) {
	initialTimeout := 1 * time.Minute
	stallTimeout := 30 * time.Second
	chSuccess := make(chan struct{})
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger.Info("load", "msg", log.NewLazySprintf("Starting transaction load (%v workers)...", workerPoolSize))
	started := time.Now()
	u := [16]byte(uuid.New()) // generate run ID on startup

	txCh := make(chan payload.Payload)
	go loadGenerate(ctx, txCh, testnet, u[:])

	for _, n := range testnet.Nodes {
		if n.SendNoLoad {
			continue
		}

		for w := 0; w < testnet.LoadTxConnections; w++ {
			go loadProcess(ctx, txCh, chSuccess, n, testnet)
		}
	}

	// Monitor successful transactions, and abort on stalls.
	success := 0
	timeout := initialTimeout
	for {
		select {
		case <-chSuccess:
			success++
			timeout = stallTimeout
			if success >= testnet.LoadTxToSend {
				logger.Info("load", "msg", log.NewLazySprintf("Ending transaction load after %v txs (%.1f tx/s)...",
					success, float64(success)/time.Since(started).Seconds()))
				return success, nil
			}
		case <-time.After(timeout):
			return 0, fmt.Errorf("unable to submit transactions for %v", timeout)
		case <-ctx.Done():
			if success == 0 {
				return 0, errors.New("failed to submit any transactions")
			}
			logger.Info("load", "msg", log.NewLazySprintf("Ending transaction load after %v txs (%.1f tx/s)...",
				success, float64(success)/time.Since(started).Seconds()))
			return success, nil
		}
	}
}

// loadGenerate generates jobs until the context is canceled or the target is attained
func loadGenerate(ctx context.Context, txCh chan<- payload.Payload, testnet *e2e.Testnet, id []byte) {
	t := time.NewTimer(0)
	defer t.Stop()
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			close(txCh)
			return
		}
		t.Reset(time.Second)

		// A context with a timeout is created here to time the createTxBatch
		// function out. If createTxBatch has not completed its work by the time
		// the next batch is set to be sent out, then the context is canceled so that
		// the current batch is halted, allowing the next batch to begin.
		tctx, cf := context.WithTimeout(ctx, time.Second)
		createTxBatch(tctx, txCh, testnet, id)
		cf()
	}
}

// createTxBatch creates new transactions and sends them into the txCh. createTxBatch
// returns when either a full batch has been sent to the txCh or the context
// is canceled.
func createTxBatch(ctx context.Context, txCh chan<- payload.Payload, testnet *e2e.Testnet, id []byte) {
	wg := &sync.WaitGroup{}
	genCh := make(chan struct{})
	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range genCh {
				tx := payload.Payload{
					Id:          id,
					Size:        uint64(testnet.LoadTxSizeBytes),
					Rate:        uint64(testnet.LoadTxBatchSize),
					Connections: uint64(testnet.LoadTxConnections),
				}
				select {
				case txCh <- tx:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	for i := 0; i < testnet.LoadTxBatchSize; i++ {
		select {
		case genCh <- struct{}{}:
		case <-ctx.Done():
			break
		}
	}
	close(genCh)
	wg.Wait()
}

// loadProcess loops over txCh, sending each transaction to the corresponding client.
func loadProcess(ctx context.Context, txCh <-chan payload.Payload, chSuccess chan<- struct{}, n *e2e.Node, testnet *e2e.Testnet) {
	var client *rpchttp.HTTP
	var err error
	s := struct{}{}
	i := 0
	for tx := range txCh {
		if client == nil {
			client, err = n.Client()
			if err != nil {
				logger.Info("non-fatal error creating node client", "error", err)
				continue
			}
		}

		if info, err := client.Status(context.TODO()); err == nil {
			if !testnet.PhysicalTimestamps {
				tx.Time = &timestamppb.Timestamp{
					Seconds: info.SyncInfo.LatestBlockHeight,
					Nanos:   0}
			} else {
				time := cmttime.Canonical(info.SyncInfo.LatestBlockTime)
				tx.Time = &timestamppb.Timestamp{
					Seconds: time.Unix(),
					Nanos:   int32(time.Nanosecond()),
				}
			}
		} else {
			logger.Info("non-fatal error fetching sync info", "error", err)
			continue
		}

		marshalled, _ := payload.NewBytes(&tx)

		if _, err = client.BroadcastTxSync(ctx, marshalled); err != nil {
			continue
		}
		chSuccess <- s
		i++
	}
}
