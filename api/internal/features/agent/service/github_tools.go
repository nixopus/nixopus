package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	gh "github.com/nixopus/nixopus/api/internal/features/github-connector/service/github"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/uptrace/bun"
)

const (
	maxResponseBody   = 50 * 1024 // 50KB truncation limit for file content
	connectorCacheTTL = 5 * time.Minute
	githubHTTPTimeout = 30 * time.Second
)

var sensitiveFilePatterns = []string{
	".env", ".pem", ".key", ".p12", ".pfx", ".jks",
	"credentials.json", "service-account.json",
	"id_rsa", "id_ed25519", ".secret",
}

type cachedConnector struct {
	connector *shared_types.GithubConnector
	token     string
	expiresAt time.Time
}

type githubClient struct {
	db    *bun.DB
	mu    sync.RWMutex
	cache map[string]*cachedConnector
}

func newGithubClient(db *bun.DB) *githubClient {
	return &githubClient{
		db:    db,
		cache: make(map[string]*cachedConnector),
	}
}

func (gc *githubClient) getInstallationToken(ctx context.Context) (string, error) {
	orgID, _ := ctx.Value(ctxKeyOrgID).(string)
	cacheKey := orgID

	gc.mu.RLock()
	if cached, ok := gc.cache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
		token := cached.token
		gc.mu.RUnlock()
		return token, nil
	}
	gc.mu.RUnlock()

	connector, err := gc.resolveConnector(ctx)
	if err != nil {
		return "", err
	}

	jwtStr := gh.GenerateJwt(connector)
	if jwtStr == "" {
		return "", fmt.Errorf("failed to generate GitHub App JWT: invalid credentials")
	}

	token, err := gh.InstallationToken(jwtStr, connector.InstallationID)
	if err != nil {
		return "", fmt.Errorf("get installation token: %w", err)
	}

	gc.mu.Lock()
	gc.cache[cacheKey] = &cachedConnector{
		connector: connector,
		token:     token,
		expiresAt: time.Now().Add(connectorCacheTTL),
	}
	gc.mu.Unlock()

	return token, nil
}

func (gc *githubClient) resolveConnector(ctx context.Context) (*shared_types.GithubConnector, error) {
	if gc.db == nil {
		return nil, fmt.Errorf("no GitHub connector with valid credentials found")
	}
	var connectors []shared_types.GithubConnector
	err := gc.db.NewSelect().
		Model(&connectors).
		Where("deleted_at IS NULL").
		Where("installation_id != ''").
		Where("pem != ''").
		Where("app_id != ''").
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("query connectors: %w", err)
	}
	if len(connectors) == 0 {
		return nil, fmt.Errorf("no GitHub connector with valid credentials found")
	}
	return &connectors[0], nil
}

func (gc *githubClient) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, int, error) {
	token, err := gc.getInstallationToken(ctx)
	if err != nil {
		return nil, 0, err
	}

	url := gh.APIBaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixopus-agent")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: githubHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("github api request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	return respBody, resp.StatusCode, nil
}

func (gc *githubClient) doJSON(ctx context.Context, method, path string, payload interface{}) (json.RawMessage, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		body = strings.NewReader(string(data))
	}

	respBody, status, err := gc.doRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	if status >= 400 {
		return nil, fmt.Errorf("GitHub API %s %s returned %d: %s", method, path, status, truncate(string(respBody), 500))
	}

	return respBody, nil
}

// Tool definitions

