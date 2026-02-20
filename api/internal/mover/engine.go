package mover

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
)

const (
	defaultDebounceMs      = 100 // Fast response with fsnotify
	largeSyncThreshold     = 50  // Use parallel sync earlier (was 100)
	chunkSize              = 64 * 1024
	manifestWaitTimeout    = 5 * time.Second
	defaultSyncWorkers     = 20 // Scale for 2000+ file projects
	defaultSyncConcurrency = 50
)

var (
	syncWorkers     int
	syncConcurrency int
)

func init() {
	if v, err := strconv.Atoi(os.Getenv("NIXOPUS_MOVER_SYNC_WORKERS")); err == nil && v > 0 {
		syncWorkers = v
	} else {
		syncWorkers = defaultSyncWorkers
	}
	if v, err := strconv.Atoi(os.Getenv("NIXOPUS_MOVER_SYNC_CONCURRENCY")); err == nil && v > 0 {
		syncConcurrency = v
	} else {
		syncConcurrency = defaultSyncConcurrency
	}
}

// ChangeType represents the type of file change
type ChangeType string

const (
	ChangeAdded    ChangeType = "A"
	ChangeModified ChangeType = "M"
	ChangeDeleted  ChangeType = "D"
)

// FileChangeEvent represents a file change event (internal use)
type FileChangeEvent struct {
	Path string
	Type ChangeType
}

// Engine orchestrates file watching and syncing.
// Uses fsnotify for real-time file system watching (like Next.js/Vite).
type Engine struct {
	rootPath      string
	fileWatcher   *Watcher
	client        *Client
	excludes      []string
	debounceDelay time.Duration
	stopChan      chan struct{}
	syncedFiles   map[string]string // path -> checksum of synced files
	syncedMu      sync.RWMutex

	syncState     *SyncState // Optional persisted sync state
	applicationID string     // Application ID for sync state (when syncState is set)
	forceFullSync bool       // When true, skip manifest merge and sync all files

	// Connection state tracking
	isConnected   bool
	connectedMu   sync.RWMutex
	onStateChange func(ConnectionEvent)

	// File sync and change tracking callbacks
	onFileSynced     func(string)      // Called when a file is successfully synced
	onChangeDetected func(string)      // Called when a file change is detected
	onServerMessage  func(SyncMessage) // Called for server-originated messages

	// Pending changes during disconnect
	pendingChanges []FileChangeEvent
	pendingMu      sync.Mutex

	// Env file (values only, never synced)
	envFilePath   string
	envStopChan   chan struct{}
	envDebounceMu sync.Mutex
	envDebounceT  *time.Timer
}

// EngineConfig holds configuration for the sync engine
type EngineConfig struct {
	RootPath         string
	Client           *Client
	Excludes         []string
	DebounceMs       int
	OnStateChange    func(ConnectionEvent)
	OnFileSynced     func(string)      // Called when a file is successfully synced
	OnChangeDetected func(string)      // Called when a file change is detected
	OnServerMessage  func(SyncMessage) // Called for server-originated messages (pipeline_progress, etc.)
	SyncStatePath    string            // Path to .nixopus-sync-state.json for persisted sync state (empty = disabled)
	ApplicationID    string            // Application ID for multi-app state (required if SyncStatePath is set)
	ForceFullSync    bool              // If true, skip loading state and clear persisted state; sync all files
	EnvFilePath      string            // Absolute path to .env file; when set, values are sent (file is never synced)
}

