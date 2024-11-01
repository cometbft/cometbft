package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	_ "embed"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/privval"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra"
	"github.com/cometbft/cometbft/types"
)

const (
	AppAddressTCP  = "tcp://127.0.0.1:30000"
	AppAddressUNIX = "unix:///var/run/app.sock"

	PrivvalAddressTCP     = "tcp://0.0.0.0:27559"
	PrivvalAddressUNIX    = "unix:///var/run/privval.sock"
	PrivvalKeyFile        = "config/priv_validator_key.json"
	PrivvalStateFile      = "data/priv_validator_state.json"
	PrivvalDummyKeyFile   = "config/dummy_validator_key.json"
	PrivvalDummyStateFile = "data/dummy_validator_state.json"

	PrometheusConfigFile = "monitoring/prometheus.yml"
)

// Setup sets up the testnet configuration.
func Setup(testnet *e2e.Testnet, infp infra.Provider) error {
	logger.Info("setup", "msg", log.NewLazySprintf("Generating testnet files in %#q", testnet.Dir))

	if err := os.MkdirAll(testnet.Dir, os.ModePerm); err != nil {
		return err
	}

	genesis, err := MakeGenesis(testnet)
	if err != nil {
		return err
	}

	for _, node := range testnet.Nodes {
		nodeDir := filepath.Join(testnet.Dir, node.Name)

		dirs := []string{
			filepath.Join(nodeDir, "config"),
			filepath.Join(nodeDir, "data"),
			filepath.Join(nodeDir, "data", "app"),
		}
		for _, dir := range dirs {
			// light clients don't need an app directory
			if node.Mode == e2e.ModeLight && strings.Contains(dir, "app") {
				continue
			}
			err := os.MkdirAll(dir, 0o755)
			if err != nil {
				return err
			}
		}

		cfg, err := MakeConfig(node)
		if err != nil {
			return err
		}
		config.WriteConfigFile(filepath.Join(nodeDir, "config", "config.toml"), cfg) // panics

		appCfg, err := MakeAppConfig(node)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(nodeDir, "config", "app.toml"), appCfg, 0o644) //nolint:gosec
		if err != nil {
			return err
		}

		if node.Mode == e2e.ModeLight {
			// stop early if a light client
			continue
		}

		err = genesis.SaveAs(filepath.Join(nodeDir, "config", "genesis.json"))
		if err != nil {
			return err
		}

		err = (&nodekey.NodeKey{PrivKey: node.NodeKey}).SaveAs(filepath.Join(nodeDir, "config", "node_key.json"))
		if err != nil {
			return err
		}

		(privval.NewFilePV(node.PrivvalKey,
			filepath.Join(nodeDir, PrivvalKeyFile),
			filepath.Join(nodeDir, PrivvalStateFile),
		)).Save()

		// Set up a dummy validator. CometBFT requires a file PV even when not used, so we
		// give it a dummy such that it will fail if it actually tries to use it.
		(privval.NewFilePV(ed25519.GenPrivKey(),
			filepath.Join(nodeDir, PrivvalDummyKeyFile),
			filepath.Join(nodeDir, PrivvalDummyStateFile),
		)).Save()

		// Generate a shell script file containing tc (traffic control) commands
		// to emulate latency to other nodes.
		tcCmds, err := tcCommands(node, infp)
		if err != nil {
			return err
		}
		latencyPath := filepath.Join(nodeDir, "emulate-latency.sh")
		//nolint: gosec // G306: Expect WriteFile permissions to be 0600 or less
		if err = os.WriteFile(latencyPath, []byte(strings.Join(tcCmds, "\n")), 0o755); err != nil {
			return err
		}
	}

	if testnet.Prometheus {
		if err := WritePrometheusConfig(testnet, PrometheusConfigFile); err != nil {
			return err
		}
		// Make a copy of the Prometheus config file in the testnet directory.
		// This should be temporary to keep it compatible with the qa-infra
		// repository.
		if err := WritePrometheusConfig(testnet, filepath.Join(testnet.Dir, "prometheus.yml")); err != nil {
			return err
		}
	}

	//nolint: revive
	if err := infp.Setup(); err != nil {
		return err
	}

	return nil
}