func (s *AgentService) githubListPullRequestsTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_list_pull_requests",
		Description: "List pull requests for a GitHub repository.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"state": {"type": "string", "enum": ["open", "closed", "all"], "description": "PR state filter (default: open)"},
				"per_page": {"type": "integer", "description": "Results per page (max 100, default 30)"}
			},
			"required": ["owner", "repo"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner   string `json:"owner"`
				Repo    string `json:"repo"`
				State   string `json:"state"`
				PerPage int    `json:"per_page"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}
			if input.State == "" {
				input.State = "open"
			}
			if input.PerPage <= 0 || input.PerPage > 100 {
				input.PerPage = 30
			}

			path := fmt.Sprintf("/repos/%s/%s/pulls?state=%s&per_page=%d", input.Owner, input.Repo, input.State, input.PerPage)
			return gc.doJSON(ctx, "GET", path, nil)
		},
	}
}

func (s *AgentService) githubListIssuesTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_list_issues",
		Description: "List issues for a repository (excludes pull requests).",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"state": {"type": "string", "enum": ["open", "closed", "all"], "description": "Issue state filter (default: open)"},
				"per_page": {"type": "integer", "description": "Results per page (max 100, default 30)"}
			},
			"required": ["owner", "repo"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner   string `json:"owner"`
				Repo    string `json:"repo"`
				State   string `json:"state"`
				PerPage int    `json:"per_page"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}
			if input.State == "" {
				input.State = "open"
			}
			if input.PerPage <= 0 || input.PerPage > 100 {
				input.PerPage = 30
			}

			path := fmt.Sprintf("/repos/%s/%s/issues?state=%s&per_page=%d", input.Owner, input.Repo, input.State, input.PerPage)
			data, err := gc.doJSON(ctx, "GET", path, nil)
			if err != nil {
				return nil, err
			}

			var items []map[string]interface{}
			if err := json.Unmarshal(data, &items); err != nil {
				return data, nil
			}

			var issues []map[string]interface{}
			for _, item := range items {
				if _, hasPR := item["pull_request"]; !hasPR {
					issues = append(issues, item)
				}
			}
			return json.Marshal(issues)
		},
	}
}

func (s *AgentService) githubCommentOnPRTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_comment_on_pr",
		Description: "Add a comment to a pull request.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"pr_number": {"type": "integer", "description": "Pull request number"},
				"body": {"type": "string", "description": "Comment text"}
			},
			"required": ["owner", "repo", "pr_number", "body"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner    string `json:"owner"`
				Repo     string `json:"repo"`
				PRNumber int    `json:"pr_number"`
				Body     string `json:"body"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", input.Owner, input.Repo, input.PRNumber)
			return gc.doJSON(ctx, "POST", path, map[string]string{"body": input.Body})
		},
	}
}

func (s *AgentService) githubCommentOnIssueTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_comment_on_issue",
		Description: "Add a comment to an issue.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"issue_number": {"type": "integer", "description": "Issue number"},
				"body": {"type": "string", "description": "Comment text"}
			},
			"required": ["owner", "repo", "issue_number", "body"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner       string `json:"owner"`
				Repo        string `json:"repo"`
				IssueNumber int    `json:"issue_number"`
				Body        string `json:"body"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", input.Owner, input.Repo, input.IssueNumber)
			return gc.doJSON(ctx, "POST", path, map[string]string{"body": input.Body})
		},
	}
}

