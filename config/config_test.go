package config_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/config"
)

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)

	// set up some defaults
	cfg := config.DefaultConfig()
	assert.NotNil(cfg.P2P)
	assert.NotNil(cfg.Mempool)
	assert.NotNil(cfg.Consensus)

	// check the root dir stuff...
	cfg.SetRoot("/foo")
	cfg.Genesis = "bar"
	cfg.DBPath = "/opt/data"
	cfg.Mempool.WalPath = "wal/mem/"

	assert.Equal("/foo/bar", cfg.GenesisFile())
	assert.Equal("/opt/data", cfg.DBDir())
	assert.Equal("/foo/wal/mem", cfg.Mempool.WalDir())
}

func TestConfigValidateBasic(t *testing.T) {
	cfg := config.DefaultConfig()
	assert.NoError(t, cfg.ValidateBasic())

	// tamper with timeout_propose
	cfg.Consensus.TimeoutPropose = -10 * time.Second
	assert.Error(t, cfg.ValidateBasic())
	cfg.Consensus.TimeoutPropose = 3 * time.Second

	cfg.Consensus.CreateEmptyBlocks = false
	cfg.Mempool.Type = config.MempoolTypeNop
	assert.Error(t, cfg.ValidateBasic())
}

func TestTLSConfiguration(t *testing.T) {
	assert := assert.New(t)
	cfg := config.DefaultConfig()
	cfg.SetRoot("/home/user")

	cfg.RPC.TLSCertFile = "file.crt"
	assert.Equal("/home/user/config/file.crt", cfg.RPC.CertFile())
	cfg.RPC.TLSKeyFile = "file.key"
	assert.Equal("/home/user/config/file.key", cfg.RPC.KeyFile())

	cfg.RPC.TLSCertFile = "/abs/path/to/file.crt"
	assert.Equal("/abs/path/to/file.crt", cfg.RPC.CertFile())
	cfg.RPC.TLSKeyFile = "/abs/path/to/file.key"
	assert.Equal("/abs/path/to/file.key", cfg.RPC.KeyFile())
}

func TestBaseConfigValidateBasic(t *testing.T) {
	cfg := config.TestBaseConfig()
	assert.NoError(t, cfg.ValidateBasic())

	// tamper with log format
	cfg.LogFormat = "invalid"
	assert.Error(t, cfg.ValidateBasic())
}

func TestRPCConfigValidateBasic(t *testing.T) {
	cfg := config.TestRPCConfig()
	assert.NoError(t, cfg.ValidateBasic())

	fieldsToTest := []string{
		"GRPCMaxOpenConnections",
		"MaxOpenConnections",
		"MaxSubscriptionClients",
		"MaxSubscriptionsPerClient",
		"TimeoutBroadcastTxCommit",
		"MaxBodyBytes",
		"MaxHeaderBytes",
		"MaxRequestBatchSize",
	}

	for _, fieldName := range fieldsToTest {
		reflect.ValueOf(cfg).Elem().FieldByName(fieldName).SetInt(-1)
		assert.Error(t, cfg.ValidateBasic())
		reflect.ValueOf(cfg).Elem().FieldByName(fieldName).SetInt(0)
	}
}