// MakeGenesis generates a genesis document.
func MakeGenesis(testnet *e2e.Testnet) (types.GenesisDoc, error) {
	genesis := types.GenesisDoc{
		GenesisTime:     time.Now(),
		ChainID:         testnet.Name,
		ConsensusParams: types.DefaultConsensusParams(),
		InitialHeight:   testnet.InitialHeight,
	}
	// set the app version to 1
	genesis.ConsensusParams.Version.App = 1
	genesis.ConsensusParams.Evidence.MaxAgeNumBlocks = e2e.EvidenceAgeHeight
	genesis.ConsensusParams.Evidence.MaxAgeDuration = e2e.EvidenceAgeTime
	genesis.ConsensusParams.Validator.PubKeyTypes = []string{testnet.KeyType}
	if testnet.BlockMaxBytes != 0 {
		genesis.ConsensusParams.Block.MaxBytes = testnet.BlockMaxBytes
	}
	if testnet.VoteExtensionsUpdateHeight == -1 {
		genesis.ConsensusParams.Feature.VoteExtensionsEnableHeight = testnet.VoteExtensionsEnableHeight
	}
	if testnet.PbtsUpdateHeight == -1 {
		genesis.ConsensusParams.Feature.PbtsEnableHeight = testnet.PbtsEnableHeight
	}
	for valName, power := range testnet.Validators {
		validator := testnet.LookupNode(valName)
		if validator == nil {
			return types.GenesisDoc{}, fmt.Errorf("unknown validator %q for genesis doc", valName)
		}
		genesis.Validators = append(genesis.Validators, types.GenesisValidator{
			Name:    valName,
			Address: validator.PrivvalKey.PubKey().Address(),
			PubKey:  validator.PrivvalKey.PubKey(),
			Power:   power,
		})
	}
	// The validator set will be sorted internally by CometBFT ranked by power,
	// but we sort it here as well so that all genesis files are identical.
	sort.Slice(genesis.Validators, func(i, j int) bool {
		return strings.Compare(genesis.Validators[i].Name, genesis.Validators[j].Name) == -1
	})
	if len(testnet.InitialState) > 0 {
		appState, err := json.Marshal(testnet.InitialState)
		if err != nil {
			return types.GenesisDoc{}, err
		}
		genesis.AppState = appState
	}

	// Customized genesis fields provided in the manifest
	if len(testnet.Genesis) > 0 {
		v := viper.New()
		v.SetConfigType("json")

		for _, field := range testnet.Genesis {
			key, value, err := e2e.ParseKeyValueField("genesis", field)
			if err != nil {
				return types.GenesisDoc{}, err
			}
			logger.Debug("Applying 'genesis' field", key, value)
			v.Set(key, value)
		}

		// We use viper because it leaves untouched keys that are not set.
		// The GenesisDoc does not use the original `mapstructure` tag.
		err := v.Unmarshal(&genesis, func(d *mapstructure.DecoderConfig) {
			d.TagName = "json"
			d.ErrorUnused = true
		})
		if err != nil {
			return types.GenesisDoc{}, fmt.Errorf("failed parsing 'genesis' field: %v", err)
		}
	}

	if err := genesis.ValidateAndComplete(); err != nil {
		return types.GenesisDoc{}, err
	}
	return genesis, nil
}

