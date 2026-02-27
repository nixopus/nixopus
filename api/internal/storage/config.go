package storage

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

// Config represents the configuration for a database connection.
type Config struct {
	Host           string
	Port           string
	Username       string
	Password       string
	DBName         string
	SSLMode        string
	MaxOpenConn    int
	Debug          bool
	MaxIdleConn    int
	MigrationsPath string
}

// getConnString generates a PostgreSQL connection string from the given configuration.
func (c Config) getConnString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.Username,
		c.Password,
		c.DBName,
		c.SSLMode,
	)
}

// NewDB creates a new database connection using the provided configuration and returns a *bun.DB.
func NewDB(c *Config) (*bun.DB, error) {
	config, err := pgx.ParseConfig(c.getConnString())
	if err != nil {
		return nil, err
	}

	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	sqldb := stdlib.OpenDB(*config)
	maxOpen := c.MaxOpenConn
	if maxOpen <= 0 {
		maxOpen = 10
	}
	sqldb.SetMaxOpenConns(maxOpen)
	maxIdle := c.MaxIdleConn
	if maxIdle <= 0 {
		maxIdle = 5
	}
	sqldb.SetMaxIdleConns(maxIdle)

	db := bun.NewDB(sqldb, pgdialect.New())

	if c.Debug {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(false)))
	}

	return db, nil
}
