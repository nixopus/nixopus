package setenv

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/cli_config"
	"github.com/spf13/cobra"
)

var SetEnvCmd = &cobra.Command{
	Use:   "set-env",
	Short: "Set the environment file path for deployment",
	Long: `Set the path to the environment file that should be synced and used during deployment.
The path should be relative to the project root (e.g., .env, .env.production, config/.env).

If an env path is set, it will be:
- Synced to the deployment environment
- Made available when services run
- Removed from the exclude list if it was previously excluded`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		envPath := args[0]

		// Start bubbletea program
		program := NewSetEnvProgram()

		// Run set-env steps in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- runSetEnvSteps(program, envPath)
		}()

		// Start UI and wait for completion
		if err := program.Start(); err != nil {
			return err
		}

		// Check if there was an error
		select {
		case err := <-done:
			if err != nil {
				return err
			}
		default:
		}

		return nil
	},
}

// runSetEnvSteps runs the set-env steps and sends updates to the UI
func runSetEnvSteps(program *SetEnvProgram, envPath string) error {
	// Validate the env path
	if err := cli_config.ValidateEnvPath(envPath); err != nil {
		program.Send(SetEnvErrorMsg{Error: fmt.Sprintf("invalid env path: %v", err)})
		program.Quit()
		return err
	}

	// Load existing config
	cfg, err := cli_config.Load()
	if err != nil {
		program.Send(SetEnvErrorMsg{Error: fmt.Sprintf("failed to load config: %v", err)})
		program.Quit()
		return err
	}

	// Update config
	cfg.EnvPath = envPath

	// Remove env path from excludes if it's in there
	cfg.Sync.Exclude = removeFromSlice(cfg.Sync.Exclude, envPath)

	// Also remove common patterns that might match
	cleanPath := filepath.Clean(envPath)
	baseName := filepath.Base(cleanPath)
	cfg.Sync.Exclude = removeFromSlice(cfg.Sync.Exclude, baseName)

	// Save config
	if err := cfg.Save(); err != nil {
		program.Send(SetEnvErrorMsg{Error: fmt.Sprintf("failed to save config: %v", err)})
		program.Quit()
		return err
	}

	// Send success message
	program.Send(SetEnvSuccessMsg{EnvPath: envPath})

	// Wait a bit for user to see success, then quit
	time.Sleep(2 * time.Second)
	program.Quit()

	return nil
}

// removeFromSlice removes an item from a string slice
func removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func init() {
}