// MakeConfig generates a CometBFT config for a node.
func MakeConfig(node *e2e.Node) (*config.Config, error) {
	cfg := config.DefaultConfig()
	cfg.Moniker = node.Name
	cfg.ProxyApp = AppAddressTCP

	cfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	cfg.RPC.PprofListenAddress = ":6060"

	cfg.GRPC.ListenAddress = "tcp://0.0.0.0:26670"
	cfg.GRPC.VersionService.Enabled = true
	cfg.GRPC.BlockService.Enabled = true
	cfg.GRPC.BlockResultsService.Enabled = true

	cfg.P2P.ExternalAddress = fmt.Sprintf("tcp://%v", node.AddressP2P(false))
	cfg.P2P.AddrBookStrict = false

	cfg.DBBackend = node.Database
	cfg.BlockSync.Version = node.BlockSyncVersion
	cfg.Consensus.PeerGossipIntraloopSleepDuration = node.Testnet.PeerGossipIntraloopSleepDuration
	cfg.Mempool.ExperimentalMaxGossipConnectionsToNonPersistentPeers = int(node.Testnet.ExperimentalMaxGossipConnectionsToNonPersistentPeers)
	cfg.Mempool.ExperimentalMaxGossipConnectionsToPersistentPeers = int(node.Testnet.ExperimentalMaxGossipConnectionsToPersistentPeers)

	// Assume that full nodes and validators will have a data companion
	// attached, which will need access to the privileged gRPC endpoint.
	if (node.Mode == e2e.ModeValidator || node.Mode == e2e.ModeFull) && node.EnableCompanionPruning {
		cfg.Storage.Pruning.DataCompanion.Enabled = true
		cfg.Storage.Pruning.DataCompanion.InitialBlockRetainHeight = 0
		cfg.Storage.Pruning.DataCompanion.InitialBlockResultsRetainHeight = 0
		cfg.GRPC.Privileged.ListenAddress = "tcp://0.0.0.0:26671"
		cfg.GRPC.Privileged.PruningService.Enabled = true
	}

	switch node.ABCIProtocol {
	case e2e.ProtocolUNIX:
		cfg.ProxyApp = AppAddressUNIX
	case e2e.ProtocolTCP:
		cfg.ProxyApp = AppAddressTCP
	case e2e.ProtocolGRPC:
		cfg.ProxyApp = AppAddressTCP
		cfg.ABCI = "grpc"
	case e2e.ProtocolBuiltin:
		cfg.ProxyApp = "e2e"
		cfg.ABCI = ""
	case e2e.ProtocolBuiltinConnSync:
		cfg.ProxyApp = "e2e_connsync"
		cfg.ABCI = ""
	default:
		return nil, fmt.Errorf("unexpected ABCI protocol setting %q", node.ABCIProtocol)
	}

	// CometBFT errors if it does not have a privval key set up, regardless of whether
	// it's actually needed (e.g. for remote KMS or non-validators). We set up a dummy
	// key here by default, and use the real key for actual validators that should use
	// the file privval.
	cfg.PrivValidatorListenAddr = ""
	cfg.PrivValidatorKey = PrivvalDummyKeyFile
	cfg.PrivValidatorState = PrivvalDummyStateFile

	switch node.Mode {
	case e2e.ModeValidator:
		switch node.PrivvalProtocol {
		case e2e.ProtocolFile:
			cfg.PrivValidatorKey = PrivvalKeyFile
			cfg.PrivValidatorState = PrivvalStateFile
		case e2e.ProtocolUNIX:
			cfg.PrivValidatorListenAddr = PrivvalAddressUNIX
		case e2e.ProtocolTCP:
			cfg.PrivValidatorListenAddr = PrivvalAddressTCP
		default:
			return nil, fmt.Errorf("invalid privval protocol setting %q", node.PrivvalProtocol)
		}
	case e2e.ModeSeed:
		cfg.P2P.SeedMode = true
		cfg.P2P.PexReactor = true
	case e2e.ModeFull, e2e.ModeLight:
		// Don't need to do anything, since we're using a dummy privval key by default.
	default:
		return nil, fmt.Errorf("unexpected mode %q", node.Mode)
	}

	if node.StateSync {
		cfg.StateSync.Enable = true
		cfg.StateSync.RPCServers = []string{}
		for _, peer := range node.Testnet.ArchiveNodes() {
			if peer.Name == node.Name {
				continue
			}
			cfg.StateSync.RPCServers = append(cfg.StateSync.RPCServers, peer.AddressRPC())
		}
		if len(cfg.StateSync.RPCServers) < 2 {
			return nil, errors.New("unable to find 2 suitable state sync RPC servers")
		}
		cfg.StateSync.MaxDiscoveryTime = 30 * time.Second
	}

	cfg.P2P.Seeds = ""
	for _, seed := range node.Seeds {
		if len(cfg.P2P.Seeds) > 0 {
			cfg.P2P.Seeds += ","
		}
		cfg.P2P.Seeds += seed.AddressP2P(true)
	}
	cfg.P2P.PersistentPeers = ""
	for _, peer := range node.PersistentPeers {
		if len(cfg.P2P.PersistentPeers) > 0 {
			cfg.P2P.PersistentPeers += ","
		}
		cfg.P2P.PersistentPeers += peer.AddressP2P(true)
	}
	if node.Testnet.DisablePexReactor {
		cfg.P2P.PexReactor = false
	}

	if node.Testnet.LogLevel != "" {
		cfg.LogLevel = node.Testnet.LogLevel
	}

	if node.Testnet.LogFormat != "" {
		cfg.LogFormat = node.Testnet.LogFormat
	}

	if node.Prometheus {
		cfg.Instrumentation.Prometheus = true
	}

	if node.ExperimentalKeyLayout != "" {
		cfg.Storage.ExperimentalKeyLayout = node.ExperimentalKeyLayout
	}

	if node.Compact {
		cfg.Storage.Compact = node.Compact
	}

	if node.DiscardABCIResponses {
		cfg.Storage.DiscardABCIResponses = node.DiscardABCIResponses
	}

	if node.Indexer != "" {
		cfg.TxIndex.Indexer = node.Indexer
	}

	if node.CompactionInterval != 0 && node.Compact {
		cfg.Storage.CompactionInterval = node.CompactionInterval
	}

	// We currently need viper in order to parse config files.
	if len(node.Config) > 0 {
		v := viper.New()
		for _, field := range node.Config {
			key, value, err := e2e.ParseKeyValueField("config", field)
			if err != nil {
				return nil, err
			}
			logger.Debug("Applying 'config' field", "node", node.Name, key, value)
			v.Set(key, value)
		}
		err := v.Unmarshal(cfg, func(d *mapstructure.DecoderConfig) {
			d.ErrorUnused = true
		})
		if err != nil {
			return nil, fmt.Errorf("failed parsing 'config' field of node %v: %v", node.Name, err)
		}
	}

	return cfg, nil
}

