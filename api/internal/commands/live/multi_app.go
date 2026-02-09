package live

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/commands/logincmd"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/mover"
)

// AppSession manages a single app's connection and sync
type AppSession struct {
	name          string
	applicationID string
	basePath      string
	domainURL     string
	client        *mover.Client
	engine        *mover.Engine
	tracker       *mover.Tracker
	poller        *DeploymentPoller
	pollerCancel  context.CancelFunc
	error         error
	mu            sync.RWMutex
}

// GetStatus returns the current status of this app session
func (s *AppSession) GetStatus() mover.ConnectionStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.tracker == nil {
		return mover.ConnectionStatusDisconnected
	}
	return s.tracker.GetConnectionStatus()
}

// GetStatusInfo returns the status info for this app
func (s *AppSession) GetStatusInfo() mover.StatusInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.tracker == nil {
		return mover.StatusInfo{
			ConnectionStatus: mover.ConnectionStatusDisconnected,
		}
	}
	info := s.tracker.GetStatusInfo()
	// Override URL with app-specific domain
	if s.domainURL != "" {
		info.URL = s.domainURL
	}
	return info
}

// Stop stops the app session
func (s *AppSession) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error
	if s.engine != nil {
		if err := s.engine.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop engine for %s: %w", s.name, err))
		}
	}
	if s.client != nil {
		s.client.Close()
	}
	if s.poller != nil {
		s.poller.Stop()
	}
	if s.pollerCancel != nil {
		s.pollerCancel()
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors stopping %s: %v", s.name, errs)
	}
	return nil
}

