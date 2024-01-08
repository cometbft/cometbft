package types

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestABCIPubKey(t *testing.T) {
	pkEd := ed25519.GenPrivKey().PubKey()
	testABCIPubKey(t, pkEd)
}

func testABCIPubKey(t *testing.T, pk crypto.PubKey) {
	t.Helper()
	abciPubKey, err := cryptoenc.PubKeyToProto(pk)
	require.NoError(t, err)
	pk2, err := cryptoenc.PubKeyFromProto(abciPubKey)
	require.NoError(t, err)
	require.Equal(t, pk, pk2)
}

func TestABCIValidators(t *testing.T) {
	pkEd := ed25519.GenPrivKey().PubKey()

	// correct validator
	cmtValExpected := NewValidator(pkEd, 10)

	cmtVal := NewValidator(pkEd, 10)

	abciVal := TM2PB.ValidatorUpdate(cmtVal)
	cmtVals, err := PB2TM.ValidatorUpdates([]abci.ValidatorUpdate{abciVal})
	require.NoError(t, err)
	assert.Equal(t, cmtValExpected, cmtVals[0])

	abciVals := TM2PB.ValidatorUpdates(NewValidatorSet(cmtVals))
	assert.Equal(t, []abci.ValidatorUpdate{abciVal}, abciVals)

	// val with address
	cmtVal.Address = pkEd.Address()

	abciVal = TM2PB.ValidatorUpdate(cmtVal)
	cmtVals, err = PB2TM.ValidatorUpdates([]abci.ValidatorUpdate{abciVal})
	require.NoError(t, err)
	assert.Equal(t, cmtValExpected, cmtVals[0])
}

type pubKeyEddie struct{}

func (pubKeyEddie) Address() Address                    { return []byte{} }
func (pubKeyEddie) Bytes() []byte                       { return []byte{} }
func (pubKeyEddie) VerifySignature([]byte, []byte) bool { return false }
func (pubKeyEddie) Equals(crypto.PubKey) bool           { return false }
func (pubKeyEddie) String() string                      { return "" }
func (pubKeyEddie) Type() string                        { return "pubKeyEddie" }

func TestABCIValidatorFromPubKeyAndPower(t *testing.T) {
	pubkey := ed25519.GenPrivKey().PubKey()

	abciVal := TM2PB.NewValidatorUpdate(pubkey, 10)
	assert.Equal(t, int64(10), abciVal.Power)

	assert.Panics(t, func() { TM2PB.NewValidatorUpdate(nil, 10) })
	assert.Panics(t, func() { TM2PB.NewValidatorUpdate(pubKeyEddie{}, 10) })
}

func TestABCIValidatorWithoutPubKey(t *testing.T) {
	pkEd := ed25519.GenPrivKey().PubKey()

	abciVal := TM2PB.Validator(NewValidator(pkEd, 10))

	// pubkey must be nil
	cmtValExpected := abci.Validator{
		Address: pkEd.Address(),
		Power:   10,
	}

	assert.Equal(t, cmtValExpected, abciVal)
}
