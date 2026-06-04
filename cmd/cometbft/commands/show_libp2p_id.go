package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/lp2p"
	"github.com/cometbft/cometbft/p2p"
)

// ShowLibp2pIDCmd dumps the node's libp2p peer ID (derived from the node key).
// This is the identity presented by the libp2p Noise privval transport and the
// lp2p host.
var ShowLibp2pIDCmd = &cobra.Command{
	Use:     "show-libp2p-id",
	Aliases: []string{"show_libp2p_id"},
	Short:   "Show this node's libp2p peer ID (from the node key)",
	RunE:    showLibp2pID,
}

func showLibp2pID(*cobra.Command, []string) error {
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}
	id, err := lp2p.IDFromPrivateKey(nodeKey.PrivKey)
	if err != nil {
		return err
	}
	fmt.Println(id.String())
	return nil
}
