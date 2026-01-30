package logincmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/spf13/cobra"
)

var (
	betterAuthURL string
	frontendURL   string
	clientID      string
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Nixopus using Device Authorization Grant",
	Long:  `Authenticate with Nixopus by opening a browser and entering a device code.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get Better Auth URL from flag or environment
		authURL := betterAuthURL
		if authURL == "" {
			authURL = os.Getenv("BETTER_AUTH_URL")
		}
		if authURL == "" {
			authURL = "https://auth.nixopus.com" // Default production URL
		}

		// Get Frontend URL from flag or environment
		frontend := frontendURL
		if frontend == "" {
			frontend = os.Getenv("FRONTEND_URL")
		}
		if frontend == "" {
			frontend = "http://localhost:3000" // Default frontend URL for development
		}

		// Get client ID from flag or environment
		cliClientID := clientID
		if cliClientID == "" {
			cliClientID = os.Getenv("OAUTH_CLIENT_ID")
		}
		if cliClientID == "" {
			cliClientID = "nixopus-cli" // Default client ID
		}

		scope := "openid profile email"

		// Start bubbletea program
		program := NewLoginProgram()

		// Run login steps in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- runLoginSteps(program, authURL, frontend, cliClientID, scope)
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

// runLoginSteps runs the login steps and sends updates to the UI
func runLoginSteps(program *LoginProgram, betterAuthURL, frontendURL, clientID, scope string) error {
	// Step 1: Request device code
	program.Send(LoginStepMsg{Step: 0, Message: "Requesting device authorization..."})
	deviceCodeResp, err := RequestDeviceCode(betterAuthURL, clientID, scope)
	if err != nil {
		program.Send(LoginErrorMsg{Error: fmt.Sprintf("Failed to request device code: %v", err)})
		program.Quit()
		return err
	}

	// Step 2: Display user code
	formattedCode := FormatUserCode(deviceCodeResp.UserCode)

	// Construct frontend verification URL (frontend serves the /device page)
	frontendVerificationURL := fmt.Sprintf("%s/device?user_code=%s", frontendURL, deviceCodeResp.UserCode)

	program.Send(LoginStepMsg{
		Step:    1,
		Message: fmt.Sprintf("Visit: %s/device", frontendURL),
	})
	program.Send(LoginStepMsg{
		Step:    2,
		Message: fmt.Sprintf("Enter code: %s", formattedCode),
	})

	// Step 3: Open browser to frontend URL
	program.Send(LoginStepMsg{Step: 3, Message: "Opening browser..."})
	if err := openBrowser(frontendVerificationURL); err != nil {
		program.Send(LoginStepMsg{
			Step:    3,
			Message: fmt.Sprintf("Could not open browser. Please visit: %s", frontendVerificationURL),
		})
	}

	program.Send(LoginStepMsg{
		Step:    4,
		Message: fmt.Sprintf("Waiting for authorization... (polling every %ds)", deviceCodeResp.Interval),
	})

	// Step 4: Poll for access token
	accessToken, refreshToken, err := PollForToken(
		betterAuthURL,
		deviceCodeResp.DeviceCode,
		clientID,
		deviceCodeResp.Interval,
		deviceCodeResp.ExpiresIn,
	)
	if err != nil {
		program.Send(LoginErrorMsg{Error: err.Error()})
		program.Quit()
		return err
	}

	// Step 5: Save access token to config
	program.Send(LoginStepMsg{Step: 5, Message: "Saving access token..."})

	// Load existing config or create new one
	cfg, err := config.Load()
	if err != nil {
		// Config doesn't exist, create a new one
		cfg = &config.Config{
			Sync: config.SyncConfig{
				DebounceMs: 300,
				Exclude: []string{
					"*.log",
					".git",
					"node_modules",
					"__pycache__",
					".env",
				},
			},
		}
	}

	// Update tokens
	cfg.AccessToken = accessToken
	if refreshToken != "" {
		cfg.RefreshToken = refreshToken
	}

	if err := cfg.Save(); err != nil {
		program.Send(LoginErrorMsg{Error: fmt.Sprintf("Failed to save access token: %v", err)})
		program.Quit()
		return err
	}

	// Send success message
	program.Send(LoginSuccessMsg{})

	// Wait a bit for user to see success, then quit
	program.Quit()

	return nil
}

// openBrowser opens the URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Run()
}

func init() {
	LoginCmd.Flags().StringVar(&betterAuthURL, "auth-url", "", "Better Auth backend URL (default: https://auth.nixopus.com or BETTER_AUTH_URL env var)")
	LoginCmd.Flags().StringVar(&frontendURL, "frontend-url", "", "Frontend URL for device verification page (default: http://localhost:3000 or FRONTEND_URL env var)")
	LoginCmd.Flags().StringVar(&clientID, "client-id", "", "OAuth client ID (default: nixopus-cli or OAUTH_CLIENT_ID env var)")
}
