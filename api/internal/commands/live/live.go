package live

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal/commands/logincmd"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/httpclient"
	"github.com/raghavyuva/nixopus-api/internal/mover"
	"github.com/spf13/cobra"
)

var (
	allFlag           bool
	envPath           string
	forceFullSyncFlag bool
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
	LiveCmd.Flags().StringVar(&envPath, "env-path", "", "Path to environment file (relative to project root, e.g., .env or .env.production). Only used during initialization if project is not initialized.")
	LiveCmd.Flags().BoolVar(&forceFullSyncFlag, "force-full-sync", false, "Bypass incremental sync; clear persisted state and sync all files")
}

func runSingleApp(args []string) error {
	// Set up terminal and event bus
	term := NewTerminal()
	bus := NewEventBus(100)

	// Start the agent UI immediately so the user sees output right away
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	agentUI := NewAgentUI(bus, term)
	uiDone := make(chan error, 1)
	go func() { uiDone <- agentUI.Run(ctx) }()

	// Header
	bus.Send(Event{Type: EventInfo, Message: "nixopus"})
	bus.Send(Event{Type: EventInfo})

	// Check authentication first - prompt for login if not authenticated
	if err := logincmd.EnsureAuthenticated(); err != nil {
		bus.Send(Event{Type: EventError, Message: fmt.Sprintf("Authentication required: %v", err)})
		cancel()
		return fmt.Errorf("authentication required: %w", err)
	}

	bus.Send(Event{Type: EventAuth, Message: "Authenticated", NeedsLLM: true})

	// Initialize status tracker (still used by poller internally)
	tracker := mover.NewTracker()

	// Check if initialization is needed and run it if necessary
	bus.Send(Event{Type: EventInfo, Message: "Checking project..."})
	cfg, err := ensureInitialized(envPath)
	if err != nil {
		bus.Send(Event{Type: EventError, Message: fmt.Sprintf("Initialization failed: %v", err)})
		cancel()
		return err
	}
	bus.Send(Event{Type: EventConfig, Message: "Project initialized", NeedsLLM: true})

	// Get app name from args (empty string means use default)
	appName := ""
	if len(args) > 0 {
		appName = args[0]
	}

	// Get application ID from config
	applicationID, err := cfg.GetApplicationID(appName)
	if err != nil {
		bus.Send(Event{Type: EventError, Message: fmt.Sprintf("Failed to get application ID: %v", err)})
		cancel()
		return fmt.Errorf("failed to get application ID: %w", err)
	}

	accessToken, err := config.GetAccessToken()
	if err != nil {
		bus.Send(Event{Type: EventError, Message: "Not authenticated"})
		cancel()
		return fmt.Errorf("not authenticated. Please run 'nixopus login' first: %w", err)
	}

	basePath := "/"

	// Build WebSocket URL
	wsURL := buildWebSocketURL(cfg.Server, applicationID, accessToken)
	if wsURL == "" {
		bus.Send(Event{Type: EventError, Message: "Failed to build connection URL"})
		cancel()
		return fmt.Errorf("failed to connect")
	}

	// Connection state handler — publishes events instead of updating Bubble Tea model
	onStateChange := func(event mover.ConnectionEvent) {
		tracker.SetConnectionStatus(connectionStatusFromMover(event.State))
		switch event.State {
		case mover.StateConnected:
			bus.Send(Event{Type: EventConnected})
		case mover.StateConnecting:
			bus.Send(Event{Type: EventConnecting})
		case mover.StateReconnecting:
			bus.Send(Event{Type: EventReconnecting})
		case mover.StateDisconnected:
			bus.Send(Event{Type: EventDisconnected})
		}
	}

	// Start WebSocket connection in background
	bus.Send(Event{Type: EventConnecting})
	clientChan := make(chan *mover.Client, 1)
	clientErrChan := make(chan error, 1)
	go func() {
		client, err := mover.NewClient(
			wsURL,
			"",
			mover.WithOnStateChange(onStateChange),
		)
		if err != nil {
			clientErrChan <- err
			return
		}
		clientChan <- client
	}()

	// Calculate domain URL (pure, no I/O — safe to do on main goroutine)
	domainURL := buildDomainURL(applicationID, cfg.DeployDomain)
	if domainURL != "" {
		tracker.SetURL(domainURL)
	}

	// Run validations in parallel
	validationErrChan := make(chan error, 1)
	repoPathChan := make(chan string, 1)
	go func() {
		wd, err := os.Getwd()
		if err != nil {
			validationErrChan <- fmt.Errorf("failed to get current directory: %w", err)
			return
		}

		if !isGitRepo(wd) {
			validationErrChan <- fmt.Errorf("git repository required: not a git repository in %s", wd)
			return
		}

		if cfg.EnvPath != "" {
			envFilePath := cfg.EnvPath
			if !strings.HasPrefix(envFilePath, "/") {
				envFilePath = fmt.Sprintf("%s/%s", wd, envFilePath)
			}
			if _, err := os.Stat(envFilePath); err == nil {
				tracker.SetEnvPath(cfg.EnvPath)
			}
		}

		repoPathChan <- wd
		validationErrChan <- nil
	}()

	// Wait for both WebSocket client and validations
	var client *mover.Client
	var validationErr error
	var clientErr error
	var repoPath string
	completed := 0

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
			bus.Send(Event{Type: EventError, Message: "Initialization timeout"})
			cancel()
			return fmt.Errorf("initialization timeout")
		}
	}

	if repoPath == "" {
		select {
		case repoPath = <-repoPathChan:
		case <-time.After(1 * time.Second):
			cancel()
			return fmt.Errorf("failed to get repository path")
		}
	}

	if validationErr != nil {
		if client != nil {
			client.Close()
		}
		bus.Send(Event{Type: EventError, Message: validationErr.Error()})
		cancel()
		return validationErr
	}
	if clientErr != nil {
		errMsg := clientErr.Error()
		if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "bad handshake") {
			bus.Send(Event{Type: EventError, Message: "Connection rejected by server. Check SSH/staging config."})
			cancel()
			return fmt.Errorf("connection rejected by server: %w", clientErr)
		}
		bus.Send(Event{Type: EventError, Message: fmt.Sprintf("Connection failed: %v", clientErr)})
		cancel()
		return fmt.Errorf("failed to connect: %w", clientErr)
	}
	if client == nil {
		cancel()
		return fmt.Errorf("client not initialized")
	}
	defer client.Close()

	// Determine watch path
	watchPath := repoPath
	if basePath != "" && basePath != "/" {
		normalizedBasePath := strings.Trim(basePath, "/")
		if normalizedBasePath != "" {
			watchPath = filepath.Join(repoPath, normalizedBasePath)
			if _, err := os.Stat(watchPath); os.IsNotExist(err) {
				cancel()
				return fmt.Errorf("base_path '%s' does not exist in repository", basePath)
			}
		}
	}

	syncStatePath, err := config.GetSyncStatePath()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to get sync state path: %w", err)
	}

	// Resolve env file path
	var envFilePath string
	if cfg.EnvPath != "" {
		cleanPath := filepath.Clean(cfg.EnvPath)
		if !filepath.IsAbs(cleanPath) {
			envFilePath = filepath.Join(repoPath, cleanPath)
		} else {
			envFilePath = cleanPath
		}
		if _, err := os.Stat(envFilePath); err != nil {
			envFilePath = ""
		}
	}

	// File sync counters — batched into periodic progress events
	var syncMu sync.Mutex
	filesSynced := 0
	changesDetected := 0

	// Workflow execution state — only one workflow per app at a time (idempotency on reconnect)
	var workflowRunningMu sync.Mutex
	workflowRunning := false

	engine, err := mover.NewEngine(mover.EngineConfig{
		RootPath:      watchPath,
		Client:        client,
		Excludes:      cfg.Sync.Exclude,
		DebounceMs:    cfg.Sync.DebounceMs,
		OnStateChange: onStateChange,
		OnFileSynced: func(path string) {
			tracker.IncrementFilesSynced()
			syncMu.Lock()
			filesSynced++
			syncMu.Unlock()
		},
		OnChangeDetected: func(path string) {
			tracker.IncrementChanges()
			syncMu.Lock()
			changesDetected++
			syncMu.Unlock()
		},
		OnServerMessage: func(msg mover.SyncMessage) {
			switch msg.Type {
			case mover.MessageTypePipelineProgress:
				if p, ok := msg.Payload.(mover.PipelineProgressPayload); ok {
					bus.Send(Event{
						Type:    EventPipelineProgress,
						Message: p.Message,
						Payload: PipelineProgressPayload{
							StageId: p.StageId,
							Message: p.Message,
						},
					})
				} else if payload, ok := msg.Payload.(map[string]interface{}); ok {
					stageId, _ := payload["stage_id"].(string)
					message, _ := payload["message"].(string)
					bus.Send(Event{
						Type:    EventPipelineProgress,
						Message: message,
						Payload: PipelineProgressPayload{
							StageId: stageId,
							Message: message,
						},
					})
				}
			case mover.MessageTypeBuildStatus:
				if p, ok := msg.Payload.(mover.BuildStatusPayload); ok {
					ev := Event{
						Type:    EventBuildStatus,
						Message: p.Message,
						Payload: BuildStatusPayload{
							Phase:   p.Phase,
							Message: p.Message,
							Error:   p.Error,
						},
					}
					if p.Phase == "error" {
						ev.NeedsLLM = true
					}
					bus.Send(ev)
				} else if payload, ok := msg.Payload.(map[string]interface{}); ok {
					phase, _ := payload["phase"].(string)
					message, _ := payload["message"].(string)
					errMsg, _ := payload["error"].(string)
					ev := Event{
						Type:    EventBuildStatus,
						Message: message,
						Payload: BuildStatusPayload{
							Phase:   phase,
							Message: message,
							Error:   errMsg,
						},
					}
					if phase == "error" {
						ev.NeedsLLM = true
					}
					bus.Send(ev)
				}
			case mover.MessageTypeBuildLog:
				if p, ok := msg.Payload.(mover.BuildLogPayload); ok {
					bus.Send(Event{
						Type:    EventBuildLog,
						Message: p.Log,
						Payload: BuildLogPayload{
							Log:       p.Log,
							Timestamp: p.Timestamp,
						},
					})
				} else if payload, ok := msg.Payload.(map[string]interface{}); ok {
					logLine, _ := payload["log"].(string)
					timestamp, _ := payload["timestamp"].(string)
					bus.Send(Event{
						Type:    EventBuildLog,
						Message: logLine,
						Payload: BuildLogPayload{
							Log:       logLine,
							Timestamp: timestamp,
						},
					})
				}
			case mover.MessageTypeDeploymentStatus:
				if p, ok := msg.Payload.(mover.DeploymentStatusPayload); ok {
					bus.Send(Event{
						Type:    EventDeploymentStatus,
						Message: fmt.Sprintf("Deployment status: %s", p.Status),
						Payload: DeploymentStatusPayload{
							Status:       p.Status,
							DeploymentID: p.DeploymentID,
						},
					})
				} else if payload, ok := msg.Payload.(map[string]interface{}); ok {
					status, _ := payload["status"].(string)
					deploymentID, _ := payload["deployment_id"].(string)
					bus.Send(Event{
						Type:    EventDeploymentStatus,
						Message: fmt.Sprintf("Deployment status: %s", status),
						Payload: DeploymentStatusPayload{
							Status:       status,
							DeploymentID: deploymentID,
						},
					})
				}
			case mover.MessageTypeCodebaseIndexed:
				workflowRunningMu.Lock()
				if workflowRunning {
					workflowRunningMu.Unlock()
					break
				}
				workflowRunning = true
				workflowRunningMu.Unlock()

				appID, orgID, source, mode := extractCodebaseIndexedPayload(msg.Payload)
				if appID == "" {
					bus.Send(Event{Type: EventError, Message: "codebase_indexed: missing application_id"})
					workflowRunningMu.Lock()
					workflowRunning = false
					workflowRunningMu.Unlock()
					break
				}
				if orgID == "" {
					if fallbackOrg, err := config.GetOrganizationID(); err == nil && fallbackOrg != "" {
						orgID = fallbackOrg
					}
				}

				wfClient := NewDeploymentWorkflowClient(accessToken, orgID)
				go func() {
					defer func() {
						workflowRunningMu.Lock()
						workflowRunning = false
						workflowRunningMu.Unlock()
					}()

					bus.Send(Event{Type: EventBuildStatus, Message: "Running deployment workflow...", Payload: BuildStatusPayload{Phase: "generating_dockerfile", Message: "Analyzing codebase..."}})

					result, err := wfClient.Run(context.Background(), appID, source, mode, func(stepID, message string) {
						bus.Send(Event{
							Type:    EventPipelineProgress,
							Message: message,
							Payload: PipelineProgressPayload{StageId: stepID, Message: message},
						})
					}, func(ctx context.Context, approval *ApprovalContext) (bool, error) {
						// Human-in-the-loop: show proposal (including prompt) via AgentUI, then wait for input
						if approval != nil {
							bus.Send(Event{
								Type: EventApprovalNeeded,
								Payload: ApprovalNeededPayload{
									Dockerfile:      approval.Dockerfile,
									Summary:         approval.Summary,
									ValidationScore: approval.ValidationScore,
									Suggestions:     approval.Suggestions,
								},
							})
						} else {
							term.Println("")
							term.Print("Approve deployment? [y/N]: ")
						}
						scanner := bufio.NewScanner(os.Stdin)
						if !scanner.Scan() {
							return false, scanner.Err()
						}
						s := strings.ToLower(strings.TrimSpace(scanner.Text()))
						return s == "y" || s == "yes", nil
					})
					if err != nil {
						bus.Send(Event{Type: EventError, Message: fmt.Sprintf("Workflow failed: %v", err)})
						bus.Send(Event{Type: EventBuildStatus, Message: err.Error(), Payload: BuildStatusPayload{Phase: "error", Message: err.Error(), Error: err.Error()}, NeedsLLM: true})
						return
					}

					triggerPayload := result.ToTriggerBuildPayload()
					if err := client.Send(mover.SyncMessage{
						Type:      mover.MessageTypeTriggerBuild,
						Timestamp: time.Now(),
						Payload:   triggerPayload,
					}); err != nil {
						bus.Send(Event{Type: EventError, Message: fmt.Sprintf("Failed to send Dockerfile: %v", err)})
						return
					}

					bus.Send(Event{Type: EventBuildStatus, Message: "Dockerfile sent, starting build...", Payload: BuildStatusPayload{Phase: "dockerfile_ready", Message: "Dockerfile sent to server"}})
				}()
			}
		},
		SyncStatePath: syncStatePath,
		ApplicationID: applicationID,
		ForceFullSync: forceFullSyncFlag,
		EnvFilePath:   envFilePath,
	})
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create sync engine: %w", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	bus.Send(Event{Type: EventSyncStart})
	if err := engine.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start sync: %w", err)
	}

	// Periodic sync progress reporter
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		lastSynced := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				syncMu.Lock()
				current := filesSynced
				syncMu.Unlock()
				if current != lastSynced {
					bus.Send(Event{Type: EventSyncProgress, Message: fmt.Sprintf("Syncing... %d files", current)})
					lastSynced = current
				}
			}
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigChan:
		cancel()
	case <-uiDone:
	}

	if err := engine.Stop(); err != nil {
		return fmt.Errorf("failed to stop sync engine: %w", err)
	}

	return nil
}