func (s *AgentService) githubCreateIssueTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_create_issue",
		Description: "Create a new issue in a repository.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"title": {"type": "string", "description": "Issue title"},
				"body": {"type": "string", "description": "Issue body"},
				"labels": {"type": "array", "items": {"type": "string"}, "description": "Labels to add"},
				"assignees": {"type": "array", "items": {"type": "string"}, "description": "Usernames to assign"}
			},
			"required": ["owner", "repo", "title"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner     string   `json:"owner"`
				Repo      string   `json:"repo"`
				Title     string   `json:"title"`
				Body      string   `json:"body"`
				Labels    []string `json:"labels"`
				Assignees []string `json:"assignees"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			payload := map[string]interface{}{
				"title": input.Title,
			}
			if input.Body != "" {
				payload["body"] = input.Body
			}
			if len(input.Labels) > 0 {
				payload["labels"] = input.Labels
			}
			if len(input.Assignees) > 0 {
				payload["assignees"] = input.Assignees
			}

			path := fmt.Sprintf("/repos/%s/%s/issues", input.Owner, input.Repo)
			return gc.doJSON(ctx, "POST", path, payload)
		},
	}
}

func (s *AgentService) githubSetCommitStatusTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_set_commit_status",
		Description: "Set a commit status (pending, success, failure, error).",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"sha": {"type": "string", "description": "Commit SHA"},
				"state": {"type": "string", "enum": ["pending", "success", "failure", "error"], "description": "Status state"},
				"description": {"type": "string", "description": "Short description"},
				"context": {"type": "string", "description": "Status context (e.g. ci/deploy)"},
				"target_url": {"type": "string", "description": "URL for details link"}
			},
			"required": ["owner", "repo", "sha", "state"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner       string `json:"owner"`
				Repo        string `json:"repo"`
				SHA         string `json:"sha"`
				State       string `json:"state"`
				Description string `json:"description"`
				Context     string `json:"context"`
				TargetURL   string `json:"target_url"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			payload := map[string]string{
				"state": input.State,
			}
			if input.Description != "" {
				payload["description"] = input.Description
			}
			if input.Context != "" {
				payload["context"] = input.Context
			}
			if input.TargetURL != "" {
				payload["target_url"] = input.TargetURL
			}

			path := fmt.Sprintf("/repos/%s/%s/statuses/%s", input.Owner, input.Repo, input.SHA)
			return gc.doJSON(ctx, "POST", path, payload)
		},
	}
}

func (s *AgentService) githubCreateDeploymentStatusTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_create_deployment_status",
		Description: "Create a GitHub deployment and set its status.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"ref": {"type": "string", "description": "Git ref (branch, tag, or SHA) to deploy"},
				"environment": {"type": "string", "description": "Deployment environment (e.g. production, staging)"},
				"state": {"type": "string", "enum": ["pending", "success", "failure", "error", "inactive", "in_progress", "queued"], "description": "Deployment status state"},
				"description": {"type": "string", "description": "Status description"},
				"environment_url": {"type": "string", "description": "URL of the deployed environment"}
			},
			"required": ["owner", "repo", "ref", "environment", "state"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner          string `json:"owner"`
				Repo           string `json:"repo"`
				Ref            string `json:"ref"`
				Environment    string `json:"environment"`
				State          string `json:"state"`
				Description    string `json:"description"`
				EnvironmentURL string `json:"environment_url"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			deployPayload := map[string]interface{}{
				"ref":               input.Ref,
				"environment":       input.Environment,
				"auto_merge":        false,
				"required_contexts": []string{},
			}
			deployPath := fmt.Sprintf("/repos/%s/%s/deployments", input.Owner, input.Repo)
			deployResp, err := gc.doJSON(ctx, "POST", deployPath, deployPayload)
			if err != nil {
				return nil, fmt.Errorf("create deployment: %w", err)
			}

			var deployment struct {
				ID int64 `json:"id"`
			}
			if err := json.Unmarshal(deployResp, &deployment); err != nil {
				return nil, fmt.Errorf("parse deployment response: %w", err)
			}

			statusPayload := map[string]string{
				"state": input.State,
			}
			if input.Description != "" {
				statusPayload["description"] = input.Description
			}
			if input.EnvironmentURL != "" {
				statusPayload["environment_url"] = input.EnvironmentURL
			}

			statusPath := fmt.Sprintf("/repos/%s/%s/deployments/%d/statuses", input.Owner, input.Repo, deployment.ID)
			statusResp, err := gc.doJSON(ctx, "POST", statusPath, statusPayload)
			if err != nil {
				return nil, fmt.Errorf("create deployment status: %w", err)
			}

			return json.Marshal(map[string]interface{}{
				"deployment_id": deployment.ID,
				"status":        json.RawMessage(statusResp),
			})
		},
	}
}

func (s *AgentService) githubSearchRepoContentTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_search_repo_content",
		Description: "Search for code in a repository.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"query": {"type": "string", "description": "Search query"},
				"per_page": {"type": "integer", "description": "Results per page (max 100, default 30)"}
			},
			"required": ["owner", "repo", "query"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner   string `json:"owner"`
				Repo    string `json:"repo"`
				Query   string `json:"query"`
				PerPage int    `json:"per_page"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}
			if input.PerPage <= 0 || input.PerPage > 100 {
				input.PerPage = 30
			}

			q := fmt.Sprintf("%s+repo:%s/%s", input.Query, input.Owner, input.Repo)
			path := fmt.Sprintf("/search/code?q=%s&per_page=%d", q, input.PerPage)
			return gc.doJSON(ctx, "GET", path, nil)
		},
	}
}