// MakeAppConfig generates an ABCI application config for a node.
func MakeAppConfig(node *e2e.Node) ([]byte, error) {
	cfg := map[string]any{
		"chain_id":                      node.Testnet.Name,
		"dir":                           "data/app",
		"listen":                        AppAddressUNIX,
		"mode":                          node.Mode,
		"protocol":                      "socket",
		"persist_interval":              node.PersistInterval,
		"snapshot_interval":             node.SnapshotInterval,
		"retain_blocks":                 node.RetainBlocks,
		"key_type":                      node.PrivvalKey.Type(),
		"prepare_proposal_delay":        node.Testnet.PrepareProposalDelay,
		"process_proposal_delay":        node.Testnet.ProcessProposalDelay,
		"check_tx_delay":                node.Testnet.CheckTxDelay,
		"vote_extension_delay":          node.Testnet.VoteExtensionDelay,
		"finalize_block_delay":          node.Testnet.FinalizeBlockDelay,
		"vote_extension_size":           node.Testnet.VoteExtensionSize,
		"vote_extensions_enable_height": node.Testnet.VoteExtensionsEnableHeight,
		"vote_extensions_update_height": node.Testnet.VoteExtensionsUpdateHeight,
		"abci_requests_logging_enabled": node.Testnet.ABCITestsEnabled,
		"pbts_enable_height":            node.Testnet.PbtsEnableHeight,
		"pbts_update_height":            node.Testnet.PbtsUpdateHeight,
		"no_lanes":                      node.Testnet.Manifest.NoLanes,
		"lanes":                         node.Testnet.Manifest.Lanes,
		"constant_flip":                 node.Testnet.ConstantFlip,
	}
	switch node.ABCIProtocol {
	case e2e.ProtocolUNIX:
		cfg["listen"] = AppAddressUNIX
	case e2e.ProtocolTCP:
		cfg["listen"] = AppAddressTCP
	case e2e.ProtocolGRPC:
		cfg["listen"] = AppAddressTCP
		cfg["protocol"] = "grpc"
	case e2e.ProtocolBuiltin, e2e.ProtocolBuiltinConnSync:
		delete(cfg, "listen")
		cfg["protocol"] = string(node.ABCIProtocol)
	default:
		return nil, fmt.Errorf("unexpected ABCI protocol setting %q", node.ABCIProtocol)
	}
	if node.Mode == e2e.ModeValidator {
		switch node.PrivvalProtocol {
		case e2e.ProtocolFile:
		case e2e.ProtocolTCP:
			cfg["privval_server"] = PrivvalAddressTCP
			cfg["privval_key"] = PrivvalKeyFile
			cfg["privval_state"] = PrivvalStateFile
		case e2e.ProtocolUNIX:
			cfg["privval_server"] = PrivvalAddressUNIX
			cfg["privval_key"] = PrivvalKeyFile
			cfg["privval_state"] = PrivvalStateFile
		default:
			return nil, fmt.Errorf("unexpected privval protocol setting %q", node.PrivvalProtocol)
		}
	}

	if len(node.Testnet.ValidatorUpdates) > 0 {
		validatorUpdates := map[string]map[string]int64{}
		for height, validators := range node.Testnet.ValidatorUpdates {
			updateVals := map[string]int64{}
			for valName, power := range validators {
				validator := node.Testnet.LookupNode(valName)
				if validator == nil {
					return nil, fmt.Errorf("unknown validator %q for validator updates in testnet, height %d", valName, height)
				}
				updateVals[base64.StdEncoding.EncodeToString(validator.PrivvalKey.PubKey().Bytes())] = power
			}
			validatorUpdates[strconv.FormatInt(height, 10)] = updateVals
		}
		cfg["validator_update"] = validatorUpdates
	}

	var buf bytes.Buffer
	err := toml.NewEncoder(&buf).Encode(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app config: %w", err)
	}
	return buf.Bytes(), nil
}

