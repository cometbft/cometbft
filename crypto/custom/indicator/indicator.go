// Package indicator is used to signal the ed25519 library that a custom crypto implementation is going to be registered.
// The main motivation for this package is that the ed25519 library uses its init function to register the key names
// for amino encoding. In a third-party library the names might be different, but the init function runs before the user
// has the chance to change the names. This indicator package is checked when the ed25519 init is run to see if the
// user implemented a third-party library on top of the ed25519 implementation, in which case the default initialization
// is skipped.
package indicator

var (
	customized = false
	once       = false
)

// IsCustomized returns true if the user implements a custom crypto library on top of the built-in ed25519 implementation.
func IsCustomized() bool {
	return customized
}

// SetCustomized sets the flag that indicates that a custom crypto library is registered.
// This allows a modular crypto implementation without changing the current ed25519 implementation.
// This function panics when called more than once or after the initialization (Golang `init`) of the ed25519 library.
func SetCustomized() {
	if !once {
		customized = true
		once = true
	} else {
		panic("cannot create custom crypto library after initialization")
	}
}

// FinishCustomize locks the customization flag and forbids changing it. It ensures that the flag is only changed up to
// once and only in the initialization phase of the modules.
func FinishCustomize() {
	once = true
}
