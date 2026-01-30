package p2p

import (
	"fmt"
	"net"
	"testing"
)

func BenchmarkPeerSetForEach(b *testing.B) {
	for _, size := range []int{64, 256, 1024} {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			ps := NewPeerSet()
			for i := 0; i < size; i++ {
				p := newMockPeer(net.IP{127, 0, 0, byte(i % 255)})
				if err := ps.Add(p); err != nil {
					b.Fatalf("add peer: %v", err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ps.ForEach(func(Peer) {})
			}
		})
	}
}