func (s *AgentService) githubCreateOrUpdateFileTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_create_or_update_file",
		Description: "Create or update a file on a feature branch. Refuses writes to main/master and blocks sensitive file patterns.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"path": {"type": "string", "description": "File path in the repository"},
				"content": {"type": "string", "description": "File content (will be base64-encoded)"},
				"message": {"type": "string", "description": "Commit message"},
				"branch": {"type": "string", "description": "Target branch (must NOT be main or master)"},
				"sha": {"type": "string", "description": "SHA of the file being replaced (required for updates)"}
			},
			"required": ["owner", "repo", "path", "content", "message", "branch"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner   string `json:"owner"`
				Repo    string `json:"repo"`
				Path    string `json:"path"`
				Content string `json:"content"`
				Message string `json:"message"`
				Branch  string `json:"branch"`
				SHA     string `json:"sha"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			if isProtectedBranch(input.Branch) {
				return nil, fmt.Errorf("refused: cannot write to protected branch %q — use a feature branch", input.Branch)
			}
			if isSensitiveFile(input.Path) {
				return nil, fmt.Errorf("refused: cannot write sensitive file %q", input.Path)
			}

			payload := map[string]string{
				"message": input.Message,
				"content": base64.StdEncoding.EncodeToString([]byte(input.Content)),
				"branch":  input.Branch,
			}
			if input.SHA != "" {
				payload["sha"] = input.SHA
			}

			apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s", input.Owner, input.Repo, input.Path)
			return gc.doJSON(ctx, "PUT", apiPath, payload)
		},
	}
}

func (s *AgentService) githubGetBranchTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_get_branch",
		Description: "Get branch details including the HEAD commit SHA.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"branch": {"type": "string", "description": "Branch name"}
			},
			"required": ["owner", "repo", "branch"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner  string `json:"owner"`
				Repo   string `json:"repo"`
				Branch string `json:"branch"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			path := fmt.Sprintf("/repos/%s/%s/branches/%s", input.Owner, input.Repo, input.Branch)
			return gc.doJSON(ctx, "GET", path, nil)
		},
	}
}

func (s *AgentService) githubCreateBranchTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_create_branch",
		Description: "Create a new branch from a given SHA.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"branch": {"type": "string", "description": "New branch name"},
				"sha": {"type": "string", "description": "SHA to branch from"}
			},
			"required": ["owner", "repo", "branch", "sha"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner  string `json:"owner"`
				Repo   string `json:"repo"`
				Branch string `json:"branch"`
				SHA    string `json:"sha"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			payload := map[string]string{
				"ref": "refs/heads/" + input.Branch,
				"sha": input.SHA,
			}

			path := fmt.Sprintf("/repos/%s/%s/git/refs", input.Owner, input.Repo)
			return gc.doJSON(ctx, "POST", path, payload)
		},
	}
}

func (s *AgentService) githubCreatePullRequestTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_create_pull_request",
		Description: "Create a pull request.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"title": {"type": "string", "description": "PR title"},
				"body": {"type": "string", "description": "PR description"},
				"head": {"type": "string", "description": "Source branch"},
				"base": {"type": "string", "description": "Target branch"},
				"draft": {"type": "boolean", "description": "Create as draft PR"}
			},
			"required": ["owner", "repo", "title", "head", "base"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner string `json:"owner"`
				Repo  string `json:"repo"`
				Title string `json:"title"`
				Body  string `json:"body"`
				Head  string `json:"head"`
				Base  string `json:"base"`
				Draft bool   `json:"draft"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			payload := map[string]interface{}{
				"title": input.Title,
				"head":  input.Head,
				"base":  input.Base,
				"draft": input.Draft,
			}
			if input.Body != "" {
				payload["body"] = input.Body
			}

			path := fmt.Sprintf("/repos/%s/%s/pulls", input.Owner, input.Repo)
			return gc.doJSON(ctx, "POST", path, payload)
		},
	}
}

