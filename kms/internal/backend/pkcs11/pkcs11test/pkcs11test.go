// Package pkcs11test provides SoftHSM2-backed helpers for exercising the PKCS#11
// signer in tests. It initializes an isolated, throwaway token and generates an
// Ed25519 key in it. Tests that use it auto-skip when SoftHSM2 is not installed.
package pkcs11test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/miekg/pkcs11"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/ed25519"
)

// PKCS#11 / SoftHSM2 constants not exported by miekg/pkcs11 v1.1.x.
const (
	ckmECEdwardsKeyPairGen = 0x00001055
	ckkECEdwards           = 0x00000040
)

// ed25519 curve OID 1.3.101.112, DER-encoded for CKA_EC_PARAMS.
var oidEd25519 = []byte{0x06, 0x03, 0x2b, 0x65, 0x70}

// Fixed token/key identifiers used by the helper.
const (
	TokenLabel = "comet"
	KeyLabel   = "validator"
	UserPIN    = "1234"
	soPIN      = "4321"
)

// KeyID is the CKA_ID of the generated key.
var KeyID = []byte{0x01}

// commonModulePaths lists where libsofthsm2.so is typically installed.
var commonModulePaths = []string{
	"/opt/homebrew/lib/softhsm/libsofthsm2.so",
	"/usr/local/lib/softhsm/libsofthsm2.so",
	"/usr/lib/softhsm/libsofthsm2.so",
	"/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so",
}

// findSlotByLabel returns the slot whose token label matches label. Token
// labels are space-padded to 32 bytes by the spec.
func findSlotByLabel(t *testing.T, ctx *pkcs11.Ctx, label string) uint {
	t.Helper()
	slots, err := ctx.GetSlotList(true)
	require.NoError(t, err)
	for _, s := range slots {
		info, err := ctx.GetTokenInfo(s)
		if err != nil {
			continue
		}
		if strings.TrimRight(info.Label, " ") == label {
			return s
		}
	}
	t.Fatalf("no slot with token label %q", label)
	return 0
}

// FindModule returns the SoftHSM2 module path, or skips the test if not found.
func FindModule(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("COMETKMS_SOFTHSM_LIB"); p != "" {
		return p
	}
	for _, p := range commonModulePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Skip("SoftHSM2 module not found; set COMETKMS_SOFTHSM_LIB to run PKCS#11 integration tests")
	return ""
}

// SetupToken initializes a fresh SoftHSM2 token in an isolated temp directory and
// generates an Ed25519 key pair in it. It returns the on-token public key so the
// caller can verify what the signer reads and signs. SOFTHSM2_CONF is pointed at
// the temp dir for the duration of the test.
func SetupToken(t *testing.T, module string) ed25519.PubKey {
	t.Helper()

	tokenDir := t.TempDir()
	confPath := filepath.Join(t.TempDir(), "softhsm2.conf")
	conf := "directories.tokendir = " + tokenDir + "\nobjectstore.backend = file\nlog.level = ERROR\n"
	require.NoError(t, os.WriteFile(confPath, []byte(conf), 0o600))
	t.Setenv("SOFTHSM2_CONF", confPath)

	ctx := pkcs11.New(module)
	require.NotNil(t, ctx, "load module")
	require.NoError(t, ctx.Initialize())
	// Finalize this provisioning context before returning so the signer under
	// test can initialize the (process-global) module cleanly. Token data
	// persists on disk via the file objectstore.
	defer func() { _ = ctx.Finalize(); ctx.Destroy() }()

	slots, err := ctx.GetSlotList(false)
	require.NoError(t, err)
	require.NotEmpty(t, slots)
	slot := slots[0]

	require.NoError(t, ctx.InitToken(slot, soPIN, TokenLabel))

	// Re-query after InitToken: the PKCS#11 spec allows the slot list to change
	// (SoftHSM2 may assign the initialized token a new slot ID). Find our token
	// by label rather than trusting the original slot ID.
	slot = findSlotByLabel(t, ctx, TokenLabel)

	// Set the user PIN: open RW, login SO, InitPIN, logout.
	sess, err := ctx.OpenSession(slot, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	require.NoError(t, err)
	require.NoError(t, ctx.Login(sess, pkcs11.CKU_SO, soPIN))
	require.NoError(t, ctx.InitPIN(sess, UserPIN))
	require.NoError(t, ctx.Logout(sess))

	// Generate the Ed25519 key pair as the user.
	require.NoError(t, ctx.Login(sess, pkcs11.CKU_USER, UserPIN))
	pubH := genKeyPair(t, ctx, sess, KeyLabel, KeyID)

	// Read the public point back so the caller can verify signatures.
	attrs, err := ctx.GetAttributeValue(sess, pubH, []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_EC_POINT, nil),
	})
	require.NoError(t, err)
	require.Len(t, attrs, 1)

	require.NoError(t, ctx.Logout(sess))
	require.NoError(t, ctx.CloseSession(sess))

	raw := attrs[0].Value
	// Unwrap a DER OCTET STRING (0x04 0x20 <32 bytes>) if present.
	if len(raw) == ed25519.PubKeySize+2 && raw[0] == 0x04 && raw[1] == ed25519.PubKeySize {
		raw = raw[2:]
	}
	require.Len(t, raw, ed25519.PubKeySize)
	pub := make(ed25519.PubKey, ed25519.PubKeySize)
	copy(pub, raw)
	return pub
}

// AddKeyPair generates an additional Ed25519 key pair with the given label and
// id on the token previously provisioned by SetupToken.
func AddKeyPair(t *testing.T, module, label string, id []byte) {
	t.Helper()

	ctx := pkcs11.New(module)
	require.NotNil(t, ctx, "load module")
	require.NoError(t, ctx.Initialize())
	defer func() { _ = ctx.Finalize(); ctx.Destroy() }()

	slot := findSlotByLabel(t, ctx, TokenLabel)
	sess, err := ctx.OpenSession(slot, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	require.NoError(t, err)
	require.NoError(t, ctx.Login(sess, pkcs11.CKU_USER, UserPIN))
	genKeyPair(t, ctx, sess, label, id)
	require.NoError(t, ctx.Logout(sess))
	require.NoError(t, ctx.CloseSession(sess))
}

// genKeyPair generates a token-resident Ed25519 key pair with the given label
// and id, returning the public key handle. The session must be logged in as
// the user.
func genKeyPair(t *testing.T, ctx *pkcs11.Ctx, sess pkcs11.SessionHandle, label string, id []byte) pkcs11.ObjectHandle {
	t.Helper()

	pubTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PUBLIC_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, ckkECEdwards),
		pkcs11.NewAttribute(pkcs11.CKA_EC_PARAMS, oidEd25519),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_VERIFY, true),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, label),
		pkcs11.NewAttribute(pkcs11.CKA_ID, id),
	}
	privTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, ckkECEdwards),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_PRIVATE, true),
		pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, label),
		pkcs11.NewAttribute(pkcs11.CKA_ID, id),
	}
	pubH, _, err := ctx.GenerateKeyPair(sess,
		[]*pkcs11.Mechanism{pkcs11.NewMechanism(ckmECEdwardsKeyPairGen, nil)},
		pubTemplate, privTemplate)
	require.NoError(t, err)
	return pubH
}
