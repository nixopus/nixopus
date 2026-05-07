package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	gh "github.com/nixopus/nixopus/api/internal/features/github-connector/service/github"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

const (
	MaxResponseBody   = 50 * 1024 // 50KB truncation limit for file content
	connectorCacheTTL = 5 * time.Minute
	httpTimeout       = 30 * time.Second
)

var SensitiveFilePatterns = []string{
	".env", ".pem", ".key", ".p12", ".pfx", ".jks",
	"credentials.json", "service-account.json",
	"id_rsa", "id_ed25519", ".secret",
}

type cachedConnector struct {
	connector *shared_types.GithubConnector
	token     string
	expiresAt time.Time
}

// Client wraps GitHub API access via installation tokens.
type Client struct {
	db           *bun.DB
	mu           sync.RWMutex
	cache        map[string]*cachedConnector
	OrgIDFromCtx func(ctx context.Context) string
}

func NewClient(db *bun.DB, orgIDFromCtx func(ctx context.Context) string) *Client {
	return &Client{
		db:           db,
		cache:        make(map[string]*cachedConnector),
		OrgIDFromCtx: orgIDFromCtx,
	}
}

// NewProbeClient returns a client with a cached installation token for tests and probes that mock the GitHub API.
// orgFromCtx determines the cache key; if nil, the empty string is used as the org key.
func NewProbeClient(token string, orgFromCtx func(ctx context.Context) string) *Client {
	if orgFromCtx == nil {
		orgFromCtx = func(context.Context) string { return "" }
	}
	key := orgFromCtx(context.Background())
	return &Client{
		cache: map[string]*cachedConnector{
			key: {token: token, expiresAt: time.Now().Add(24 * time.Hour)},
		},
		OrgIDFromCtx: orgFromCtx,
	}
}

// RedirectAPIToTestServer points the connector package's GitHub API base URL at baseURL until restore runs.
func RedirectAPIToTestServer(baseURL string) (restore func()) {
	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(baseURL)
	return func() { gh.SetAPIBaseURL(prev) }
}

func (c *Client) GetInstallationToken(ctx context.Context) (string, error) {
	orgID := ""
	if c.OrgIDFromCtx != nil {
		orgID = c.OrgIDFromCtx(ctx)
	}
	cacheKey := orgID

	c.mu.RLock()
	if cached, ok := c.cache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
		token := cached.token
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	connector, err := c.resolveConnector(ctx)
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

	c.mu.Lock()
	c.cache[cacheKey] = &cachedConnector{
		connector: connector,
		token:     token,
		expiresAt: time.Now().Add(connectorCacheTTL),
	}
	c.mu.Unlock()

	return token, nil
}

func (c *Client) resolveConnector(ctx context.Context) (*shared_types.GithubConnector, error) {
	if c.db == nil {
		return nil, fmt.Errorf("no GitHub connector with valid credentials found")
	}
	var connectors []shared_types.GithubConnector
	err := c.db.NewSelect().
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

func (c *Client) DoRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, int, error) {
	token, err := c.GetInstallationToken(ctx)
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

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("github api request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	return respBody, resp.StatusCode, nil
}

func (c *Client) DoJSON(ctx context.Context, method, path string, payload interface{}) (json.RawMessage, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		body = strings.NewReader(string(data))
	}

	respBody, status, err := c.DoRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	if status >= 400 {
		return nil, fmt.Errorf("GitHub API %s %s returned %d: %s", method, path, status, Truncate(string(respBody), 500))
	}

	return respBody, nil
}

func IsProtectedBranch(branch string) bool {
	b := strings.ToLower(strings.TrimSpace(branch))
	return b == "main" || b == "master"
}

func IsSensitiveFile(path string) bool {
	lower := strings.ToLower(path)
	for _, pattern := range SensitiveFilePatterns {
		if strings.HasSuffix(lower, pattern) || strings.Contains(lower, pattern+"/") {
			return true
		}
	}
	return false
}

func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
