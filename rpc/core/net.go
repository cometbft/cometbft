package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/p2p"
	na "github.com/cometbft/cometbft/p2p/netaddress"
	ni "github.com/cometbft/cometbft/p2p/nodeinfo"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

// NetInfo returns network info.
// More: https://docs.cometbft.com/main/rpc/#/Info/net_info
func (env *Environment) NetInfo(*rpctypes.Context) (*ctypes.ResultNetInfo, error) {
	peers := make([]ctypes.Peer, 0)
	var err error
	env.P2PPeers.Peers().ForEach(func(peer p2p.Peer) {
		nodeInfo, ok := peer.NodeInfo().(ni.DefaultNodeInfo)
		if !ok {
			err = ErrInvalidNodeType{
				PeerID:   string(peer.ID()),
				Expected: fmt.Sprintf("%T", ni.DefaultNodeInfo{}),
				Actual:   fmt.Sprintf("%T", peer.NodeInfo()),
			}
			return
		}
		peers = append(peers, ctypes.Peer{
			NodeInfo:         nodeInfo,
			IsOutbound:       peer.IsOutbound(),
			ConnectionStatus: peer.Status(),
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
	if len(env.genChunks) > 1 {
		return nil, ErrGenesisRespSize
	}

	return &ctypes.ResultGenesis{Genesis: env.GenDoc}, nil
}

func (env *Environment) GenesisChunked(_ *rpctypes.Context, chunk uint) (*ctypes.ResultGenesisChunk, error) {
	if env.genChunks == nil {
		return nil, ErrServiceConfig{ErrChunkNotInitialized}
	}

	if len(env.genChunks) == 0 {
		return nil, ErrServiceConfig{ErrNoChunks}
	}

	id := int(chunk)

	if id > len(env.genChunks)-1 {
		return nil, ErrInvalidChunkID{id, len(env.genChunks) - 1}
	}

	return &ctypes.ResultGenesisChunk{
		TotalChunks: len(env.genChunks),
		ChunkNumber: id,
		Data:        env.genChunks[id],
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
