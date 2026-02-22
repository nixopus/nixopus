package cliconfig

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/pkg/cli/mover"
)

// Build-time variables (set via -ldflags). Must be present at build time.
var (
	// APIURL is the Nixopus API server URL (CLI_API_URL)
	APIURL string
	// BetterAuthURL is the Better Auth backend URL (CLI_BETTER_AUTH_URL)
	BetterAuthURL string
	// FrontendURL is the frontend URL for device code verification (CLI_FRONTEND_URL)
	FrontendURL string
	// OAuthClientID is the OAuth client ID for device flow (CLI_OAUTH_CLIENT_ID)
	OAuthClientID string
	// AgentEndpoint is the deployment workflow/agent endpoint (CLI_AGENT_ENDPOINT)
	AgentEndpoint string
	// InitTimeout is the initialization timeout duration (CLI_INIT_TIMEOUT, e.g. 30s)
	InitTimeout string
	// WorkflowTimeout is the deployment workflow timeout (CLI_WORKFLOW_TIMEOUT, e.g. 30m)
	WorkflowTimeout string
	// WorkflowID is the Mastra workflow ID (CLI_WORKFLOW_ID)
	WorkflowID string
	// DebugStream enables verbose stream logging when "1" (CLI_DEBUG_STREAM, optional)
	DebugStream string
	// ConfigFileName is the project config file name (CLI_CONFIG_FILE, optional, default .nixopus)
	ConfigFileName string
	// AuthFileName is the auth file name in ~/.config/nixopus/ (CLI_AUTH_FILE, optional, default auth.json)
	AuthFileName string
	// SyncStateFileName is the sync state file name (CLI_SYNC_STATE_FILE, optional, default sync-state.json)
	SyncStateFileName string
	// Mover vars (optional, see GetMoverConfig)
	MoverSendBuffer, MoverReceiveBuffer                                 string
	MoverWriteWait, MoverPongWait, MoverPingPeriod                      string
	MoverMaxMessageSize, MoverHandshakeTimeout                          string
	MoverInitialReconnectDelay, MoverMaxReconnectDelay                  string
	MoverReconnectBackoffRate, MoverMaxReconnectAttempts                string
	MoverCloseFlushDelay                                                string
	MoverDebounceMs, MoverLargeSyncThreshold, MoverChunkSize            string
	MoverManifestWaitTimeout                                            string
	MoverSyncWorkers, MoverSyncConcurrency                              string
	MoverEventsBufferSize, MoverWatcherDebounceMs, MoverGitCheckTimeout string
	MoverSyncStateDebounceMs                                            string
)

func init() {
	var missing []string
	if strings.TrimSpace(APIURL) == "" {
		missing = append(missing, "CLI_API_URL")
	}
	if strings.TrimSpace(BetterAuthURL) == "" {
		missing = append(missing, "CLI_BETTER_AUTH_URL")
	}
	if strings.TrimSpace(FrontendURL) == "" {
		missing = append(missing, "CLI_FRONTEND_URL")
	}
	if strings.TrimSpace(OAuthClientID) == "" {
		missing = append(missing, "CLI_OAUTH_CLIENT_ID")
	}
	if strings.TrimSpace(AgentEndpoint) == "" {
		missing = append(missing, "CLI_AGENT_ENDPOINT")
	}
	if strings.TrimSpace(InitTimeout) == "" {
		missing = append(missing, "CLI_INIT_TIMEOUT")
	}
	if strings.TrimSpace(WorkflowTimeout) == "" {
		missing = append(missing, "CLI_WORKFLOW_TIMEOUT")
	}
	if strings.TrimSpace(WorkflowID) == "" {
		missing = append(missing, "CLI_WORKFLOW_ID")
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "nixopus: build-time config missing: %s\nSet these at build time, e.g.:\n  make cli-build CLI_API_URL=https://api.example.com CLI_BETTER_AUTH_URL=...\nSee api/.env.cli for all required variables.\n", strings.Join(missing, ", "))
		os.Exit(1)
	}
}

// GetAPIURL returns the API URL (trimmed, no trailing slash).
func GetAPIURL() string {
	return strings.TrimRight(strings.TrimSpace(APIURL), "/")
}

// GetBetterAuthURL returns the Better Auth URL.
func GetBetterAuthURL() string {
	return strings.TrimRight(strings.TrimSpace(BetterAuthURL), "/")
}

