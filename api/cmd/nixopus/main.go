package main

import (
	"os"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/pkg/cli"
	"github.com/raghavyuva/nixopus-api/pkg/cli/cliconfig" // validates build-time config on import
)

func main() {
	config.ServerURLProvider = cliconfig.GetAPIURL
	config.ConfigFileNameProvider = cliconfig.GetConfigFileName
	config.AuthFileNameProvider = cliconfig.GetAuthFileName
	config.SyncStateFileNameProvider = cliconfig.GetSyncStateFileName
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
