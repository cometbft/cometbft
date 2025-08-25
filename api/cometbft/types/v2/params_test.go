package v2

import (
    "math/rand"
    "testing"
    "github.com/cosmos/gogoproto/proto"
)

// testRand implements the randyParams interface for deterministic random number generation.
type testRand struct {
    forceMaxFieldNumber bool // For overflow test
}

func (r testRand) Intn(n int) int {
    if r.forceMaxFieldNumber {
        return n // Return max value for overflow test
    }
    return rand.New(rand.NewSource(1)).Intn(n) + 1 // Ensure non-zero
}

func (testRand) Float32() float32 {
    return rand.New(rand.NewSource(1)).Float32()
}

func (testRand) Float64() float64 {
    return rand.New(rand.NewSource(1)).Float64()
}

func (testRand) Int31() int32 {
    return rand.New(rand.NewSource(1)).Int31()
}

func (testRand) Int63() int64 {
    return rand.New(rand.NewSource(1)).Int63()
}

func (testRand) Uint32() uint32 {
    return rand.New(rand.NewSource(1)).Uint32()
}

func (testRand) Uint64() uint64 {
    return rand.New(rand.NewSource(1)).Uint64()
}

// TestSafeRandUnrecognizedParamsNoOverflow verifies that SafeRandUnrecognizedParams
// generates field numbers within the valid Protocol Buffers range [1, 536870911]
// and prevents uint32 overflow in the key calculation.
func TestSafeRandUnrecognizedParamsNoOverflow(t *testing.T) {
    r := testRand{}
    maxFieldNumber := 536870911
    dAtA, err := SafeRandUnrecognizedParams(r, maxFieldNumber)
    if err != nil {
        t.Fatalf("SafeRandUnrecognizedParams failed: %v", err)
    }
    if len(dAtA) == 0 {
        t.Fatal("SafeRandUnrecognizedParams returned empty data")
    }
    for len(dAtA) > 0 {
        key, n := proto.DecodeVarint(dAtA)
        if n == 0 {
            t.Fatalf("proto.DecodeVarint returned n=0, invalid varint at dAtA=%v", dAtA)
        }
        fieldNumber := int(key >> 3)
        if fieldNumber > 536870911 || fieldNumber < 1 {
            t.Errorf("Invalid fieldNumber detected: %d; expected range [1, 536870911]", fieldNumber)
        }
        dAtA = dAtA[n:]
    }
}

// TestSafeRandUnrecognizedParamsInvalidInput verifies that SafeRandUnrecognizedParams
// rejects maxFieldNumber values that would cause overflow.
func TestSafeRandUnrecognizedParamsInvalidInput(t *testing.T) {
    r := testRand{}
    maxFieldNumber := 536870912 // Exceeds maximum
    _, err := SafeRandUnrecognizedParams(r, maxFieldNumber)
    if err == nil {
        t.Errorf("Expected error for maxFieldNumber %d, got nil", maxFieldNumber)
    }
    expectedErr := "maxFieldNumber 536870912 exceeds maximum (536870911)"
    if err.Error() != expectedErr {
        t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
    }
}

// TestRandUnrecognizedParamsOverflow reproduces the original overflow issue for reference.
func TestRandUnrecognizedParamsOverflow(t *testing.T) {
    r := testRand{forceMaxFieldNumber: true} // Force max field number
    maxFieldNumber := 536870912             // Intentionally trigger overflow
    dAtA := randUnrecognizedParams(r, maxFieldNumber)
    if len(dAtA) == 0 {
        t.Fatal("randUnrecognizedParams returned empty data")
    }
    for len(dAtA) > 0 {
        key, n := proto.DecodeVarint(dAtA)
        if n == 0 {
            t.Fatalf("proto.DecodeVarint returned n=0, invalid varint at dAtA=%v", dAtA)
        }
        fieldNumber := int(key >> 3)
        if fieldNumber > 536870911 || fieldNumber < 1 {
            t.Logf("Overflow detected: invalid fieldNumber %d; expected range [1, 536870911]", fieldNumber)
            return // Pass if overflow is detected
        }
        dAtA = dAtA[n:]
    }
    t.Error("Expected overflow with invalid fieldNumber, but none detected")
}