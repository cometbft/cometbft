package pex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
)

// findAddressesForSameOldBucket draws random routable addresses until at
// least `need` of them hash (via calcOldBucket) into the same old bucket,
// and returns that bucket index plus the colliding addresses.
func findAddressesForSameOldBucket(t *testing.T, a *addrBook, need int) (int, []*p2p.NetAddress) {
	t.Helper()
	byBucket := make(map[int][]*p2p.NetAddress, oldBucketCount)
	for i := 0; i < 50000; i++ {
		addr := randIPv4Address(t)
		idx, err := a.calcOldBucket(addr)
		require.NoError(t, err)
		byBucket[idx] = append(byBucket[idx], addr)
		if len(byBucket[idx]) >= need {
			return idx, byBucket[idx]
		}
	}
	t.Fatalf("failed to find %d addresses hashing to the same old bucket after 50000 draws", need)
	return 0, nil
}

// TestOldBucketRejectsEntryAtCapacity pins that an old bucket already at its
// documented capacity (oldBucketSize) refuses a further address, rather than
// silently admitting a 65th entry into a 64-entry bucket.
func TestOldBucketRejectsEntryAtCapacity(t *testing.T) {
	fname := createTempFileName("addrbook_test")
	defer deleteTempFile(fname)
	book := NewAddrBook(fname, true)
	book.SetLogger(log.TestingLogger())
	a := book.(*addrBook)

	target, addrs := findAddressesForSameOldBucket(t, a, oldBucketSize+1)

	for i, addr := range addrs[:oldBucketSize] {
		ka := newKnownAddress(addr, addr)
		ka.BucketType = bucketTypeOld
		require.True(t, a.addToOldBucket(ka, target), "failed to fill old bucket, entry %d/%d", i+1, oldBucketSize)
	}
	require.Len(t, a.bucketsOld[target], oldBucketSize)

	overflowAddr := addrs[oldBucketSize]
	overflow := newKnownAddress(overflowAddr, overflowAddr)
	overflow.BucketType = bucketTypeOld
	added := a.addToOldBucket(overflow, target)

	assert.False(t, added, "old bucket admitted a %dth entry past its documented %d-address capacity", oldBucketSize+1, oldBucketSize)
	assert.Len(t, a.bucketsOld[target], oldBucketSize, "old bucket size must never exceed its capacity")
}

// TestMoveToOldDemotionDoesNotLoseAddress pins that when moveToOld's
// overflow-recovery path demotes the oldest entry of a full old bucket back
// to a new bucket, that entry actually lands somewhere -- it must not
// disappear from the address book entirely.
//
// The old bucket is seeded directly to oldBucketSize+1 entries (bypassing
// addToOldBucket's own capacity check), so addToOldBucket's "is this bucket
// full" comparison rejects the promotee under BOTH the buggy `>` and the
// fixed `>=` form -- the demotion branch is forced to run regardless of
// which addrbook.go is under test, isolating this defect from the separate
// off-by-one capacity bug pinned by TestOldBucketRejectsEntryAtCapacity.
func TestMoveToOldDemotionDoesNotLoseAddress(t *testing.T) {
	fname := createTempFileName("addrbook_test")
	defer deleteTempFile(fname)
	book := NewAddrBook(fname, true)
	book.SetLogger(log.TestingLogger())
	a := book.(*addrBook)

	const seeded = oldBucketSize + 1
	target, addrs := findAddressesForSameOldBucket(t, a, seeded+1)

	bucket := a.getBucket(bucketTypeOld, target)
	var oldest *knownAddress
	for i, addr := range addrs[:seeded] {
		ka := newKnownAddress(addr, addr)
		ka.BucketType = bucketTypeOld
		ka.LastAttempt = time.Now().Add(time.Duration(i) * time.Second) // strictly increasing
		bucket[ka.Addr.String()] = ka
		ka.addBucketRef(target)
		a.addrLookup[ka.ID()] = ka
		a.nOld++
		if i == 0 {
			oldest = ka
		}
	}
	require.Len(t, bucket, seeded)
	require.NotNil(t, oldest)
	require.Contains(t, a.addrLookup, oldest.ID())

	// A brand-new address destined for the same (already-overfull) old
	// bucket forces the overflow-recovery/demotion branch of moveToOld.
	promoteeAddr := addrs[seeded]
	promotee := newKnownAddress(promoteeAddr, promoteeAddr)
	newIdx, err := a.calcNewBucket(promotee.Addr, promotee.Src)
	require.NoError(t, err)
	require.NoError(t, a.addToNewBucket(promotee, newIdx))

	require.NoError(t, a.moveToOld(promotee))

	assert.Contains(t, a.addrLookup, oldest.ID(),
		"the demoted address must still be reachable from the book, not silently dropped")
	if ka, ok := a.addrLookup[oldest.ID()]; ok {
		assert.True(t, ka.isNew(), "a demoted address must be marked new, not left as old with nowhere to live")
	}
}

