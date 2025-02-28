// Package rand provides a pseudo-random number generator seeded with OS randomness.
//
// Deprecated: This package will be removed in a future release. Users should migrate
// off this package, as its functionality will no longer be part of the exported interface.
package rand

import (
	crand "crypto/rand"
	mrand "math/rand"
	"time"

	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

const (
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
)

// Rand is a prng, that is seeded with OS randomness.
// The OS randomness is obtained from crypto/rand, however none of the provided
// methods are suitable for cryptographic usage.
// They all utilize math/rand's prng internally.
//
// All of the methods here are suitable for concurrent use.
// This is achieved by using a mutex lock on all of the provided methods.
// Deprecated: This struct will be removed in a future release. Do not use.
type Rand struct {
	cmtsync.Mutex
	rand *mrand.Rand
}

var grand *Rand

func init() {
	grand = NewRand()
	grand.init()
}

// Deprecated: This function will be removed in a future release. Do not use.
func NewRand() *Rand {
	rand := &Rand{}
	rand.init()
	return rand
}

// Make a new stdlib rand source. Its up to the caller to ensure
// that the rand source is not called in parallel.
// The failure mode of calling the returned rand multiple times in parallel is
// repeated values across threads.
// Deprecated: This function will be removed in a future release. Do not use.
func NewStdlibRand() *mrand.Rand {
	// G404: Use of weak random number generator (math/rand instead of crypto/rand)
	//nolint:gosec
	return mrand.New(mrand.NewSource(newSeed()))
}

func newSeed() int64 {
	bz := cRandBytes(8)
	var seed uint64
	for i := 0; i < 8; i++ {
		seed |= uint64(bz[i])
		seed <<= 8
	}
	return int64(seed)
}

func (r *Rand) init() {
	r.reset(newSeed())
}

func (r *Rand) reset(seed int64) {
	// G404: Use of weak random number generator (math/rand instead of crypto/rand)
	//nolint:gosec
	r.rand = mrand.New(mrand.NewSource(seed))
}

// ----------------------------------------
// Global functions

// Deprecated: This function will be removed in a future release. Do not use.
func Seed(seed int64) {
	grand.Seed(seed)
}

// Deprecated: This function will be removed in a future release. Do not use.
func Str(length int) string {
	return grand.Str(length)
}

// Deprecated: This function will be removed in a future release. Do not use.
func Uint16() uint16 {
	return grand.Uint16()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Uint32() uint32 {
	return grand.Uint32()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Uint64() uint64 {
	return grand.Uint64()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Uint() uint {
	return grand.Uint()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Int16() int16 {
	return grand.Int16()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Int32() int32 {
	return grand.Int32()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Int64() int64 {
	return grand.Int64()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Int() int {
	return grand.Int()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Int31() int32 {
	return grand.Int31()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Int31n(n int32) int32 {
	return grand.Int31n(n)
}

// Deprecated: This function will be removed in a future release. Do not use.
func Int63() int64 {
	return grand.Int63()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Int63n(n int64) int64 {
	return grand.Int63n(n)
}

// Deprecated: This function will be removed in a future release. Do not use.
func Bool() bool {
	return grand.Bool()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Float32() float32 {
	return grand.Float32()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Float64() float64 {
	return grand.Float64()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Time() time.Time {
	return grand.Time()
}

// Deprecated: This function will be removed in a future release. Do not use.
func Bytes(n int) []byte {
	return grand.Bytes(n)
}

// Deprecated: This function will be removed in a future release. Do not use.
func Intn(n int) int {
	return grand.Intn(n)
}

// Deprecated: This function will be removed in a future release. Do not use.
func Perm(n int) []int {
	return grand.Perm(n)
}

// ----------------------------------------
// Rand methods

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Seed(seed int64) {
	r.Lock()
	r.reset(seed)
	r.Unlock()
}

// Str constructs a random alphanumeric string of given length.
// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Str(length int) string {
	if length <= 0 {
		return ""
	}

	chars := []byte{}
MAIN_LOOP:
	for {
		val := r.Int63()
		for i := 0; i < 10; i++ {
			v := int(val & 0x3f) // rightmost 6 bits
			if v >= 62 {         // only 62 characters in strChars
				val >>= 6
				continue
			}
			chars = append(chars, strChars[v])
			if len(chars) == length {
				break MAIN_LOOP
			}
			val >>= 6
		}
	}

	return string(chars)
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Uint16() uint16 {
	return uint16(r.Uint32() & (1<<16 - 1))
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Uint32() uint32 {
	r.Lock()
	u32 := r.rand.Uint32()
	r.Unlock()
	return u32
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Uint64() uint64 {
	return uint64(r.Uint32())<<32 + uint64(r.Uint32())
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Uint() uint {
	r.Lock()
	i := r.rand.Int()
	r.Unlock()
	return uint(i)
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Int16() int16 {
	return int16(r.Uint32() & (1<<16 - 1))
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Int32() int32 {
	return int32(r.Uint32())
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Int64() int64 {
	return int64(r.Uint64())
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Int() int {
	r.Lock()
	i := r.rand.Int()
	r.Unlock()
	return i
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Int31() int32 {
	r.Lock()
	i31 := r.rand.Int31()
	r.Unlock()
	return i31
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Int31n(n int32) int32 {
	r.Lock()
	i31n := r.rand.Int31n(n)
	r.Unlock()
	return i31n
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Int63() int64 {
	r.Lock()
	i63 := r.rand.Int63()
	r.Unlock()
	return i63
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Int63n(n int64) int64 {
	r.Lock()
	i63n := r.rand.Int63n(n)
	r.Unlock()
	return i63n
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Float32() float32 {
	r.Lock()
	f32 := r.rand.Float32()
	r.Unlock()
	return f32
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Float64() float64 {
	r.Lock()
	f64 := r.rand.Float64()
	r.Unlock()
	return f64
}

// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Time() time.Time {
	return time.Unix(int64(r.Uint64()), 0)
}

// Bytes returns n random bytes generated from the internal
// prng.
// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Bytes(n int) []byte {
	// cRandBytes isn't guaranteed to be fast so instead
	// use random bytes generated from the internal PRNG
	bs := make([]byte, n)
	for i := 0; i < len(bs); i++ {
		bs[i] = byte(r.Int() & 0xFF)
	}
	return bs
}

// Intn returns, as an int, a uniform pseudo-random number in the range [0, n).
// It panics if n <= 0.
// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Intn(n int) int {
	r.Lock()
	i := r.rand.Intn(n)
	r.Unlock()
	return i
}

// Bool returns a uniformly random boolean.
// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Bool() bool {
	// See https://github.com/golang/go/issues/23804#issuecomment-365370418
	// for reasoning behind computing like this
	return r.Int63()%2 == 0
}

// Perm returns a pseudo-random permutation of n integers in [0, n).
// Deprecated: This function will be removed in a future release. Do not use.
func (r *Rand) Perm(n int) []int {
	r.Lock()
	perm := r.rand.Perm(n)
	r.Unlock()
	return perm
}

// NOTE: This relies on the os's random number generator.
// For real security, we should salt that with some seed.
// See github.com/cometbft/cometbft/crypto for a more secure reader.
// This function is thread safe, see:
// https://stackoverflow.com/questions/75685374/is-golang-crypto-rand-thread-safe
func cRandBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, err := crand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
