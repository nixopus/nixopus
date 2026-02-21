package main

import (
	"os"

	"github.com/raghavyuva/nixopus-api/internal/cliconfig" // validates build-time config on import
	"github.com/raghavyuva/nixopus-api/internal/commands"
	"github.com/raghavyuva/nixopus-api/internal/config"
)

func main() {
	config.ServerURLProvider = cliconfig.GetAPIURL
	config.ConfigFileNameProvider = cliconfig.GetConfigFileName
	config.AuthFileNameProvider = cliconfig.GetAuthFileName
	config.SyncStateFileNameProvider = cliconfig.GetSyncStateFileName
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