//go:embed templates/prometheus-yml.tmpl
var prometheusYamlTemplate string

func WritePrometheusConfig(testnet *e2e.Testnet, path string) error {
	tmpl, err := template.New("prometheus-yaml").Parse(prometheusYamlTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, testnet)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, buf.Bytes(), 0o644) //nolint:gosec
	if err != nil {
		return err
	}
	return nil
}

// UpdateConfigStateSync updates the state sync config for a node.
func UpdateConfigStateSync(node *e2e.Node, height int64, hash []byte) error {
	cfgPath := filepath.Join(node.Testnet.Dir, node.Name, "config", "config.toml")

	// FIXME Apparently there's no function to simply load a config file without
	// involving the entire Viper apparatus, so we'll just resort to regexps.
	bz, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}
	bz = regexp.MustCompile(`(?m)^trust_height =.*`).ReplaceAll(bz, []byte(fmt.Sprintf(`trust_height = %v`, height)))
	bz = regexp.MustCompile(`(?m)^trust_hash =.*`).ReplaceAll(bz, []byte(fmt.Sprintf(`trust_hash = "%X"`, hash)))
	return os.WriteFile(cfgPath, bz, 0o644) //nolint:gosec
}

// tcCommands generates a list of tc (traffic control) commands to emulate
// latency from the node to all other nodes.
func tcCommands(node *e2e.Node, infp infra.Provider) ([]string, error) {
	allZones, zoneMatrix, err := e2e.LoadZoneLatenciesMatrix()
	if err != nil {
		return nil, err
	}
	nodeZoneIndex := slices.Index(allZones, node.Zone)

	tcCmds := []string{
		"#!/bin/sh",

		// Delete any existing qdisc on the root of the eth0 interface.
		"tc qdisc del dev eth0 root 2> /dev/null",

		// Add a new root qdisc of type HTB with a default class of 10.
		"tc qdisc add dev eth0 root handle 1: htb default 10",

		// Add a root class with identifier 1:1 and a rate limit of 1 gigabit per second.
		"tc class add dev eth0 parent 1: classid 1:1 htb rate 1gbit 2> /dev/null",

		// Add a default class under the root class with identifier 1:10 and a rate limit of 1 gigabit per second.
		"tc class add dev eth0 parent 1:1 classid 1:10 htb rate 1gbit 2> /dev/null",

		// Add an SFQ qdisc to the default class with handle 10: to manage traffic with fairness.
		"tc qdisc add dev eth0 parent 1:10 handle 10: sfq perturb 10",
	}

	// handle must be unique for each rule; start from one higher than last handle used above (10).
	handle := 11
	for _, targetZone := range allZones {
		// Get latency from node's zone to target zone (note that the matrix is symmetric).
		latency := zoneMatrix[targetZone][nodeZoneIndex]
		if latency <= 0 {
			continue
		}

		// Assign latency +/- 0.05% to handle.
		delta := latency / 20
		if delta == 0 {
			// Zero is not allowed in normal distribution.
			delta = 1
		}

		// Add a class with the calculated handle, under the root class, with the specified rate.
		tcCmds = append(tcCmds, fmt.Sprintf("tc class add dev eth0 parent 1:1 classid 1:%d htb rate 1gbit 2> /dev/null", handle))

		// Add a netem qdisc to simulate the specified delay with normal distribution.
		tcCmds = append(tcCmds, fmt.Sprintf("tc qdisc add dev eth0 parent 1:%d handle %d: netem delay %dms %dms distribution normal", handle, handle, latency, delta))

		// Set emulated latency to nodes in the target zone.
		for _, otherNode := range node.Testnet.Nodes {
			if otherNode.Zone == targetZone || node.Name == otherNode.Name {
				continue
			}
			otherNodeIP := infp.NodeIP(otherNode)
			// Assign latency handle to target node.
			tcCmds = append(tcCmds, fmt.Sprintf("tc filter add dev eth0 protocol ip parent 1: prio 1 u32 match ip dst %s/32 flowid 1:%d", otherNodeIP, handle))
		}

		handle++
	}

	// Display tc configuration for debugging.
	tcCmds = append(tcCmds, []string{
		fmt.Sprintf("echo Traffic Control configuration on %s:", node.Name),
		"tc qdisc show",
		"tc class show dev eth0",
		// "tc filter show dev eth0", // too verbose
	}...)

	return tcCmds, nil
}
