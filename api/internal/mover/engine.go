package mover

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	defaultDebounceMs  = 100 // Fast response with fsnotify
	largeSyncThreshold = 100
	chunkSize          = 64 * 1024 // 64KB
)

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

	// Connection state tracking
	isConnected   bool
	connectedMu   sync.RWMutex
	onStateChange func(ConnectionEvent)

	// File sync and change tracking callbacks
	onFileSynced     func(string) // Called when a file is successfully synced
	onChangeDetected func(string) // Called when a file change is detected

	// Pending changes during disconnect
	pendingChanges []FileChangeEvent
	pendingMu      sync.Mutex
}

// EngineConfig holds configuration for the sync engine
type EngineConfig struct {
	RootPath         string
	Client           *Client
	Excludes         []string
	DebounceMs       int
	OnStateChange    func(ConnectionEvent)
	OnFileSynced     func(string) // Called when a file is successfully synced
	OnChangeDetected func(string) // Called when a file change is detected
}

// NewEngine creates a new sync engine using fsnotify for file watching.
func NewEngine(cfg EngineConfig) (*Engine, error) {
	debounceMs := cfg.DebounceMs
	if debounceMs <= 0 {
		debounceMs = defaultDebounceMs
	}

	// Create file system watcher
	fw, err := New(Config{
		RootPath:       cfg.RootPath,
		DebounceMs:     debounceMs,
		IgnorePatterns: cfg.Excludes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	engine := &Engine{
		rootPath:         cfg.RootPath,
		client:           cfg.Client,
		excludes:         cfg.Excludes,
		debounceDelay:    time.Duration(debounceMs) * time.Millisecond,
		fileWatcher:      fw,
		stopChan:         make(chan struct{}),
		syncedFiles:      make(map[string]string),
		isConnected:      true, // Assume connected initially
		pendingChanges:   make([]FileChangeEvent, 0),
		onStateChange:    cfg.OnStateChange,
		onFileSynced:     cfg.OnFileSynced,
		onChangeDetected: cfg.OnChangeDetected,
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

	// Start file system watcher
	if err := e.fileWatcher.Start(); err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	go e.watchLoop()
	return nil
}

// InitialSync performs the initial full sync by walking the directory
// Uses parallel processing for better performance
func (e *Engine) InitialSync() error {
	files, err := e.getAllSyncableFiles()
	if err != nil {
		return fmt.Errorf("failed to get files: %w", err)
	}

	totalFiles := len(files)

	// Use parallel processing for large syncs
	if totalFiles > largeSyncThreshold {
		return e.parallelSync(files)
	}

	// Sequential sync for small syncs (less overhead)
	for _, file := range files {
		change := FileChangeEvent{
			Path: file,
			Type: ChangeAdded,
		}
		if err := e.sendChange(change); err != nil {
			return fmt.Errorf("failed to send file %s: %w", file, err)
		}
		// Notify file synced (sendChange already calls onFileSynced on success)
	}

	return nil
}

// parallelSync performs parallel file syncing using worker pool pattern
func (e *Engine) parallelSync(files []string) error {
	const maxWorkers = 10     // Optimal balance between speed and resource usage
	const maxConcurrency = 20 // Maximum concurrent file operations

	// Create worker pool
	type job struct {
		index int
		file  string
	}

	jobs := make(chan job, maxConcurrency)
	results := make(chan error, len(files))
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				change := FileChangeEvent{
					Path: j.file,
					Type: ChangeAdded,
				}
				if err := e.sendChange(change); err != nil {
					results <- fmt.Errorf("failed to send file %s: %w", j.file, err)
				} else {
					results <- nil
					// sendChange already calls onFileSynced on success
				}
				// Continue processing other jobs even if one fails
			}
		}()
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for i, file := range files {
			jobs <- job{index: i, file: file}
		}
	}()

	// Wait for completion
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results - count to ensure all files were processed
	var firstErr error
	resultCount := 0
	for err := range results {
		resultCount++
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// Verify all files were processed
	if resultCount != len(files) {
		if firstErr == nil {
			return fmt.Errorf("not all files were processed: expected %d, got %d results", len(files), resultCount)
		}
		return fmt.Errorf("sync incomplete: %w (processed %d/%d files)", firstErr, resultCount, len(files))
	}

	if firstErr != nil {
		return firstErr
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
	if e.fileWatcher != nil {
		return e.fileWatcher.Stop()
	}
	return nil
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

// getAllSyncableFiles walks the directory and returns all files to sync
func (e *Engine) getAllSyncableFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(e.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		// Get relative path
		relPath, err := filepath.Rel(e.rootPath, path)
		if err != nil {
			return nil
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		// Skip directories (but continue walking into them)
		if info.IsDir() {
			// Skip ignored directories entirely
			if e.shouldIgnoreDir(relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip excluded files
		if e.shouldExclude(relPath) {
			return nil
		}

		// Check if file is gitignored
		if e.isGitIgnored(relPath) {
			return nil
		}

		files = append(files, relPath)
		return nil
	})

	return files, err
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

// isGitIgnored checks if a file is ignored by git
func (e *Engine) isGitIgnored(relPath string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "check-ignore", "-q", relPath)
	cmd.Dir = e.rootPath
	err := cmd.Run()
	// Exit code 0 = ignored, 1 = not ignored
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
	fullPath := filepath.Join(e.rootPath, change.Path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	if info.IsDir() {
		return nil
	}

	content, checksum, err := e.readFileWithChecksum(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Store checksum for deduplication
	e.syncedMu.Lock()
	e.syncedFiles[change.Path] = checksum
	e.syncedMu.Unlock()

	if err := e.sendFileChangeNotification(change, info, checksum); err != nil {
		return err
	}

	if err = e.sendFileContentChunks(change.Path, content, checksum); err != nil {
		return err
	}

	// File successfully synced
	if e.onFileSynced != nil {
		e.onFileSynced(change.Path)
	}

	return nil
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
