package config

import (
	"context"
	"fmt"

	cmtdb "github.com/cometbft/cometbft/db"
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
	// DB name
	ID string
	// Node configuration
	Config *Config
}

// DBProvider takes a DBContext and returns an instantiated DB.
type DBProvider func(*DBContext) (cmtdb.DB, error)

// DefaultDBProvider creates a DB using the given ctx.
func DefaultDBProvider(ctx *DBContext) (cmtdb.DB, error) {
	var (
		dbName = ctx.ID
		dbDir  = ctx.Config.DBDir()
	)
	db, err := cmtdb.New(dbName, dbDir)
	if err != nil {
		return nil, fmt.Errorf("database provider: %w", err)
	}

	return db, nil
}
