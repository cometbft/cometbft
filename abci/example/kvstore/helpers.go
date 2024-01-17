package kvstore

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/abci/types"
	cmtrand "github.com/tendermint/tendermint/libs/rand"
)

// RandVal creates one random validator, with a key derived
// from the input value
func RandVal(i int) types.ValidatorUpdate {
	pubkey := cmtrand.Bytes(32)
	power := cmtrand.Uint16() + 1
	v := types.UpdateValidator(pubkey, int64(power), "")
	return v
}

// RandVals returns a list of cnt validators for initializing
// the application. Note that the keys are deterministically
// derived from the index in the array, while the power is
// random (Change this if not desired)
func RandVals(cnt int) []types.ValidatorUpdate {
	res := make([]types.ValidatorUpdate, cnt)
	for i := 0; i < cnt; i++ {
		res[i] = RandVal(i)
	}
	return res
}

// InitKVStore initializes the kvstore app with some data,
// which allows tests to pass and is fine as long as you
// don't make any tx that modify the validator state
func InitKVStore(app *PersistentKVStoreApplication) {
	app.InitChain(types.RequestInitChain{
		Validators: RandVals(1),
	})
}

// Create a new transaction
func NewTx(key, value string) []byte {
	return []byte(strings.Join([]string{key, value}, "="))
}

func NewRandomTx(size int) []byte {
	if size < 4 {
		panic("random tx size must be greater than 3")
	}
	return NewTx(cmtrand.Str(2), cmtrand.Str(size-3))
}

func NewRandomTxs(n int) [][]byte {
	txs := make([][]byte, n)
	for i := 0; i < n; i++ {
		txs[i] = NewRandomTx(10)
	}
	return txs
}

func NewTxFromID(i int) []byte {
	return []byte(fmt.Sprintf("%d=%d", i, i))
}
