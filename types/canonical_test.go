package types

import (
	"reflect"
	"testing"

	"github.com/tendermint/tendermint/crypto/tmhash"
	cmtrand "github.com/tendermint/tendermint/libs/rand"
	cmtproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestCanonicalizeBlockID(t *testing.T) {
	randhash := cmtrand.Bytes(tmhash.Size)
	block1 := cmtproto.BlockID{Hash: randhash,
		PartSetHeader: cmtproto.PartSetHeader{Total: 5, Hash: randhash}}
	block2 := cmtproto.BlockID{Hash: randhash,
		PartSetHeader: cmtproto.PartSetHeader{Total: 10, Hash: randhash}}
	cblock1 := cmtproto.CanonicalBlockID{Hash: randhash,
		PartSetHeader: cmtproto.CanonicalPartSetHeader{Total: 5, Hash: randhash}}
	cblock2 := cmtproto.CanonicalBlockID{Hash: randhash,
		PartSetHeader: cmtproto.CanonicalPartSetHeader{Total: 10, Hash: randhash}}

	tests := []struct {
		name string
		args cmtproto.BlockID
		want *cmtproto.CanonicalBlockID
	}{
		{"first", block1, &cblock1},
		{"second", block2, &cblock2},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := CanonicalizeBlockID(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CanonicalizeBlockID() = %v, want %v", got, tt.want)
			}
		})
	}
}