// NewEngine creates a new sync engine using fsnotify for file watching.
func NewEngine(cfg EngineConfig) (*Engine, error) {
	debounceMs := cfg.DebounceMs
	if debounceMs <= 0 {
		debounceMs = defaultDebounceMs
	}

	// Create file system watcher
	fw, err := New(Config{
		RootPath:         cfg.RootPath,
		DebounceMs:       debounceMs,
		IgnorePatterns:   cfg.Excludes,
		EventsBufferSize: 512, // Handle bursty events (npm install, large refactors)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	syncedFiles := make(map[string]string)
	var syncState *SyncState
	if cfg.SyncStatePath != "" && cfg.ApplicationID != "" {
		syncState = NewSyncState(cfg.SyncStatePath)
		if !cfg.ForceFullSync {
			if err := syncState.Load(); err != nil {
				// Log but continue with empty state (full sync)
			}
			if paths := syncState.GetPaths(cfg.ApplicationID); paths != nil {
				for p, c := range paths {
					syncedFiles[p] = c
				}
			}
		} else {
			// Clear persisted state for this app so next run is also fresh
			syncState.SetAllPaths(cfg.ApplicationID, nil)
			_ = syncState.Save()
		}
	}

	engine := &Engine{
		rootPath:         cfg.RootPath,
		client:           cfg.Client,
		excludes:         cfg.Excludes,
		debounceDelay:    time.Duration(debounceMs) * time.Millisecond,
		fileWatcher:      fw,
		stopChan:         make(chan struct{}),
		syncedFiles:      syncedFiles,
		syncState:        syncState,
		applicationID:    cfg.ApplicationID,
		forceFullSync:    cfg.ForceFullSync,
		isConnected:      true, // Assume connected initially
		pendingChanges:   make([]FileChangeEvent, 0),
		onStateChange:    cfg.OnStateChange,
		onFileSynced:     cfg.OnFileSynced,
		onChangeDetected: cfg.OnChangeDetected,
		onServerMessage:  cfg.OnServerMessage,
		envFilePath:      cfg.EnvFilePath,
	}
	if cfg.EnvFilePath != "" {
		engine.envStopChan = make(chan struct{})
	}

	return engine, nil
}

// HandleConnectionEvent handles connection state changes from the transport client
func (e *Engine) HandleConnectionEvent(event ConnectionEvent) {
	e.connectedMu.Lock()
	wasConnected := e.isConnected
	e.isConnected = event.State == StateConnected
	isNowConnected := e.isConnected
	e.connectedMu.Unlock()

	// Forward to user callback if set
	if e.onStateChange != nil {
		e.onStateChange(event)
	}

	// If we just reconnected, flush pending changes
	if !wasConnected && isNowConnected {
		go e.flushPendingChanges()
	}
}

// flushPendingChanges sends any changes that were queued while disconnected
func (e *Engine) flushPendingChanges() {
	e.pendingMu.Lock()
	changes := e.pendingChanges
	e.pendingChanges = make([]FileChangeEvent, 0)
	e.pendingMu.Unlock()

	if len(changes) == 0 {
		return
	}

	// Deduplicate changes - keep only the latest change per file
	changeMap := make(map[string]FileChangeEvent)
	for _, change := range changes {
		changeMap[change.Path] = change
	}

	for _, change := range changeMap {
		// Check connection before attempting to send
		e.connectedMu.RLock()
		connected := e.isConnected
		e.connectedMu.RUnlock()

		if !connected {
			// Connection dropped, re-queue
			e.pendingMu.Lock()
			e.pendingChanges = append(e.pendingChanges, change)
			e.pendingMu.Unlock()
			continue
		}

		if err := e.sendChange(change); err != nil {
			// Only re-queue if connection is still up (transient error)
			// If connection dropped, it will be handled by connection event
			if strings.Contains(err.Error(), "client closed") {
				e.pendingMu.Lock()
				e.pendingChanges = append(e.pendingChanges, change)
				e.pendingMu.Unlock()
			}
			// Error handled via connection state callback
		}
	}
}

// Start begins watching for changes and syncing
func (e *Engine) Start() error {
	if err := e.InitialSync(); err != nil {
		return fmt.Errorf("initial sync failed: %w", err)
	}

	// Send initial env vars if configured (file is never synced, values only)
	if e.envFilePath != "" {
		e.sendEnvVars()
		go e.watchEnvFile()
	}

	// Start file system watcher
	if err := e.fileWatcher.Start(); err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	go e.watchLoop()
	go e.drainReceiveChannel() // Prevent receive channel from blocking read loop
	return nil
}

// InitialSync performs the initial full sync by walking the directory
// Skips files whose local checksum matches persisted/server state.
func (e *Engine) InitialSync() error {
	// Wait for server manifest and merge into syncedFiles (server is source of truth)
	// Skip merge when ForceFullSync so we sync everything
	manifest := e.waitForManifest()
	if !e.forceFullSync && manifest != nil && manifest.paths != nil {
		e.syncedMu.Lock()
		for p, c := range manifest.paths {
			e.syncedFiles[p] = c
		}
		e.syncedMu.Unlock()
	}

	files, err := e.getAllSyncableFiles()
	if err != nil {
		return fmt.Errorf("failed to get files: %w", err)
	}

	// Prune non-existent paths from persisted state (e.g. after git branch switch)
	if e.syncState != nil && e.applicationID != "" {
		existingSet := make(map[string]bool, len(files))
		for _, f := range files {
			existingSet[filepath.ToSlash(filepath.Clean(f))] = true
		}
		e.syncState.PruneNonexistentPaths(e.applicationID, existingSet)
		_ = e.syncState.Save()
	}

	// Build local Merkle tree and diff against server state
	filesToSync, filesToDelete := e.merkleDiffWithRootCheck(files, manifest)

	// Send deletes for files server has but we don't (e.g. deleted locally)
	for _, file := range filesToDelete {
		if err := e.sendDeleteMessage(file); err != nil {
			return fmt.Errorf("failed to send delete %s: %w", file, err)
		}
		if e.syncState != nil && e.applicationID != "" {
			e.syncState.RemovePath(e.applicationID, file)
		}
		e.syncedMu.Lock()
		delete(e.syncedFiles, file)
		e.syncedMu.Unlock()
	}

	// Use parallel processing for large syncs
	if len(filesToSync) > largeSyncThreshold {
		if err := e.parallelSync(filesToSync); err != nil {
			return err
		}
	} else {
		// Sequential sync for small syncs (less overhead)
		for _, file := range filesToSync {
			change := FileChangeEvent{Path: file, Type: ChangeAdded}
			if err := e.sendChange(change); err != nil {
				return fmt.Errorf("failed to send file %s: %w", file, err)
			}
		}
	}

	// Persist Merkle root for cache hint (Phase 3b)
	if e.syncState != nil && e.applicationID != "" {
		e.syncedMu.RLock()
		tree := BuildFromPaths(e.syncedFiles)
		e.syncedMu.RUnlock()
		if tree.RootHash != "" {
			e.syncState.SetRootHash(e.applicationID, tree.RootHash)
			_ = e.syncState.Save()
		}
	}

	// Signal server that initial sync is complete; triggers immediate build (no debounce).
	if err := e.client.Send(e.newSyncMessage(MessageTypeSyncComplete, nil)); err != nil {
		return fmt.Errorf("failed to send sync_complete: %w", err)
	}

	return nil
}

// manifestResult holds both paths and root_hash from server (Phase 3b).
type manifestResult struct {
	paths    map[string]string
	rootHash string
}

// waitForManifest waits for the server's manifest message (or timeout).
// Returns paths and root_hash from server, or nil paths if not received.
func (e *Engine) waitForManifest() *manifestResult {
	receive := e.client.Receive()
	deadline := time.After(manifestWaitTimeout)
	for {
		select {
		case msg, ok := <-receive:
			if !ok {
				return nil
			}
			if msg.Type == MessageTypeManifest {
				if m := extractManifestPayload(msg.Payload); m != nil {
					return m
				}
			}
		case <-deadline:
			return nil
		}
	}
}

// extractManifestPayload extracts ManifestPayload from message.
// Payload is already ManifestPayload from parseAndForwardMessage (no double encode).
func extractManifestPayload(payload interface{}) *manifestResult {
	if payload == nil {
		return nil
	}
	switch m := payload.(type) {
	case ManifestPayload:
		return &manifestResult{paths: m.Paths, rootHash: m.RootHash}
	case *ManifestPayload:
		if m != nil {
			return &manifestResult{paths: m.Paths, rootHash: m.RootHash}
		}
		return nil
	default:
		return nil
	}
}

// computeChecksumsParallel computes file checksums in parallel for initial sync diff.
// Returns path->checksum for successful reads; paths that fail are omitted (caller adds to toSync).
func (e *Engine) computeChecksumsParallel(files []string) map[string]string {
	if len(files) == 0 {
		return make(map[string]string)
	}
	workers := syncWorkers
	if workers > len(files) {
		workers = len(files)
	}
	type result struct {
		normPath string
		checksum string
	}
	results := make(chan result, len(files))
	jobs := make(chan string, len(files))
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				fullPath := filepath.Join(e.rootPath, file)
				checksum := e.computeFileChecksum(fullPath)
				if checksum != "" {
					results <- result{filepath.ToSlash(filepath.Clean(file)), checksum}
				}
			}
		}()
	}
	go func() {
		for _, f := range files {
			jobs <- f
		}
		close(jobs)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()
	localLeaves := make(map[string]string, len(files))
	for r := range results {
		localLeaves[r.normPath] = r.checksum
	}
	return localLeaves
}

