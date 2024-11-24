package mempool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/types"
)

// mempoolTx is an entry in the mempool.
type mempoolTx struct {
	height    int64    // height that this tx had been validated in
	gasWanted int64    // amount of gas this tx states it will require
	tx        types.Tx // validated by the application
	lane      LaneID
	seq       int64
	timestamp time.Time // time when entry was created
	// signatures of peers who've sent us this tx (as a map for quick lookups).
	// signatures: PubKey -> signature
	signatures sync.Map
	// Number of valid signatures
	signatureCount int32 // atomic counter for signatures
	// ids of peers who've sent us this tx (as a map for quick lookups).
	// senders: PeerID -> struct{}
	senders sync.Map
}

func (memTx *mempoolTx) Tx() types.Tx {
	return memTx.tx
}

func (memTx *mempoolTx) Height() int64 {
	return atomic.LoadInt64(&memTx.height)
}

func (memTx *mempoolTx) GasWanted() int64 {
	return memTx.gasWanted
}

func (memTx *mempoolTx) IsSender(peerID nodekey.ID) bool {
	_, ok := memTx.senders.Load(peerID)
	return ok
}
// Add a signature to the list of signatures. // TODO : return true to follow code convention.
func (memTx *mempoolTx) AddSignature(pubKey crypto.PubKey, signature []byte) error {
	// Validate the signature before adding it
	if !pubKey.VerifySignature(memTx.tx, signature) {
		return fmt.Errorf("invalid signature for transaction")
	}

	// Attempt to store the signature
	_, loaded := memTx.signatures.LoadOrStore(pubKey, signature)
	if !loaded {
		// Increment the counter only if it's a new signature
		atomic.AddInt32(&memTx.signatureCount, 1)
	}

	return nil
}

// Validates All Signatures check if each signature is matching with the pubKey
// If a invalidSignatures return an error;
// TODO : Use the right signature system
func (memTx *mempoolTx) ValidateSignatures() error {
	var invalidSignatures []crypto.PubKey

	memTx.signatures.Range(func(key, value interface{}) bool {
		pubKey, ok := key.(crypto.PubKey)
		if !ok {
			return false
		}
		signature, ok := value.([]byte)
		if !ok {
			return false
		}

		if !pubKey.VerifySignature(memTx.tx, signature) {
			invalidSignatures = append(invalidSignatures, pubKey)
		}
		return true
	})

	if len(invalidSignatures) > 0 {
		return fmt.Errorf("invalid signatures found for pubKeys: %v", invalidSignatures)
	}
	return nil
}

// Count the number of signature to check if we reach the number required to stop broadcasting
func (memTx *mempoolTx) SignatureCount() int {
	return int(atomic.LoadInt32(&memTx.signatureCount))
}

// Add the peer ID to the list of senders. Return true iff it exists already in the list.
func (memTx *mempoolTx) addSender(peerID nodekey.ID) bool {
	if len(peerID) == 0 {
		return false
	}
	if _, loaded := memTx.senders.LoadOrStore(peerID, struct{}{}); loaded {
		return true
	}
	return false
}
