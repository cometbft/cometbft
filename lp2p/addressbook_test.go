package lp2p

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestAddressBook(t *testing.T) {
	t.Run("valid address book", func(t *testing.T) {
		// ARRANGE
		// Given 2 private keys
		pk1 := ed25519.GenPrivKey()
		pk2 := ed25519.GenPrivKey()

		pkID := func(pk ed25519.PrivKey) peer.ID {
			id, err := IDFromPrivateKey(pk)
			require.NoError(t, err)
			return id
		}

		// Given an address book
		ab := &AddressBookConfig{
			Peers: []PeerConfig{
				{Host: "127.0.0.1:26656", ID: pkID(pk1).String()},
				{Host: "127.0.0.1:26657", ID: pkID(pk2).String()},
			},
		}

		// ACT
		// Validate the address book
		err := ab.Validate()

		// ASSERT
		require.NoError(t, err)
	})

	t.Run("decode from file", func(t *testing.T) {
		// ARRANGE
		// Given an address book file
		const contents = `
		[[peers]]
		host = "127.0.0.1:26656"
		id = "12D3KooWRqqKwyNnjwukrxXTUXLiNK838WN5tc8Nk2DnMVPbpVPV"
		[[peers]]
		host = "127.0.0.1:26657"
		id = "12D3KooWRqqKwyNnjwukrxXTUXLiNK838WN5tc8Nk2DnMVPbpVPM"
		`

		tempFile := filepath.Join(t.TempDir(), "addressbook.toml")
		err := os.WriteFile(tempFile, []byte(contents), 0o644)
		require.NoError(t, err)

		// ACT
		ab, err := AddressBookFromFilePath(tempFile)

		// ASSERT
		require.NoError(t, err)
		require.NoError(t, ab.Validate())

		require.Equal(t, ab.Peers[0].Host, "127.0.0.1:26656")
		require.Equal(t, ab.Peers[0].ID, "12D3KooWRqqKwyNnjwukrxXTUXLiNK838WN5tc8Nk2DnMVPbpVPV")

		require.Equal(t, ab.Peers[1].Host, "127.0.0.1:26657")
		require.Equal(t, ab.Peers[1].ID, "12D3KooWRqqKwyNnjwukrxXTUXLiNK838WN5tc8Nk2DnMVPbpVPM")
	})
}
