package cli

import (
	"github.com/raghavyuva/nixopus-api/pkg/cli/addcmd"
	"github.com/raghavyuva/nixopus-api/pkg/cli/listcmd"
	"github.com/raghavyuva/nixopus-api/pkg/cli/live"
	"github.com/raghavyuva/nixopus-api/pkg/cli/pause"
	"github.com/raghavyuva/nixopus-api/pkg/cli/removecmd"
	setenv "github.com/raghavyuva/nixopus-api/pkg/cli/set_env"
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
	rootCmd.AddCommand(live.LiveCmd)
	rootCmd.AddCommand(pause.PauseCmd)
	rootCmd.AddCommand(removecmd.RemoveCmd)
	rootCmd.AddCommand(setenv.SetEnvCmd)
}
