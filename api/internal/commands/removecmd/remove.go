package removecmd

import (
	"fmt"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/spf13/cobra"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an application from the family",
	Long:  `Remove an application from the current family. This will delete the application from the server.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Start bubbletea program
		program := NewRemoveProgram()

		// Run remove steps in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- runRemoveSteps(program, name)
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

// runRemoveSteps runs the remove steps and sends updates to the UI
func runRemoveSteps(program *RemoveProgram, name string) error {
	step := 0

	// Step 1: Load config and find application
	program.Send(RemoveStepMsg{Step: step, Message: "Loading configuration..."})
	cfg, err := config.Load()
	if err != nil {
		program.Send(RemoveErrorMsg{Error: fmt.Sprintf("failed to load config: %v", err)})
		program.Quit()
		return err
	}

	// Find application ID by name
	var applicationID string
	if len(cfg.Applications) > 0 {
		if appID, ok := cfg.Applications[name]; ok {
			applicationID = appID
		}
	}

	// Fallback: check if name is "root" or "default" and use ProjectID
	if applicationID == "" {
		if (name == "root" || name == "default") && cfg.ProjectID != "" {
			applicationID = cfg.ProjectID
		}
	}

	if applicationID == "" {
		program.Send(RemoveErrorMsg{Error: fmt.Sprintf("application '%s' not found in family", name)})
		program.Quit()
		return fmt.Errorf("application '%s' not found", name)
	}
	step++

	// Step 2: Delete application from server
	program.Send(RemoveStepMsg{Step: step, Message: "Deleting application from server..."})
	server := config.GetServerURL()

	// Check for access token
	if cfg.AccessToken == "" {
		program.Send(RemoveErrorMsg{Error: "not authenticated. Please run 'nixopus login' first"})
		program.Quit()
		return fmt.Errorf("not authenticated. Please run 'nixopus login' first")
	}

	if err := deleteApplication(server, cfg.AccessToken, applicationID); err != nil {
		program.Send(RemoveErrorMsg{Error: fmt.Sprintf("failed to delete application: %v", err)})
		program.Quit()
		return err
	}
	step++

	// Step 3: Update config
	program.Send(RemoveStepMsg{Step: step, Message: "Updating configuration..."})
	if len(cfg.Applications) > 0 {
		delete(cfg.Applications, name)
		// Also remove "default" if it points to the same app
		if cfg.Applications["default"] == applicationID {
			delete(cfg.Applications, "default")
		}
		// Also remove "root" if it points to the same app
		if cfg.Applications["root"] == applicationID {
			delete(cfg.Applications, "root")
		}
	}

	// If we removed the last app and ProjectID matches, clear ProjectID too
	if len(cfg.Applications) == 0 && cfg.ProjectID == applicationID {
		cfg.ProjectID = ""
	}

	if err := cfg.Save(); err != nil {
		program.Send(RemoveErrorMsg{Error: fmt.Sprintf("failed to save config: %v", err)})
		program.Quit()
		return err
	}

	// Send success message
	program.Send(RemoveSuccessMsg{AppName: name})

	// Wait a bit for user to see success, then quit
	time.Sleep(2 * time.Second)
	program.Quit()

	return nil
}
