// Package config provides public access to configuration initialization from the internal package.
// This allows other modules (like cloud) to use the config initialization without importing internal packages.
package config

import (
	internalConfig "github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/pkg/storage"
)

// Init initializes the app configuration using Viper to load values from config files,
// environment variables, and defaults. It then creates a new PostgreSQL client using
// the loaded configuration and initializes the storage.Store.
// This is a wrapper around the internal config.Init() function.
func Init() *storage.Store {
	return internalConfig.Init()
}

// AppConfig provides access to the global AppConfig from the internal config package.
var AppConfig = &internalConfig.AppConfig
