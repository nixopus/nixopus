package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/codebase"
)

const (
	maxFileSize  = 512 * 1024       // 512KB per file
	maxTotalSize = 50 * 1024 * 1024 // 50MB total
	maxFileCount = 5000
	cloneTimeout = 2 * time.Minute
)

var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "dist": true, "build": true,
	".next": true, ".nuxt": true, "coverage": true, "__pycache__": true,
	".venv": true, "vendor": true,
}

var binaryExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
	".webp": true, ".svg": true, ".bmp": true, ".tiff": true,
	".woff": true, ".woff2": true, ".ttf": true, ".eot": true, ".otf": true,
	".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true, ".7z": true, ".rar": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".exe": true, ".dll": true, ".so": true, ".dylib": true,
	".wasm": true, ".map": true,
	".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".wav": true, ".ogg": true,
	".pyc": true, ".pyo": true, ".class": true, ".o": true, ".a": true,
	".DS_Store": true, ".lock": true,
}

func (s *AgentService) analyzeRepositoryTool() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "analyze_repository",
		Description: "Analyze a GitHub repository to detect ecosystem, framework, port, Dockerfile presence, and deployment hints. Uses the Nixopus GitHub connector for private repo access.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner":        {"type": "string", "description": "GitHub repository owner"},
				"repo":         {"type": "string", "description": "GitHub repository name"},
				"branch":       {"type": "string", "description": "Branch to analyze (default: default branch)"},
				"connector_id": {"type": "string", "description": "Nixopus GitHub connector ID (uses first available if omitted)"}
			},
			"required": ["owner", "repo"]
		}`),
		Handler: s.analyzeRepositoryHandler,
	}
}

func (s *AgentService) analyzeRepositoryHandler(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Owner       string `json:"owner"`
		Repo        string `json:"repo"`
		Branch      string `json:"branch"`
		ConnectorID string `json:"connector_id"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if input.Owner == "" || input.Repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	authToken, _ := ctx.Value(ctxKeyAuthToken).(string)
	orgID, _ := ctx.Value(ctxKeyOrgID).(string)
	baseURL, _ := ctx.Value(ctxKeyBaseURL).(string)
	if baseURL == "" {
		baseURL = "http://localhost:" + getEnvOrDefault("PORT", "2089")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	connectorID := input.ConnectorID
	if connectorID == "" {
		var err error
		connectorID, err = fetchDefaultConnectorID(ctx, client, baseURL, authToken, orgID)
		if err != nil {
			return nil, fmt.Errorf("get connector: %w", err)
		}
	}

	treeURL := fmt.Sprintf("%s/api/v1/github-connector/repository/tree?owner=%s&repo=%s&connector_id=%s",
		baseURL, input.Owner, input.Repo, connectorID)
	if input.Branch != "" {
		treeURL += "&branch=" + input.Branch
	}

	req, err := http.NewRequestWithContext(ctx, "GET", treeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create tree request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	if orgID != "" {
		req.Header.Set("X-Organization-ID", orgID)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch tree: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if resp.StatusCode >= 400 {
		return json.Marshal(map[string]interface{}{
			"error":       fmt.Sprintf("GitHub tree API returned %d", resp.StatusCode),
			"status_code": resp.StatusCode,
			"body":        string(body),
		})
	}

	var treeResp struct {
		SHA  string `json:"sha"`
		Tree []struct {
			Path    string `json:"path"`
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"tree"`
	}
	if err := json.Unmarshal(body, &treeResp); err != nil {
		return nil, fmt.Errorf("parse tree response: %w", err)
	}

	var files []codebase.FileEntry
	for _, entry := range treeResp.Tree {
		if entry.Type != "blob" {
			continue
		}
		if shouldSkipPath(entry.Path) {
			continue
		}
		files = append(files, codebase.FileEntry{
			Path:    entry.Path,
			Content: entry.Content,
		})
	}

	if len(files) == 0 {
		return json.Marshal(map[string]interface{}{
			"error":      "no analyzable files found in repository",
			"file_count": 0,
		})
	}

	hints := codebase.AnalyzeFiles(files)

	return json.Marshal(map[string]interface{}{
		"file_count": len(files),
		"commit_sha": treeResp.SHA,
		"hints":      hints,
		"message":    confidenceMessage(hints.Confidence),
	})
}

func loadRemoteRepositoryTool() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "load_remote_repository",
		Description: "Clone a public git repository and analyze it to detect ecosystem, framework, port, Dockerfile, and deployment configuration.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"repo_url": {"type": "string", "description": "Public HTTPS git URL (e.g. https://github.com/owner/repo.git)"},
				"branch":   {"type": "string", "description": "Branch to clone (default: default branch)"}
			},
			"required": ["repo_url"]
		}`),
		Handler: loadRemoteRepositoryHandler,
	}
}

func loadRemoteRepositoryHandler(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input struct {
		RepoURL string `json:"repo_url"`
		Branch  string `json:"branch"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if input.RepoURL == "" {
		return nil, fmt.Errorf("repo_url is required")
	}
	if !strings.HasPrefix(input.RepoURL, "https://") && !strings.HasPrefix(input.RepoURL, "http://") && !strings.HasPrefix(input.RepoURL, "file://") {
		return nil, fmt.Errorf("only HTTPS git URLs are supported")
	}

	tmpDir, err := os.MkdirTemp("", "nixopus-repo-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneCtx, cancel := context.WithTimeout(ctx, cloneTimeout)
	defer cancel()

	cloneArgs := []string{"clone", "--depth", "1"}
	if input.Branch != "" {
		cloneArgs = append(cloneArgs, "--branch", input.Branch)
	}
	cloneArgs = append(cloneArgs, input.RepoURL, tmpDir)

	cmd := exec.CommandContext(cloneCtx, "git", cloneArgs...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git clone failed: %w\n%s", err, string(output))
	}

	commitSHA, branch := gitMetadata(tmpDir)

	files, err := collectFiles(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("collect files: %w", err)
	}

	if len(files) == 0 {
		return json.Marshal(map[string]interface{}{
			"error":      "no analyzable text files found in repository",
			"file_count": 0,
		})
	}

	hints := codebase.AnalyzeFiles(files)

	return json.Marshal(map[string]interface{}{
		"file_count": len(files),
		"commit":     commitSHA,
		"branch":     branch,
		"hints":      hints,
		"message":    confidenceMessage(hints.Confidence),
	})
}

func collectFiles(root string) ([]codebase.FileEntry, error) {
	var files []codebase.FileEntry
	var totalSize int64

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if len(files) >= maxFileCount {
			return filepath.SkipAll
		}
		if totalSize >= int64(maxTotalSize) {
			return filepath.SkipAll
		}
		if info.Size() > maxFileSize || info.Size() == 0 {
			return nil
		}
		if binaryExts[filepath.Ext(info.Name())] || binaryExts[info.Name()] {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if isBinaryContent(content) {
			return nil
		}

		totalSize += int64(len(content))
		files = append(files, codebase.FileEntry{
			Path:    rel,
			Content: string(content),
		})
		return nil
	})

	return files, err
}

func isBinaryContent(data []byte) bool {
	checkLen := 8192
	if len(data) < checkLen {
		checkLen = len(data)
	}
	for _, b := range data[:checkLen] {
		if b == 0 {
			return true
		}
	}
	return false
}

func gitMetadata(repoDir string) (sha, branch string) {
	if out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD").Output(); err == nil {
		sha = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		branch = strings.TrimSpace(string(out))
	}
	return
}

func shouldSkipPath(path string) bool {
	parts := strings.Split(path, "/")
	for _, p := range parts {
		if skipDirs[p] {
			return true
		}
	}
	return binaryExts[filepath.Ext(path)] || binaryExts[filepath.Base(path)]
}

func confidenceMessage(c codebase.Confidence) string {
	switch c {
	case codebase.ConfidenceHigh:
		return "Repository analyzed with high confidence. Deployment configuration can be auto-generated."
	case codebase.ConfidenceMedium:
		return "Repository analyzed with medium confidence. Some settings may need manual verification."
	default:
		return "Repository analyzed with low confidence. Manual configuration review is recommended."
	}
}

func fetchDefaultConnectorID(ctx context.Context, client *http.Client, baseURL, authToken, orgID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/github-connector", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	if orgID != "" {
		req.Header.Set("X-Organization-ID", orgID)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("connectors API returned %d: %s", resp.StatusCode, string(body))
	}

	var connectors []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &connectors); err != nil {
		var wrapper struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err2 := json.Unmarshal(body, &wrapper); err2 != nil {
			return "", fmt.Errorf("parse connectors: %w", err)
		}
		connectors = wrapper.Data
	}

	if len(connectors) == 0 {
		return "", fmt.Errorf("no GitHub connectors configured")
	}
	return connectors[0].ID, nil
}
