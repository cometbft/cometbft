package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/cometbft/cometbft/test/loadtime/payload"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
)

type Stats struct {
	peers     map[*e2e.Node]int
	bandwidth map[*e2e.Node]map[*e2e.Node]int
	added     map[*e2e.Node]int
	redundant map[*e2e.Node]int
	cpuLoad   map[*e2e.Node]float32
	sent      map[*e2e.Node]int
	latencies map[float64]int
}

func ComputeStats(testnet *e2e.Testnet) (Stats, error) {
	timeout := 1 * time.Second

	peers := map[*e2e.Node]int{}
	bw := map[*e2e.Node]map[*e2e.Node]int{}
	seen := map[*e2e.Node]int{}
	redundant := map[*e2e.Node]int{}
	cpuLoad := map[*e2e.Node]float32{}
	sent := map[*e2e.Node]int{}
	latencies := map[float64]int{}

	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090",
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
	}
	v1api := v1.NewAPI(client)

	for _, n := range testnet.Nodes {

		seen[n] = 0

		if seen[n], err = queryInt(v1api, timeout, "cometbft_mempool_added_txs", n.String(), ""); err != nil {
			return Stats{}, err
		}

		if redundant[n], err = queryInt(v1api, timeout, "cometbft_mempool_already_received_txs", n.String(), ""); err != nil {
			return Stats{}, err
		}

		if peers[n], err = queryInt(v1api, timeout, "cometbft_p2p_peers", n.String(), ""); err != nil {
			return Stats{}, err
		}

		if cpuLoad[n], err = queryFloat(v1api, timeout, "process_cpu_seconds_total", n.String(), ""); err != nil {
			return Stats{}, err
		}

		if sent[n], err = queryInt(v1api, timeout, "cometbft_gossip_sent_txs", n.String(), ""); err != nil {
			return Stats{}, err
		}

		bw[n] = map[*e2e.Node]int{}
		for _, m := range testnet.Nodes {
			if n == m {
				continue
			}
			if bw[n][m], err = queryInt(v1api, timeout, "cometbft_p2p_peer_receive_bytes_total", n.String(), "chID="+"'"+"0x30"+"', "+"peer_id="+"'"+string(m.ID)+"'"); err != nil {
				return Stats{}, err
			}
		}
	}

	if latencies, err = fetchLatencies(testnet); err != nil {
		return Stats{}, err
	}

	return Stats{peers: peers, bandwidth: bw, added: seen, redundant: redundant, cpuLoad: cpuLoad, sent: sent, latencies: latencies}, err
}

func (t *Stats) Output() string {
	jsn, err := json.Marshal(t)
	if err != nil {
		fmt.Errorf("failed to marshal state: %w", err)
		return ""
	}
	return string(jsn)
}

func (t *Stats) String() string {
	return fmt.Sprintf(`bandwidth="%v", added="%v"`, t.bandwidth, t.added)
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

func (t *Stats) TxsAdded(testnet *e2e.Testnet) float32 {
	count := 0
	for _, n := range testnet.Nodes {
		count += t.added[n]
	}
	return float32(count) / float32(len(testnet.Nodes))
}

func (t *Stats) txsSent(testnet *e2e.Testnet) float32 {
	rtotal := float32(0)
	for _, n := range testnet.Nodes {
		rtotal += float32(t.sent[n])
	}
	return rtotal / float32(len(testnet.Nodes))
}

func (t *Stats) Completion(testnet *e2e.Testnet, txsSent int) float32 {
	total := float32(0)
	for _, n := range testnet.Nodes {
		total += float32(t.added[n]) / float32(txsSent)
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
		stotal += float32(t.added[n])
	}
	return rtotal / stotal
}

func (t *Stats) Degree(testnet *e2e.Testnet) float32 {
	rtotal := float32(0)
	for _, n := range testnet.Nodes {
		rtotal += float32(t.peers[n])
	}
	return rtotal / float32(len(testnet.Nodes))
}

func (t *Stats) CPULoad(testnet *e2e.Testnet) float32 {
	count := float32(0)
	for _, n := range testnet.Nodes {
		count += t.cpuLoad[n]
	}
	return count / float32(len(testnet.Nodes))
}

// Latency returns the average latency in sec. (or the number of elapsed blocks when timestamps are logical)
func (t *Stats) Latency() float64 {
	sum := float64(0)
	count := float64(0)
	for k, v := range t.latencies {
		for i := 1; i <= v; i++ {
			sum += k
			count++
		}
	}

	if sum == 0 {
		return 0 // no latency found (by convention)
	}

	return sum / count
}

func queryInt(v1api v1.API, timeout time.Duration, field string, node string, extra string) (int, error) {
	result, err := doQuery(v1api, timeout, field, node, extra)
	if err != nil {
		return 0, err
	}
	if len(result.(model.Vector)) == 0 {
		return 0, nil
	}
	return strconv.Atoi(result.(model.Vector)[0].Value.String())
}

func queryFloat(v1api v1.API, timeout time.Duration, field string, node string, extra string) (float32, error) {
	result, err := doQuery(v1api, timeout, field, node, extra)
	if err != nil {
		return 0, err
	}
	if len(result.(model.Vector)) == 0 {
		return 0, nil
	}
	convert, err := strconv.ParseFloat(result.(model.Vector)[0].Value.String(), 32)
	if err != nil {
		return 0, err
	}
	return float32(convert), nil
}

func doQuery(v1api v1.API, timeout time.Duration, field string, node string, extra string) (model.Value, error) {
	if extra != "" {
		extra = ", " + extra
	}

	q := field +
		"{" +
		"job=" + "'" + node + "'" +
		extra +
		"}"

	if result, _, err := v1api.Query(context.TODO(), q, time.Now(), v1.WithTimeout(timeout)); err == nil {
		return result, nil
	}

	return nil, fmt.Errorf("Query (" + q + ") has failed")

}

func fetchLatencies(testnet *e2e.Testnet) (map[float64]int, error) {
	latencies := map[float64]int{}

	archiveNode := testnet.ArchiveNodes()[0]
	c, err := archiveNode.Client()
	if err != nil {
		return nil, err
	}

	s, err := c.Status(context.TODO())
	if err != nil {
		return nil, err
	}

	to := s.SyncInfo.LatestBlockHeight
	from := testnet.InitialHeight

	for from < to {
		resp, err := c.Block(context.TODO(), &from)
		if err != nil {
			return nil, err
		}

		now := float64(resp.Block.Height)
		if testnet.PhysicalTimestamps {
			now = float64(resp.Block.Header.Time.UTC().UnixNano())
		}

		for _, tx := range resp.Block.Txs {

			if payload, err := payload.FromBytes(tx); err == nil {

				start := float64(payload.Time.Seconds)
				if testnet.PhysicalTimestamps {
					start = float64(payload.Time.Nanos) + float64(payload.Time.Seconds)*math.Pow(10, 9)
				}

				latency := now - start

				if testnet.PhysicalTimestamps {
					latency = latency / math.Pow(10, 9)
				}

				if latency < 0 {
					// should never happen
					logger.Error("skipping invalid latency: ",
						"lat", latency,
						"block", resp.Block.Header.Height)
					continue
				}

				if _, found := latencies[latency]; !found {
					latencies[latency] = 0
				}

				latencies[latency]++

			} else {
				return nil, err
			}
		}
		from++
	}

	return latencies, nil
}
