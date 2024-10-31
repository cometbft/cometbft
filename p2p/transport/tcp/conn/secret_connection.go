package conn

import (
	"bufio"
	"bytes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/oasisprotocol/curve25519-voi/primitives/merlin"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/box"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/internal/async"
	"github.com/cometbft/cometbft/libs/protoio"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

// 4 + 1024 == 1028 total frame size.
const (
	dataLenSize      = 4
	dataMaxSize      = 1024
	totalFrameSize   = dataMaxSize + dataLenSize
	aeadSizeOverhead = 16 // overhead of poly 1305 authentication tag
	aeadKeySize      = chacha20poly1305.KeySize
	aeadNonceSize    = chacha20poly1305.NonceSize

	labelEphemeralLowerPublicKey = "EPHEMERAL_LOWER_PUBLIC_KEY"
	labelEphemeralUpperPublicKey = "EPHEMERAL_UPPER_PUBLIC_KEY"
	labelDHSecret                = "DH_SECRET"
	labelSecretConnectionMac     = "SECRET_CONNECTION_MAC"

	defaultWriteBufferSize = 128 * 1024
	// try to read the biggest logical packet we can get, in one read.
	// biggest logical packet is encoding_overhead(64kb).
	defaultReadBufferSize = 65 * 1024
)

var (
	ErrSmallOrderRemotePubKey    = errors.New("detected low order point from remote peer")
	secretConnKeyAndChallengeGen = []byte("TENDERMINT_SECRET_CONNECTION_KEY_AND_CHALLENGE_GEN")
)

// SecretConnection implements net.Conn.
// It is an implementation of the STS protocol.
// For more details regarding this implementation of the STS protocol, please refer to:
// https://github.com/cometbft/cometbft/blob/main/spec/p2p/legacy-docs/peer.md#authenticated-encryption-handshake.
//
// The original STS protocol, which inspired this implementation:
// https://citeseerx.ist.psu.edu/document?rapid=rep1&type=pdf&doi=b852bc961328ce74f7231a4b569eec1ab6c3cf50. # codespell:ignore
//
// Consumers of the SecretConnection are responsible for authenticating
// the remote peer's pubkey against known information, like a nodeID.
type SecretConnection struct {
	// immutable
	recvAead cipher.AEAD
	sendAead cipher.AEAD

	remPubKey crypto.PubKey

	conn       io.ReadWriteCloser
	connWriter *bufio.Writer
	connReader io.Reader

	// net.Conn must be thread safe:
	// https://golang.org/pkg/net/#Conn.
	// Since we have internal mutable state,
	// we need mtxs. But recv and send states
	// are independent, so we can use two mtxs.
	// All .Read are covered by recvMtx,
	// all .Write are covered by sendMtx.
	recvMtx         cmtsync.Mutex
	recvBuffer      []byte
	recvNonce       *[aeadNonceSize]byte
	recvFrame       []byte
	recvSealedFrame []byte

	sendMtx         cmtsync.Mutex
	sendNonce       *[aeadNonceSize]byte
	sendFrame       []byte
	sendSealedFrame []byte
}

// MakeSecretConnection performs handshake and returns a new authenticated
// SecretConnection.
// Returns nil if there is an error in handshake.
// Caller should call conn.Close().
func MakeSecretConnection(conn io.ReadWriteCloser, locPrivKey crypto.PrivKey) (*SecretConnection, error) {
	locPubKey := locPrivKey.PubKey()

	// Generate ephemeral keys for perfect forward secrecy.
	locEphPub, locEphPriv := genEphKeys()

	// Write local ephemeral pubkey and receive one too.
	// NOTE: every 32-byte string is accepted as a Curve25519 public key (see
	// DJB's Curve25519 paper: http://cr.yp.to/ecdh/curve25519-20060209.pdf)
	remEphPub, err := shareEphPubKey(conn, locEphPub)
	if err != nil {
		return nil, err
	}

	// Sort by lexical order.
	loEphPub, hiEphPub := sort32(locEphPub, remEphPub)

	transcript := merlin.NewTranscript("TENDERMINT_SECRET_CONNECTION_TRANSCRIPT_HASH")

	transcript.AppendMessage(labelEphemeralLowerPublicKey, loEphPub[:])
	transcript.AppendMessage(labelEphemeralUpperPublicKey, hiEphPub[:])

	// Check if the local ephemeral public key was the least,
	// lexicographically sorted.
	locIsLeast := bytes.Equal(locEphPub[:], loEphPub[:])

	// Compute common diffie hellman secret using X25519.
	dhSecret, err := computeDHSecret(remEphPub, locEphPriv)
	if err != nil {
		return nil, err
	}

	transcript.AppendMessage(labelDHSecret, dhSecret[:])

	// Generate the secret used for receiving, sending, challenge via
	// HKDF-SHA2 on the dhSecret.
	recvSecret, sendSecret := deriveSecrets(dhSecret, locIsLeast)

	const challengeSize = 32
	var challenge [challengeSize]byte
	transcript.ExtractBytes(challenge[:], labelSecretConnectionMac)

	sendAead, err := chacha20poly1305.New(sendSecret[:])
	if err != nil {
		return nil, ErrInvalidSecretConnKeySend
	}

	recvAead, err := chacha20poly1305.New(recvSecret[:])
	if err != nil {
		return nil, ErrInvalidSecretConnKeyRecv
	}

	sc := &SecretConnection{
		conn:            conn,
		connWriter:      bufio.NewWriterSize(conn, defaultWriteBufferSize),
		connReader:      bufio.NewReaderSize(conn, defaultReadBufferSize),
		recvBuffer:      nil,
		recvNonce:       new([aeadNonceSize]byte),
		sendNonce:       new([aeadNonceSize]byte),
		recvAead:        recvAead,
		sendAead:        sendAead,
		recvFrame:       make([]byte, totalFrameSize),
		recvSealedFrame: make([]byte, aeadSizeOverhead+totalFrameSize),
		sendFrame:       make([]byte, totalFrameSize),
		sendSealedFrame: make([]byte, aeadSizeOverhead+totalFrameSize),
	}

	// Sign the challenge bytes for authentication.
	locSignature, err := signChallenge(&challenge, locPrivKey)
	if err != nil {
		return nil, err
	}

	// Share (in secret) each other's pubkey & challenge signature
	authSigMsg, err := shareAuthSignature(sc, locPubKey, locSignature)
	if err != nil {
		return nil, err
	}

	remPubKey, remSignature := authSigMsg.Key, authSigMsg.Sig
	// Usage in your function
	if _, ok := remPubKey.(ed25519.PubKey); !ok {
		return nil, ErrUnexpectedPubKeyType{
			Expected: ed25519.KeyType,
			Got:      remPubKey.Type(),
		}
	}

	if !remPubKey.VerifySignature(challenge[:], remSignature) {
		return nil, ErrChallengeVerification
	}

	// We've authorized.
	sc.remPubKey = remPubKey
	return sc, nil
}

// RemotePubKey returns authenticated remote pubkey.
func (sc *SecretConnection) RemotePubKey() crypto.PubKey {
	return sc.remPubKey
}

// Writes encrypted frames of `totalFrameSize + aeadSizeOverhead`.
// CONTRACT: data smaller than dataMaxSize is written atomically.
func (sc *SecretConnection) Write(data []byte) (n int, err error) {
	sc.sendMtx.Lock()
	defer sc.sendMtx.Unlock()
	sealedFrame, frame := sc.sendSealedFrame, sc.sendFrame

	for 0 < len(data) {
		if err := func() error {
			var chunk []byte
			if dataMaxSize < len(data) {
				chunk = data[:dataMaxSize]
				data = data[dataMaxSize:]
			} else {
				chunk = data
				data = nil
			}
			chunkLength := len(chunk)
			binary.LittleEndian.PutUint32(frame, uint32(chunkLength))
			copy(frame[dataLenSize:], chunk)

			// encrypt the frame
			sc.sendAead.Seal(sealedFrame[:0], sc.sendNonce[:], frame, nil)
			incrNonce(sc.sendNonce)
			// end encryption

			_, err = sc.connWriter.Write(sealedFrame)
			if err != nil {
				return err
			}

			n += len(chunk)
			return nil
		}(); err != nil {
			return n, err
		}
	}
	sc.connWriter.Flush()
	return n, err
}

// CONTRACT: data smaller than dataMaxSize is read atomically.
func (sc *SecretConnection) Read(data []byte) (n int, err error) {
	sc.recvMtx.Lock()
	defer sc.recvMtx.Unlock()

	// read off and update the recvBuffer, if non-empty
	if 0 < len(sc.recvBuffer) {
		n = copy(data, sc.recvBuffer)
		sc.recvBuffer = sc.recvBuffer[n:]
		return n, err
	}

	// read off the conn
	sealedFrame := sc.recvSealedFrame
	_, err = io.ReadFull(sc.connReader, sealedFrame)
	if err != nil {
		return n, err
	}

	// decrypt the frame.
	// reads and updates the sc.recvNonce
	frame := sc.recvFrame
	_, err = sc.recvAead.Open(frame[:0], sc.recvNonce[:], sealedFrame, nil)
	if err != nil {
		return n, ErrDecryptFrame{Source: err}
	}

	incrNonce(sc.recvNonce)
	// end decryption

	// copy checkLength worth into data,
	// set recvBuffer to the rest.
	chunkLength := binary.LittleEndian.Uint32(frame) // read the first four bytes
	if chunkLength > dataMaxSize {
		return 0, ErrChunkTooBig{
			Received: int(chunkLength),
			Max:      dataMaxSize,
		}
	}

	chunk := frame[dataLenSize : dataLenSize+chunkLength]
	n = copy(data, chunk)
	if n < len(chunk) {
		sc.recvBuffer = make([]byte, len(chunk)-n)
		copy(sc.recvBuffer, chunk[n:])
	}
	return n, err
}

// Implements net.Conn.
func (sc *SecretConnection) Close() error                  { return sc.conn.Close() }
func (sc *SecretConnection) LocalAddr() net.Addr           { return sc.conn.(net.Conn).LocalAddr() }
func (sc *SecretConnection) RemoteAddr() net.Addr          { return sc.conn.(net.Conn).RemoteAddr() }
func (sc *SecretConnection) SetDeadline(t time.Time) error { return sc.conn.(net.Conn).SetDeadline(t) }
func (sc *SecretConnection) SetReadDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetReadDeadline(t)
}