// connectionStatusFromMover maps mover connection states to tracker connection statuses.
func connectionStatusFromMover(state mover.ConnectionState) mover.ConnectionStatus {
	switch state {
	case mover.StateConnected:
		return mover.ConnectionStatusConnected
	case mover.StateConnecting:
		return mover.ConnectionStatusConnecting
	case mover.StateReconnecting:
		return mover.ConnectionStatusReconnecting
	case mover.StateDisconnected:
		return mover.ConnectionStatusDisconnected
	default:
		return mover.ConnectionStatusDisconnected
	}
}

// buildWebSocketURL builds the WebSocket URL from server URL
// buildWebSocketURL constructs the WebSocket URL for live deployment
func buildWebSocketURL(server, projectID, accessToken string) string {
	// Convert http:// to ws:// and https:// to wss://
	wsURL := server
	if strings.HasPrefix(wsURL, "http://") {
		wsURL = "ws://" + wsURL[7:]
	} else if strings.HasPrefix(wsURL, "https://") {
		wsURL = "wss://" + wsURL[8:]
	}

	// Add WebSocket path and query params (using projectID as application_id)
	if accessToken != "" {
		wsURL += "/ws/live/" + projectID + "?token=" + accessToken
	} else {
		wsURL += "/ws/live/" + projectID
	}
	return wsURL
}

