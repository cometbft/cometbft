package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/version"
)

// VersionCmd ...
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		tmVersion := version.TMCoreSemVer
		if version.TMGitCommitHash != "" {
			tmVersion += "+" + version.TMGitCommitHash
		}

		if verbose {
			values, _ := json.MarshalIndent(struct {
				CometBFT      string `json:"cometbft"`
				ABCI          string `json:"abci"`
				BlockProtocol uint64 `json:"block_protocol"`
				P2PProtocol   uint64 `json:"p2p_protocol"`
			}{
<<<<<<< HEAD:cmd/tendermint/commands/version.go
				Tendermint:    tmVersion,
=======
				CometBFT:      cmtVersion,
>>>>>>> 1cb55d49b (Rename Tendermint to CometBFT: further actions (#224)):cmd/cometbft/commands/version.go
				ABCI:          version.ABCISemVer,
				BlockProtocol: version.BlockProtocol,
				P2PProtocol:   version.P2PProtocol,
			}, "", "  ")
			fmt.Println(string(values))
		} else {
			fmt.Println(tmVersion)
		}
	},
}

func init() {
	VersionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show protocol and library versions")
}
