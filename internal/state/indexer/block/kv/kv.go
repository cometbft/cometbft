package kv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/google/orderedcode"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	idxutil "github.com/cometbft/cometbft/internal/indexer"
	"github.com/cometbft/cometbft/internal/pubsub/query"
	"github.com/cometbft/cometbft/internal/pubsub/query/syntax"
	"github.com/cometbft/cometbft/internal/state"
	"github.com/cometbft/cometbft/internal/state/indexer"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/types"
)

var (
	LastBlockIndexerRetainHeightKey = []byte("LastBlockIndexerRetainHeightKey")
	BlockIndexerRetainHeightKey     = []byte("BlockIndexerRetainHeightKey")
	ErrInvalidHeightValue           = errors.New("invalid height value")
)

// BlockerIndexer implements a block indexer, indexing FinalizeBlock
// events with an underlying KV store. Block events are indexed by their height,
// such that matching search criteria returns the respective block height(s).
type BlockerIndexer struct {
	store dbm.DB

	// Add unique event identifier to use when querying
	// Matching will be done both on height AND eventSeq
	eventSeq int64
	log      log.Logger
}

func New(store dbm.DB) *BlockerIndexer {
	return &BlockerIndexer{
		store: store,
	}
}

func (idx *BlockerIndexer) SetLogger(l log.Logger) {
	idx.log = l
}

// Has returns true if the given height has been indexed. An error is returned
// upon database query failure.
func (idx *BlockerIndexer) Has(height int64) (bool, error) {
	key, err := heightKey(height)
	if err != nil {
		return false, fmt.Errorf("failed to create block height index key: %w", err)
	}

	return idx.store.Has(key)
}

// Index indexes FinalizeBlock events for a given block by its height.
// The following is indexed:
//
// primary key: encode(block.height | height) => encode(height)
// FinalizeBlock events: encode(eventType.eventAttr|eventValue|height|finalize_block|eventSeq) => encode(height).
func (idx *BlockerIndexer) Index(bh types.EventDataNewBlockEvents) error {
	batch := idx.store.NewBatch()
	defer batch.Close()

	height := bh.Height

	// 1. index by height
	key, err := heightKey(height)
	if err != nil {
		return fmt.Errorf("failed to create block height index key: %w", err)
	}
	if err := batch.Set(key, int64ToBytes(height)); err != nil {
		return err
	}

	// 2. index block events
	if err := idx.indexEvents(batch, bh.Events, height); err != nil {
		return fmt.Errorf("failed to index FinalizeBlock events: %w", err)
	}
	return batch.WriteSync()
}

func getKeys(indexer BlockerIndexer) [][]byte {
	var keys [][]byte

	itr, err := indexer.store.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	for ; itr.Valid(); itr.Next() {
		keys = append(keys, itr.Key())
	}
	return keys
}

func (idx *BlockerIndexer) Prune(retainHeight int64) (numPruned int64, newRetainHeight int64, err error) {
	// Returns numPruned, newRetainHeight, err
	// numPruned: the number of heights pruned or 0 in case of error. E.x. if heights {1, 3, 7} were pruned and there was no error, numPruned == 3
	// newRetainHeight: new retain height after pruning or lastRetainHeight in case of error
	// err: error

	lastRetainHeight, err := idx.getLastRetainHeight()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to look up last block indexer retain height: %w", err)
	}
	if lastRetainHeight == 0 {
		lastRetainHeight = 1
	}

	batch := idx.store.NewBatch()
	closeBatch := func(batch dbm.Batch) {
		err := batch.Close()
		if err != nil {
			idx.log.Error(fmt.Sprintf("Error when closing block indexer pruning batch: %v", err))
		}
	}
	defer closeBatch(batch)

	flush := func(batch dbm.Batch) error {
		err := batch.WriteSync()
		if err != nil {
			return fmt.Errorf("failed to flush block indexer pruning batch %w", err)
		}
		err = batch.Close()
		if err != nil {
			idx.log.Error(fmt.Sprintf("Error when closing block indexer pruning batch: %v", err))
		}
		return nil
	}

	itr, err := idx.store.Iterator(nil, nil)
	if err != nil {
		return 0, lastRetainHeight, err
	}
	deleted := 0
	affectedHeights := make(map[int64]struct{})
	for ; itr.Valid(); itr.Next() {
		if keyBelongsToHeightRange(itr.Key(), lastRetainHeight, retainHeight) {
			err := batch.Delete(itr.Key())
			if err != nil {
				return 0, lastRetainHeight, err
			}
			height := getHeightFromKey(itr.Key())
			affectedHeights[height] = struct{}{}
			deleted++
		}
		if deleted%1000 == 0 && deleted != 0 {
			err = flush(batch)
			if err != nil {
				return 0, lastRetainHeight, err
			}
			batch = idx.store.NewBatch()
			defer closeBatch(batch)
		}
	}

	errSetLastRetainHeight := idx.setLastRetainHeight(retainHeight, batch)
	if errSetLastRetainHeight != nil {
		return 0, lastRetainHeight, errSetLastRetainHeight
	}
	errWriteBatch := batch.WriteSync()
	if errWriteBatch != nil {
		return 0, lastRetainHeight, errWriteBatch
	}

	return int64(len(affectedHeights)), retainHeight, err
}

