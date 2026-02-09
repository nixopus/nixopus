package commands

import (
	"github.com/raghavyuva/nixopus-api/internal/commands/addcmd"
	"github.com/raghavyuva/nixopus-api/internal/commands/listcmd"
	"github.com/raghavyuva/nixopus-api/internal/commands/live"
	"github.com/raghavyuva/nixopus-api/internal/commands/removecmd"
	setenv "github.com/raghavyuva/nixopus-api/internal/commands/set_env"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "nixopus",
	Short:         "Nixopus CLI - Live deploy to cloud",
	Long:          `Nixopus CLI enables hot-reload deployments to your cloud server.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(addcmd.AddCmd)
	rootCmd.AddCommand(listcmd.ListCmd)
	rootCmd.AddCommand(removecmd.RemoveCmd)
	rootCmd.AddCommand(live.LiveCmd)
	rootCmd.AddCommand(setenv.SetEnvCmd)
}
