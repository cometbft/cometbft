package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/cometbft/cometbft/libs/log"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/loadtime/payload"
	"github.com/cometbft/cometbft/types"
)

const workerPoolSize = 16

// Load generates transactions against the network until the given context is
// canceled.
func Load(ctx context.Context, loads []*e2e.Load) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, load := range loads {
		wg := &sync.WaitGroup{}
		logger.Info("load", "step", log.NewLazySprintf("Starting transaction load #%v", i), "workers", workerPoolSize)

		if load.WaitToStart > 0 {
			waitTostart := time.Duration(load.WaitToStart) * time.Second
			logger.Info("load", "step", log.NewLazySprintf("Waiting %v before starting load instance", waitTostart))
			time.Sleep(waitTostart)
		}

		for runName, run := range load.Runs {
			wg.Add(1)
			go func(runName string, run *e2e.LoadRun) {
				defer wg.Done()
				runID := [16]byte(uuid.New()) // generate run ID on startup
				if err := loadRun(ctx, runName, runID[:], run); err != nil {
					logger.Error("load", "err", err)
				}
			}(runName, run)
		}
		wg.Wait()

		if load.WaitUntil == e2e.LoadConditionMempoolsAreEmpty {
			waitMempoolsAreEmpty(ctx, load.Testnet.Nodes)
		}

		if load.WaitAtEnd > 0 {
			waitToFinish := time.Duration(load.WaitAtEnd) * time.Second
			logger.Info("load", "step", log.NewLazySprintf("Waiting %s to finish load instance", waitToFinish))
			time.Sleep(waitToFinish)
		}
		logger.Info("load", "step", "Finished transaction load", "#load", i)
	}
	return nil
}

func loadRun(ctx context.Context, runName string, runID []byte, run *e2e.LoadRun) error {
	logger := logger.With("run", runName)

	if run.WaitToRun > 0 {
		waitToRun := time.Duration(run.WaitToRun) * time.Second
		logger.Info("load", "step", log.NewLazySprintf("Waiting %v to start load run instance", waitToRun))
		time.Sleep(waitToRun)
	}

	logger.Info("load", "step", "Start",
		"tx_size", run.TxBytes, "batch_size", run.BatchSize, "connections", run.Connections,
		"max_duration", run.MaxDuration, "max_txs", run.MaxTxs)

	initialTimeout := 1 * time.Minute
	stallTimeout := 30 * time.Second
	chSuccess := make(chan struct{})
	chFailed := make(chan struct{})

	// Spawn the transaction generation routine.
	txCh := make(chan types.Tx)
	go loadGenerate(ctx, txCh, run, runID)

	// Spawn one load routine per node, per connection.
	started := time.Now()
	for _, n := range run.TargetNodes {
		if n.SendNoLoad {
			continue
		}

		for w := 0; w < run.Connections; w++ {
			go loadProcess(ctx, txCh, chSuccess, chFailed, n)
		}
	}

	var maxDurationCh <-chan time.Time
	if run.MaxDuration > 0 {
		maxDurationCh = time.After(time.Duration(run.MaxDuration) * time.Second)
	}

	// Monitor successful and failed transactions, and abort on stalls.
	success, failed := 0, 0
	timeout := initialTimeout
	for {
		rate := log.NewLazySprintf("%.1f", float64(success)/time.Since(started).Seconds())

		select {
		case <-chSuccess:
			success++
			timeout = stallTimeout
		case <-chFailed:
			failed++
		case <-time.After(timeout):
			return fmt.Errorf("unable to submit transactions for %v", timeout)
		case <-maxDurationCh:
			logger.Info("load", "step", log.NewLazySprintf("Finished after reaching %v seconds", run.MaxDuration), "success", success, "tx/s", rate)
			return nil
		case <-ctx.Done():
			if success == 0 {
				return errors.New("failed to submit any transactions")
			}
			logger.Info("load", "step", "Finished transaction load", "success", success, "tx/s", rate)
			return nil
		}

		// Log every ~1 second the number of sent transactions.
		total := success + failed
		if total%run.BatchSize == 0 {
			succcessRate := log.NewLazySprintf("%.1f", float64(success)/float64(total))
			logger.Debug("load", "success", success, "failed", failed, "success/total", succcessRate, "tx/s", rate)
		}

		// Check if reached max number of allowed transactions to send.
		if run.MaxTxs > 0 && success >= run.MaxTxs {
			logger.Info("load", "step", log.NewLazySprintf("Finished after sending %v txs", success), "tx/s", rate)
			return nil
		}
	}
}

// loadGenerate generates jobs until the context is canceled.
func loadGenerate(ctx context.Context, txCh chan<- types.Tx, run *e2e.LoadRun, id []byte) {
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
		createTxBatch(tctx, txCh, run, id)
		cf()
	}
}

// createTxBatch creates new transactions and sends them into the txCh. createTxBatch
// returns when either a full batch has been sent to the txCh or the context
// is canceled.
func createTxBatch(ctx context.Context, txCh chan<- types.Tx, run *e2e.LoadRun, id []byte) {
	wg := &sync.WaitGroup{}
	genCh := make(chan struct{})
	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range genCh {
				tx, err := payload.NewBytes(&payload.Payload{
					Id:          id,
					Size:        uint64(run.TxBytes),
					Rate:        uint64(run.BatchSize),
					Connections: uint64(run.Connections),
				})
				if err != nil {
					panic(fmt.Sprintf("Failed to generate tx: %v", err))
				}

				select {
				case txCh <- tx:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
LOOP:
	for i := 0; i < run.BatchSize; i++ {
		select {
		case genCh <- struct{}{}:
		case <-ctx.Done():
			break LOOP
		}
	}
	close(genCh)
	wg.Wait()
}

// loadProcess processes transactions by sending transactions received on the txCh
// to the client.
func loadProcess(ctx context.Context, txCh <-chan types.Tx, chSuccess chan<- struct{}, chFailed chan<- struct{}, n *e2e.Node) {
	var client *rpchttp.HTTP
	var err error
	s := struct{}{}
	for tx := range txCh {
		if client == nil {
			client, err = n.Client()
			if err != nil {
				logger.Info("non-fatal error creating node client", "error", err)
				continue
			}
		}
		if _, err = client.BroadcastTxSync(ctx, tx); err != nil {
			logger.Error("failed to send transaction", "err", err)
			chFailed <- s
			continue
		}
		chSuccess <- s
	}
}

// waitMempoolsAreEmpty will block until the mempools of all nodes are empty.
func waitMempoolsAreEmpty(ctx context.Context, nodes []*e2e.Node) {
	logger.Info("load", "step", "Wait until mempools are empty")
	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(node *e2e.Node) {
			defer wg.Done()
			for {
				isEmpty, err := mempoolIsEmpty(ctx, node)
				if err != nil || isEmpty {
					return
				}
				time.Sleep(5 * time.Second)
			}
		}(node)
	}
	wg.Wait()
}

func mempoolIsEmpty(ctx context.Context, node *e2e.Node) (bool, error) {
	client, err := node.Client()
	if err != nil {
		logger.Error("non-fatal error creating node client", "error", err)
		return false, err
	}
	limit := 1
	res, err := client.UnconfirmedTxs(ctx, &limit)
	if err != nil {
		logger.Error("failed to request unconfirmed txs", "err", err)
		return false, err
	}
	return res.Count == 0, nil
}
