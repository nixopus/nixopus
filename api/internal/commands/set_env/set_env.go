package setenv

import (
	"fmt"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/spf13/cobra"
)

var SetEnvCmd = &cobra.Command{
	Use:   "set-env",
	Short: "Set the environment file path for deployment",
	Long: `Set the path to the environment file (e.g., .env, .env.production) relative to project root.
The file itself is never synced. Its values are read locally and sent to the server when you run 'nixopus live'.
Changes to the env file update the container's environment automatically.`,
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
	if err := config.ValidateEnvPath(envPath); err != nil {
		program.Send(SetEnvErrorMsg{Error: fmt.Sprintf("invalid env path: %v", err)})
		program.Quit()
		return err
	}

	// Load existing config
	cfg, err := config.Load()
	if err != nil {
		program.Send(SetEnvErrorMsg{Error: fmt.Sprintf("failed to load config: %v", err)})
		program.Quit()
		return err
	}

	// Update config
	cfg.EnvPath = envPath
	// .env stays excluded; values are sent from client, file is never synced

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

func init() {
}