func (sc *SecretConnection) SetWriteDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetWriteDeadline(t)
}

func genEphKeys() (ephPub, ephPriv *[32]byte) {
	var err error
	ephPub, ephPriv, err = box.GenerateKey(crand.Reader)
	if err != nil {
		panic("failed to generate ephemeral key-pair")
	}
	return ephPub, ephPriv
}

func shareEphPubKey(conn io.ReadWriter, locEphPub *[32]byte) (remEphPub *[32]byte, err error) {
	// Send our pubkey and receive theirs in tandem.
	trs, _ := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			lc := *locEphPub
			_, err = protoio.NewDelimitedWriter(conn).WriteMsg(&gogotypes.BytesValue{Value: lc[:]})
			if err != nil {
				return nil, true, err // abort
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			var bytes gogotypes.BytesValue
			_, err = protoio.NewDelimitedReader(conn, 1024*1024).ReadMsg(&bytes)
			if err != nil {
				return nil, true, err // abort
			}

			var _remEphPub [32]byte
			copy(_remEphPub[:], bytes.Value)
			return _remEphPub, false, nil
		},
	)

	// If error:
	if trs.FirstError() != nil {
		err = trs.FirstError()
		return remEphPub, err
	}

	// Otherwise:
	_remEphPub := trs.FirstValue().([32]byte)
	return &_remEphPub, nil
}

