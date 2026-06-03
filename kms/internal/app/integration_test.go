package app_test

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/privval"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/kms/internal/backend/softsign"
	"github.com/cometbft/cometbft/kms/internal/manager"
	"github.com/cometbft/cometbft/kms/internal/signer"
)

// writeKey writes a softsign key file and returns its path.
func writeKey(t *testing.T, dir string) string {
	t.Helper()
	raw, err := cmtjson.MarshalIndent(struct {
		PrivKey ed25519.PrivKey `json:"priv_key"`
	}{PrivKey: ed25519.GenPrivKey()}, "", "  ")
	require.NoError(t, err)
	p := filepath.Join(dir, "key.json")
	require.NoError(t, os.WriteFile(p, raw, 0o600))
	return p
}

// startListener starts a cometbft validator-side signer listener on ln and
// returns the endpoint. The KMS dials into this.
func startListener(t *testing.T, logger log.Logger, ln net.Listener) *privval.SignerListenerEndpoint {
	t.Helper()
	ep := privval.NewSignerListenerEndpoint(logger, privval.NewTCPListener(ln, ed25519.GenPrivKey()))
	require.NoError(t, ep.Start())
	return ep
}

func TestEndToEndSigning(t *testing.T) {
	const chainID = "integration-chain"
	dir := t.TempDir()
	logger := log.TestingLogger()

	// cometkms (signer) side.
	be, err := softsign.Load(writeKey(t, dir))
	require.NoError(t, err)
	cs, err := signer.NewChainSigner(chainID, be, filepath.Join(dir, "state.json"))
	require.NoError(t, err)

	// validator (listener) side.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	listener := startListener(t, logger, ln)
	defer func() { _ = listener.Stop() }()

	// cometkms Manager dials in.
	mgr := manager.New(logger, []manager.ValidatorConn{{
		ChainID:     chainID,
		Addr:        "tcp://" + addr,
		IdentityKey: ed25519.GenPrivKey(),
		Signer:      cs,
		Reconnect:   true,
	}})
	require.NoError(t, mgr.Start())
	defer mgr.Stop()

	// Drive requests as the validator would.
	client, err := privval.NewSignerClient(listener, chainID)
	require.NoError(t, err)
	require.NoError(t, client.WaitForConnection(5*time.Second))

	pub, err := client.GetPubKey()
	require.NoError(t, err)
	require.NotNil(t, pub)

	// Vote — signature must verify against canonical sign-bytes.
	vote := &cmtproto.Vote{Type: cmtproto.PrevoteType, Height: 5, Round: 0}
	require.NoError(t, client.SignVote(chainID, vote))
	require.True(t, pub.VerifySignature(types.VoteSignBytes(chainID, vote), vote.Signature))

	// Proposal.
	prop := &cmtproto.Proposal{Type: cmtproto.ProposalType, Height: 6, Round: 0}
	require.NoError(t, client.SignProposal(chainID, prop))
	require.True(t, pub.VerifySignature(types.ProposalSignBytes(chainID, prop), prop.Signature))

	// Double-sign: a lower height must be refused.
	lower := &cmtproto.Vote{Type: cmtproto.PrevoteType, Height: 4, Round: 0}
	require.Error(t, client.SignVote(chainID, lower))
}

func TestReconnectAfterListenerRestart(t *testing.T) {
	const chainID = "rc-chain"
	dir := t.TempDir()
	logger := log.TestingLogger()

	be, err := softsign.Load(writeKey(t, dir))
	require.NoError(t, err)
	cs, err := signer.NewChainSigner(chainID, be, filepath.Join(dir, "state.json"))
	require.NoError(t, err)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	listener := startListener(t, logger, ln)

	mgr := manager.New(logger, []manager.ValidatorConn{{
		ChainID: chainID, Addr: "tcp://" + addr, IdentityKey: ed25519.GenPrivKey(), Signer: cs, Reconnect: true,
	}})
	require.NoError(t, mgr.Start())
	defer mgr.Stop()

	client, err := privval.NewSignerClient(listener, chainID)
	require.NoError(t, err)
	require.NoError(t, client.WaitForConnection(5*time.Second))
	_, err = client.GetPubKey()
	require.NoError(t, err)

	// Gracefully stop the validator side (sends FIN -> KMS sees EOF) and free the
	// port. listener.Stop() closes the underlying net.Listener (ln), so we do not
	// close ln again here.
	require.NoError(t, listener.Stop())
	time.Sleep(100 * time.Millisecond)

	// Bring the validator back on the SAME address; the KMS must redial and resume.
	ln2, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	listener2 := startListener(t, logger, ln2)
	defer func() { _ = listener2.Stop() }()

	client2, err := privval.NewSignerClient(listener2, chainID)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		if err := client2.WaitForConnection(500 * time.Millisecond); err != nil {
			return false
		}
		_, err := client2.GetPubKey()
		return err == nil
	}, 20*time.Second, 200*time.Millisecond, "manager did not reconnect after graceful restart")
}

func TestReconnectDisabledStopsAfterDrop(t *testing.T) {
	const chainID = "rc-disabled-chain"
	dir := t.TempDir()
	logger := log.TestingLogger()

	be, err := softsign.Load(writeKey(t, dir))
	require.NoError(t, err)
	cs, err := signer.NewChainSigner(chainID, be, filepath.Join(dir, "state.json"))
	require.NoError(t, err)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	listener := startListener(t, logger, ln)

	mgr := manager.New(logger, []manager.ValidatorConn{{
		ChainID: chainID, Addr: "tcp://" + addr, IdentityKey: ed25519.GenPrivKey(), Signer: cs, Reconnect: false,
	}})
	require.NoError(t, mgr.Start())
	defer mgr.Stop()

	client, err := privval.NewSignerClient(listener, chainID)
	require.NoError(t, err)
	require.NoError(t, client.WaitForConnection(5*time.Second))
	_, err = client.GetPubKey()
	require.NoError(t, err)

	// Gracefully stop the validator side (KMS sees EOF) and free the port.
	require.NoError(t, listener.Stop())
	time.Sleep(100 * time.Millisecond)

	// Bring the validator back on the SAME address.
	ln2, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	listener2 := startListener(t, logger, ln2)
	defer func() { _ = listener2.Stop() }()

	client2, err := privval.NewSignerClient(listener2, chainID)
	require.NoError(t, err)

	// With reconnect disabled, the KMS must NOT redial after the drop.
	require.Never(t, func() bool {
		if err := client2.WaitForConnection(300 * time.Millisecond); err != nil {
			return false
		}
		_, err := client2.GetPubKey()
		return err == nil
	}, 3*time.Second, 300*time.Millisecond)
}
