package config

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/internal/storage"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
)

// ServiceProvider takes a config and a logger and returns a ready to go Node.
type ServiceProvider func(
	context.Context,
	*Config,
	log.Logger,
) (service.Service, error)

// DBContext specifies config information for loading a new DB.
type DBContext struct {
	ID     string
	Config *Config
}

// DBProvider takes a DBContext and returns an instantiated DB.
type DBProvider func(*DBContext) (storage.DB, error)

// DefaultDBProvider returns a database using the DBBackend and DBDir
// specified in the Config.
func DefaultDBProvider(ctx *DBContext) (storage.DB, error) {
	var (
		dbName = ctx.ID
		dbDir  = ctx.Config.DBDir()
	)
	db, err := storage.NewDB(dbName, dbDir)
	if err != nil {
		return nil, fmt.Errorf("database provider: %w", err)
	}

	return db, nil
}