// merkleDiffWithRootCheck builds a Merkle tree from local files and diffs against server state.
// Phase 3b: If server root_hash matches local root, skip full diff (0 files to sync).
// Uses parallel checksum computation for 2000+ file projects.
func (e *Engine) merkleDiffWithRootCheck(files []string, manifest *manifestResult) (toSync []string, toDelete []string) {
	localLeaves := e.computeChecksumsParallel(files)
	for _, file := range files {
		normPath := filepath.ToSlash(filepath.Clean(file))
		if _, ok := localLeaves[normPath]; !ok {
			toSync = append(toSync, normPath)
		}
	}

	tree := BuildFromPaths(localLeaves)

	// Phase 3b fast path: if roots match, nothing to sync or delete
	if manifest != nil && manifest.rootHash != "" && tree.RootHash != "" && manifest.rootHash == tree.RootHash {
		return nil, nil
	}
	// Both empty: server and local have no files
	if manifest != nil && manifest.rootHash == "" && tree.RootHash == "" && len(localLeaves) == 0 {
		e.syncedMu.RLock()
		serverEmpty := len(e.syncedFiles) == 0
		e.syncedMu.RUnlock()
		if serverEmpty {
			return nil, nil
		}
	}

	e.syncedMu.RLock()
	serverLeaves := make(map[string]string, len(e.syncedFiles))
	for p, c := range e.syncedFiles {
		serverLeaves[p] = c
	}
	e.syncedMu.RUnlock()

	return DiffAgainst(tree, serverLeaves)
}

