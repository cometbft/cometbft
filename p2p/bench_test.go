package p2p

import (
	"testing"
)

var sink any = nil

func BenchmarkPeerSetRemoveOne(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testPeerSetAddRemoveOne(b)
		sink = i
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}

	// Reset the sink.
	sink = nil
}

func BenchmarkPeerSetRemoveMany(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testPeerSetAddRemoveMany(b)
		sink = i
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}

	// Reset the sink.
	sink = nil
}
