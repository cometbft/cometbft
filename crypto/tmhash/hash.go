package tmhash

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"regexp"
)

const (
	Size      = sha256.Size
	BlockSize = sha256.BlockSize
)

// New returns a new hash.Hash.
func New() hash.Hash {
	return sha256.New()
}

// Sum returns the SHA256 of the bz.
func Sum(bz []byte) []byte {
	h := sha256.Sum256(bz)
	return h[:]
}

// SumMany takes at least 1 byteslice along with a variadic
// number of other byteslices and produces the SHA256 sum from
// hashing them as if they were 1 joined slice.
func SumMany(data []byte, rest ...[]byte) []byte {
	h := sha256.New()
	h.Write(data)
	for _, data := range rest {
		h.Write(data)
	}
	return h.Sum(nil)
}

// ValidateSHA256 checks if the given string is a syntactically valid SHA256 hash.
// A valid SHA256 hash is a hex-encoded 64-character string.
// If the hash isn't valid, it returns an error explaining why.
func ValidateSHA256(hashStr string) error {
	const sha256Pattern = `^[a-fA-F0-9]{64}$`

	if len(hashStr) != 64 {
		return fmt.Errorf("expected 64 characters, but have %d", len(hashStr))
	}

	match, err := regexp.MatchString(sha256Pattern, hashStr)
	if err != nil {
		// if this happens, there is a bug in the regex or some internal regexp
		// package error.
		return fmt.Errorf("can't run regex %q: %s", sha256Pattern, err)
	}

	if !match {
		return fmt.Errorf("contains non-hexadecimal characters")
	}

	return nil
}

// -------------------------------------------------------------

const (
	TruncatedSize = 20
)

type sha256trunc struct {
	sha256 hash.Hash
}

func (h sha256trunc) Write(p []byte) (n int, err error) {
	return h.sha256.Write(p)
}

func (h sha256trunc) Sum(b []byte) []byte {
	shasum := h.sha256.Sum(b)
	return shasum[:TruncatedSize]
}

func (h sha256trunc) Reset() {
	h.sha256.Reset()
}

func (sha256trunc) Size() int {
	return TruncatedSize
}

func (h sha256trunc) BlockSize() int {
	return h.sha256.BlockSize()
}

// NewTruncated returns a new hash.Hash.
func NewTruncated() hash.Hash {
	return sha256trunc{
		sha256: sha256.New(),
	}
}

// SumTruncated returns the first 20 bytes of SHA256 of the bz.
func SumTruncated(bz []byte) []byte {
	hash := sha256.Sum256(bz)
	return hash[:TruncatedSize]
}