// parallelSyncJob is pooled to reduce allocations during large syncs.
type parallelSyncJob struct {
	index int
	file  string
}

// parallelSyncResult carries per-file result for batch state update
type parallelSyncResult struct {
	path     string
	checksum string
	err      error
}

var parallelSyncJobPool = sync.Pool{
	New: func() interface{} { return &parallelSyncJob{} },
}

// parallelSync performs parallel file syncing using worker pool pattern.
// Batches syncedFiles and syncState updates at the end to reduce lock contention.
func (e *Engine) parallelSync(files []string) error {
	jobs := make(chan *parallelSyncJob, syncConcurrency)
	results := make(chan parallelSyncResult, len(files))
	var wg sync.WaitGroup

	// Start workers - send files without per-file state updates (batch at end)
	for w := 0; w < syncWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				change := FileChangeEvent{Path: j.file, Type: ChangeAdded}
				checksum, err := e.sendFileContentNoStateUpdate(change)
				if err != nil {
					results <- parallelSyncResult{path: j.file, err: fmt.Errorf("failed to send file %s: %w", j.file, err)}
				} else {
					results <- parallelSyncResult{path: j.file, checksum: checksum}
				}
				parallelSyncJobPool.Put(j)
			}
		}()
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for i, file := range files {
			j := parallelSyncJobPool.Get().(*parallelSyncJob)
			j.index, j.file = i, file
			jobs <- j
		}
	}()

	// Wait for completion
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and batch state update (one Lock for syncedFiles, one merge for syncState)
	var firstErr error
	batchedPaths := make(map[string]string)
	resultCount := 0
	for r := range results {
		resultCount++
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		if r.checksum != "" {
			normPath := filepath.ToSlash(filepath.Clean(r.path))
			batchedPaths[normPath] = r.checksum
		}
	}

	if resultCount != len(files) {
		if firstErr == nil {
			return fmt.Errorf("not all files were processed: expected %d, got %d results", len(files), resultCount)
		}
		return fmt.Errorf("sync incomplete: %w (processed %d/%d files)", firstErr, resultCount, len(files))
	}

	if firstErr != nil {
		return firstErr
	}

	// Batch update syncedFiles and syncState (avoids N locks during sync)
	if len(batchedPaths) > 0 {
		e.syncedMu.Lock()
		for p, c := range batchedPaths {
			e.syncedFiles[p] = c
		}
		e.syncedMu.Unlock()
		if e.syncState != nil && e.applicationID != "" {
			existing := e.syncState.GetPaths(e.applicationID)
			merged := make(map[string]string, len(existing)+len(batchedPaths))
			for p, c := range existing {
				merged[p] = c
			}
			for p, c := range batchedPaths {
				merged[p] = c
			}
			e.syncState.SetAllPaths(e.applicationID, merged)
		}
	}

	return nil
}