func TestP2PConfigValidateBasic(t *testing.T) {
	cfg := config.TestP2PConfig()
	assert.NoError(t, cfg.ValidateBasic())

	fieldsToTest := []string{
		"MaxNumInboundPeers",
		"MaxNumOutboundPeers",
		"FlushThrottleTimeout",
		"MaxPacketMsgPayloadSize",
		"SendRate",
		"RecvRate",
	}

	for _, fieldName := range fieldsToTest {
		reflect.ValueOf(cfg).Elem().FieldByName(fieldName).SetInt(-1)
		assert.Error(t, cfg.ValidateBasic())
		reflect.ValueOf(cfg).Elem().FieldByName(fieldName).SetInt(0)
	}

	t.Run("libp2p", func(t *testing.T) {
		for _, tt := range []struct {
			name        string
			mutate      func(*config.P2PConfig)
			errContains string
		}{
			{
				name: "disabled",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Enabled = false
				},
			},
			{
				name: "disabledWithInvalidLimitsStillPasses",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Enabled = false
					// When LibP2P is disabled, limits validation is skipped
					cfg.LibP2PConfig.Limits.Mode = config.LibP2PLimitsModeCustom
					cfg.LibP2PConfig.Limits.MaxPeers = 0
					cfg.LibP2PConfig.Limits.MaxPeerStreams = 0
				},
			},
			{
				name:   "enabled-default",
				mutate: func(cfg *config.P2PConfig) {},
			},
			{
				name: "allowsEnabledConfigWithEmptyScaler",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Scaler = config.LibP2PScaler{}
				},
			},
			{
				name: "requiresBootstrapPeerHost",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.BootstrapPeers = []config.LibP2PBootstrapPeer{
						{ID: "peer-id"},
					}
				},
				errContains: "p2p.libp2p.bootstrap_peers.0.host is required",
			},
			{
				name: "requiresBootstrapPeerID",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.BootstrapPeers = []config.LibP2PBootstrapPeer{
						{Host: "192.0.2.1:26656"},
					}
				},
				errContains: "p2p.libp2p.bootstrap_peers.0.id is required",
			},
			{
				name: "rejectsNegativeScalerMinWorkers",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Scaler.MinWorkers = -1
				},
				errContains: "p2p.libp2p.scaler.min_workers can't be negative",
			},
			{
				name: "rejectsScalerMinWorkersGreaterThanMax",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Scaler.MinWorkers = 10
					cfg.LibP2PConfig.Scaler.MaxWorkers = 1
				},
				errContains: "invalid field p2p.libp2p.scaler.min_workers must be less than max_workers",
			},
			{
				name: "requiresOverrideReactor",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Scaler.Overrides = []config.LibP2PScalerOverride{
						{MinWorkers: 1, MaxWorkers: 2},
					}
				},
				errContains: "p2p.libp2p.scaler.overrides.0.reactor is required",
			},
			{
				name: "rejectsNegativeOverrideThresholdLatency",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Scaler.Overrides = []config.LibP2PScalerOverride{
						{
							Reactor:          "MEMPOOL",
							MinWorkers:       1,
							MaxWorkers:       2,
							ThresholdLatency: -1,
						},
					}
				},
				errContains: "p2p.libp2p.scaler.overrides.0.threshold_latency can't be negative",
			},
			{
				name: "disabledLimits",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Limits.Mode = config.LibP2PLimitsModeDisabled
				},
			},
			{
				name: "customLimits",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Limits.Mode = config.LibP2PLimitsModeCustom
					cfg.LibP2PConfig.Limits.MaxPeers = 128
					cfg.LibP2PConfig.Limits.MaxPeerStreams = 32
				},
			},
			{
				name: "rejectsCustomLimitsWithZeroMaxPeers",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Limits.Mode = config.LibP2PLimitsModeCustom
					cfg.LibP2PConfig.Limits.MaxPeers = 0
					cfg.LibP2PConfig.Limits.MaxPeerStreams = 1
				},
				errContains: "p2p.libp2p.limits.max_peers is required",
			},
			{
				name: "rejectsCustomLimitsWithZeroMaxPeerStreams",
				mutate: func(cfg *config.P2PConfig) {
					cfg.LibP2PConfig.Limits.Mode = config.LibP2PLimitsModeCustom
					cfg.LibP2PConfig.Limits.MaxPeers = 1
					cfg.LibP2PConfig.Limits.MaxPeerStreams = 0
				},
				errContains: "p2p.libp2p.limits.max_peer_streams is required",
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				// ARRANGE
				cfg := config.TestP2PConfig()
				cfg.LibP2PConfig.Enabled = true

				tt.mutate(cfg)

				// ACT
				err := cfg.ValidateBasic()

				// ASSERT
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
					return
				}
				require.NoError(t, err)
			})
		}
	})
}

func TestMempoolConfigValidateBasic(t *testing.T) {
	cfg := config.TestMempoolConfig()
	assert.NoError(t, cfg.ValidateBasic())

	fieldsToTest := []string{
		"Size",
		"MaxTxsBytes",
		"CacheSize",
		"MaxTxBytes",
	}

	for _, fieldName := range fieldsToTest {
		reflect.ValueOf(cfg).Elem().FieldByName(fieldName).SetInt(-1)
		assert.Error(t, cfg.ValidateBasic())
		reflect.ValueOf(cfg).Elem().FieldByName(fieldName).SetInt(0)
	}

	reflect.ValueOf(cfg).Elem().FieldByName("Type").SetString("invalid")
	assert.Error(t, cfg.ValidateBasic())
}

