// Watcher provides a file system watcher that monitors directory changes using OS-level notifications.
// It supports debouncing, pattern-based ignoring (including .gitignore), and recursive directory watching.
package mover

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// EventType represents the type of file system event
type EventType int

const (
	EventCreate EventType = iota
	EventModify
	EventDelete
	EventRename
)

var eventTypeNames = map[EventType]string{
	EventCreate: "create",
	EventModify: "modify",
	EventDelete: "delete",
	EventRename: "rename",
}

func (e EventType) String() string {
	if name, ok := eventTypeNames[e]; ok {
		return name
	}
	return "unknown"
}

// Event represents a file system change event
type Event struct {
	Path string
	Type EventType
}

// Config holds watcher configuration
type Config struct {
	RootPath       string
	DebounceMs     int
	IgnorePatterns []string
}

// Watcher watches a directory for file changes using OS-level notifications.
type Watcher struct {
	rootPath      string
	watcher       *fsnotify.Watcher
	events        chan Event
	errors        chan error
	done          chan struct{}
	debounceDelay time.Duration
	ignorer       *pathIgnorer
}

// New creates a new file system watcher
func New(cfg Config) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	w := &Watcher{
		rootPath:      cfg.RootPath,
		watcher:       fsWatcher,
		events:        make(chan Event, 100),
		errors:        make(chan error, 10),
		done:          make(chan struct{}),
		debounceDelay: parseDebounceDuration(cfg.DebounceMs),
		ignorer:       newPathIgnorer(cfg.RootPath, cfg.IgnorePatterns),
	}

	return w, nil
}

func parseDebounceDuration(ms int) time.Duration {
	if ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return 100 * time.Millisecond
}

// Start begins watching for file changes
func (w *Watcher) Start() error {
	if err := w.addWatchRecursive(w.rootPath); err != nil {
		return fmt.Errorf("failed to add watch paths: %w", err)
	}
	go w.runEventLoop()
	return nil
}

// Events returns the channel of file change events
func (w *Watcher) Events() <-chan Event {
	return w.events
}

// Errors returns the channel of watcher errors
func (w *Watcher) Errors() <-chan error {
	return w.errors
}

// Stop stops the watcher
func (w *Watcher) Stop() error {
	close(w.done)
	return w.watcher.Close()
}

func (w *Watcher) addWatchRecursive(root string) error {
	return filepath.Walk(root, w.visitDir)
}

func (w *Watcher) visitDir(path string, info os.FileInfo, err error) error {
	if err != nil || !info.IsDir() {
		return nil
	}

	relPath, _ := filepath.Rel(w.rootPath, path)
	if w.ignorer.shouldIgnore(relPath) {
		return filepath.SkipDir
	}

	_ = w.watcher.Add(path)
	return nil
}

func (w *Watcher) runEventLoop() {
	debouncer := newDebouncer(w.debounceDelay, w.flushEvents)

	for {
		select {
		case <-w.done:
			return
		case fsEvent, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleFsEvent(fsEvent, debouncer)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.forwardError(err)
		}
	}
}

func (w *Watcher) handleFsEvent(fsEvent fsnotify.Event, debouncer *debouncer) {
	relPath, err := filepath.Rel(w.rootPath, fsEvent.Name)
	if err != nil {
		return
	}

	if w.ignorer.shouldIgnore(relPath) {
		return
	}

	eventType, ok := w.convertEvent(fsEvent)
	if !ok {
		return
	}

	isDir := isDirectory(fsEvent.Name)
	if eventType == EventCreate && isDir {
		w.addWatchRecursive(fsEvent.Name)
	}

	if isDir && eventType != EventDelete {
		return
	}

	debouncer.add(relPath, Event{Path: relPath, Type: eventType})
}