func (idx *BlockerIndexer) SetRetainHeight(retainHeight int64) error {
	return idx.store.SetSync(BlockIndexerRetainHeightKey, int64ToBytes(retainHeight))
}

func (idx *BlockerIndexer) GetRetainHeight() (int64, error) {
	buf, err := idx.store.Get(BlockIndexerRetainHeightKey)
	if err != nil {
		return 0, err
	}
	if buf == nil {
		return 0, state.ErrKeyNotFound
	}
	height := int64FromBytes(buf)

	if height < 0 {
		return 0, state.ErrInvalidHeightValue
	}

	return height, nil
}

func (*BlockerIndexer) setLastRetainHeight(height int64, batch dbm.Batch) error {
	return batch.Set(LastBlockIndexerRetainHeightKey, int64ToBytes(height))
}

func (idx *BlockerIndexer) getLastRetainHeight() (int64, error) {
	bz, err := idx.store.Get(LastBlockIndexerRetainHeightKey)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil
	}
	height := int64FromBytes(bz)
	if height < 0 {
		return 0, ErrInvalidHeightValue
	}
	return height, nil
}

// Search performs a query for block heights that match a given FinalizeBlock
// event search criteria. The given query can match against zero,
// one or more block heights. In the case of height queries, i.e. block.height=H,
// if the height is indexed, that height alone will be returned. An error and
// nil slice is returned. Otherwise, a non-nil slice and nil error is returned.
func (idx *BlockerIndexer) Search(ctx context.Context, q *query.Query) ([]int64, error) {
	results := make([]int64, 0)
	select {
	case <-ctx.Done():
		return results, nil

	default:
	}

	conditions := q.Syntax()

	// conditions to skip because they're handled before "everything else"
	skipIndexes := make([]int, 0)

	var ok bool

	var heightInfo HeightInfo
	// If we are not matching events and block.height occurs more than once, the later value will
	// overwrite the first one.
	conditions, heightInfo, ok = dedupHeight(conditions)

	// Extract ranges. If both upper and lower bounds exist, it's better to get
	// them in order as to not iterate over kvs that are not within range.
	ranges, rangeIndexes, heightRange := indexer.LookForRangesWithHeight(conditions)
	heightInfo.heightRange = heightRange

	// If we have additional constraints and want to query per event
	// attributes, we cannot simply return all blocks for a height.
	// But we remember the height we want to find and forward it to
	// match(). If we only have the height constraint
	// in the query (the second part of the ||), we don't need to query
	// per event conditions and return all events within the height range.
	if ok && heightInfo.onlyHeightEq {
		ok, err := idx.Has(heightInfo.height)
		if err != nil {
			return nil, err
		}

		if ok {
			return []int64{heightInfo.height}, nil
		}

		return results, nil
	}

	var heightsInitialized bool
	filteredHeights := make(map[string][]byte)
	if heightInfo.heightEqIdx != -1 {
		skipIndexes = append(skipIndexes, heightInfo.heightEqIdx)
	}

	if len(ranges) > 0 {
		skipIndexes = append(skipIndexes, rangeIndexes...)

		for _, qr := range ranges {
			// If we have a query range over height and want to still look for
			// specific event values we do not want to simply return all
			// blocks in this height range. We remember the height range info
			// and pass it on to match() to take into account when processing events.
			if qr.Key == types.BlockHeightKey && !heightInfo.onlyHeightRange {
				// If the query contains ranges other than the height then we need to treat the height
				// range when querying the conditions of the other range.
				// Otherwise we can just return all the blocks within the height range (as there is no
				// additional constraint on events)

				continue
			}
			prefix, err := orderedcode.Append(nil, qr.Key)
			if err != nil {
				return nil, fmt.Errorf("failed to create prefix key: %w", err)
			}

			if !heightsInitialized {
				filteredHeights, err = idx.matchRange(ctx, qr, prefix, filteredHeights, true, heightInfo)
				if err != nil {
					return nil, err
				}

				heightsInitialized = true

				// Ignore any remaining conditions if the first condition resulted in no
				// matches (assuming implicit AND operand).
				if len(filteredHeights) == 0 {
					break
				}
			} else {
				filteredHeights, err = idx.matchRange(ctx, qr, prefix, filteredHeights, false, heightInfo)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// for all other conditions
	for i, c := range conditions {
		if intInSlice(i, skipIndexes) {
			continue
		}

		var err error
		filteredHeights, err = idx.processCondition(ctx, c, filteredHeights, !heightsInitialized, heightInfo)
		if err != nil {
			return nil, err
		}

		heightsInitialized = heightsInitialized || len(filteredHeights) > 0

		// Ignore any remaining conditions if the first condition resulted in no matches (assuming AND operand).
		if len(filteredHeights) == 0 {
			break
		}
	}

	// fetch matching heights
	results = make([]int64, 0, len(filteredHeights))
	resultMap := make(map[int64]struct{})
	for _, hBz := range filteredHeights {
		h := int64FromBytes(hBz)

		ok, err := idx.Has(h)
		if err != nil {
			return nil, err
		}
		if ok {
			if _, ok := resultMap[h]; !ok {
				resultMap[h] = struct{}{}
				results = append(results, h)
			}
		}

		select {
		case <-ctx.Done():

		default:
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i] < results[j] })

	return results, nil
}

// matchRange returns all matching block heights that match a given QueryRange
// and start key. An already filtered result (filteredHeights) is provided such
// that any non-intersecting matches are removed.
//
// NOTE: The provided filteredHeights may be empty if no previous condition has
// matched.
func (idx *BlockerIndexer) matchRange(
	ctx context.Context,
	qr indexer.QueryRange,
	startKey []byte,
	filteredHeights map[string][]byte,
	firstRun bool,
	heightInfo HeightInfo,
) (map[string][]byte, error) {
	// If this is not the first run and there are no previous matches, return an empty result.
	if !firstRun && len(filteredHeights) == 0 {
		return filteredHeights, nil
	}

	tmpHeights := make(map[string][]byte)
	it, err := dbm.IteratePrefix(idx.store, startKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create prefix iterator: %w", err)
	}
	defer it.Close()

	for ; it.Valid(); it.Next() {
		include, err := idx.processIteratorItem(it, qr, tmpHeights, heightInfo)
		if err != nil {
			continue // Skip this iteration if there was an error processing the item.
		}

		if include {
			idx.setTmpHeights(tmpHeights, it)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	if err := it.Error(); err != nil {
		return nil, err
	}

	if firstRun {
		return tmpHeights, nil
	}

	// Merge the new filtered heights with the existing ones.
	return idx.mergeFilteredHeights(filteredHeights, tmpHeights), nil
}

func (*BlockerIndexer) setTmpHeights(tmpHeights map[string][]byte, it dbm.Iterator) {
	// If we return attributes that occur within the same events, then store the event sequence in the
	// result map as well
	eventSeq, _ := parseEventSeqFromEventKey(it.Key())
	retVal := it.Value()
	tmpHeights[string(retVal)+strconv.FormatInt(eventSeq, 10)] = it.Value()
}

func (idx *BlockerIndexer) indexEvents(batch dbm.Batch, events []abci.Event, height int64) error {
	heightBz := int64ToBytes(height)

	for _, event := range events {
		idx.eventSeq++
		// only index events with a non-empty type
		if len(event.Type) == 0 {
			continue
		}

		for _, attr := range event.Attributes {
			if len(attr.Key) == 0 {
				continue
			}

			// index iff the event specified index:true and it's not a reserved event
			compositeKey := fmt.Sprintf("%s.%s", event.Type, attr.Key)
			if compositeKey == types.BlockHeightKey {
				return fmt.Errorf("event type and attribute key \"%s\" is reserved; please use a different key", compositeKey)
			}

			if attr.GetIndex() {
				key, err := eventKey(compositeKey, attr.Value, height, idx.eventSeq)
				if err != nil {
					return fmt.Errorf("failed to create block index key: %w", err)
				}

				if err := batch.Set(key, heightBz); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// HELPER FUNCTIONS TO REDUCE COMPLEXITY

// matchTEq returns all matching block heights that match a given QueryRange.
func (idx *BlockerIndexer) matchTEq(
	ctx context.Context,
	startKeyBz []byte,
	heightInfo HeightInfo,
) (map[string][]byte, error) {
	tmpHeights := make(map[string][]byte)

	it, err := dbm.IteratePrefix(idx.store, startKeyBz)
	if err != nil {
		return nil, fmt.Errorf("failed to create prefix iterator: %w", err)
	}
	defer it.Close()

	for ; it.Valid(); it.Next() {
		keyHeight, err := parseHeightFromEventKey(it.Key())
		if err != nil {
			idx.log.Error("failure to parse height from key:", err)
			continue
		}
		withinHeight, err := checkHeightConditions(heightInfo, keyHeight)
		if err != nil {
			idx.log.Error("failure checking for height bounds:", err)
			continue
		}
		if !withinHeight {
			continue
		}

		idx.setTmpHeights(tmpHeights, it)

		if err := ctx.Err(); err != nil {
			break
		}
	}

	if err := it.Error(); err != nil {
		return nil, err
	}

	return tmpHeights, nil
}

// matchTExists returns all matching block heights that match a given QueryRange.
func (idx *BlockerIndexer) matchTExists(
	ctx context.Context,
	c syntax.Condition,
	heightInfo HeightInfo,
) (map[string][]byte, error) {
	tmpHeights := make(map[string][]byte)

	// Create prefix based on the event tag (condition tag)
	prefix, err := orderedcode.Append(nil, c.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to create prefix key: %w", err)
	}

	// Create a prefix iterator to go through all events with the specified tag
	it, err := dbm.IteratePrefix(idx.store, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create prefix iterator: %w", err)
	}
	defer it.Close()

	for ; it.Valid(); it.Next() {
		// Extract the height from the event key
		keyHeight, err := parseHeightFromEventKey(it.Key())
		if err != nil {
			idx.log.Error("failure to parse height from key:", err)
			continue
		}

		// Check if the height is within specified bounds
		withinHeight, err := checkHeightConditions(heightInfo, keyHeight)
		if err != nil {
			idx.log.Error("failure checking for height bounds:", err)
			continue
		}
		if !withinHeight {
			continue
		}

		// Add to temporary heights map
		idx.setTmpHeights(tmpHeights, it)

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return tmpHeights, ctx.Err()

		default:
		}
	}

	if err := it.Error(); err != nil {
		return nil, err
	}

	return tmpHeights, nil
}

// matchTContains returns all matching block heights that match a given QueryRange.
func (idx *BlockerIndexer) matchTContains(
	ctx context.Context,
	c syntax.Condition,
	heightInfo HeightInfo,
) (map[string][]byte, error) {
	tmpHeights := make(map[string][]byte)

	// Create prefix based on the event tag (condition tag)
	prefix, err := orderedcode.Append(nil, c.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to create prefix key: %w", err)
	}

	// Create a prefix iterator to go through all events with the specified tag
	it, err := dbm.IteratePrefix(idx.store, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create prefix iterator: %w", err)
	}
	defer it.Close()

	for ; it.Valid(); it.Next() {
		// Extract the event value from the key
		eventValue, err := parseValueFromEventKey(it.Key())
		if err != nil {
			continue
		}

		// Check if the event value contains the condition argument
		if strings.Contains(eventValue, c.Arg.Value()) {
			// Extract the height from the event key
			keyHeight, err := parseHeightFromEventKey(it.Key())
			if err != nil {
				idx.log.Error("failure to parse height from key:", err)
				continue
			}

			// Check if the height is within specified bounds
			withinHeight, err := checkHeightConditions(heightInfo, keyHeight)
			if err != nil {
				idx.log.Error("failure checking for height bounds:", err)
				continue
			}
			if !withinHeight {
				continue
			}

			// Add to temporary heights map
			idx.setTmpHeights(tmpHeights, it)
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return tmpHeights, ctx.Err()

		default:
		}
	}

	if err := it.Error(); err != nil {
		return nil, err
	}

	return tmpHeights, nil
}

// processIteratorItem processes a single iterator item for matchRange.
func (idx *BlockerIndexer) processIteratorItem(
	it dbm.Iterator,
	qr indexer.QueryRange,
	tmpHeights map[string][]byte,
	heightInfo HeightInfo,
) (bool, error) {
	var eventValue string
	var err error
	if qr.Key == types.BlockHeightKey {
		eventValue, err = parseValueFromPrimaryKey(it.Key())
	} else {
		eventValue, err = parseValueFromEventKey(it.Key())
	}
	if err != nil {
		return false, err
	}

	// For numeric bounds, we handle them differently
	if _, isNumeric := qr.AnyBound().(*big.Float); isNumeric {
		return idx.processNumericBounds(it, qr, tmpHeights, eventValue, heightInfo)
	}

	if qr.Key != types.BlockHeightKey {
		keyHeight, err := parseHeightFromEventKey(it.Key())
		if err != nil {
			idx.log.Error("failure to parse height from key:", err)
			return false, err
		}
		withinHeight, err := checkHeightConditions(heightInfo, keyHeight)
		if err != nil {
			idx.log.Error("failure checking for height bounds:", err)
			return false, err
		}
		if !withinHeight {
			return false, nil
		}
	}

	idx.setTmpHeights(tmpHeights, it)
	return true, nil
}

// processNumericBounds checks if the numeric value of an event meets the query range bounds.
func (idx *BlockerIndexer) processNumericBounds(
	it dbm.Iterator,
	qr indexer.QueryRange,
	tmpHeights map[string][]byte,
	eventValue string,
	heightInfo HeightInfo,
) (bool, error) {
	v := new(big.Int)
	v, ok := v.SetString(eventValue, 10)

	if !ok {
		vF, _, err := big.ParseFloat(eventValue, 10, 125, big.ToNearestEven)
		if err != nil {
			return false, err
		}
		withinBounds, err := idxutil.CheckBounds(qr, vF)
		if err != nil || !withinBounds {
			return false, err
		}
	} else {
		withinBounds, err := idxutil.CheckBounds(qr, v)
		if err != nil || !withinBounds {
			return false, err
		}
	}

	if qr.Key != types.BlockHeightKey {
		keyHeight, err := parseHeightFromEventKey(it.Key())
		if err != nil {
			idx.log.Error("failure to parse height from key:", err)
			return false, err
		}
		withinHeight, err := checkHeightConditions(heightInfo, keyHeight)
		if err != nil || !withinHeight {
			return false, err
		}
	}

	idx.setTmpHeights(tmpHeights, it)
	return true, nil
}

func (*BlockerIndexer) mergeFilteredHeights(
	existingFilteredHeights map[string][]byte,
	newFilteredHeights map[string][]byte,
) map[string][]byte {
	// If the existing filtered heights map is empty, return the new filtered heights as there's nothing to merge.
	if len(existingFilteredHeights) == 0 {
		return newFilteredHeights
	}

	// If the new filtered heights map is empty, return an empty map as there are no common heights.
	if len(newFilteredHeights) == 0 {
		return make(map[string][]byte)
	}

	// Initialize a map to store the merged heights.
	mergedHeights := make(map[string][]byte)

	// Iterate over the existing filtered heights and check if they exist in the new filtered heights.
	for key, value := range existingFilteredHeights {
		if newVal, ok := newFilteredHeights[key]; ok && bytes.Equal(value, newVal) {
			mergedHeights[key] = value
		}
	}

	// Return the merged heights.
	return mergedHeights
}

func (idx *BlockerIndexer) processCondition(
	ctx context.Context,
	c syntax.Condition,
	filteredHeights map[string][]byte,
	firstRun bool,
	heightInfo HeightInfo,
) (map[string][]byte, error) {
	var tmpHeights map[string][]byte
	var err error
	startKey, err := orderedcode.Append(nil, c.Tag, c.Arg.Value())
	if err != nil {
		return nil, err
	}

	switch {
	case c.Op == syntax.TEq:
		tmpHeights, err = idx.matchTEq(ctx, startKey, heightInfo)
	case c.Op == syntax.TExists:
		tmpHeights, err = idx.matchTExists(ctx, c, heightInfo)
	case c.Op == syntax.TContains:
		tmpHeights, err = idx.matchTContains(ctx, c, heightInfo)
	default:
		return nil, errors.New("unsupported operator")
	}
	if err != nil {
		return nil, err
	}

	if len(tmpHeights) == 0 || firstRun {
		return tmpHeights, nil
	}

	return idx.mergeFilteredHeights(filteredHeights, tmpHeights), nil
}
