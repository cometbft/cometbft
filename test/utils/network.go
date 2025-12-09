package utils

import (
	"net"
	"testing"
)

// GetFreePorts returns n free ports
func GetFreePorts(t *testing.T, n int) []int {
	var (
		ports     = make([]int, 0, n)
		listeners = make([]net.Listener, 0, n)
	)

	for i := 0; i < n; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			t.Fatalf("unable to resolve port: %v", err)
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			t.Fatalf("unable to listen to tcp: %v", err)
		}

		port := l.Addr().(*net.TCPAddr).Port
		ports = append(ports, port)

		// keep the listener open to avoid port collisions
		listeners = append(listeners, l)
	}

	// close the listeners to free the ports
	for _, l := range listeners {
		if err := l.Close(); err != nil {
			t.Fatalf("unable to close listener: %v", err)
		}
	}

	return ports
}