func TestStateSyncConfigValidateBasic(t *testing.T) {
	cfg := config.TestStateSyncConfig()
	require.NoError(t, cfg.ValidateBasic())
}

func TestBlockSyncConfigValidateBasic(t *testing.T) {
	cfg := config.TestBlockSyncConfig()
	assert.NoError(t, cfg.ValidateBasic())

	// tamper with version
	cfg.Version = "v1"
	assert.Error(t, cfg.ValidateBasic())

	cfg.Version = "invalid"
	assert.Error(t, cfg.ValidateBasic())
}

func TestConsensusConfig_ValidateBasic(t *testing.T) {
	//nolint: lll
	testcases := map[string]struct {
		modify    func(*config.ConsensusConfig)
		expectErr bool
	}{
		"TimeoutPropose":                       {func(c *config.ConsensusConfig) { c.TimeoutPropose = time.Second }, false},
		"TimeoutPropose negative":              {func(c *config.ConsensusConfig) { c.TimeoutPropose = -1 }, true},
		"TimeoutProposeDelta":                  {func(c *config.ConsensusConfig) { c.TimeoutProposeDelta = time.Second }, false},
		"TimeoutProposeDelta negative":         {func(c *config.ConsensusConfig) { c.TimeoutProposeDelta = -1 }, true},
		"TimeoutPrevote":                       {func(c *config.ConsensusConfig) { c.TimeoutPrevote = time.Second }, false},
		"TimeoutPrevote negative":              {func(c *config.ConsensusConfig) { c.TimeoutPrevote = -1 }, true},
		"TimeoutPrevoteDelta":                  {func(c *config.ConsensusConfig) { c.TimeoutPrevoteDelta = time.Second }, false},
		"TimeoutPrevoteDelta negative":         {func(c *config.ConsensusConfig) { c.TimeoutPrevoteDelta = -1 }, true},
		"TimeoutPrecommit":                     {func(c *config.ConsensusConfig) { c.TimeoutPrecommit = time.Second }, false},
		"TimeoutPrecommit negative":            {func(c *config.ConsensusConfig) { c.TimeoutPrecommit = -1 }, true},
		"TimeoutPrecommitDelta":                {func(c *config.ConsensusConfig) { c.TimeoutPrecommitDelta = time.Second }, false},
		"TimeoutPrecommitDelta negative":       {func(c *config.ConsensusConfig) { c.TimeoutPrecommitDelta = -1 }, true},
		"TimeoutCommit":                        {func(c *config.ConsensusConfig) { c.TimeoutCommit = time.Second }, false},
		"TimeoutCommit negative":               {func(c *config.ConsensusConfig) { c.TimeoutCommit = -1 }, true},
		"PeerGossipSleepDuration":              {func(c *config.ConsensusConfig) { c.PeerGossipSleepDuration = time.Second }, false},
		"PeerGossipSleepDuration negative":     {func(c *config.ConsensusConfig) { c.PeerGossipSleepDuration = -1 }, true},
		"PeerQueryMaj23SleepDuration":          {func(c *config.ConsensusConfig) { c.PeerQueryMaj23SleepDuration = time.Second }, false},
		"PeerQueryMaj23SleepDuration negative": {func(c *config.ConsensusConfig) { c.PeerQueryMaj23SleepDuration = -1 }, true},
		"DoubleSignCheckHeight negative":       {func(c *config.ConsensusConfig) { c.DoubleSignCheckHeight = -1 }, true},
		"BlockTimeTolerance":                   {func(c *config.ConsensusConfig) { c.BlockTimeTolerance = time.Second }, false},
		"BlockTimeTolerance zero":              {func(c *config.ConsensusConfig) { c.BlockTimeTolerance = 0 }, true},
		"BlockTimeTolerance negative":          {func(c *config.ConsensusConfig) { c.BlockTimeTolerance = -1 }, true},
	}
	for desc, tc := range testcases {
		// appease linter
		t.Run(desc, func(t *testing.T) {
			cfg := config.DefaultConsensusConfig()
			tc.modify(cfg)

			err := cfg.ValidateBasic()
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstrumentationConfigValidateBasic(t *testing.T) {
	cfg := config.TestInstrumentationConfig()
	assert.NoError(t, cfg.ValidateBasic())

	// tamper with maximum open connections
	cfg.MaxOpenConnections = -1
	assert.Error(t, cfg.ValidateBasic())
}
