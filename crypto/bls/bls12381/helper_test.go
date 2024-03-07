//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && bls12381

package blst

// Note: These functions are for tests to access private globals, such as pubkeyCache.

// DisableCaches sets the cache sizes to 0.
func DisableCaches() {
	pubkeyCache.Resize(0)
}

// EnableCaches sets the cache sizes to the default values.
func EnableCaches() {
	pubkeyCache.Resize(maxKeys)
}