// Sync triggers a manual sync (re-syncs all files)
func (e *Engine) Sync() error {
	return e.InitialSync()
}

// Stop stops the sync engine
func (e *Engine) Stop() error {
	close(e.stopChan)
	if e.envFilePath != "" && e.envStopChan != nil {
		close(e.envStopChan)
	}
	if e.syncState != nil {
		_ = e.syncState.Save()
	}
	if e.fileWatcher != nil {
		return e.fileWatcher.Stop()
	}
	return nil
}

// sendEnvVars reads the env file, parses it, and sends the values to the server (file is never synced).
func (e *Engine) sendEnvVars() {
	if e.envFilePath == "" {
		return
	}
	data, err := os.ReadFile(e.envFilePath)
	if err != nil {
		return // File may not exist yet
	}
	envVars, err := godotenv.Unmarshal(string(data))
	if err != nil || len(envVars) == 0 {
		return
	}
	e.connectedMu.RLock()
	connected := e.isConnected
	e.connectedMu.RUnlock()
	if !connected {
		return
	}
	msg := e.newSyncMessage(MessageTypeEnvVars, EnvVarsPayload{Vars: envVars})
	_ = e.client.Send(msg)
}

// watchEnvFile watches the env file for changes and sends parsed values (file content is never synced).
func (e *Engine) watchEnvFile() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer w.Close()
	dir := filepath.Dir(e.envFilePath)
	if err := w.Add(dir); err != nil {
		return
	}
	for {
		select {
		case <-e.envStopChan:
			return
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Name != e.envFilePath {
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}
			// Debounce
			e.envDebounceMu.Lock()
			if e.envDebounceT != nil {
				e.envDebounceT.Stop()
			}
			e.envDebounceT = time.AfterFunc(e.debounceDelay, func() {
				e.envDebounceMu.Lock()
				e.envDebounceT = nil
				e.envDebounceMu.Unlock()
				e.sendEnvVars()
			})
			e.envDebounceMu.Unlock()
		}
	}
}

// drainReceiveChannel reads from client.Receive() to prevent the channel from filling
// and blocking the WebSocket read loop. Server-originated messages (like pipeline_progress,
// build_status, build_log, deployment_status) are forwarded via onServerMessage; other
// messages are discarded (manifest already consumed).
func (e *Engine) drainReceiveChannel() {
	receive := e.client.Receive()
	for {
		select {
		case <-e.stopChan:
			return
		case msg, ok := <-receive:
			if !ok {
				return
			}
			if e.onServerMessage != nil {
				switch msg.Type {
				case MessageTypePipelineProgress, MessageTypeBuildStatus,
					MessageTypeBuildLog, MessageTypeDeploymentStatus,
					MessageTypeCodebaseIndexed:
					e.onServerMessage(msg)
				}
			}
		}
	}
}

// watchLoop processes file system events from the watcher
func (e *Engine) watchLoop() {
	for {
		select {
		case <-e.stopChan:
			return

		case event := <-e.fileWatcher.Events():
			if err := e.handleWatcherEvent(event); err != nil {
				// Don't exit on "client closed" - the client will reconnect
				// and we'll flush pending changes then
				// Error handled via connection state callback
			}

		case err := <-e.fileWatcher.Errors():
			// Error handled via connection state callback
			_ = err
		}
	}
}

