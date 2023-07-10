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

type Stats struct {
	peers     map[*e2e.Node]int
	bandwidth map[*e2e.Node]map[*e2e.Node]int
	seen      map[*e2e.Node]int
	redundant map[*e2e.Node]int
}

func Fetch(testnet *e2e.Testnet) (Stats, error) {
	timeout := 1 * time.Second

	peers := map[*e2e.Node]int{}
	bw := map[*e2e.Node]map[*e2e.Node]int{}
	seen := map[*e2e.Node]int{}
	redundant := map[*e2e.Node]int{}

	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090",
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
	}
	v1api := v1.NewAPI(client)

	for _, n := range testnet.Nodes {

		seen[n] = 0

		if seen[n], err = query(v1api, timeout, "cometbft_mempool_size", n.String(), ""); err != nil {
			return Stats{}, err
		}

		if redundant[n], err = query(v1api, timeout, "cometbft_mempool_already_received_txs", n.String(), ""); err != nil {
			return Stats{}, err
		}

		if peers[n], err = query(v1api, timeout, "cometbft_p2p_peers", n.String(), ""); err != nil {
			return Stats{}, err
		}

		bw[n] = map[*e2e.Node]int{}
		for _, m := range testnet.Nodes {
			if n == m {
				continue
			}
			if bw[n][m], err = query(v1api, timeout, "cometbft_p2p_peer_receive_bytes_total", n.String(), "chID="+"'"+"0x30"+"', "+"peer_id="+"'"+string(m.ID)+"'"); err != nil {
				return Stats{}, err
			}
		}
	}
	return Stats{peers: peers, bandwidth: bw, seen: seen, redundant: redundant}, err
}

func (t *Stats) Output() string {
	jsn, err := json.Marshal(t)
	if err != nil {
		fmt.Errorf("failed to marshal state: %w")
		return ""
	}
	return string(jsn)
}

func (t *Stats) String() string {
	return fmt.Sprintf(`bandwidth="%v", seen="%v"`, t.bandwidth, t.seen)
}

func (t *Stats) TotalBandwidth(testnet *e2e.Testnet) int {
	total := 0
	for _, n := range testnet.Nodes {
		for _, m := range testnet.Nodes {
			total += t.bandwidth[n][m]
		}
	}
	return total
}

func (t *Stats) TxsSeen(testnet *e2e.Testnet) float32 {
	count := 0
	for _, n := range testnet.Nodes {
		count += t.seen[n]
	}
	return float32(count) / float32(len(testnet.Nodes))
}

func (t *Stats) Completion(testnet *e2e.Testnet, txsSent int) float32 {
	total := float32(0)
	for _, n := range testnet.Nodes {
		total += float32(t.seen[n]) / float32(txsSent)
	}
	return float32(100) * (total / float32(len(testnet.Nodes)))
}

func (t *Stats) BandwidthGraph(testnet *e2e.Testnet, computeRatio bool) map[string]map[string]float32 {
	result := map[string]map[string]float32{}
	for i, n := range testnet.Nodes {
		sum := 1
		if computeRatio {
			sum := 0
			for _, m := range testnet.Nodes {
				sum += t.bandwidth[n][m]
			}
		}
		result[strconv.Itoa(i)] = map[string]float32{}
		for j, m := range testnet.Nodes {
			result[strconv.Itoa(i)][strconv.Itoa(j)] = float32(t.bandwidth[n][m]) / float32(sum)
		}
	}
	return result
}

func (t *Stats) Duplicates(testnet *e2e.Testnet) map[string]int {
	result := map[string]int{}
	for i, n := range testnet.Nodes {
		result[strconv.Itoa(i)] = t.redundant[n]
	}
	return result
}

func (t *Stats) Peers(testnet *e2e.Testnet) map[string]int {
	result := map[string]int{}
	for i, n := range testnet.Nodes {
		result[strconv.Itoa(i)] = t.peers[n]
	}
	return result
}

func (t *Stats) Redundancy(testnet *e2e.Testnet) float32 {
	rtotal := float32(0)
	stotal := float32(0)
	for _, n := range testnet.Nodes {
		rtotal += float32(t.redundant[n])
		stotal += float32(t.seen[n])
	}
	return rtotal / stotal
}

func query(v1api v1.API, timeout time.Duration, field string, node string, extra string) (int, error) {

	if extra != "" {
		extra = ", " + extra
	}

	q := field +
		"{" +
		"job=" + "'" + node + "'" +
		extra +
		"}"

	if result, _, err := v1api.Query(context.TODO(), q, time.Now(), v1.WithTimeout(timeout)); err == nil {
		if len(result.(model.Vector)) != 0 {
			return strconv.Atoi(result.(model.Vector)[0].Value.String())
		} else {
			return 0, nil
		}
	}

	return 0, fmt.Errorf("Query (" + q + ") has failed")

}
