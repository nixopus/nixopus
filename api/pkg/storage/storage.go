// Package storage provides public access to storage types from the internal package.
package storage

import (
	"context"

	internalStorage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

// App is a type alias for the internal storage.App type.
// This allows other modules to reference the App type without importing internal packages.
type App = internalStorage.App

// Store is a type alias for the internal storage.Store type.
type Store = internalStorage.Store

// NewApp creates a new App instance.
// Note: This function is a wrapper around the internal NewApp function.
func NewApp(config *types.Config, store *Store, ctx context.Context) *App {
	return internalStorage.NewApp(config, store, ctx)
}

// NewStore creates a new Store instance.
// Note: This function is a wrapper around the internal NewStore function.
func NewStore(db *bun.DB) *Store {
	return internalStorage.NewStore(db)
}

// RunMigrations runs database migrations from the specified path.
// This is a wrapper around the internal storage.RunMigrations function.
func RunMigrations(db *bun.DB, migrationsPath string) error {
	return internalStorage.RunMigrations(db, migrationsPath)
}