// handleWatcherEvent processes a single file system event
func (e *Engine) handleWatcherEvent(event Event) error {
	// Skip excluded paths
	if e.shouldExclude(event.Path) {
		return nil
	}

	var change FileChangeEvent
	change.Path = event.Path

	switch event.Type {
	case EventCreate:
		change.Type = ChangeAdded
	case EventModify:
		change.Type = ChangeModified
	case EventDelete:
		change.Type = ChangeDeleted
	case EventRename:
		change.Type = ChangeDeleted // Rename source is effectively deleted
	default:
		return nil
	}

	// For non-delete events, verify file exists and check if actually changed
	if change.Type != ChangeDeleted {
		fullPath := filepath.Join(e.rootPath, event.Path)
		info, err := os.Stat(fullPath)
		if err != nil {
			// File doesn't exist (maybe deleted between event and now)
			change.Type = ChangeDeleted
		} else if info.IsDir() {
			// Skip directories
			return nil
		} else {
			// Check if file actually changed (avoid duplicate syncs)
			checksum := e.computeFileChecksum(fullPath)
			if checksum == "" {
				// Failed to compute checksum, sync anyway
			} else {
				// Hold lock during check-and-update to prevent race condition
				e.syncedMu.Lock()
				prevChecksum := e.syncedFiles[event.Path]
				if checksum == prevChecksum {
					// File hasn't actually changed
					e.syncedMu.Unlock()
					return nil
				}
				// Update checksum while holding lock
				e.syncedFiles[event.Path] = checksum
				e.syncedMu.Unlock()
			}
		}
	} else {
		// For deletes, remove from synced files map
		e.syncedMu.Lock()
		delete(e.syncedFiles, event.Path)
		e.syncedMu.Unlock()
		if e.syncState != nil && e.applicationID != "" {
			e.syncState.RemovePath(e.applicationID, event.Path)
		}
	}

	// Check if connected - if not, queue the change
	e.connectedMu.RLock()
	connected := e.isConnected
	e.connectedMu.RUnlock()

	if !connected {
		e.pendingMu.Lock()
		e.pendingChanges = append(e.pendingChanges, change)
		e.pendingMu.Unlock()
		return nil
	}

	// Notify that a change was detected
	if e.onChangeDetected != nil {
		e.onChangeDetected(change.Path)
	}

	// Send change - if it fails due to disconnection, it will be handled
	// by the connection event handler which will queue it
	err := e.sendChange(change)
	if err != nil && strings.Contains(err.Error(), "client closed") {
		// Connection dropped, queue the change
		e.pendingMu.Lock()
		e.pendingChanges = append(e.pendingChanges, change)
		e.pendingMu.Unlock()
		return nil
	}

	return err
}

// getAllSyncableFiles walks the directory and returns all files to sync.
// Uses batched git check-ignore for 1000+ file projects (single exec instead of N).
func (e *Engine) getAllSyncableFiles() ([]string, error) {
	var candidates []string

	err := filepath.Walk(e.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		relPath, err := filepath.Rel(e.rootPath, path)
		if err != nil {
			return nil
		}
		if relPath == "." {
			return nil
		}

		if info.IsDir() {
			if e.shouldIgnoreDir(relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		if e.shouldExclude(relPath) {
			return nil
		}

		candidates = append(candidates, relPath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Batch git check-ignore: single exec for all paths (handles 2000+ files)
	ignored := e.batchGitCheckIgnore(candidates)
	if len(ignored) == 0 {
		return candidates, nil
	}
	ignoredSet := make(map[string]bool, len(ignored))
	for _, p := range ignored {
		ignoredSet[p] = true
	}

	files := make([]string, 0, len(candidates)-len(ignored))
	for _, p := range candidates {
		if !ignoredSet[p] {
			files = append(files, p)
		}
	}
	return files, nil
}

// batchGitCheckIgnore runs git check-ignore --stdin for all paths in one exec.
// Returns paths that are ignored. Empty on error (fallback: include all).
func (e *Engine) batchGitCheckIgnore(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "check-ignore", "--stdin")
	cmd.Dir = e.rootPath
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil
	}
	if err := cmd.Start(); err != nil {
		return nil
	}

	// Write paths, one per line
	go func() {
		defer stdin.Close()
		for _, p := range paths {
			stdin.Write([]byte(p + "\n"))
		}
	}()

	// Read ignored paths (output is one path per line)
	var ignored []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			ignored = append(ignored, line)
		}
	}
	cmd.Wait()
	return ignored
}

// shouldIgnoreDir checks if a directory should be skipped entirely
func (e *Engine) shouldIgnoreDir(relPath string) bool {
	// Always skip .git
	if relPath == ".git" || strings.HasPrefix(relPath, ".git/") {
		return true
	}

	// Always skip node_modules
	if relPath == "node_modules" || strings.HasPrefix(relPath, "node_modules/") {
		return true
	}

	// Check custom excludes
	for _, pattern := range e.excludes {
		if matched, _ := filepath.Match(pattern, filepath.Base(relPath)); matched {
			return true
		}
	}

	return false
}

// isGitIgnored checks if a file is ignored by git (used by watch path, not initial sync)
func (e *Engine) isGitIgnored(relPath string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "check-ignore", "-q", relPath)
	cmd.Dir = e.rootPath
	err := cmd.Run()
	return err == nil
}