func (w *Watcher) convertEvent(fsEvent fsnotify.Event) (EventType, bool) {
	switch {
	case fsEvent.Op&fsnotify.Create != 0:
		return EventCreate, true
	case fsEvent.Op&fsnotify.Write != 0:
		return EventModify, true
	case fsEvent.Op&fsnotify.Remove != 0:
		return EventDelete, true
	case fsEvent.Op&fsnotify.Rename != 0:
		return EventDelete, true
	default:
		return 0, false
	}
}

func (w *Watcher) flushEvents(events map[string]Event) {
	for _, event := range events {
		select {
		case w.events <- event:
		case <-w.done:
			return
		}
	}
}

func (w *Watcher) forwardError(err error) {
	select {
	case w.errors <- err:
	default:
	}
}

type debouncer struct {
	delay   time.Duration
	pending map[string]Event
	mu      sync.Mutex
	timer   *time.Timer
	onFlush func(map[string]Event)
}

func newDebouncer(delay time.Duration, onFlush func(map[string]Event)) *debouncer {
	return &debouncer{
		delay:   delay,
		pending: make(map[string]Event),
		onFlush: onFlush,
	}
}

func (d *debouncer) add(path string, event Event) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if existing, exists := d.pending[path]; exists && existing.Type == EventDelete {
		return
	}

	d.pending[path] = event
	d.resetTimer()
}

func (d *debouncer) resetTimer() {
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.delay, d.flush)
}

func (d *debouncer) flush() {
	d.mu.Lock()
	toSend := d.pending
	d.pending = make(map[string]Event)
	d.mu.Unlock()

	d.onFlush(toSend)
}


type pathIgnorer struct {
	rootPath   string
	patterns   []string
	gitCache   map[string]bool
	gitCacheMu sync.RWMutex
}

func newPathIgnorer(rootPath string, patterns []string) *pathIgnorer {
	p := &pathIgnorer{
		rootPath: rootPath,
		patterns: patterns,
		gitCache: make(map[string]bool),
	}
	p.loadGitIgnorePatterns()
	return p
}

func (p *pathIgnorer) shouldIgnore(relPath string) bool {
	if p.isBuiltinIgnored(relPath) {
		return true
	}

	if p.matchesPattern(relPath) {
		return true
	}

	return p.isGitIgnored(relPath)
}

func (p *pathIgnorer) isBuiltinIgnored(relPath string) bool {
	return pathContains(relPath, ".git") || pathContains(relPath, "node_modules")
}

func (p *pathIgnorer) matchesPattern(relPath string) bool {
	baseName := filepath.Base(relPath)
	for _, pattern := range p.patterns {
		if matched, _ := filepath.Match(pattern, baseName); matched {
			return true
		}
	}
	return false
}

func (p *pathIgnorer) isGitIgnored(relPath string) bool {
	p.gitCacheMu.RLock()
	ignored, cached := p.gitCache[relPath]
	p.gitCacheMu.RUnlock()

	if cached {
		return ignored
	}

	ignored = p.checkGitIgnore(relPath)
	p.gitCacheMu.Lock()
	p.gitCache[relPath] = ignored
	p.gitCacheMu.Unlock()

	return ignored
}

func (p *pathIgnorer) checkGitIgnore(relPath string) bool {
	cmd := exec.Command("git", "check-ignore", "-q", relPath)
	cmd.Dir = p.rootPath
	return cmd.Run() == nil
}

func (p *pathIgnorer) loadGitIgnorePatterns() {
	gitignorePath := filepath.Join(p.rootPath, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if p.isValidGitIgnoreLine(line) {
			p.patterns = append(p.patterns, line)
		}
	}
}

func (p *pathIgnorer) isValidGitIgnoreLine(line string) bool {
	if line == "" {
		return false
	}
	if strings.HasPrefix(line, "#") {
		return false
	}
	if strings.HasPrefix(line, "!") {
		return false
	}
	return true
}

func pathContains(path, segment string) bool {
	if path == segment {
		return true
	}
	return strings.Contains(path, segment+"/") || strings.Contains(path, segment+"\\")
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
