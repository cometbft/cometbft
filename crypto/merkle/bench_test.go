package merkle

import (
	"crypto/sha256"
	"strings"
	"testing"
)

var sink any

type innerHashTest struct {
	left, right string
}

var innerHashTests = []*innerHashTest{
	{"aaaaaaaaaaaaaaa", "                    "},
	{"", ""},
	{"                        ", "a    ff     b    f1    a"},
	{"ffff122fff", "ffff122fff"},
	{"😎💡✅alalalalalalalalalallalallaallalaallalalalalalalalaallalalalalalala", "😎💡✅alalalalalalalalalallalallaallalaallalalalalalalalaallalalalalalalaffff122fff"},
	{strings.Repeat("ff", 1<<10), strings.Repeat("00af", 4<<10)},
	{strings.Repeat("f", sha256.Size), strings.Repeat("00af", 10<<10)},
	{"aaaaaaaaaaaaaaaaaaaaaaaaaaaffff122fffaaaaaaaaa", "aaaaaaaaaffff1aaaaaaaaaaaaaaaaaa22fffaaaaaaaaa"},
}

func BenchmarkInnerHash(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, tt := range innerHashTests {
			got := innerHash([]byte(tt.left), []byte(tt.right))
			if g, w := len(got), sha256.Size; g != w {
				b.Fatalf("size discrepancy: got %d, want %d", g, w)
			}
			sink = got
		}
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}
}