// runAllApps runs all apps in the family simultaneously
func runAllApps(args []string) error {
	// Check authentication first - prompt for login if not authenticated
	if err := logincmd.EnsureAuthenticated(); err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	// Load configuration
	// Ensure initialization before loading config
	cfg, err := ensureInitialized("")
	if err != nil {
		return fmt.Errorf("failed to initialize or load config: %w", err)
	}

	// Get all apps from config
	if len(cfg.Applications) == 0 {
		return fmt.Errorf("no applications found in config. Use 'nixopus add' to add applications")
	}

	// Deduplicate: if multiple names point to same app ID, prefer "default" over others
	// This prevents showing the same app twice (e.g., "default" and "root" pointing to same ID)
	preferredNames := []string{"default", "root"} // preference order
	appIDToName := make(map[string]string)        // appID -> best name found so far

	// First pass: collect all app IDs and track the most preferred name for each
	for name, appID := range cfg.Applications {
		if appID == "" {
			continue // Skip empty app IDs
		}
		if bestName, exists := appIDToName[appID]; exists {
			// Already seen this app ID, check if current name is more preferred
			currentPreference := getPreferenceIndex(name, preferredNames)
			bestPreference := getPreferenceIndex(bestName, preferredNames)
			if currentPreference < bestPreference {
				appIDToName[appID] = name
			}
		} else {
			appIDToName[appID] = name
		}
	}

	// Second pass: create sessions only for the most preferred name of each app
	// This ensures we have one session per unique application ID
	sessions := make([]*AppSession, 0, len(appIDToName))
	for appID, preferredName := range appIDToName {
		if appID == "" {
			continue // Skip empty app IDs
		}
		session := &AppSession{
			name:          preferredName,
			applicationID: appID,
		}
		sessions = append(sessions, session)
	}

	// Ensure we have at least one session
	if len(sessions) == 0 {
		return fmt.Errorf("no valid applications found in config. Use 'nixopus add' to add applications")
	}

	// Create multi-app tracker
	tracker := mover.NewMultiAppTracker()

	// Initialize tracker with all sessions upfront (so UI shows them immediately)
	for _, session := range sessions {
		tracker.UpdateSession(session.name, mover.AppSessionInfo{
			Name:          session.name,
			ApplicationID: session.applicationID,
			Status:        mover.ConnectionStatusConnecting,
			Error:         nil,
		})
	}

	// Start UI immediately
	program := NewMultiAppProgram(tracker)
	done := make(chan error, 1)
	programStarted := make(chan bool, 1)
	go func() {
		programStarted <- true
		done <- program.Start()
	}()

	// Wait for program to start
	<-programStarted
	time.Sleep(200 * time.Millisecond)

	// Initialize all sessions in parallel
	var wg sync.WaitGroup
	for _, session := range sessions {
		wg.Add(1)
		go func(s *AppSession) {
			defer wg.Done()
			if err := initializeAppSession(s, cfg, tracker); err != nil {
				s.mu.Lock()
				s.error = err
				s.mu.Unlock()
				// Update tracker with error
				tracker.UpdateSession(s.name, mover.AppSessionInfo{
					Name:          s.name,
					ApplicationID: s.applicationID,
					Status:        mover.ConnectionStatusDisconnected,
					Error:         err,
				})
			}
		}(session)
	}
	wg.Wait()

	// Start all sync engines
	for _, session := range sessions {
		if session.error == nil && session.engine != nil {
			go func(s *AppSession) {
				if err := s.engine.Start(); err != nil {
					s.mu.Lock()
					s.error = err
					s.mu.Unlock()
					// Update tracker with error
					info := s.GetStatusInfo()
					tracker.UpdateSession(s.name, mover.AppSessionInfo{
						Name:            s.name,
						ApplicationID:   s.applicationID,
						Status:          mover.ConnectionStatusDisconnected,
						FilesSynced:     info.FilesSynced,
						ChangesDetected: info.ChangesDetected,
						URL:             info.URL,
						Deployment:      info.Deployment,
						Error:           err,
					})
				}
			}(session)
		} else if session.error != nil {
			// Session failed to initialize, make sure tracker is updated
			tracker.UpdateSession(session.name, mover.AppSessionInfo{
				Name:          session.name,
				ApplicationID: session.applicationID,
				Status:        mover.ConnectionStatusDisconnected,
				Error:         session.error,
			})
		}
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

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

	// Stop all sessions
	var stopErrs []error
	for _, session := range sessions {
		if err := session.Stop(); err != nil {
			stopErrs = append(stopErrs, err)
		}
	}
	if len(stopErrs) > 0 {
		return fmt.Errorf("errors stopping sessions: %v", stopErrs)
	}

	return nil
}

// initializeAppSession initializes a single app session
func initializeAppSession(session *AppSession, cfg *config.Config, tracker *mover.MultiAppTracker) error {
	// Get access token from global auth storage
	accessToken, err := config.GetAccessToken()
	if err != nil {
		return fmt.Errorf("not authenticated: %w", err)
	}

	// Fetch application details
	basePath, domainURL, err := getApplicationDetailsWithURL(cfg.Server, session.applicationID, accessToken)
	if err != nil {
		return fmt.Errorf("failed to fetch application details: %w", err)
	}

	session.mu.Lock()
	session.basePath = basePath
	session.domainURL = domainURL
	session.tracker = mover.NewTracker()
	session.mu.Unlock()

	// Set domain URL in tracker
	if domainURL != "" {
		session.tracker.SetURL(domainURL)
	}

	// Create WebSocket client
	wsURL := buildWebSocketURL(cfg.Server, session.applicationID, accessToken)
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
		session.tracker.SetConnectionStatus(statusState)

		// Update multi-app tracker
		info := session.GetStatusInfo()
		tracker.UpdateSession(session.name, mover.AppSessionInfo{
			Name:            session.name,
			ApplicationID:   session.applicationID,
			Status:          info.ConnectionStatus,
			FilesSynced:     info.FilesSynced,
			ChangesDetected: info.ChangesDetected,
			URL:             info.URL,
			Deployment:      info.Deployment,
		})
	}

	session.tracker.SetConnectionStatus(mover.ConnectionStatusConnecting)
	client, err := mover.NewClient(
		wsURL,
		"", // TODO: Add session token
		mover.WithOnStateChange(onStateChange),
	)
	if err != nil {
		return fmt.Errorf("failed to create WebSocket client: %w", err)
	}

	session.mu.Lock()
	session.client = client
	session.mu.Unlock()

	// Get repository root
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if !isGitRepo(wd) {
		return fmt.Errorf("git repository required: not a git repository in %s", wd)
	}

	// Determine watch path based on base_path
	watchPath := wd
	if basePath != "" && basePath != "/" {
		normalizedBasePath := strings.Trim(basePath, "/")
		if normalizedBasePath != "" {
			watchPath = filepath.Join(wd, normalizedBasePath)
			if _, err := os.Stat(watchPath); os.IsNotExist(err) {
				return fmt.Errorf("base_path '%s' does not exist in repository", basePath)
			}
		}
	}

	// Create sync engine
	engine, err := mover.NewEngine(mover.EngineConfig{
		RootPath:      watchPath,
		Client:        client,
		Excludes:      cfg.Sync.Exclude,
		DebounceMs:    cfg.Sync.DebounceMs,
		OnStateChange: onStateChange,
		OnFileSynced: func(path string) {
			session.tracker.IncrementFilesSynced()
			// Note: Tracker updates are batched via the 5-second ticker in poller goroutine
			// This reduces mutex contention when many apps are syncing files simultaneously
		},
		OnChangeDetected: func(path string) {
			session.tracker.IncrementChanges()
			// Note: Tracker updates are batched via the 5-second ticker in poller goroutine
			// This reduces mutex contention when many apps are detecting changes simultaneously
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create sync engine: %w", err)
	}

	session.mu.Lock()
	session.engine = engine
	session.mu.Unlock()

	// Start deployment poller with callback to update multi-app tracker
	pollerCtx, pollerCancel := context.WithCancel(context.Background())
	poller := NewDeploymentPoller(cfg, session.tracker, session.applicationID)

	// Wrap the poller's update function to also update multi-app tracker
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				info := session.GetStatusInfo()
				tracker.UpdateSession(session.name, mover.AppSessionInfo{
					Name:            session.name,
					ApplicationID:   session.applicationID,
					Status:          info.ConnectionStatus,
					FilesSynced:     info.FilesSynced,
					ChangesDetected: info.ChangesDetected,
					URL:             info.URL,
					Deployment:      info.Deployment,
				})
			case <-pollerCtx.Done():
				return
			}
		}
	}()

	go poller.Start(pollerCtx)

	session.mu.Lock()
	session.poller = poller
	session.pollerCancel = pollerCancel
	session.mu.Unlock()

	// Initial tracker update
	info := session.GetStatusInfo()
	tracker.UpdateSession(session.name, mover.AppSessionInfo{
		Name:            session.name,
		ApplicationID:   session.applicationID,
		Status:          info.ConnectionStatus,
		FilesSynced:     info.FilesSynced,
		ChangesDetected: info.ChangesDetected,
		URL:             info.URL,
		Deployment:      info.Deployment,
	})

	return nil
}

// getApplicationDetailsWithURL fetches application details and returns base_path and domain URL
func getApplicationDetailsWithURL(server, applicationID, accessToken string) (string, string, error) {
	basePath, err := getApplicationDetails(server, applicationID, accessToken)
	if err != nil {
		return "", "", err
	}

	domainURL := buildDomainURL(applicationID)
	return basePath, domainURL, nil
}

// getPreferenceIndex returns the index of name in preferredNames, or a large number if not found
func getPreferenceIndex(name string, preferredNames []string) int {
	for i, preferred := range preferredNames {
		if name == preferred {
			return i
		}
	}
	return len(preferredNames) // Not in preferred list, lowest priority
}
