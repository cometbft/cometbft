package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
)

func TestWaitForNodeTimeoutDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name    string
		rpcFail bool
		want    string
	}{
		{name: "RPC unavailable", rpcFail: true, want: "last RPC error:"},
		{name: "node behind", want: "last height 7, catching_up true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.rpcFail {
					http.Error(w, "node unavailable", http.StatusServiceUnavailable)
					return
				}
				var request struct {
					ID json.RawMessage `json:"id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":` + string(request.ID) +
					`,"result":{"sync_info":{"latest_block_height":"7","catching_up":true}}}`))
			}))
			defer server.Close()
			host, port, err := net.SplitHostPort(server.Listener.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			portNumber, err := strconv.Atoi(port)
			if err != nil {
				t.Fatal(err)
			}
			node := &e2e.Node{Name: "validator02", ExternalIP: net.ParseIP(host), ProxyPort: uint32(portNumber)}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			// An expired deadline exercises the first status response without a
			// timing-sensitive sleep or changing the polling interval.
			_, err = waitForNode(ctx, node, 12, -time.Second)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q in timeout error, got %v", tc.want, err)
			}
			if !strings.Contains(err.Error(), "validator02 to reach height 12") {
				t.Fatalf("missing node and target height: %v", err)
			}
		})
	}
}
