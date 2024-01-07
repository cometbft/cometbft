package e2e

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Manifest represents a TOML testnet manifest.
type Manifest struct {
	InitialState                                         map[string]string           `toml:"initial_state"`
	Validators                                           *map[string]int64           `toml:"validators"`
	ValidatorUpdates                                     map[string]map[string]int64 `toml:"validator_update"`
	Nodes                                                map[string]*ManifestNode    `toml:"node"`
	ABCIProtocol                                         string                      `toml:"abci_protocol"`
	DefaultZone                                          string                      `toml:"default_zone"`
	UpgradeVersion                                       string                      `toml:"upgrade_version"`
	KeyType                                              string                      `toml:"key_type"`
	LoadTxBatchSize                                      int                         `toml:"load_tx_batch_size"`
	LoadTxSizeBytes                                      int                         `toml:"load_tx_size_bytes"`
	Evidence                                             int                         `toml:"evidence"`
	PrepareProposalDelay                                 time.Duration               `toml:"prepare_proposal_delay"`
	ProcessProposalDelay                                 time.Duration               `toml:"process_proposal_delay"`
	CheckTxDelay                                         time.Duration               `toml:"check_tx_delay"`
	VoteExtensionDelay                                   time.Duration               `toml:"vote_extension_delay"`
	FinalizeBlockDelay                                   time.Duration               `toml:"finalize_block_delay"`
	InitialHeight                                        int64                       `toml:"initial_height"`
	VoteExtensionsEnableHeight                           int64                       `toml:"vote_extensions_enable_height"`
	ExperimentalMaxGossipConnectionsToNonPersistentPeers uint                        `toml:"experimental_max_gossip_connections_to_non_persistent_peers"`
	LoadTxConnections                                    int                         `toml:"load_tx_connections"`
	ExperimentalMaxGossipConnectionsToPersistentPeers    uint                        `toml:"experimental_max_gossip_connections_to_persistent_peers"`
	VoteExtensionSize                                    uint                        `toml:"vote_extension_size"`
	PeerGossipIntraloopSleepDuration                     time.Duration               `toml:"peer_gossip_intraloop_sleep_duration"`
	Prometheus                                           bool                        `toml:"prometheus"`
	IPv6                                                 bool                        `toml:"ipv6"`
	ABCITestsEnabled                                     bool                        `toml:"abci_tests_enabled"`
	DisablePexReactor                                    bool                        `toml:"disable_pex"`
}

// ManifestNode represents a node in a testnet manifest.
type ManifestNode struct {
	PersistInterval        *uint64  `toml:"persist_interval"`
	BlockSyncVersion       string   `toml:"block_sync_version"`
	Version                string   `toml:"version"`
	Zone                   string   `toml:"zone"`
	Database               string   `toml:"database"`
	PrivvalProtocol        string   `toml:"privval_protocol"`
	Mode                   string   `toml:"mode"`
	PersistentPeers        []string `toml:"persistent_peers"`
	Perturb                []string `toml:"perturb"`
	Seeds                  []string `toml:"seeds"`
	StartAt                int64    `toml:"start_at"`
	SnapshotInterval       uint64   `toml:"snapshot_interval"`
	RetainBlocks           uint64   `toml:"retain_blocks"`
	StateSync              bool     `toml:"state_sync"`
	SendNoLoad             bool     `toml:"send_no_load"`
	EnableCompanionPruning bool     `toml:"enable_companion_pruning"`
}

// Save saves the testnet manifest to a file.
func (m Manifest) Save(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create manifest file %q: %w", file, err)
	}
	return toml.NewEncoder(f).Encode(m)
}

// LoadManifest loads a testnet manifest from a file.
func LoadManifest(file string) (Manifest, error) {
	manifest := Manifest{}
	_, err := toml.DecodeFile(file, &manifest)
	if err != nil {
		return manifest, fmt.Errorf("failed to load testnet manifest %q: %w", file, err)
	}
	return manifest, nil
}
