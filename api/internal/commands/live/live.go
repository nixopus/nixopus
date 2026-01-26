package live

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/cli_config"
	"github.com/raghavyuva/nixopus-api/internal/mover"
	"github.com/spf13/cobra"
)

var (
	allFlag bool
)

var LiveCmd = &cobra.Command{
	Use:   "live [app-name]",
	Short: "Start a live deploy session",
	Long:  `Start watching for file changes and hot reload. Optionally specify an app name to use a specific application. Use --all to run all apps in the family simultaneously.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if --all flag is set
		if allFlag {
			return runAllApps(args)
		}
		return runSingleApp(args)
	},
}

func init() {
	LiveCmd.Flags().BoolVar(&allFlag, "all", false, "Run all apps in the family simultaneously")
}

func runSingleApp(args []string) error {
		// Initialize status tracker
		tracker := mover.NewTracker()

		// Start bubbletea program IMMEDIATELY to show connecting UI before any initialization
		program := NewProgram(tracker)

		// Run program in background
		done := make(chan error, 1)
		programStarted := make(chan bool, 1)
		go func() {
			programStarted <- true
			done <- program.Start()
		}()

		// Wait for program to start and render
		<-programStarted
		// Give UI a moment to render the connecting box
		time.Sleep(200 * time.Millisecond)

		// Now load configuration (UI is already showing)
		cfg, err := cli_config.Load()
		if err != nil {
			program.Quit()
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get app name from args (empty string means use default)
		appName := ""
		if len(args) > 0 {
			appName = args[0]
		}

		// Get application ID from config
		applicationID, err := cfg.GetApplicationID(appName)
		if err != nil {
			program.Quit()
			return fmt.Errorf("failed to get application ID: %w", err)
		}

		// Fetch application details to get base_path
		basePath, err := getApplicationDetails(cfg.Server, applicationID, cfg.APIKey)
		if err != nil {
			program.Quit()
			return fmt.Errorf("failed to fetch application details: %w", err)
		}

		// OPTIMIZATION: Start WebSocket connection IMMEDIATELY after config load
		// Do other validations in parallel while connection is establishing
		wsURL := buildWebSocketURL(cfg.Server, applicationID, cfg.APIKey)
		if wsURL == "" {
			program.Quit()
			return fmt.Errorf("failed to connect")
		}

		// Connection state handler - maps mover states to status states
		onStateChange := func(event mover.ConnectionEvent) {
			var statusState mover.ConnectionStatus
			switch event.State {
			case mover.StateConnected:
				statusState = mover.ConnectionStatusConnected
			case mover.StateConnecting:
				statusState = mover.ConnectionStatusConnecting
			case mover.StateReconnecting:
				statusState = mover.ConnectionStatusReconnecting
			case mover.StateDisconnected:
				statusState = mover.ConnectionStatusDisconnected
			}
			tracker.SetConnectionStatus(statusState)
		}

		// Start WebSocket connection in background (non-blocking)
		// This begins the handshake immediately while we do other validations
		tracker.SetConnectionStatus(mover.ConnectionStatusConnecting)
		clientChan := make(chan *mover.Client, 1)
		clientErrChan := make(chan error, 1)
		go func() {
			client, err := mover.NewClient(
				wsURL,
				cfg.APIKey,
				mover.WithOnStateChange(onStateChange),
			)
			if err != nil {
				clientErrChan <- err
				return
			}
			clientChan <- client
		}()

		// Do other validations in parallel while connection is establishing
		// Run validations concurrently to maximize overlap with connection attempt
		validationErrChan := make(chan error, 1)
		repoPathChan := make(chan string, 1)
		go func() {
			// Calculate and set domain URL immediately (client-side calculation)
			// Domain format: https://{first-8-chars-of-application-id}.nixopus.com
			domainURL := buildDomainURL(applicationID)
			if domainURL != "" {
				tracker.SetURL(domainURL)
			}

			// Get current working directory
			wd, err := os.Getwd()
			if err != nil {
				validationErrChan <- fmt.Errorf("failed to get current directory: %w", err)
				return
			}

			// Git repository is required
			if !isGitRepo(wd) {
				validationErrChan <- fmt.Errorf("git repository required: not a git repository in %s", wd)
				return
			}

			// Check and set environment file path if configured
			if cfg.EnvPath != "" {
				// Resolve env path relative to repo root
				envFilePath := cfg.EnvPath
				if !strings.HasPrefix(envFilePath, "/") {
					envFilePath = fmt.Sprintf("%s/%s", wd, envFilePath)
				}
				// Verify env file exists
				if _, err := os.Stat(envFilePath); err == nil {
					tracker.SetEnvPath(cfg.EnvPath) // Store relative path for display
				}
			}

			repoPathChan <- wd
			validationErrChan <- nil
		}()

		// Wait for both WebSocket client and validations to complete
		// This allows connection and validations to happen in parallel
		var client *mover.Client
		var validationErr error
		var clientErr error
		var repoPath string
		completed := 0

		// Wait for both to complete (whichever finishes first)
		for completed < 2 {
			select {
			case client = <-clientChan:
				completed++
			case err := <-clientErrChan:
				clientErr = err
				completed++
			case err := <-validationErrChan:
				validationErr = err
				completed++
			case path := <-repoPathChan:
				repoPath = path
			case <-time.After(30 * time.Second):
				program.Quit()
				return fmt.Errorf("initialization timeout")
			}
		}

		// Ensure we got repoPath (it might have been sent before we started waiting)
		if repoPath == "" {
			select {
			case repoPath = <-repoPathChan:
			case <-time.After(1 * time.Second):
				program.Quit()
				return fmt.Errorf("failed to get repository path")
			}
		}

		// Check for errors
		if validationErr != nil {
			if client != nil {
				client.Close()
			}
			program.Quit()
			return validationErr
		}
		if clientErr != nil {
			program.Quit()
			return fmt.Errorf("failed to connect: %w", clientErr)
		}
		if client == nil {
			program.Quit()
			return fmt.Errorf("client not initialized")
		}
		defer client.Close()

		// Determine root path based on base_path
		// If base_path is "/", watch entire repo
		// Otherwise, watch only the subdirectory
		watchPath := repoPath
		if basePath != "" && basePath != "/" {
			// Normalize base_path (remove leading/trailing slashes)
			normalizedBasePath := strings.Trim(basePath, "/")
			if normalizedBasePath != "" {
				watchPath = filepath.Join(repoPath, normalizedBasePath)
				// Verify the path exists
				if _, err := os.Stat(watchPath); os.IsNotExist(err) {
					program.Quit()
					return fmt.Errorf("base_path '%s' does not exist in repository", basePath)
				}
			}
		}

		engine, err := mover.NewEngine(mover.EngineConfig{
			RootPath:         watchPath,
			Client:           client,
			Excludes:         cfg.Sync.Exclude,
			DebounceMs:       cfg.Sync.DebounceMs,
			OnStateChange:    onStateChange,
			OnFileSynced:     func(path string) { tracker.IncrementFilesSynced() },
			OnChangeDetected: func(path string) { tracker.IncrementChanges() },
		})
		if err != nil {
			program.Quit()
			return fmt.Errorf("failed to create sync engine: %w", err)
		}

		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		if err := engine.Start(); err != nil {
			program.Quit()
			return fmt.Errorf("failed to start sync: %w", err)
		}

		// Start deployment status poller
		pollerCtx, pollerCancel := context.WithCancel(context.Background())
		poller := NewDeploymentPoller(cfg, tracker, applicationID)
		go poller.Start(pollerCtx)
		defer func() {
			poller.Stop()
			pollerCancel()
		}()

		// The UI will automatically switch from connecting box to status box
		// when connection status becomes Connected (handled in TickMsg)

		// Wait for either interrupt or program exit
		select {
		case <-sigChan:
			// User pressed Ctrl+C
			program.Quit()
			<-done // Wait for program to exit
		case err := <-done:
			// Program exited
			if err != nil {
				return fmt.Errorf("UI program error: %w", err)
			}
		}

		// Stop engine
		if err := engine.Stop(); err != nil {
			return fmt.Errorf("failed to stop sync engine: %w", err)
		}

		return nil
	}

// buildWebSocketURL builds the WebSocket URL from server URL
func buildWebSocketURL(server, projectID, apiKey string) string {
	// Convert http:// to ws:// and https:// to wss://
	wsURL := server
	if strings.HasPrefix(wsURL, "http://") {
		wsURL = "ws://" + wsURL[7:]
	} else if strings.HasPrefix(wsURL, "https://") {
		wsURL = "wss://" + wsURL[8:]
	}

	// Add WebSocket path and query params (using projectID as application_id)
	wsURL += "/ws/live/" + projectID + "?token=" + apiKey
	return wsURL
}

// buildDomainURL builds the domain URL from project ID
// Format: https://{first-8-chars-of-project-id}.nixopus.com
func buildDomainURL(projectID string) string {
	if projectID == "" || len(projectID) < 8 {
		return ""
	}
	// Take first 8 characters of project ID (UUID format)
	subdomain := projectID[:8]
	return "https://" + subdomain + ".nixopus.com"
}

// isGitRepo checks if the given path is a git repository
func isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}
