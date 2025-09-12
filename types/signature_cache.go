package types

// The value type for the verified signature cache.
type SignatureCacheValue struct {
	ValidatorAddress []byte
	VoteSignBytes    []byte
}

type SignatureCache struct {
	cache map[string]SignatureCacheValue
}

func NewSignatureCache() *SignatureCache {
	return &SignatureCache{
		cache: make(map[string]SignatureCacheValue),
	}
}

func (sc *SignatureCache) Add(key string, value SignatureCacheValue) {
	sc.cache[key] = value
}

func (sc *SignatureCache) Get(key string) (SignatureCacheValue, bool) {
	value, ok := sc.cache[key]
	return value, ok
}

func (sc *SignatureCache) Len() int {
	return len(sc.cache)
}