// GetFrontendURL returns the frontend URL.
func GetFrontendURL() string {
	return strings.TrimRight(strings.TrimSpace(FrontendURL), "/")
}

// GetOAuthClientID returns the OAuth client ID.
func GetOAuthClientID() string {
	return strings.TrimSpace(OAuthClientID)
}

// GetAgentEndpoint returns the agent endpoint.
func GetAgentEndpoint() string {
	return strings.TrimRight(strings.TrimSpace(AgentEndpoint), "/")
}

// GetInitTimeout returns the init timeout duration.
func GetInitTimeout() (time.Duration, error) {
	d, err := time.ParseDuration(strings.TrimSpace(InitTimeout))
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("CLI_INIT_TIMEOUT must be a valid positive duration (e.g. 30s)")
	}
	return d, nil
}

// GetWorkflowTimeout returns the workflow timeout duration.
func GetWorkflowTimeout() (time.Duration, error) {
	d, err := time.ParseDuration(strings.TrimSpace(WorkflowTimeout))
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("CLI_WORKFLOW_TIMEOUT must be a valid positive duration (e.g. 30m)")
	}
	return d, nil
}

// GetWorkflowID returns the workflow ID.
func GetWorkflowID() string {
	return strings.TrimSpace(WorkflowID)
}

// IsDebugStream returns true when CLI_DEBUG_STREAM=1 (optional, defaults to false).
func IsDebugStream() bool {
	return DebugStream == "1"
}

// GetConfigFileName returns the project config file name (optional, default .nixopus).
func GetConfigFileName() string {
	if s := strings.TrimSpace(ConfigFileName); s != "" {
		return s
	}
	return ".nixopus"
}

// GetAuthFileName returns the auth file name (optional, default auth.json).
func GetAuthFileName() string {
	if s := strings.TrimSpace(AuthFileName); s != "" {
		return s
	}
	return "auth.json"
}

// GetSyncStateFileName returns the sync state file name (optional, default sync-state.json).
func GetSyncStateFileName() string {
	if s := strings.TrimSpace(SyncStateFileName); s != "" {
		return s
	}
	return "sync-state.json"
}

// GetMoverConfig returns mover configuration from build-time vars (optional, uses defaults when unset).
func GetMoverConfig() mover.MoverConfig {
	parseDur := func(s string) time.Duration {
		d, err := time.ParseDuration(strings.TrimSpace(s))
		if err != nil || d <= 0 {
			return 0
		}
		return d
	}
	atoi := func(s string) int {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n < 0 {
			return 0
		}
		return n
	}
	atof := func(s string) float64 {
		f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil || f <= 0 {
			return 0
		}
		return f
	}
	atoi64 := func(s string) int64 {
		n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil || n <= 0 {
			return 0
		}
		return n
	}
	return mover.MoverConfig{
		SendBufferSize:        atoi(MoverSendBuffer),
		ReceiveBufferSize:     atoi(MoverReceiveBuffer),
		WriteWait:             parseDur(MoverWriteWait),
		PongWait:              parseDur(MoverPongWait),
		PingPeriod:            parseDur(MoverPingPeriod),
		MaxMessageSize:        atoi64(MoverMaxMessageSize),
		HandshakeTimeout:      parseDur(MoverHandshakeTimeout),
		InitialReconnectDelay: parseDur(MoverInitialReconnectDelay),
		MaxReconnectDelay:     parseDur(MoverMaxReconnectDelay),
		ReconnectBackoffRate:  atof(MoverReconnectBackoffRate),
		MaxReconnectAttempts:  atoi(MoverMaxReconnectAttempts),
		CloseFlushDelay:       parseDur(MoverCloseFlushDelay),
		DebounceMs:            atoi(MoverDebounceMs),
		LargeSyncThreshold:    atoi(MoverLargeSyncThreshold),
		ChunkSize:             atoi(MoverChunkSize),
		ManifestWaitTimeout:   parseDur(MoverManifestWaitTimeout),
		SyncWorkers:           atoi(MoverSyncWorkers),
		SyncConcurrency:       atoi(MoverSyncConcurrency),
		EventsBufferSize:      atoi(MoverEventsBufferSize),
		WatcherDebounceMs:     atoi(MoverWatcherDebounceMs),
		GitCheckTimeout:       parseDur(MoverGitCheckTimeout),
		SyncStateDebounceMs:   atoi(MoverSyncStateDebounceMs),
	}
}
