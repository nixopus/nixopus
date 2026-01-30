package live

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// LogEntry represents a single log entry with timestamp and source
type LogEntry struct {
	Message   string
	Timestamp time.Time
	Source    string // "build" or "container"
}

// LogFetcher handles fetching and combining logs from different sources
type LogFetcher struct {
	client *http.Client
	config *Config
}

// Config holds configuration for log fetching
type Config struct {
	Server      string
	AccessToken string // Bearer token for API authentication
	Timeout     time.Duration
}

// NewLogFetcher creates a new log fetcher
func NewLogFetcher(config *Config) *LogFetcher {
	return &LogFetcher{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// FetchContainerLogs fetches container logs from the API
func (lf *LogFetcher) FetchContainerLogs(containerID string, tail int) ([]LogEntry, error) {
	if containerID == "" {
		return []LogEntry{}, nil
	}

	url := fmt.Sprintf("%s/api/v1/container/%s/logs", lf.config.Server, containerID)

	reqBody := map[string]interface{}{
		"id":     containerID,
		"tail":   tail,
		"stdout": true,
		"stderr": true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// TODO: Add session token to Authorization header
	if lf.config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+lf.config.AccessToken)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := lf.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch container logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var containerLogsResp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&containerLogsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse container logs (they come as a single string with newlines)
	logs := parseContainerLogs(containerLogsResp.Data)
	return logs, nil
}

// parseContainerLogs parses container logs string into LogEntry slice
func parseContainerLogs(logsString string) []LogEntry {
	if logsString == "" {
		return []LogEntry{}
	}

	lines := strings.Split(logsString, "\n")
	entries := make([]LogEntry, 0, len(lines))
	now := time.Now()

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to extract timestamp from log line
		// Container logs might have timestamps in various formats
		timestamp := extractTimestamp(line, now)

		entries = append(entries, LogEntry{
			Message:   line,
			Timestamp: timestamp,
			Source:    "container",
		})

		// If we couldn't parse timestamp, use a slightly earlier time for each line
		// to maintain relative ordering
		if timestamp.Equal(now) {
			timestamp = now.Add(-time.Duration(len(lines)-i) * time.Second)
			entries[len(entries)-1].Timestamp = timestamp
		}
	}

	return entries
}

// extractTimestamp tries to extract timestamp from log line
// Returns the provided defaultTime if no timestamp is found
func extractTimestamp(line string, defaultTime time.Time) time.Time {
	// Try common timestamp formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000000000Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
	}

	// Look for timestamp at the beginning of the line (common in Docker logs)
	parts := strings.Fields(line)
	if len(parts) > 0 {
		for _, format := range formats {
			if t, err := time.Parse(format, parts[0]); err == nil {
				return t
			}
		}
	}

	return defaultTime
}

// CombineLogs combines build logs and container logs, sorted by timestamp (latest first)
func CombineLogs(buildLogs []BuildLog, containerLogs []LogEntry) []string {
	allEntries := make([]LogEntry, 0)

	// Convert build logs to LogEntry
	for _, log := range buildLogs {
		timestamp := time.Now()
		if log.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, log.CreatedAt); err == nil {
				timestamp = t
			}
		}

		allEntries = append(allEntries, LogEntry{
			Message:   log.Log,
			Timestamp: timestamp,
			Source:    "build",
		})
	}

	// Add container logs
	allEntries = append(allEntries, containerLogs...)

	// Sort by timestamp descending (latest first)
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Timestamp.After(allEntries[j].Timestamp)
	})

	// Convert to string slice with source prefix
	result := make([]string, len(allEntries))
	for i, entry := range allEntries {
		prefix := "[BUILD]"
		if entry.Source == "container" {
			prefix = "[CONTAINER]"
		}
		result[i] = fmt.Sprintf("%s %s", prefix, entry.Message)
	}

	return result
}

// BuildLog represents a build log entry from the deployment API
type BuildLog struct {
	Log       string `json:"log"`
	CreatedAt string `json:"created_at"`
}

// FetchApplicationLogs fetches logs directly from the application logs API endpoint
func (lf *LogFetcher) FetchApplicationLogs(applicationID string, maxLogs int) ([]BuildLog, error) {
	if applicationID == "" {
		return []BuildLog{}, nil
	}

	url := fmt.Sprintf("%s/api/v1/deploy/application/logs/%s?page=1&page_size=%d", lf.config.Server, applicationID, maxLogs)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// TODO: Add session token to Authorization header
	if lf.config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+lf.config.AccessToken)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := lf.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch application logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var logsResp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Logs []struct {
				ID        string `json:"id"`
				Log       string `json:"log"`
				CreatedAt string `json:"created_at"`
			} `json:"logs"`
			TotalCount int64 `json:"total_count"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&logsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	buildLogs := make([]BuildLog, len(logsResp.Data.Logs))
	for i, log := range logsResp.Data.Logs {
		buildLogs[i] = BuildLog{
			Log:       log.Log,
			CreatedAt: log.CreatedAt,
		}
	}

	return buildLogs, nil
}

// FetchCombinedLogs fetches both build logs and container logs in parallel and combines them
func (lf *LogFetcher) FetchCombinedLogs(buildLogs []BuildLog, containerID string, applicationID string, maxLogs int) []string {
	// If build logs are empty and we have an application ID, try fetching from application logs API
	if len(buildLogs) == 0 && applicationID != "" {
		appLogs, err := lf.FetchApplicationLogs(applicationID, maxLogs)
		if err == nil {
			buildLogs = appLogs
		}
	}

	// Fetch container logs in parallel (non-blocking)
	containerLogsChan := make(chan []LogEntry, 1)

	go func() {
		if containerID == "" {
			containerLogsChan <- []LogEntry{}
			return
		}

		logs, err := lf.FetchContainerLogs(containerID, maxLogs)
		if err != nil {
			containerLogsChan <- []LogEntry{}
			return
		}
		containerLogsChan <- logs
	}()

	// Wait for container logs (with timeout)
	var containerLogs []LogEntry
	select {
	case logs := <-containerLogsChan:
		containerLogs = logs
	case <-time.After(3 * time.Second):
		containerLogs = []LogEntry{}
	}

	// Combine and sort logs
	combined := CombineLogs(buildLogs, containerLogs)

	// Limit to maxLogs (already sorted latest first)
	if len(combined) > maxLogs {
		combined = combined[:maxLogs]
	}

	return combined
}
