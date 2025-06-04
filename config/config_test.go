package config_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/v2/config"
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

	assert.Equal("/foo/bar", cfg.GenesisFile())
	assert.Equal("/opt/data", cfg.DBDir())
}

func TestConfigValidateBasic(t *testing.T) {
	cfg := config.DefaultConfig()
	require.NoError(t, cfg.ValidateBasic())

	// tamper with timeout_propose
	cfg.Consensus.TimeoutPropose = -10 * time.Second
	require.Error(t, cfg.ValidateBasic())
	cfg.Consensus.TimeoutPropose = 3 * time.Second

	cfg.Consensus.CreateEmptyBlocks = false
	cfg.Mempool.Type = config.MempoolTypeNop
	require.Error(t, cfg.ValidateBasic())
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
	require.NoError(t, cfg.ValidateBasic())

	// tamper with log format
	cfg.LogFormat = "invalid"
	require.Error(t, cfg.ValidateBasic())
}

func TestBaseConfigProxyApp_ValidateBasic(t *testing.T) {
	testcases := map[string]struct {
		proxyApp  string
		expectErr bool
	}{
		"empty":                  {"", true},
		"valid":                  {"kvstore", false},
		"invalid static":         {"kvstore1", true},
		"invalid tcp":            {"127.0.0.1", true},
		"invalid tcp with proto": {"tcp://127.0.0.1", true},
		"valid tcp":              {"tcp://127.0.0.1:80", false},
		"invalid proto":          {"unix1://local", true},
		"valid unix":             {"unix://local", false},
	}
	for desc, tc := range testcases {
		t.Run(desc, func(t *testing.T) {
			cfg := config.DefaultBaseConfig()
			cfg.ProxyApp = tc.proxyApp

			err := cfg.ValidateBasic()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRPCConfigValidateBasic(t *testing.T) {
	cfg := config.TestRPCConfig()
	require.NoError(t, cfg.ValidateBasic())

	fieldsToTest := []string{
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
		require.Error(t, cfg.ValidateBasic())
		reflect.ValueOf(cfg).Elem().FieldByName(fieldName).SetInt(0)
	}
}

func TestP2PConfigValidateBasic(t *testing.T) {
	cfg := config.TestP2PConfig()
	require.NoError(t, cfg.ValidateBasic())

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
		require.Error(t, cfg.ValidateBasic())
		reflect.ValueOf(cfg).Elem().FieldByName(fieldName).SetInt(0)
	}
}

func TestMempoolConfigValidateBasic(t *testing.T) {
	cfg := config.TestMempoolConfig()
	require.NoError(t, cfg.ValidateBasic())

	// tamper with type
	reflect.ValueOf(cfg).Elem().FieldByName("Type").SetString("invalid")
	require.Error(t, cfg.ValidateBasic())
	reflect.ValueOf(cfg).Elem().FieldByName("Type").SetString(config.MempoolTypeFlood)
	reflect.ValueOf(cfg).Elem().FieldByName("DOGProtocolEnabled").SetBool(false)

	setFieldTo := func(fieldName string, value int64) {
		reflect.ValueOf(cfg).Elem().FieldByName(fieldName).SetInt(value)
	}

	// tamper with numbers
	fields2values := []struct {
		Name             string
		AllowedValues    []int64
		DisallowedValues []int64
	}{
		{"Size", []int64{1}, []int64{-1, 0}},
		{"MaxTxsBytes", []int64{1}, []int64{-1, 0}},
		{"CacheSize", []int64{0, 1}, []int64{-1}},
		{"MaxTxBytes", []int64{1}, []int64{-1, 0}},
		{"ExperimentalMaxGossipConnectionsToPersistentPeers", []int64{0, 1}, []int64{-1}},
		{"ExperimentalMaxGossipConnectionsToNonPersistentPeers", []int64{0, 1}, []int64{-1}},
	}
	for _, field := range fields2values {
		for _, value := range field.AllowedValues {
			setFieldTo(field.Name, value)
			require.NoError(t, cfg.ValidateBasic())
			setFieldTo(field.Name, 1) // reset
		}

		for _, value := range field.DisallowedValues {
			setFieldTo(field.Name, value)
			require.Error(t, cfg.ValidateBasic())
			setFieldTo(field.Name, 1) // reset
		}
	}

	// with noop mempool, zero values are allowed for the fields below
	reflect.ValueOf(cfg).Elem().FieldByName("Type").SetString(config.MempoolTypeNop)
	fieldNames := []string{
		"Size",
		"MaxTxsBytes",
		"MaxTxBytes",
	}
	for _, name := range fieldNames {
		setFieldTo(name, 0)
		require.NoError(t, cfg.ValidateBasic())
		setFieldTo(name, 1) // reset
	}

	// with DOG protocol only works with Flood and no MaxGossip feature.
	reflect.ValueOf(cfg).Elem().FieldByName("DOGProtocolEnabled").SetBool(true)
	require.Error(t, cfg.ValidateBasic())
	reflect.ValueOf(cfg).Elem().FieldByName("Type").SetString(config.MempoolTypeFlood)
	reflect.ValueOf(cfg).Elem().FieldByName("ExperimentalMaxGossipConnectionsToPersistentPeers").SetInt(0)
	reflect.ValueOf(cfg).Elem().FieldByName("ExperimentalMaxGossipConnectionsToNonPersistentPeers").SetInt(0)
	require.NoError(t, cfg.ValidateBasic())
}

func TestStateSyncConfigValidateBasic(t *testing.T) {
	cfg := config.TestStateSyncConfig()
	require.NoError(t, cfg.ValidateBasic())
}

func TestBlockSyncConfigValidateBasic(t *testing.T) {
	cfg := config.TestBlockSyncConfig()
	require.NoError(t, cfg.ValidateBasic())

	// tamper with version
	cfg.Version = "v1"
	require.Error(t, cfg.ValidateBasic())

	cfg.Version = "invalid"
	require.Error(t, cfg.ValidateBasic())
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
		"TimeoutVote":                          {func(c *config.ConsensusConfig) { c.TimeoutVote = time.Second }, false},
		"TimeoutVote negative":                 {func(c *config.ConsensusConfig) { c.TimeoutVote = -1 }, true},
		"TimeoutVoteDelta":                     {func(c *config.ConsensusConfig) { c.TimeoutVoteDelta = time.Second }, false},
		"TimeoutVoteDelta negative":            {func(c *config.ConsensusConfig) { c.TimeoutVoteDelta = -1 }, true},
		"TimeoutCommit":                        {func(c *config.ConsensusConfig) { c.TimeoutCommit = time.Second }, false},
		"TimeoutCommit negative":               {func(c *config.ConsensusConfig) { c.TimeoutCommit = -1 }, true},
		"PeerGossipSleepDuration":              {func(c *config.ConsensusConfig) { c.PeerGossipSleepDuration = time.Second }, false},
		"PeerGossipSleepDuration negative":     {func(c *config.ConsensusConfig) { c.PeerGossipSleepDuration = -1 }, true},
		"PeerQueryMaj23SleepDuration":          {func(c *config.ConsensusConfig) { c.PeerQueryMaj23SleepDuration = time.Second }, false},
		"PeerQueryMaj23SleepDuration negative": {func(c *config.ConsensusConfig) { c.PeerQueryMaj23SleepDuration = -1 }, true},
		"DoubleSignCheckHeight negative":       {func(c *config.ConsensusConfig) { c.DoubleSignCheckHeight = -1 }, true},
	}
	for desc, tc := range testcases {
		t.Run(desc, func(t *testing.T) {
			cfg := config.DefaultConsensusConfig()
			tc.modify(cfg)

			err := cfg.ValidateBasic()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInstrumentationConfigValidateBasic(t *testing.T) {
	cfg := config.TestInstrumentationConfig()
	require.NoError(t, cfg.ValidateBasic())

	// tamper with maximum open connections
	cfg.MaxOpenConnections = -1
	require.Error(t, cfg.ValidateBasic())
}

func TestConfigPossibleMisconfigurations(t *testing.T) {
	cfg := config.DefaultConfig()
	require.Len(t, cfg.PossibleMisconfigurations(), 0)
	// providing rpc_servers while enable = false is a possible misconfiguration
	cfg.StateSync.RPCServers = []string{"first_rpc"}
	require.Equal(t, []string{"[statesync] section: rpc_servers specified but enable = false"}, cfg.PossibleMisconfigurations())
	// enabling statesync deletes possible misconfiguration
	cfg.StateSync.Enable = true
	require.Len(t, cfg.PossibleMisconfigurations(), 0)
}

func TestStateSyncPossibleMisconfigurations(t *testing.T) {
	cfg := config.DefaultStateSyncConfig()
	require.Len(t, cfg.PossibleMisconfigurations(), 0)
	// providing rpc_servers while enable = false is a possible misconfiguration
	cfg.RPCServers = []string{"first_rpc"}
	require.Equal(t, []string{"rpc_servers specified but enable = false"}, cfg.PossibleMisconfigurations())
	// enabling statesync deletes possible misconfiguration
	cfg.Enable = true
	require.Len(t, cfg.PossibleMisconfigurations(), 0)
}
