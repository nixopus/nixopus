package addcmd

import (
	"fmt"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:   "add <path> <name>",
	Short: "Add a new application to the family",
	Long:  `Add a new application from a subdirectory to the current family.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		name := args[1]

		// Start bubbletea program
		program := NewAddProgram()

		// Run add steps in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- runAddSteps(program, path, name)
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

// runAddSteps runs the add steps and sends updates to the UI
func runAddSteps(program *AddProgram, path, name string) error {
	step := 0

	// Step 1: Load and validate config
	program.Send(AddStepMsg{Step: step, Message: "Loading configuration..."})
	cfg, err := config.Load()
	if err != nil {
		program.Send(AddErrorMsg{Error: fmt.Sprintf("failed to load config: %v", err)})
		program.Quit()
		return err
	}

	// Validate family_id exists
	if cfg.FamilyID == "" {
		program.Send(AddErrorMsg{Error: "family_id not found in config. Run 'nixopus live' to initialize first"})
		program.Quit()
		return fmt.Errorf("family_id not found in config")
	}
	step++

	// Step 2: Get repository info and add application
	program.Send(AddStepMsg{Step: step, Message: "Getting repository info..."})
	_, repoURL, err := getGitInfo()
	if err != nil {
		program.Send(AddErrorMsg{Error: fmt.Sprintf("failed to get git info: %v", err)})
		program.Quit()
		return err
	}

	branch := getGitBranch()

	// Normalize base_path (remove leading ./ and ensure it's relative)
	basePath := normalizeBasePath(path)

	program.Send(AddStepMsg{Step: step, Message: "Adding application to family..."})
	server := config.GetServerURL()

	// Get access token from global auth storage
	accessToken, err := config.GetAccessToken()
	if err != nil {
		program.Send(AddErrorMsg{Error: "not authenticated. Please run 'nixopus login' first"})
		program.Quit()
		return fmt.Errorf("not authenticated. Please run 'nixopus login' first: %w", err)
	}

	applicationID, err := addApplicationToFamily(server, accessToken, cfg.FamilyID, name, basePath, repoURL, branch, 0)
	if err != nil {
		program.Send(AddErrorMsg{Error: fmt.Sprintf("failed to add application: %v", err)})
		program.Quit()
		return err
	}
	step++

	// Step 3: Update config
	program.Send(AddStepMsg{Step: step, Message: "Updating configuration..."})
	if cfg.Applications == nil {
		cfg.Applications = make(map[string]string)
	}
	cfg.Applications[name] = applicationID

	if err := cfg.Save(); err != nil {
		program.Send(AddErrorMsg{Error: fmt.Sprintf("failed to save config: %v", err)})
		program.Quit()
		return err
	}

	// Send success message
	program.Send(AddSuccessMsg{AppName: name, BasePath: basePath})

	// Wait a bit for user to see success, then quit
	time.Sleep(2 * time.Second)
	program.Quit()

	return nil
}

func init() {
	// No flags needed for now
}
