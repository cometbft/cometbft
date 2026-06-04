// Package pkcs11 implements a backend.Signer backed by a PKCS#11 token or HSM.
// The private key never leaves the token: signing is performed on-device via the
// PKCS#11 C_Sign operation. Ed25519 (CKM_EDDSA) is the only key algorithm today;
// see algo.go for the per-algorithm seam.
package pkcs11

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/miekg/pkcs11"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/kms/internal/backend"
)

var _ backend.Signer = (*Signer)(nil)

// Config describes how to open a key on a PKCS#11 token. Exactly one of
// TokenLabel/Slot selects the token, at least one of KeyLabel/KeyID selects the
// key, and exactly one of PIN/PINEnv/PINFile supplies the user PIN. Algorithm
// defaults to "ed25519" when empty.
type Config struct {
	Module     string
	TokenLabel string
	Slot       *uint
	KeyLabel   string
	KeyID      []byte
	PIN        string
	PINEnv     string
	PINFile    string
	Algorithm  string
}

// Signer is a backend.Signer that signs on a PKCS#11 token. It owns a single
// long-lived session; the mutex serializes signing (PKCS#11 sessions are not
// safe for concurrent use) and guards Close.
type Signer struct {
	mod     *pkcs11.Ctx
	module  string // module path, used to release the shared context on Close
	session pkcs11.SessionHandle
	privH   pkcs11.ObjectHandle
	pub     crypto.PubKey
	algo    keyAlgo

	mu     sync.Mutex
	closed bool
}

// Open loads the PKCS#11 module, logs into the selected token, locates the key,
// and caches its public key. Any failure is returned (fatal at startup for the
// chain). On success the returned Signer holds an open, logged-in session that
// must be released with Close.
func Open(cfg Config) (s *Signer, err error) {
	algoName := cfg.Algorithm
	if algoName == "" {
		algoName = algoEd25519
	}
	algo, ok := algos[algoName]
	if !ok {
		return nil, fmt.Errorf("pkcs11: unknown algorithm %q", algoName)
	}

	pin, err := resolvePIN(cfg)
	if err != nil {
		return nil, err
	}

	mod, err := acquireModule(cfg.Module)
	if err != nil {
		return nil, err
	}
	// Release our module reference on any error past this point.
	defer func() {
		if err != nil {
			releaseModule(cfg.Module)
		}
	}()

	slot, err := selectSlot(mod, cfg)
	if err != nil {
		return nil, err
	}

	session, err := mod.OpenSession(slot, pkcs11.CKF_SERIAL_SESSION)
	if err != nil {
		return nil, fmt.Errorf("pkcs11: open session on slot %d: %w", slot, err)
	}
	defer func() {
		if err != nil {
			_ = mod.CloseSession(session)
		}
	}()

	// Login is per-application (shared across sessions on a slot): a concurrent
	// signer on the same token may already hold the login, which is fine.
	if err = mod.Login(session, pkcs11.CKU_USER, pin); err != nil {
		if ce, ok := err.(pkcs11.Error); !ok || ce != pkcs11.CKR_USER_ALREADY_LOGGED_IN {
			return nil, fmt.Errorf("pkcs11: login: %w", err)
		}
		err = nil
	}

	privH, err := findObject(mod, session, pkcs11.CKO_PRIVATE_KEY, cfg)
	if err != nil {
		return nil, fmt.Errorf("pkcs11: find private key: %w", err)
	}
	pubH, err := findObject(mod, session, pkcs11.CKO_PUBLIC_KEY, cfg)
	if err != nil {
		return nil, fmt.Errorf("pkcs11: find public key: %w", err)
	}

	attrs, err := mod.GetAttributeValue(session, pubH, []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_EC_POINT, nil),
	})
	if err != nil {
		return nil, fmt.Errorf("pkcs11: read public key: %w", err)
	}
	if len(attrs) == 0 {
		return nil, fmt.Errorf("pkcs11: public key has no CKA_EC_POINT")
	}
	pub, err := algo.decodePub(attrs[0].Value)
	if err != nil {
		return nil, fmt.Errorf("pkcs11: decode public key: %w", err)
	}

	return &Signer{mod: mod, module: cfg.Module, session: session, privH: privH, pub: pub, algo: algo}, nil
}

