package mover

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const syncStateVersion = 1

// persistableState is the JSON structure written to disk.
type persistableState struct {
	Applications map[string]appState `json:"applications"`
	Version      int                 `json:"version"`
}

type appState struct {
	Paths     map[string]string `json:"paths"`
	RootHash  string            `json:"root_hash,omitempty"` // Merkle root for cache hint
	UpdatedAt time.Time         `json:"updated_at"`
}

// SyncState persists path→checksum state per application for incremental sync.
// State is stored in .nixopus-sync-state.json in the project root.
type SyncState struct {
	path   string
	mu     sync.RWMutex
	state  *persistableState
	dirty  bool
	saveMu sync.Mutex
}

// NewSyncState creates a SyncState that reads/writes to the given path.
// Path should be the full path to .nixopus-sync-state.json.
func NewSyncState(path string) *SyncState {
	return &SyncState{
		path:  path,
		state: &persistableState{Applications: make(map[string]appState), Version: syncStateVersion},
	}
}

// Load reads persisted state from disk. On failure (missing file, corrupt JSON),
// returns an empty state and no error so the caller can proceed with full sync.
func (s *SyncState) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // First run, empty state
		}
		return fmt.Errorf("failed to read sync state: %w", err)
	}

	var parsed persistableState
	if err := json.Unmarshal(data, &parsed); err != nil {
		log.Printf("sync state: corrupt or invalid JSON, treating as empty: %v", err)
		return nil
	}

	if parsed.Applications == nil {
		parsed.Applications = make(map[string]appState)
	}
	if parsed.Version != syncStateVersion {
		// Future: migration logic
		return nil
	}

	s.state = &parsed
	return nil
}

// GetRootHash returns the persisted Merkle root for the application (if any).
func (s *SyncState) GetRootHash(applicationID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, ok := s.state.Applications[applicationID]
	if !ok {
		return ""
	}
	return app.RootHash
}

// SetRootHash records the Merkle root for the application (cache hint).
func (s *SyncState) SetRootHash(applicationID, rootHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	app := s.state.Applications[applicationID]
	if app.Paths == nil {
		app.Paths = make(map[string]string)
	}
	app.RootHash = rootHash
	app.UpdatedAt = time.Now().UTC()
	s.state.Applications[applicationID] = app
	s.dirty = true
}

// GetPaths returns a copy of path→checksum for the given application ID.
// Paths are normalized (forward slashes, clean).
func (s *SyncState) GetPaths(applicationID string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	app, ok := s.state.Applications[applicationID]
	if !ok || app.Paths == nil {
		return nil
	}

	out := make(map[string]string, len(app.Paths))
	for p, c := range app.Paths {
		out[normalizePath(p)] = c
	}
	return out
}

// SetPath records a path and its checksum for the application.
func (s *SyncState) SetPath(applicationID, path, checksum string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app := s.state.Applications[applicationID]
	if app.Paths == nil {
		app.Paths = make(map[string]string)
	}
	app.Paths[normalizePath(path)] = checksum
	app.UpdatedAt = time.Now().UTC()
	s.state.Applications[applicationID] = app
	s.dirty = true
}

// PruneNonexistentPaths removes paths from state that are not in existingPaths.
// Call after load when current file list is known (e.g. after git branch switch).
func (s *SyncState) PruneNonexistentPaths(applicationID string, existingPaths map[string]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	app, ok := s.state.Applications[applicationID]
	if !ok || app.Paths == nil {
		return
	}
	for p := range app.Paths {
		norm := normalizePath(p)
		if !existingPaths[norm] && !existingPaths[p] {
			delete(app.Paths, p)
			app.UpdatedAt = time.Now().UTC()
			s.dirty = true
		}
	}
	s.state.Applications[applicationID] = app
}

// RemovePath removes a path from the application's state (e.g. on delete).
func (s *SyncState) RemovePath(applicationID, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, ok := s.state.Applications[applicationID]
	if !ok || app.Paths == nil {
		return
	}
	delete(app.Paths, normalizePath(path))
	app.UpdatedAt = time.Now().UTC()
	s.state.Applications[applicationID] = app
	s.dirty = true
}

// SetAllPaths replaces all paths for an application (e.g. after full sync).
func (s *SyncState) SetAllPaths(applicationID string, paths map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	norm := make(map[string]string, len(paths))
	for p, c := range paths {
		norm[normalizePath(p)] = c
	}
	s.state.Applications[applicationID] = appState{
		Paths:     norm,
		RootHash:  s.state.Applications[applicationID].RootHash,
		UpdatedAt: time.Now().UTC(),
	}
	s.dirty = true
}

// Save persists state to disk. Uses atomic write (temp file + rename).
func (s *SyncState) Save() error {
	s.mu.Lock()
	if !s.dirty {
		s.mu.Unlock()
		return nil
	}
	stateCopy := s.copyStateLocked()
	s.dirty = false
	s.mu.Unlock()

	s.saveMu.Lock()
	defer s.saveMu.Unlock()

	data, err := json.MarshalIndent(stateCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sync state: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create sync state directory: %w", err)
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write sync state: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to atomically save sync state: %w", err)
	}

	return nil
}

func (s *SyncState) copyStateLocked() *persistableState {
	apps := make(map[string]appState)
	for id, app := range s.state.Applications {
		paths := make(map[string]string)
		for p, c := range app.Paths {
			paths[p] = c
		}
		apps[id] = appState{Paths: paths, RootHash: app.RootHash, UpdatedAt: app.UpdatedAt}
	}
	return &persistableState{Applications: apps, Version: syncStateVersion}
}

func normalizePath(p string) string {
	return filepath.ToSlash(filepath.Clean(p))
}