func (s *AgentService) githubMergePullRequestTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_merge_pull_request",
		Description: "Merge a pull request.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"pr_number": {"type": "integer", "description": "Pull request number"},
				"merge_method": {"type": "string", "enum": ["merge", "squash", "rebase"], "description": "Merge method (default: merge)"},
				"commit_title": {"type": "string", "description": "Custom merge commit title"},
				"commit_message": {"type": "string", "description": "Custom merge commit message"}
			},
			"required": ["owner", "repo", "pr_number"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner         string `json:"owner"`
				Repo          string `json:"repo"`
				PRNumber      int    `json:"pr_number"`
				MergeMethod   string `json:"merge_method"`
				CommitTitle   string `json:"commit_title"`
				CommitMessage string `json:"commit_message"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}
			if input.MergeMethod == "" {
				input.MergeMethod = "merge"
			}

			payload := map[string]string{
				"merge_method": input.MergeMethod,
			}
			if input.CommitTitle != "" {
				payload["commit_title"] = input.CommitTitle
			}
			if input.CommitMessage != "" {
				payload["commit_message"] = input.CommitMessage
			}

			path := fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", input.Owner, input.Repo, input.PRNumber)
			return gc.doJSON(ctx, "PUT", path, payload)
		},
	}
}

func (s *AgentService) githubGetRepoFileTool(gc *githubClient) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "github_get_repo_file",
		Description: "Read a file from a repository. Large content is truncated to 50KB.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"owner": {"type": "string", "description": "Repository owner"},
				"repo": {"type": "string", "description": "Repository name"},
				"path": {"type": "string", "description": "File path"},
				"ref": {"type": "string", "description": "Branch, tag, or SHA (default: default branch)"}
			},
			"required": ["owner", "repo", "path"]
		}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Owner string `json:"owner"`
				Repo  string `json:"repo"`
				Path  string `json:"path"`
				Ref   string `json:"ref"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s", input.Owner, input.Repo, input.Path)
			if input.Ref != "" {
				apiPath += "?ref=" + input.Ref
			}

			data, err := gc.doJSON(ctx, "GET", apiPath, nil)
			if err != nil {
				return nil, err
			}

			var fileResp struct {
				Content  string `json:"content"`
				Encoding string `json:"encoding"`
				Name     string `json:"name"`
				Path     string `json:"path"`
				SHA      string `json:"sha"`
				Size     int    `json:"size"`
			}
			if err := json.Unmarshal(data, &fileResp); err != nil {
				return data, nil
			}

			if fileResp.Encoding == "base64" && fileResp.Content != "" {
				decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(fileResp.Content, "\n", ""))
				if err == nil {
					content := string(decoded)
					truncated := false
					if len(content) > maxResponseBody {
						content = content[:maxResponseBody]
						truncated = true
					}
					return json.Marshal(map[string]interface{}{
						"name":      fileResp.Name,
						"path":      fileResp.Path,
						"sha":       fileResp.SHA,
						"size":      fileResp.Size,
						"content":   content,
						"truncated": truncated,
					})
				}
			}

			return data, nil
		},
	}
}

// Helpers

func isProtectedBranch(branch string) bool {
	b := strings.ToLower(strings.TrimSpace(branch))
	return b == "main" || b == "master"
}

func isSensitiveFile(path string) bool {
	lower := strings.ToLower(path)
	for _, pattern := range sensitiveFilePatterns {
		if strings.HasSuffix(lower, pattern) || strings.Contains(lower, pattern+"/") {
			return true
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
