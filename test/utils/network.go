package utils

import (
	"net"
	"testing"
)

// GetFreePorts returns n free ports
func GetFreePorts(t *testing.T, n int) []int {
	ports := make([]int, 0, n)

	for i := 0; i < n; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			t.Fatalf("unable to resolve port: %v", err)
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			t.Fatalf("unable to listen to tcp: %v", err)
		}

		// This is done on purpose - we want to keep ports
		// busy to avoid collisions when getting the next one
		defer func() { _ = l.Close() }()

		port := l.Addr().(*net.TCPAddr).Port
		ports = append(ports, port)
	}

	return ports
}
