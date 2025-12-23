package statesync_test

import (
	"fmt"
	"math"
	"testing"
	"unsafe"
)

const (
	Chunks = math.MaxUint32
)

func TestMemory(t *testing.T) {
	chunkFiles := make(map[uint32]string, Chunks)

	fmt.Printf("size of chunk files: %d\n", unsafe.Sizeof(chunkFiles))
	// fmt.Printf("size of chunk senders: %d\n", unsafe.Sizeof(chunkSenders))
	// fmt.Printf("size of chunks allocated: %d\n", unsafe.Sizeof(chunkAllocated))
	// fmt.Printf("size of chunks returned: %d\n", unsafe.Sizeof(chunkReturned))
	fmt.Println("done")
}