func deriveSecrets(
	dhSecret *[32]byte,
	locIsLeast bool,
) (recvSecret, sendSecret *[aeadKeySize]byte) {
	hash := sha256.New
	hkdf := hkdf.New(hash, dhSecret[:], nil, secretConnKeyAndChallengeGen)
	// get enough data for 2 aead keys, and a 32 byte challenge
	res := new([2*aeadKeySize + 32]byte)
	_, err := io.ReadFull(hkdf, res[:])
	if err != nil {
		panic(err)
	}

	recvSecret = new([aeadKeySize]byte)
	sendSecret = new([aeadKeySize]byte)

	// bytes 0 through aeadKeySize - 1 are one aead key.
	// bytes aeadKeySize through 2*aeadKeySize -1 are another aead key.
	// which key corresponds to sending and receiving key depends on whether
	// the local key is less than the remote key.
	if locIsLeast {
		copy(recvSecret[:], res[0:aeadKeySize])
		copy(sendSecret[:], res[aeadKeySize:aeadKeySize*2])
	} else {
		copy(sendSecret[:], res[0:aeadKeySize])
		copy(recvSecret[:], res[aeadKeySize:aeadKeySize*2])
	}

	return recvSecret, sendSecret
}

// computeDHSecret computes a Diffie-Hellman shared secret key
// from our own local private key and the other's public key.
func computeDHSecret(remPubKey, locPrivKey *[32]byte) (*[32]byte, error) {
	shrKey, err := curve25519.X25519(locPrivKey[:], remPubKey[:])
	if err != nil {
		return nil, err
	}
	var shrKeyArray [32]byte
	copy(shrKeyArray[:], shrKey)
	return &shrKeyArray, nil
}

