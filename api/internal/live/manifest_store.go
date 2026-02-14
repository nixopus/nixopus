package live

import (
	"path/filepath"
	"sync"
)

// ManifestStore tracks path→checksum for each application.
// Updated on file write and delete. Sent to client on WebSocket connect.
type ManifestStore struct {
	mu   sync.RWMutex
	data map[string]map[string]string // appID -> path -> checksum
}

// NewManifestStore creates a new manifest store.
func NewManifestStore() *ManifestStore {
	return &ManifestStore{
		data: make(map[string]map[string]string),
	}
}

// Set records a path and its checksum for an application.
func (s *ManifestStore) Set(appID string, path string, checksum string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data[appID] == nil {
		s.data[appID] = make(map[string]string)
	}
	s.data[appID][normalizePath(path)] = checksum
}

// Remove removes a path from an application's manifest.
func (s *ManifestStore) Remove(appID string, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data[appID] != nil {
		delete(s.data[appID], normalizePath(path))
	}
}

// GetPaths returns a copy of path→checksum for the given application.
func (s *ManifestStore) GetPaths(appID string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.data[appID] == nil {
		return nil
	}
	out := make(map[string]string, len(s.data[appID]))
	for p, c := range s.data[appID] {
		out[p] = c
	}
	return out
}

func normalizePath(p string) string {
	return filepath.ToSlash(filepath.Clean(p))
}
