package commands

import (
	"fmt"
	"os"

	"github.com/raghavyuva/nixopus-api/internal/commands/addcmd"
	"github.com/raghavyuva/nixopus-api/internal/commands/initcmd"
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
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "%s\n", rootCmd.UsageString())
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(initcmd.InitCmd)
	rootCmd.AddCommand(addcmd.AddCmd)
	rootCmd.AddCommand(listcmd.ListCmd)
	rootCmd.AddCommand(removecmd.RemoveCmd)
	rootCmd.AddCommand(live.LiveCmd)
	rootCmd.AddCommand(setenv.SetEnvCmd)
}
