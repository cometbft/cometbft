package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/viper"

	db "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/abci/tutorials/abci-v2-forum-app/abci"
	cfg "github.com/cometbft/cometbft/config"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
)

var homeDir string

func init() {
	flag.StringVar(&homeDir, "home", "", "Path to the CometBFT config directory (if empty, uses $HOME/.forumapp)")
}

func main() {
	flag.Parse()
	if homeDir == "" {
		homeDir = os.ExpandEnv("$HOME/.forumapp")
	}

	config := cfg.DefaultConfig()
	config.SetRoot(homeDir)
	viper.SetConfigFile(fmt.Sprintf("%s/%s", homeDir, "config/config.toml"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	store, err := db.NewGoLevelDB(filepath.Join(homeDir, "forum-db"), ".")
	if err != nil {
		log.Printf("failed to create database: %v", err)
		defer os.Exit(1)
	}
	defer store.Close()

	dbPath := "forum-db"
	appConfigPath := "app.toml"
	app, err := abci.NewForumApp(dbPath, appConfigPath)
	if err != nil {
		log.Printf("failed to create ForumApp instance: %v\n", err)
		defer os.Exit(1)
	}

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	logger, err = cmtflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel)
	if err != nil {
		log.Printf("failed to read genesis doc: %v", err)
		defer os.Exit(1)
	}

	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		log.Printf("failed to load node key: %v", err)
		defer os.Exit(1)
	}

	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)

	node, err := nm.NewNode(
		context.Background(),
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger,
	)
	if err != nil {
		log.Printf("failed to create CometBFT node: %v", err)
		defer os.Exit(1)
	}

	if err := node.Start(); err != nil {
		log.Printf("failed to start CometBFT node: %v", err)
		defer os.Exit(1)
	}
	defer func() {
		_ = node.Stop()
		node.Wait()
	}()

	httpAddr := "127.0.0.1:8080"

	server := &http.Server{
		Addr:              httpAddr,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Printf("failed to start HTTP server: %v", err)
		defer os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Forum application stopped")
}
