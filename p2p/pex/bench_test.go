package pex

import (
	"testing"

	"github.com/cometbft/cometbft/v2/p2p/internal/nodekey"
)

func BenchmarkAddrBook_hash(b *testing.B) {
	book := &addrBook{
		ourAddrs:          make(map[string]struct{}),
		privateIDs:        make(map[nodekey.ID]struct{}),
		addrLookup:        make(map[nodekey.ID]*knownAddress),
		badPeers:          make(map[nodekey.ID]*knownAddress),
		filePath:          "",
		routabilityStrict: true,
	}
	book.init()
	msg := []byte(`foobar`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = book.hash(msg)
	}
}