func sort32(foo, bar *[32]byte) (lo, hi *[32]byte) {
	if bytes.Compare(foo[:], bar[:]) < 0 {
		lo = foo
		hi = bar
	} else {
		lo = bar
		hi = foo
	}
	return lo, hi
}

func signChallenge(challenge *[32]byte, locPrivKey crypto.PrivKey) ([]byte, error) {
	signature, err := locPrivKey.Sign(challenge[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

type authSigMessage struct {
	Key crypto.PubKey
	Sig []byte
}

func shareAuthSignature(sc io.ReadWriter, pubKey crypto.PubKey, signature []byte) (recvMsg authSigMessage, err error) {
	// Send our info and receive theirs in tandem.
	trs, _ := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			pbpk, err := cryptoenc.PubKeyToProto(pubKey)
			if err != nil {
				return nil, true, err
			}
			_, err = protoio.NewDelimitedWriter(sc).WriteMsg(&tmp2p.AuthSigMessage{PubKey: pbpk, Sig: signature})
			if err != nil {
				return nil, true, err // abort
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			var pba tmp2p.AuthSigMessage
			_, err = protoio.NewDelimitedReader(sc, 1024*1024).ReadMsg(&pba)
			if err != nil {
				return nil, true, err // abort
			}

			pk, err := cryptoenc.PubKeyFromProto(pba.PubKey)
			if err != nil {
				return nil, true, err // abort
			}

			_recvMsg := authSigMessage{
				Key: pk,
				Sig: pba.Sig,
			}
			return _recvMsg, false, nil
		},
	)

	// If error:
	if trs.FirstError() != nil {
		err = trs.FirstError()
		return recvMsg, err
	}

	_recvMsg := trs.FirstValue().(authSigMessage)
	return _recvMsg, nil
}

// --------------------------------------------------------------------------------

// Increment nonce little-endian by 1 with wraparound.
// Due to chacha20poly1305 expecting a 12 byte nonce we do not use the first four
// bytes. We only increment a 64 bit unsigned int in the remaining 8 bytes
// (little-endian in nonce[4:]).
func incrNonce(nonce *[aeadNonceSize]byte) {
	counter := binary.LittleEndian.Uint64(nonce[4:])
	if counter == math.MaxUint64 {
		// Terminates the session and makes sure the nonce would not re-used.
		// See https://github.com/tendermint/tendermint/issues/3531
		panic("can't increase nonce without overflow")
	}
	counter++
	binary.LittleEndian.PutUint64(nonce[4:], counter)
}