// selectSlot returns the slot to use: the explicit Slot when set, otherwise the
// slot whose token CKA_LABEL matches TokenLabel.
func selectSlot(mod *pkcs11.Ctx, cfg Config) (uint, error) {
	if cfg.Slot != nil {
		return *cfg.Slot, nil
	}
	slots, err := mod.GetSlotList(true)
	if err != nil {
		return 0, fmt.Errorf("pkcs11: list slots: %w", err)
	}
	for _, slot := range slots {
		info, err := mod.GetTokenInfo(slot)
		if err != nil {
			continue
		}
		// Token labels are space-padded to 32 bytes by the spec.
		if strings.TrimRight(info.Label, " ") == cfg.TokenLabel {
			return slot, nil
		}
	}
	return 0, fmt.Errorf("pkcs11: no token with label %q", cfg.TokenLabel)
}

// findObject locates exactly one key object of the given class matching the
// configured label and/or id.
func findObject(mod *pkcs11.Ctx, session pkcs11.SessionHandle, class uint, cfg Config) (pkcs11.ObjectHandle, error) {
	template := []*pkcs11.Attribute{pkcs11.NewAttribute(pkcs11.CKA_CLASS, class)}
	if cfg.KeyLabel != "" {
		template = append(template, pkcs11.NewAttribute(pkcs11.CKA_LABEL, cfg.KeyLabel))
	}
	if len(cfg.KeyID) > 0 {
		template = append(template, pkcs11.NewAttribute(pkcs11.CKA_ID, cfg.KeyID))
	}

	if err := mod.FindObjectsInit(session, template); err != nil {
		return 0, err
	}
	handles, _, err := mod.FindObjects(session, 1)
	if finErr := mod.FindObjectsFinal(session); finErr != nil && err == nil {
		err = finErr
	}
	if err != nil {
		return 0, err
	}
	if len(handles) == 0 {
		return 0, fmt.Errorf("no matching object")
	}
	return handles[0], nil
}

// PubKey returns the validator public key cached at Open.
func (s *Signer) PubKey(context.Context) (crypto.PubKey, error) { return s.pub, nil }

// Sign signs the canonical consensus sign-bytes on the token.
func (s *Signer) Sign(_ context.Context, signBytes []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fmt.Errorf("pkcs11: signer is closed")
	}
	if err := s.mod.SignInit(s.session, s.algo.mechanism(), s.privH); err != nil {
		return nil, fmt.Errorf("pkcs11: sign init: %w", err)
	}
	raw, err := s.mod.Sign(s.session, signBytes)
	if err != nil {
		return nil, fmt.Errorf("pkcs11: sign: %w", err)
	}
	return s.algo.fixSig(raw)
}

// Close logs out, closes the session, and tears down the module. It is
// idempotent.
func (s *Signer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	_ = s.mod.CloseSession(s.session)
	// releaseModule finalizes and unloads the module once the last signer using
	// it has closed (Finalize tears down login state and sessions).
	releaseModule(s.module)
	return nil
}

// resolvePIN returns the user PIN from whichever source the config specifies.
// The PIN is read at open time (not stored in config files): an env var is read
// from the process environment; a file is read and stripped of trailing
// whitespace. An empty resolved PIN is an error.
func resolvePIN(cfg Config) (string, error) {
	switch {
	case cfg.PIN != "":
		return cfg.PIN, nil
	case cfg.PINEnv != "":
		v := os.Getenv(cfg.PINEnv)
		if v == "" {
			return "", fmt.Errorf("pkcs11: pin_env %q is empty or unset", cfg.PINEnv)
		}
		return v, nil
	case cfg.PINFile != "":
		raw, err := os.ReadFile(cfg.PINFile)
		if err != nil {
			return "", fmt.Errorf("pkcs11: read pin_file %q: %w", cfg.PINFile, err)
		}
		v := strings.TrimRight(string(raw), " \t\r\n")
		if v == "" {
			return "", fmt.Errorf("pkcs11: pin_file %q is empty", cfg.PINFile)
		}
		return v, nil
	default:
		return "", fmt.Errorf("pkcs11: no PIN source configured (set pin, pin_env, or pin_file)")
	}
}
