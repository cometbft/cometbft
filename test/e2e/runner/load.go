package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"

	cmtrand "github.com/cometbft/cometbft/internal/rand"
	"github.com/cometbft/cometbft/libs/log"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/loadtime/payload"
	"github.com/cometbft/cometbft/types"
)

const workerPoolSize = 16

// Load generates transactions against the network until the given context is
// canceled.
func Load(ctx context.Context, testnet *e2e.Testnet, useInternalIP bool) error {
	initialTimeout := 1 * time.Minute
	stallTimeout := 30 * time.Second
	chSuccess := make(chan struct{})
	chFailed := make(chan error)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger.Info("load", "msg", log.NewLazySprintf("Starting transaction load (%v workers)...", workerPoolSize),
		"tx/s", testnet.LoadTxBatchSize, "tx-bytes", testnet.LoadTxSizeBytes, "conn", testnet.LoadTxConnections,
		"max-seconds", testnet.LoadMaxSeconds, "target-nodes", testnet.LoadTargetNodes)
	started := time.Now()
	u := [16]byte(uuid.New()) // generate run ID on startup

	// Nodes that will receive load.
	targetNodes := make([]*e2e.Node, 0)
	for _, n := range testnet.Nodes {
		if len(testnet.LoadTargetNodes) == 0 {
			if n.SendNoLoad {
				continue
			}
		} else if !slices.Contains(testnet.LoadTargetNodes, n.Name) {
			continue
		}
		targetNodes = append(targetNodes, n)
	}

	// Create one channel per target node.
	txChs := make([](chan types.Tx), len(targetNodes))
	for i := range targetNodes {
		txChs[i] = make(chan types.Tx)
	}
	go loadGenerate(ctx, txChs, testnet, targetNodes, u[:])

	// Create a loading goroutine per target node and per connection.
	for i, n := range targetNodes {
		for w := 0; w < testnet.LoadTxConnections; w++ {
			go loadProcess(ctx, txChs[i], chSuccess, chFailed, n, useInternalIP)
		}
	}

	maxTimer := time.NewTimer(time.Duration(testnet.LoadMaxSeconds) * time.Second)
	if testnet.LoadMaxSeconds <= 0 {
		<-maxTimer.C
	}

	// Monitor successful and failed transactions, and abort on stalls.
	success, failed := 0, 0
	errorCounter := make(map[string]int)
	timeout := initialTimeout
	for {
		rate := log.NewLazySprintf("%.1f", float64(success)/time.Since(started).Seconds())

		select {
		case <-chSuccess:
			success++
			timeout = stallTimeout
		case err := <-chFailed:
			failed++
			errorCounter[err.Error()]++
		case <-time.After(timeout):
			return fmt.Errorf("unable to submit transactions for %v", timeout)
		case <-maxTimer.C:
			logger.Info("load", "msg", log.NewLazySprintf("Transaction load finished after reaching %v seconds (%v tx/s)", testnet.LoadMaxSeconds, rate))
			return nil
		case <-ctx.Done():
			if success == 0 {
				return errors.New("failed to submit any transactions")
			}
			logger.Info("load", "msg", log.NewLazySprintf("Ending transaction load after %v txs (%v tx/s)...", success, rate))
			return nil
		}

		// Log every ~1 second the number of sent transactions.
		total := success + failed
		if total%testnet.LoadTxBatchSize == 0 {
			successRate := float64(success) / float64(total)
			logger.Debug("load", "success", success, "failed", failed, "success/total", log.NewLazySprintf("%.2f", successRate), "tx/s", rate)
			if len(errorCounter) > 0 {
				for err, c := range errorCounter {
					if c == 1 {
						logger.Error("failed to send transaction", "err", err)
					} else {
						logger.Error("failed to send multiple transactions", "count", c, "err", err)
					}
				}
				errorCounter = make(map[string]int)
			}
		}

		// Check if reached max number of allowed transactions to send.
		if testnet.LoadMaxTxs > 0 && success >= testnet.LoadMaxTxs {
			logger.Info("load", "msg", log.NewLazySprintf("Transaction load finished after reaching %v txs (%v tx/s)", success, rate))
			return nil
		}
	}
}

// loadGenerate generates jobs until the context is canceled.
func loadGenerate(ctx context.Context, txChs []chan types.Tx, testnet *e2e.Testnet, targetNodes []*e2e.Node, id []byte) {
	t := time.NewTimer(0)
	defer t.Stop()
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			for _, ch := range txChs {
				close(ch)
			}
			return
		}
		t.Reset(time.Second)

		// A context with a timeout is created here to time the createTxBatch
		// function out. If createTxBatch has not completed its work by the time
		// the next batch is set to be sent out, then the context is canceled so that
		// the current batch is halted, allowing the next batch to begin.
		tctx, cf := context.WithTimeout(ctx, time.Second)
		createTxBatch(tctx, txChs, testnet, targetNodes, id)
		cf()
	}
}

// createTxBatch creates new transactions and sends them into the txCh. createTxBatch
// returns when either a full batch has been sent to the txCh or the context
// is canceled.
func createTxBatch(ctx context.Context, txChs []chan types.Tx, testnet *e2e.Testnet, targetNodes []*e2e.Node, id []byte) {
	wg := &sync.WaitGroup{}
	genCh := make(chan struct{})
	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range genCh {
				tx, err := payload.NewBytes(&payload.Payload{
					Id:          id,
					Size:        uint64(testnet.LoadTxSizeBytes),
					Rate:        uint64(testnet.LoadTxBatchSize),
					Connections: uint64(testnet.LoadTxConnections),
					Lane:        testnet.WeightedRandomLane(),
				})
				if err != nil {
					panic(fmt.Sprintf("Failed to generate tx: %v", err))
				}

				var nodeIndices []int
				if testnet.LoadNumNodesPerTx <= 1 {
					// Pick one random node to send the transaction.
					nodeIndices = []int{cmtrand.Intn(len(targetNodes))}
				} else {
					// Pick LoadNumNodesPerTx random nodes (channel indices) to
					// send the transaction.
					nodeIndices = cmtrand.Perm(len(targetNodes))[:testnet.LoadNumNodesPerTx]
				}
				for _, i := range nodeIndices {
					select {
					case txChs[i] <- tx:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
FOR_LOOP:
	for i := 0; i < testnet.LoadTxBatchSize; i++ {
		select {
		case genCh <- struct{}{}:
		case <-ctx.Done():
			break FOR_LOOP
		}
	}
	close(genCh)
	wg.Wait()
}

// loadProcess processes transactions by sending transactions received on the txCh
// to the client.
func loadProcess(ctx context.Context, txCh <-chan types.Tx, chSuccess chan<- struct{}, chFailed chan<- error, n *e2e.Node, useInternalIP bool) {
	var client *rpchttp.HTTP
	var err error
	s := struct{}{}
	for tx := range txCh {
		if client == nil {
			if useInternalIP {
				client, err = n.ClientInternalIP()
			} else {
				client, err = n.Client()
			}
			if err != nil {
				logger.Info("non-fatal error creating node client", "error", err)
				continue
			}
		}
		if _, err = client.BroadcastTxSync(ctx, tx); err != nil {
			chFailed <- err
			continue
		}
		chSuccess <- s
	}
}
