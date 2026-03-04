package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/lp2p"
	"github.com/cometbft/cometbft/p2p"
)

var showNodeIDAsLibP2P bool

func init() {
	ShowNodeIDCmd.Flags().BoolVar(&showNodeIDAsLibP2P, "libp2p", false, "show node ID as libp2p peer ID")
}

// ShowNodeIDCmd dumps node's ID to the standard output.
var ShowNodeIDCmd = &cobra.Command{
	Use:     "show-node-id",
	Aliases: []string{"show_node_id"},
	Short:   "Show this node's ID",
	RunE:    showNodeID,
}

func showNodeID(*cobra.Command, []string) error {
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}

	if showNodeIDAsLibP2P {
		id, err := lp2p.IDFromPrivateKey(nodeKey.PrivKey)
		if err != nil {
			return err
		}

		fmt.Println(id.String())
		return nil
	}

	fmt.Println(nodeKey.ID())
	return nil
}
