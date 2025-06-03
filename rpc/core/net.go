package core

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	cmtjson "github.com/cometbft/cometbft/v2/libs/json"
	"github.com/cometbft/cometbft/v2/p2p"
	na "github.com/cometbft/cometbft/v2/p2p/netaddr"
	ctypes "github.com/cometbft/cometbft/v2/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/v2/rpc/jsonrpc/types"
	"github.com/cometbft/cometbft/v2/types"
)

// NetInfo returns network info.
// More: https://docs.cometbft.com/main/rpc/#/Info/net_info
func (env *Environment) NetInfo(*rpctypes.Context) (*ctypes.ResultNetInfo, error) {
	peers := make([]ctypes.Peer, 0)
	var err error
	env.P2PPeers.Peers().ForEach(func(peer p2p.Peer) {
		nodeInfo, ok := peer.NodeInfo().(p2p.NodeInfoDefault)
		if !ok {
			err = ErrInvalidNodeType{
				PeerID:   peer.ID(),
				Expected: fmt.Sprintf("%T", p2p.NodeInfoDefault{}),
				Actual:   fmt.Sprintf("%T", peer.NodeInfo()),
			}
			return
		}
		peers = append(peers, ctypes.Peer{
			NodeInfo:         nodeInfo,
			IsOutbound:       peer.IsOutbound(),
			ConnectionStatus: peer.ConnState(),
			RemoteIP:         peer.RemoteIP().String(),
		})
	})
	if err != nil {
		return nil, err
	}
	// TODO: Should we include PersistentPeers and Seeds in here?
	// PRO: useful info
	// CON: privacy
	return &ctypes.ResultNetInfo{
		Listening: env.P2PTransport.IsListening(),
		Listeners: env.P2PTransport.Listeners(),
		NPeers:    len(peers),
		Peers:     peers,
	}, nil
}

// UnsafeDialSeeds dials the given seeds (comma-separated id@IP:PORT).
func (env *Environment) UnsafeDialSeeds(_ *rpctypes.Context, seeds []string) (*ctypes.ResultDialSeeds, error) {
	if len(seeds) == 0 {
		return &ctypes.ResultDialSeeds{}, errors.New("no seeds provided")
	}
	env.Logger.Info("DialSeeds", "seeds", seeds)
	if err := env.P2PPeers.DialPeersAsync(seeds); err != nil {
		return &ctypes.ResultDialSeeds{}, err
	}
	return &ctypes.ResultDialSeeds{Log: "Dialing seeds in progress. See /net_info for details"}, nil
}

// UnsafeDialPeers dials the given peers (comma-separated id@IP:PORT),
// optionally making them persistent.
func (env *Environment) UnsafeDialPeers(
	_ *rpctypes.Context,
	peers []string,
	persistent, unconditional, private bool,
) (*ctypes.ResultDialPeers, error) {
	if len(peers) == 0 {
		return &ctypes.ResultDialPeers{}, errors.New("no peers provided")
	}

	ids, err := getIDs(peers)
	if err != nil {
		return &ctypes.ResultDialPeers{}, err
	}

	env.Logger.Info("DialPeers", "peers", peers, "persistent",
		persistent, "unconditional", unconditional, "private", private)

	if persistent {
		if err := env.P2PPeers.AddPersistentPeers(peers); err != nil {
			return &ctypes.ResultDialPeers{}, err
		}
	}

	if private {
		if err := env.P2PPeers.AddPrivatePeerIDs(ids); err != nil {
			return &ctypes.ResultDialPeers{}, err
		}
	}

	if unconditional {
		if err := env.P2PPeers.AddUnconditionalPeerIDs(ids); err != nil {
			return &ctypes.ResultDialPeers{}, err
		}
	}

	if err := env.P2PPeers.DialPeersAsync(peers); err != nil {
		return &ctypes.ResultDialPeers{}, err
	}

	return &ctypes.ResultDialPeers{Log: "Dialing peers in progress. See /net_info for details"}, nil
}

// Genesis returns genesis file.
// More: https://docs.cometbft.com/main/rpc/#/Info/genesis
func (env *Environment) Genesis(*rpctypes.Context) (*ctypes.ResultGenesis, error) {
	if len(env.genesisChunksFiles) > 0 {
		return nil, ErrGenesisRespSize
	}

	fGenesis, err := os.ReadFile(env.GenesisFilePath)
	if err != nil {
		return nil, fmt.Errorf("retrieving genesis file from disk: %s", err)
	}

	genDoc := types.GenesisDoc{}
	if err = cmtjson.Unmarshal(fGenesis, &genDoc); err != nil {
		formatStr := "genesis file JSON format is invalid: %s"
		return nil, fmt.Errorf(formatStr, err)
	}

	return &ctypes.ResultGenesis{Genesis: &genDoc}, nil
}

func (env *Environment) GenesisChunked(
	_ *rpctypes.Context,
	chunkID uint,
) (*ctypes.ResultGenesisChunk, error) {
	if len(env.genesisChunksFiles) == 0 {
		// See discussion in the following PR for why we still serve chunk 0 even
		// if env.genChunks is nil:
		// https://github.com/cometbft/cometbft/v2/pull/4235#issuecomment-2389109521
		if chunkID == 0 {
			fGenesis, err := os.ReadFile(env.GenesisFilePath)
			if err != nil {
				return nil, fmt.Errorf("retrieving genesis file from disk: %w", err)
			}

			genesisBase64 := base64.StdEncoding.EncodeToString(fGenesis)

			resp := &ctypes.ResultGenesisChunk{
				TotalChunks: 1,
				ChunkNumber: 0,
				Data:        genesisBase64,
			}

			return resp, nil
		}

		return nil, ErrServiceConfig{ErrNoChunks}
	}

	id := int(chunkID)

	if id > len(env.genesisChunksFiles)-1 {
		return nil, ErrInvalidChunkID{id, len(env.genesisChunksFiles) - 1}
	}

	chunkPath := env.genesisChunksFiles[id]
	chunk, err := os.ReadFile(chunkPath)
	if err != nil {
		return nil, fmt.Errorf("retrieving chunk %d from disk: %w", id, err)
	}

	chunkBase64 := base64.StdEncoding.EncodeToString(chunk)

	return &ctypes.ResultGenesisChunk{
		TotalChunks: len(env.genesisChunksFiles),
		ChunkNumber: id,
		Data:        chunkBase64,
	}, nil
}

func getIDs(peers []string) ([]string, error) {
	ids := make([]string, 0, len(peers))

	for _, peer := range peers {
		spl := strings.Split(peer, "@")
		if len(spl) != 2 {
			return nil, na.ErrNoID{Addr: peer}
		}
		ids = append(ids, spl[0])
	}
	return ids, nil
}
