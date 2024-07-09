package model

import (
	"github.com/cometbft/cometbft/crypto/ed25519"
)

type User struct {
	Name          string         `json:"name"`
	PubKey        ed25519.PubKey `badgerhold:"index"   json:"pubKey"` // this is just a wrapper around bytes
	Moderator     bool           `json:"moderator"`
	Banned        bool           `json:"banned"`
	NumMessages   int64          `json:"numMessages"`
	Version       uint64         `json:"version"`
	SchemaVersion int            `json:"schemaVersion"`
}
