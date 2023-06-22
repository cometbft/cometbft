package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"strconv"
	"time"

	e2e "github.com/cometbft/cometbft/test/simulator/pkg"
)

// Mempool is a simple function for fetching, calculating and printing the following mempool metrics:
// - complete : % nodes receiving all txs
// - total bandwidth: sum of all the bandwidth used by at the nodes
// - useful bandwidth: total bandwidth / (#txs * tx_size * #nodes)
// - overhead: (total bandwdith - useful bandwidth) / (useful bandwidth)
func Mempool(ctx context.Context, loadCancel context.CancelFunc, testnet *e2e.Testnet, benchmarkDuration time.Duration) (MempoolStats, error) {
	logger.Info("Beginning benchmark period...")
	startAt := time.Now()

	// wait for the length of the benchmark
	timer := time.NewTimer(benchmarkDuration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return MempoolStats{}, ctx.Err()
	case <-timer.C:
		if time.Since(startAt) < benchmarkDuration {
			return MempoolStats{}, fmt.Errorf("timed out without reason")
		}
	}

	logger.Info("Ending benchmark period")

	loadCancel()

	// grace period
	timer = time.NewTimer(1 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		if time.Since(startAt) < benchmarkDuration {
			return MempoolStats{}, fmt.Errorf("timed out without reason")
		}
	}

	logger.Info("Fetching stats")

	stats, err := fetchStats(testnet)
	if err != nil {
		return MempoolStats{}, nil
	}

	return stats, nil
}

type MempoolStats struct {
	bandwidth map[*e2e.Node]map[*e2e.Node]int
	seen      map[*e2e.Node]int
}

func fetchStats(testnet *e2e.Testnet) (MempoolStats, error) {
	timeout := 1 * time.Second

	bw := map[*e2e.Node]map[*e2e.Node]int{}
	seen := map[*e2e.Node]int{}

	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090",
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
	}
	v1api := v1.NewAPI(client)

	for _, n := range testnet.Nodes {

		seen[n] = 0

		result, _, err := v1api.Query(context.TODO(), "cometbft_mempool_size", time.Now(), v1.WithTimeout(timeout))
		if err != nil {
			fmt.Printf("Error querying Prometheus: %v\n", err)
			return MempoolStats{}, err
		}

		if len(result.(model.Vector)) != 0 {
			seen[n], err = strconv.Atoi(result.(model.Vector)[0].Value.String())
			if err != nil {
				fmt.Printf("Error querying Prometheus: %v\n", err)
				return MempoolStats{}, err
			}
		}

		bw[n] = map[*e2e.Node]int{}
		for _, m := range testnet.Nodes {
			bw[n][m] = 0

			q := "cometbft_p2p_message_receive_bytes_total{job='" + m.String() + "',message_type='mempool_Txs'}"
			result, _, err = v1api.Query(context.TODO(), q, time.Now(), v1.WithTimeout(timeout))
			if err != nil {
				fmt.Printf("Error querying Prometheus: %v\n", err)
				return MempoolStats{}, err
			}

			if len(result.(model.Vector)) != 0 {
				bw[n][m], err = strconv.Atoi(result.(model.Vector)[0].Value.String())
				if err != nil {
					fmt.Printf("Error querying Prometheus: %v\n", err)
					return MempoolStats{}, err
				}
			}
		}
	}
	return MempoolStats{bandwidth: bw, seen: seen}, err
}

func (t *MempoolStats) Output() string {
	jsn, err := json.Marshal(t)
	if err != nil {
		fmt.Errorf("failed to marshal state: %w")
		return ""
	}
	return string(jsn)
}

func (t *MempoolStats) String() string {
	return fmt.Sprintf(`bandwidth="%v", seen="%v"`, t.bandwidth, t.seen)
}

func (t *MempoolStats) TotalBandwidth(testnet *e2e.Testnet) int {
	total := 0
	for _, n := range testnet.Nodes {
		for _, m := range testnet.Nodes {
			total += t.bandwidth[n][m]
		}
	}
	return total
}

func (t *MempoolStats) TxsSeen(testnet *e2e.Testnet) float32 {
	count := 0
	for _, n := range testnet.Nodes {
		count += t.seen[n]
	}
	return float32(count) / float32(len(testnet.Nodes))
}

func (t *MempoolStats) Complete(testnet *e2e.Testnet, txsSent int) float32 {
	count := 0
	for _, n := range testnet.Nodes {
		if t.seen[n] == txsSent {
			count++
		}
	}
	return float32(100) * (float32(count) / float32(len(testnet.Nodes)))
}
