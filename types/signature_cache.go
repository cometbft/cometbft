package types

// The value type for the verified signature cache.
type SignatureCacheValue struct {
	ValidatorPubKeyBytes []byte
	VoteSignBytes        []byte
}

type SignatureCache interface {
	Add(key string, value SignatureCacheValue)
	Get(key string) (SignatureCacheValue, bool)
	Len() int
}

type signatureCache struct {
	cache map[string]SignatureCacheValue
}

func NewSignatureCache() SignatureCache {
	return &signatureCache{
		cache: make(map[string]SignatureCacheValue),
	}
}

func (sc *signatureCache) Add(key string, value SignatureCacheValue) {
	sc.cache[key] = value
}

func (sc *signatureCache) Get(key string) (SignatureCacheValue, bool) {
	value, ok := sc.cache[key]
	return value, ok
}

func (sc *signatureCache) Len() int {
	return len(sc.cache)
}
