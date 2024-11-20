package types

import "github.com/cometbft/cometbft/crypto"

type TxSignature struct {
    PubKey    crypto.PubKey `json:"pub_key"`
    Signature []byte        `json:"signature"`
}

type TxSignatures []TxSignature
