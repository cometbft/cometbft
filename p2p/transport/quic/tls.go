package quic

import (
	"crypto/tls"
	"fmt"
)

// CreateTLSConfig creates a tls.Config using provided certificate and key files.
func CreateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("tls.LoadX509KeyPair: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert}, // Use the loaded certificate
		NextProtos:   []string{"h3"},          // QUIC requires ALPN; "h3" for HTTP/3
	}, nil
}