// buildDomainURL builds the domain URL from project ID
// deployDomain is optional; when empty, uses config.GetDeployDomain()
// Format: https://{first-8-chars-of-project-id}.{deploy_domain}
func buildDomainURL(projectID, deployDomain string) string {
	if projectID == "" || len(projectID) < 8 {
		return ""
	}
	if deployDomain == "" {
		deployDomain = config.GetDeployDomain()
	}
	return "https://" + projectID[:8] + "." + deployDomain
}

// extractCodebaseIndexedPayload extracts app ID, org ID, source, and mode from codebase_indexed payload.
func extractCodebaseIndexedPayload(payload interface{}) (appID, orgID, source, mode string) {
	if p, ok := payload.(mover.CodebaseIndexedPayload); ok {
		return p.ApplicationID, p.OrganizationID, p.Source, p.Mode
	}
	if m, ok := payload.(map[string]interface{}); ok {
		if v, ok := m["application_id"].(string); ok {
			appID = v
		}
		if v, ok := m["organization_id"].(string); ok {
			orgID = v
		}
		if v, ok := m["source"].(string); ok {
			source = v
		}
		if v, ok := m["mode"].(string); ok {
			mode = v
		}
		return appID, orgID, source, mode
	}
	return "", "", "", ""
}

// isGitRepo checks if the given path is a git repository
func isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// ensureInitialized checks if the project is initialized and initializes it if needed
func ensureInitialized(envPathFlag string) (*config.Config, error) {
	// Try to load config
	cfg, err := config.Load()
	if err == nil {
		// Config exists, check if it's complete
		if cfg.FamilyID != "" && len(cfg.Applications) > 0 {
			// Already initialized, return config
			return cfg, nil
		}
		// Config exists but incomplete, need to initialize
	}

	// Check if user is authenticated before attempting initialization
	// This should not happen if EnsureAuthenticated was called, but double-check
	accessToken, err := config.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("not authenticated: %w", err)
	}

	// Need to initialize - run init steps
	server := config.GetServerURL()

	// Validate env path if provided
	if envPathFlag != "" {
		if err := config.ValidateEnvPath(envPathFlag); err != nil {
			return nil, fmt.Errorf("invalid env path: %w", err)
		}
	}

	// Step 1: Parse env file if provided
	var envVars map[string]string
	if envPathFlag != "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		fullPath := filepath.Join(cwd, filepath.Clean(envPathFlag))
		envVars, err = godotenv.Read(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read env file %s: %w", envPathFlag, err)
		}
	}

	// Step 2: Create project (use access token from global auth)
	projectID, familyID, deployDomain, err := createProject(server, envVars, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Step 3: Save config (keep .env excluded - values are sent separately, not the file)
	exclude := []string{
		"*.log",
		".git",
		"node_modules",
		"__pycache__",
		".env",
	}

	applications := map[string]string{
		"default": projectID,
	}

	newCfg := &config.Config{
		Server:       config.GetServerURL(),
		FamilyID:     familyID,
		Applications: applications,
		Sync: config.SyncConfig{
			DebounceMs: 300,
			Exclude:    exclude,
		},
		EnvPath:      envPathFlag,
		DeployDomain: deployDomain,
	}

	cfg = newCfg

	if err := cfg.Save(); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return cfg, nil
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name                 string            `json:"name"`
	Repository           string            `json:"repository"`
	Branch               string            `json:"branch,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
}

// CreateProjectResponse represents the response from project creation endpoint
type CreateProjectResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	ProjectID    string `json:"project_id"`
	FamilyID     string `json:"family_id"`
	DeployDomain string `json:"deploy_domain,omitempty"`
}

// getGitBranch gets the current git branch
func getGitBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getGitInfo gets the git repository name and remote URL
func getGitInfo() (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	repoName := filepath.Base(cwd)

	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return repoName, repoName, nil
	}

	repoURL := strings.TrimSpace(string(output))
	if repoURL == "" {
		return repoName, repoName, nil
	}

	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		lastPart := strings.TrimSuffix(parts[len(parts)-1], ".git")
		if lastPart != "" {
			repoName = lastPart
		}
	}

	return repoName, repoURL, nil
}

// createProject creates a draft project on the server using the CLI init endpoint
// Returns: projectID, familyID, deployDomain, error
func createProject(serverURL string, envVars map[string]string, accessToken string) (string, string, string, error) {
	repoName, repoURL, err := getGitInfo()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get git info: %w", err)
	}

	branch := getGitBranch()

	if accessToken == "" {
		return "", "", "", fmt.Errorf("not authenticated")
	}

	// Use authenticated HTTP client
	client := httpclient.NewAuthenticatedHTTPClient(accessToken)
	url := httpclient.BuildURL(serverURL, "/api/v1/auth/cli/init")

	reqBody := CreateProjectRequest{
		Name:                 repoName,
		Repository:           repoURL,
		Branch:               branch,
		EnvironmentVariables: envVars,
	}

	resp, err := client.Post(url, reqBody)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to connect to server: %w", err)
	}

	bodyBytes, err := httpclient.ReadResponseBody(resp)
	if err != nil {
		return "", "", "", err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", "", "", httpclient.HandleErrorResponse(resp, bodyBytes, "failed to create project")
	}

	var projectResp CreateProjectResponse
	if err := httpclient.ParseJSONResponse(bodyBytes, &projectResp); err != nil {
		return "", "", "", err
	}

	if projectResp.ProjectID == "" {
		return "", "", "", fmt.Errorf("project ID not found in response")
	}

	if projectResp.FamilyID == "" {
		return "", "", "", fmt.Errorf("family ID not found in response")
	}

	return projectResp.ProjectID, projectResp.FamilyID, projectResp.DeployDomain, nil
}
