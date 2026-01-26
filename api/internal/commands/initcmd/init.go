package initcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal/cli_config"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project for live deploy",
	Long:  `Initialize the current directory as a Nixopus live deploy project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("api-key")
		envPath, _ := cmd.Flags().GetString("env-path")
		server := cli_config.GetServerURL()

		// Start bubbletea program
		program := NewInitProgram()

		// Run init steps in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- runInitSteps(program, server, apiKey, envPath)
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

// parseEnvFile parses a .env file and returns a map of environment variables
func parseEnvFile(envPath string) (map[string]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	fullPath := filepath.Join(cwd, filepath.Clean(envPath))
	envMap, err := godotenv.Read(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file %s: %w", envPath, err)
	}

	return envMap, nil
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

// runInitSteps runs the initialization steps and sends updates to the UI
func runInitSteps(program *InitProgram, server, apiKey, envPath string) error {
	step := 0

	// Validate env path if provided (before UI starts, so we can show error immediately)
	if envPath != "" {
		if err := cli_config.ValidateEnvPath(envPath); err != nil {
			program.Send(InitErrorMsg{Error: fmt.Sprintf("invalid env path: %v", err)})
			program.Quit()
			return err
		}
	}

	// Step 1: Validate API key
	program.Send(InitStepMsg{Step: step, Message: "Connecting to server..."})
	if err := ValidateAPIKey(server, apiKey); err != nil {
		program.Send(InitErrorMsg{Error: fmt.Sprintf("API key validation failed: %v", err)})
		program.Quit()
		return err
	}
	step++

	// Step 2: Parse env file if provided
	var envVars map[string]string
	if envPath != "" {
		program.Send(InitStepMsg{Step: step, Message: fmt.Sprintf("Reading %s...", envPath)})
		parsedEnvVars, err := parseEnvFile(envPath)
		if err != nil {
			program.Send(InitErrorMsg{Error: fmt.Sprintf("failed to parse env file: %v", err)})
			program.Quit()
			return err
		}
		envVars = parsedEnvVars
	} else {
		// Skip this step visually
		program.Send(InitStepMsg{Step: step, Message: "Skipping environment file (not provided)"})
	}
	step++

	// Step 3: Create project
	program.Send(InitStepMsg{Step: step, Message: "Creating project on server..."})
	projectID, familyID, err := CreateProject(server, apiKey, envVars)
	if err != nil {
		program.Send(InitErrorMsg{Error: fmt.Sprintf("failed to create project: %v", err)})
		program.Quit()
		return err
	}
	step++

	// Step 4: Save config
	program.Send(InitStepMsg{Step: step, Message: "Saving configuration..."})

	// Build exclude list - remove env path from excludes if it's set
	exclude := []string{
		"*.log",
		".git",
		"node_modules",
		"__pycache__",
		".env",
	}
	if envPath != "" {
		exclude = removeFromSlice(exclude, envPath)
	}

	// Create applications map with root app as default
	// Only use "default" - "root" can be added later if user wants that name
	applications := map[string]string{
		"default": projectID,
	}

	cfg := &cli_config.Config{
		APIKey:       apiKey,
		FamilyID:     familyID,
		Applications: applications,
		Sync: cli_config.SyncConfig{
			DebounceMs: 300,
			Exclude:    exclude,
		},
		EnvPath: envPath,
	}

	if err := cfg.Save(); err != nil {
		program.Send(InitErrorMsg{Error: fmt.Sprintf("failed to save config: %v", err)})
		program.Quit()
		return err
	}

	// Send success message
	program.Send(InitSuccessMsg{ProjectID: projectID, EnvPath: envPath})

	// Wait a bit for user to see success, then quit
	time.Sleep(3 * time.Second)
	program.Quit()

	return nil
}

func init() {
	InitCmd.Flags().StringP("api-key", "k", "", "API key for authentication")
	InitCmd.MarkFlagRequired("api-key")
	InitCmd.Flags().String("env-path", "", "Path to environment file (relative to project root, e.g., .env or .env.production)")
}
