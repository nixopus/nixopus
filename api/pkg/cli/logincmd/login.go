package logincmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/raghavyuva/nixopus-api/pkg/cli/cliconfig"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/spf13/cobra"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Nixopus using Device Authorization Grant",
	Long:  `Authenticate with Nixopus by opening a browser and entering a device code.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		authURL := cliconfig.GetBetterAuthURL()
		frontend := cliconfig.GetFrontendURL()
		cliClientID := cliconfig.GetOAuthClientID()
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

// ForceLogin performs a fresh login flow, replacing any existing credentials.
// Use with --login flag on live/deploy commands to re-authenticate.
func ForceLogin() error {
	return performLogin()
}

// EnsureAuthenticated checks if the user is authenticated and prompts for login if not.
// Returns an error if authentication fails or is cancelled.
func EnsureAuthenticated() error {
	// Check global auth storage for existing token
	accessToken, err := config.GetAccessToken()
	if err == nil && accessToken != "" {
		// Already authenticated
		return nil
	}

	// Not authenticated, prompt for login
	return performLogin()
}

// performLogin performs the login flow
func performLogin() error {
	authURL := cliconfig.GetBetterAuthURL()
	frontend := cliconfig.GetFrontendURL()
	cliClientID := cliconfig.GetOAuthClientID()
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

	// Wait for login to complete
	if err := <-done; err != nil {
		return err
	}

	return nil
}

// runLoginSteps runs the login steps and sends updates to the UI
func runLoginSteps(program *LoginProgram, betterAuthURL, frontendURL, clientID, scope string) error {
	// Request device code
	deviceCodeResp, err := RequestDeviceCode(betterAuthURL, clientID, scope)
	if err != nil {
		program.Send(LoginErrorMsg{Error: fmt.Sprintf("Failed to request device code: %v", err)})
		program.Quit()
		return err
	}

	// Construct frontend verification URL (frontend serves the /device page)
	frontendVerificationURL := fmt.Sprintf("%s/device?user_code=%s", frontendURL, deviceCodeResp.UserCode)

	// Display URL - user can click if browser opening fails
	program.Send(LoginStepMsg{
		Step:    0,
		Message: frontendVerificationURL,
	})

	// Try to open browser (silently fail if it doesn't work)
	_ = openBrowser(frontendVerificationURL)

	// Poll for access token
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

	// Save tokens to global auth storage (not project config)
	if err := config.SaveAuth(accessToken, refreshToken); err != nil {
		program.Send(LoginErrorMsg{Error: fmt.Sprintf("Failed to save access token: %v", err)})
		program.Quit()
		return err
	}

	// Fetch user's organizations and save the first org ID
	orgs, err := FetchUserOrganizations(betterAuthURL, accessToken)
	if err != nil {
		program.Send(LoginStepMsg{Step: 2, Message: fmt.Sprintf("Warning: could not fetch organizations: %v", err)})
	} else if len(orgs) > 0 {
		if err := config.SaveOrganizationID(orgs[0].ID); err != nil {
			program.Send(LoginStepMsg{Step: 2, Message: fmt.Sprintf("Warning: could not save organization ID: %v", err)})
		}
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