// TestMoveToOldSurvivesLegacyOverfullOldBucket pins that promoting into an
// old bucket that already holds MORE than one entry past capacity (state a
// book persisted by a build predating the addToOldBucket capacity fix could
// have on disk) does not lose the promotee itself. A single demote-one
// recovery pass only brings such a bucket down to exactly its capacity, one
// short of the room the promotee needs; the promotee must not end up marked
// BucketType=old with no bucket membership as a result.
func TestMoveToOldSurvivesLegacyOverfullOldBucket(t *testing.T) {
	fname := createTempFileName("addrbook_test")
	defer deleteTempFile(fname)
	book := NewAddrBook(fname, true)
	book.SetLogger(log.TestingLogger())
	a := book.(*addrBook)

	// oldBucketSize+2 pre-existing entries: one entry past what a single
	// demotion can clear back down to capacity.
	const seeded = oldBucketSize + 2
	target, addrs := findAddressesForSameOldBucket(t, a, seeded+1)

	bucket := a.getBucket(bucketTypeOld, target)
	for i, addr := range addrs[:seeded] {
		ka := newKnownAddress(addr, addr)
		ka.BucketType = bucketTypeOld
		ka.LastAttempt = time.Now().Add(time.Duration(i) * time.Second)
		bucket[ka.Addr.String()] = ka
		ka.addBucketRef(target)
		a.addrLookup[ka.ID()] = ka
		a.nOld++
	}
	require.Len(t, bucket, seeded)

	promoteeAddr := addrs[seeded]
	promotee := newKnownAddress(promoteeAddr, promoteeAddr)
	newIdx, err := a.calcNewBucket(promotee.Addr, promotee.Src)
	require.NoError(t, err)
	require.NoError(t, a.addToNewBucket(promotee, newIdx))

	require.NoError(t, a.moveToOld(promotee))

	got, ok := a.addrLookup[promotee.ID()]
	require.True(t, ok, "the promoted address itself must still be reachable from the book")
	assert.False(t, got.isOld() && len(got.Buckets) == 0,
		"the promoted address must not be left marked old with zero bucket membership (orphaned)")
	assert.LessOrEqual(t, len(a.getBucket(bucketTypeOld, target)), oldBucketSize,
		"the old bucket must not remain over its documented capacity")
}

// TestAddToNewBucketShrinksLegacyOverfullBucket pins that a new bucket
// already holding more than one entry past capacity (again, legacy
// pre-fix-persisted state) is brought back down to capacity by a single
// addToNewBucket call, rather than staying permanently over-full because
// only one expiry ran before the new entry was unconditionally added.
func TestAddToNewBucketShrinksLegacyOverfullBucket(t *testing.T) {
	fname := createTempFileName("addrbook_test")
	defer deleteTempFile(fname)
	book := NewAddrBook(fname, true)
	book.SetLogger(log.TestingLogger())
	a := book.(*addrBook)

	seedAddr := randIPv4Address(t)
	bucketIdx, err := a.calcNewBucket(seedAddr, seedAddr)
	require.NoError(t, err)

	// Directly seed the target bucket with synthetic entries. Their own
	// calcNewBucket destination is irrelevant here: addToNewBucket places an
	// entry into whatever bucketIdx the caller passes, not one it derives
	// from the address itself, so this is a faithful way to simulate a
	// bucket a legacy build already over-filled.
	bucket := a.getBucket(bucketTypeNew, bucketIdx)
	const seeded = newBucketSize + 2
	for i := 0; i < seeded; i++ {
		ka := newKnownAddress(randIPv4Address(t), randIPv4Address(t))
		bucket[ka.Addr.String()] = ka
		ka.addBucketRef(bucketIdx)
		a.addrLookup[ka.ID()] = ka
		a.nNew++
	}
	require.Len(t, bucket, seeded)

	newAddr := randIPv4Address(t)
	newKa := newKnownAddress(newAddr, newAddr)
	require.NoError(t, a.addToNewBucket(newKa, bucketIdx))

	assert.LessOrEqual(t, len(a.getBucket(bucketTypeNew, bucketIdx)), newBucketSize,
		"a legacy over-capacity new bucket must be brought back down to its documented capacity, not left permanently over-full")
}