// sendChange sends a file change to the server
func (e *Engine) sendChange(change FileChangeEvent) error {
	switch change.Type {
	case ChangeDeleted:
		return e.sendDeleteMessage(change.Path)
	case ChangeAdded, ChangeModified:
		return e.sendFileContent(change)
	default:
		return e.sendFileContent(change)
	}
}

// sendDeleteMessage sends a file delete message
func (e *Engine) sendDeleteMessage(path string) error {
	msg := e.newSyncMessage(MessageTypeFileDelete, FileChange{
		Path:      path,
		Operation: "delete",
	})
	return e.client.Send(msg)
}

// sendFileContent reads and sends file content to the server
func (e *Engine) sendFileContent(change FileChangeEvent) error {
	checksum, err := e.sendFileContentNoStateUpdate(change)
	if err != nil {
		return err
	}
	if checksum == "" {
		return nil // directory, skipped
	}

	// Store checksum for deduplication (used by watcher path, single-file updates)
	e.syncedMu.Lock()
	e.syncedFiles[change.Path] = checksum
	e.syncedMu.Unlock()

	if e.syncState != nil && e.applicationID != "" {
		e.syncState.SetPath(e.applicationID, change.Path, checksum)
	}

	if e.onFileSynced != nil {
		e.onFileSynced(change.Path)
	}
	return nil
}

// sendFileContentNoStateUpdate sends file content without updating syncedFiles or syncState.
// Used by parallelSync for batch state updates. Returns (checksum, error).
func (e *Engine) sendFileContentNoStateUpdate(change FileChangeEvent) (string, error) {
	fullPath := filepath.Join(e.rootPath, change.Path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}
	if info.IsDir() {
		return "", nil
	}

	content, checksum, err := e.readFileWithChecksum(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	if err := e.sendFileChangeNotification(change, info, checksum); err != nil {
		return "", err
	}
	if err = e.sendFileContentChunks(change.Path, content, checksum); err != nil {
		return "", err
	}

	if e.onFileSynced != nil {
		e.onFileSynced(change.Path)
	}
	return checksum, nil
}

// readFileWithChecksum reads a file and calculates its checksum
// Reads file once and computes checksum from the same buffer for efficiency
func (e *Engine) readFileWithChecksum(fullPath string) ([]byte, string, error) {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	hasher := sha256.New()
	hasher.Write(content)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	return content, checksum, nil
}

// computeFileChecksum is a helper that returns empty string on error
func (e *Engine) computeFileChecksum(fullPath string) string {
	file, err := os.Open(fullPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// sendFileChangeNotification sends the file change metadata
func (e *Engine) sendFileChangeNotification(change FileChangeEvent, info os.FileInfo, checksum string) error {
	msg := e.newSyncMessage(MessageTypeFileChange, FileChange{
		Path:      change.Path,
		Operation: changeTypeToOperation(change.Type),
		Size:      info.Size(),
		Checksum:  checksum,
		ModTime:   info.ModTime(),
	})
	return e.client.Send(msg)
}

// sendFileContentChunks sends file content in chunks
func (e *Engine) sendFileContentChunks(path string, content []byte, checksum string) error {
	totalChunks := (len(content) + chunkSize - 1) / chunkSize
	if totalChunks == 0 {
		totalChunks = 1
	}

	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(content) {
			end = len(content)
		}

		msg := e.newSyncMessage(MessageTypeFileContent, FileContent{
			Path:        path,
			ChunkIndex:  i,
			TotalChunks: totalChunks,
			Data:        content[start:end],
			Checksum:    checksum,
		})

		if err := e.client.Send(msg); err != nil {
			return err
		}
	}

	return nil
}

// newSyncMessage creates a new sync message with common fields
func (e *Engine) newSyncMessage(msgType MessageType, payload interface{}) SyncMessage {
	return SyncMessage{
		Type:      msgType,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

// shouldExclude checks if a path should be excluded
func (e *Engine) shouldExclude(path string) bool {
	// Always exclude .git
	if path == ".git" || strings.HasPrefix(path, ".git/") || strings.HasPrefix(path, ".git\\") {
		return true
	}

	// Always exclude node_modules
	if strings.Contains(path, "node_modules") {
		return true
	}

	for _, pattern := range e.excludes {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// changeTypeToOperation converts ChangeType to operation string
func changeTypeToOperation(ct ChangeType) string {
	switch ct {
	case ChangeAdded:
		return "create"
	case ChangeModified:
		return "modify"
	case ChangeDeleted:
		return "delete"
	default:
		return "modify"
	}
}
