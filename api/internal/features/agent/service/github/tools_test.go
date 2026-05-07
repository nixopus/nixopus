package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ctxKeyTestOrgID struct{}

func testCtxWithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, ctxKeyTestOrgID{}, orgID)
}

func testOrgIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyTestOrgID{}).(string)
	return v
}

// setupGithubTestServer creates a mock GitHub API server and returns a Client
// that bypasses connector resolution by pre-caching a token.
func setupGithubTestServer(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	restore := RedirectAPIToTestServer(srv.URL)
	t.Cleanup(restore)

	gc := NewProbeClient("test-token", testOrgIDFromCtx)
	return gc, srv
}

func TestGithubListPullRequests_Success(t *testing.T) {
	expected := []map[string]interface{}{
		{"number": 1, "title": "Fix bug", "state": "open"},
		{"number": 2, "title": "Add feature", "state": "open"},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "open", r.URL.Query().Get("state"))
		assert.Equal(t, "30", r.URL.Query().Get("per_page"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		json.NewEncoder(w).Encode(expected)
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := ListPullRequestsTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo"})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var prs []map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &prs))
	assert.Len(t, prs, 2)
	assert.Equal(t, "Fix bug", prs[0]["title"])
}

func TestGithubListPullRequests_CustomState(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "closed", r.URL.Query().Get("state"))
		assert.Equal(t, "50", r.URL.Query().Get("per_page"))
		json.NewEncoder(w).Encode([]interface{}{})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := ListPullRequestsTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo", "state": "closed", "per_page": 50})
	_, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
}

func TestGithubListIssues_FiltersPRs(t *testing.T) {
	items := []map[string]interface{}{
		{"number": 1, "title": "Bug report"},
		{"number": 2, "title": "Feature PR", "pull_request": map[string]string{"url": "https://..."}},
		{"number": 3, "title": "Another bug"},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(items)
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := ListIssuesTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo"})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var issues []map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &issues))
	assert.Len(t, issues, 2)
	assert.Equal(t, "Bug report", issues[0]["title"])
	assert.Equal(t, "Another bug", issues[1]["title"])
}

func TestGithubCommentOnPR_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues/42/comments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Great work!", body["body"])
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 123, "body": body["body"]})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := CommentOnPRTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo", "pr_number": 42, "body": "Great work!"})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, float64(123), resp["id"])
}

func TestGithubCommentOnIssue_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues/10/comments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 456})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := CommentOnIssueTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo", "issue_number": 10, "body": "Noted"})
	_, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
}

func TestGithubCreateIssue_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "New bug", body["title"])
		assert.Equal(t, "Description here", body["body"])
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"number": 55, "title": "New bug"})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := CreateIssueTool(gc)

	args, _ := json.Marshal(map[string]interface{}{
		"owner": "owner", "repo": "repo",
		"title": "New bug", "body": "Description here",
		"labels": []string{"bug"},
	})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, float64(55), resp["number"])
}

func TestGithubSetCommitStatus_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/statuses/abc123", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "success", body["state"])
		assert.Equal(t, "ci/nixopus", body["context"])
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"state": "success"})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := SetCommitStatusTool(gc)

	args, _ := json.Marshal(map[string]interface{}{
		"owner": "owner", "repo": "repo", "sha": "abc123",
		"state": "success", "context": "ci/nixopus",
	})
	_, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
}

func TestGithubCreateDeploymentStatus_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/deployments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"id": 999})
		}
	})
	mux.HandleFunc("/repos/owner/repo/deployments/999/statuses", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "success", body["state"])
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"state": "success"})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := CreateDeploymentStatusTool(gc)

	args, _ := json.Marshal(map[string]interface{}{
		"owner": "owner", "repo": "repo", "ref": "main",
		"environment": "production", "state": "success",
	})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, float64(999), resp["deployment_id"])
}

func TestGithubSearchRepoContent_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/search/code", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		assert.Contains(t, q, "TODO")
		assert.Contains(t, q, "repo:owner/repo")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count": 2,
			"items":       []map[string]string{{"name": "main.go"}, {"name": "util.go"}},
		})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := SearchRepoContentTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo", "query": "TODO"})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, float64(2), resp["total_count"])
}

func TestGithubCreateOrUpdateFile_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/contents/src/hello.go", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "feature-x", body["branch"])
		assert.Equal(t, "add hello", body["message"])

		decoded, err := base64.StdEncoding.DecodeString(body["content"])
		require.NoError(t, err)
		assert.Equal(t, "package main", string(decoded))

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"content": map[string]string{"sha": "newsha"}})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := CreateOrUpdateFileTool(gc)

	args, _ := json.Marshal(map[string]interface{}{
		"owner": "owner", "repo": "repo", "path": "src/hello.go",
		"content": "package main", "message": "add hello", "branch": "feature-x",
	})
	_, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
}

func TestGithubCreateOrUpdateFile_BlocksProtectedBranch(t *testing.T) {
	gc := &Client{
		cache: map[string]*cachedConnector{
			"": {token: "t", expiresAt: time.Now().Add(time.Hour)},
		},
		OrgIDFromCtx: testOrgIDFromCtx,
	}
	tool := CreateOrUpdateFileTool(gc)

	for _, branch := range []string{"main", "master", "Main", "MASTER"} {
		args, _ := json.Marshal(map[string]interface{}{
			"owner": "o", "repo": "r", "path": "f.txt",
			"content": "x", "message": "m", "branch": branch,
		})
		_, err := tool.Handler(context.Background(), args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "protected branch")
	}
}

func TestGithubCreateOrUpdateFile_BlocksSensitiveFiles(t *testing.T) {
	gc := &Client{
		cache: map[string]*cachedConnector{
			"": {token: "t", expiresAt: time.Now().Add(time.Hour)},
		},
		OrgIDFromCtx: testOrgIDFromCtx,
	}
	tool := CreateOrUpdateFileTool(gc)

	sensitiveFiles := []string{
		".env", "config/.env", "server.pem", "private.key",
		"credentials.json", "id_rsa", "cert.p12",
	}

	for _, file := range sensitiveFiles {
		args, _ := json.Marshal(map[string]interface{}{
			"owner": "o", "repo": "r", "path": file,
			"content": "x", "message": "m", "branch": "feature",
		})
		_, err := tool.Handler(context.Background(), args)
		require.Error(t, err, "expected error for file: %s", file)
		assert.Contains(t, err.Error(), "sensitive file", "file: %s", file)
	}
}

func TestGithubCreateOrUpdateFile_AllowsNonSensitiveOnFeatureBranch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/contents/readme.md", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"sha": "abc"})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := CreateOrUpdateFileTool(gc)

	args, _ := json.Marshal(map[string]interface{}{
		"owner": "o", "repo": "r", "path": "readme.md",
		"content": "# Hi", "message": "docs", "branch": "docs-update",
	})
	_, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
}

func TestGithubGetBranch_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/branches/develop", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":   "develop",
			"commit": map[string]string{"sha": "deadbeef"},
		})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := GetBranchTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo", "branch": "develop"})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, "develop", resp["name"])
}

func TestGithubCreateBranch_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/git/refs", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "refs/heads/feature-new", body["ref"])
		assert.Equal(t, "abc123", body["sha"])
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"ref": body["ref"]})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := CreateBranchTool(gc)

	args, _ := json.Marshal(map[string]interface{}{
		"owner": "owner", "repo": "repo", "branch": "feature-new", "sha": "abc123",
	})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, "refs/heads/feature-new", resp["ref"])
}

func TestGithubCreatePullRequest_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Add feature", body["title"])
		assert.Equal(t, "feature-x", body["head"])
		assert.Equal(t, "main", body["base"])
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"number": 77, "html_url": "https://github.com/owner/repo/pull/77"})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := CreatePullRequestTool(gc)

	args, _ := json.Marshal(map[string]interface{}{
		"owner": "owner", "repo": "repo", "title": "Add feature",
		"head": "feature-x", "base": "main",
	})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, float64(77), resp["number"])
}

func TestGithubMergePullRequest_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls/10/merge", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "squash", body["merge_method"])
		json.NewEncoder(w).Encode(map[string]interface{}{"merged": true, "sha": "mergesha"})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := MergePullRequestTool(gc)

	args, _ := json.Marshal(map[string]interface{}{
		"owner": "owner", "repo": "repo", "pr_number": 10, "merge_method": "squash",
	})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, true, resp["merged"])
}

func TestGithubGetRepoFile_Success(t *testing.T) {
	content := "package main\n\nfunc main() {}"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/contents/main.go", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":     "main.go",
			"path":     "main.go",
			"sha":      "filesha",
			"size":     len(content),
			"encoding": "base64",
			"content":  encoded,
		})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := GetRepoFileTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo", "path": "main.go"})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, "main.go", resp["name"])
	assert.Equal(t, content, resp["content"])
	assert.Equal(t, false, resp["truncated"])
}

func TestGithubGetRepoFile_Truncation(t *testing.T) {
	largeContent := strings.Repeat("x", MaxResponseBody+1000)
	encoded := base64.StdEncoding.EncodeToString([]byte(largeContent))

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/contents/big.txt", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":     "big.txt",
			"path":     "big.txt",
			"sha":      "bigsha",
			"size":     len(largeContent),
			"encoding": "base64",
			"content":  encoded,
		})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := GetRepoFileTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo", "path": "big.txt"})
	result, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, true, resp["truncated"])
	assert.Len(t, resp["content"], MaxResponseBody)
}

func TestGithubGetRepoFile_WithRef(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/contents/f.txt", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "develop", r.URL.Query().Get("ref"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "f.txt", "path": "f.txt", "sha": "s",
			"size": 2, "encoding": "base64",
			"content": base64.StdEncoding.EncodeToString([]byte("hi")),
		})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := GetRepoFileTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo", "path": "f.txt", "ref": "develop"})
	_, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
}

func TestGithubAPIError_Handling(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := ListPullRequestsTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "owner", "repo": "repo"})
	_, err := tool.Handler(context.Background(), args)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestGithubInvalidJSON_Args(t *testing.T) {
	gc := &Client{
		cache: map[string]*cachedConnector{
			"": {token: "t", expiresAt: time.Now().Add(time.Hour)},
		},
		OrgIDFromCtx: testOrgIDFromCtx,
	}
	tools := []llm.ToolDefinition{
		ListPullRequestsTool(gc),
		ListIssuesTool(gc),
		CommentOnPRTool(gc),
		CommentOnIssueTool(gc),
		CreateIssueTool(gc),
		SetCommitStatusTool(gc),
		CreateDeploymentStatusTool(gc),
		SearchRepoContentTool(gc),
		CreateOrUpdateFileTool(gc),
		GetBranchTool(gc),
		CreateBranchTool(gc),
		CreatePullRequestTool(gc),
		MergePullRequestTool(gc),
		GetRepoFileTool(gc),
	}

	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			_, err := tool.Handler(context.Background(), json.RawMessage(`{invalid`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid")
		})
	}
}

func TestIsProtectedBranch(t *testing.T) {
	assert.True(t, IsProtectedBranch("main"))
	assert.True(t, IsProtectedBranch("master"))
	assert.True(t, IsProtectedBranch("Main"))
	assert.True(t, IsProtectedBranch("MASTER"))
	assert.True(t, IsProtectedBranch(" main "))
	assert.False(t, IsProtectedBranch("develop"))
	assert.False(t, IsProtectedBranch("feature/main-page"))
	assert.False(t, IsProtectedBranch("release/v1"))
}

func TestIsSensitiveFile(t *testing.T) {
	assert.True(t, IsSensitiveFile(".env"))
	assert.True(t, IsSensitiveFile("config/.env"))
	assert.True(t, IsSensitiveFile("server.pem"))
	assert.True(t, IsSensitiveFile("private.key"))
	assert.True(t, IsSensitiveFile("credentials.json"))
	assert.True(t, IsSensitiveFile("id_rsa"))
	assert.True(t, IsSensitiveFile("id_ed25519"))
	assert.True(t, IsSensitiveFile("app.p12"))
	assert.True(t, IsSensitiveFile("cert.pfx"))
	assert.True(t, IsSensitiveFile("store.jks"))
	assert.True(t, IsSensitiveFile("my.secret"))
	assert.False(t, IsSensitiveFile("readme.md"))
	assert.False(t, IsSensitiveFile("src/main.go"))
	assert.False(t, IsSensitiveFile("package.json"))
	assert.False(t, IsSensitiveFile("Dockerfile"))
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", Truncate("hello", 10))
	assert.Equal(t, "hel...", Truncate("hello world", 3))
	assert.Equal(t, "", Truncate("", 5))
}

func TestConnectorCache_Hit(t *testing.T) {
	gc := &Client{
		cache: map[string]*cachedConnector{
			"org1": {token: "cached-token", expiresAt: time.Now().Add(5 * time.Minute)},
		},
		OrgIDFromCtx: testOrgIDFromCtx,
	}

	ctx := testCtxWithOrgID(context.Background(), "org1")
	token, err := gc.GetInstallationToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "cached-token", token)
}

func TestConnectorCache_Expired(t *testing.T) {
	gc := &Client{
		cache: map[string]*cachedConnector{
			"org1": {token: "old-token", expiresAt: time.Now().Add(-1 * time.Minute)},
		},
		OrgIDFromCtx: testOrgIDFromCtx,
	}

	ctx := testCtxWithOrgID(context.Background(), "org1")
	_, err := gc.GetInstallationToken(ctx)
	// Should fail because db is nil and cache is expired, causing a panic-recovered error
	// or a nil pointer. We just confirm it doesn't return a valid token.
	require.Error(t, err)
}

func TestConnectorCache_NoDB(t *testing.T) {
	gc := &Client{
		cache:        map[string]*cachedConnector{},
		OrgIDFromCtx: testOrgIDFromCtx,
	}
	ctx := context.Background()
	_, err := gc.resolveConnector(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no GitHub connector")
}

func TestGithubToolDefinitions_HaveRequiredFields(t *testing.T) {
	gc := &Client{
		cache:        map[string]*cachedConnector{},
		OrgIDFromCtx: testOrgIDFromCtx,
	}

	tools := []llm.ToolDefinition{
		ListPullRequestsTool(gc),
		ListIssuesTool(gc),
		CommentOnPRTool(gc),
		CommentOnIssueTool(gc),
		CreateIssueTool(gc),
		SetCommitStatusTool(gc),
		CreateDeploymentStatusTool(gc),
		SearchRepoContentTool(gc),
		CreateOrUpdateFileTool(gc),
		GetBranchTool(gc),
		CreateBranchTool(gc),
		CreatePullRequestTool(gc),
		MergePullRequestTool(gc),
		GetRepoFileTool(gc),
	}

	expectedNames := []string{
		"github_list_pull_requests", "github_list_issues",
		"github_comment_on_pr", "github_comment_on_issue",
		"github_create_issue", "github_set_commit_status",
		"github_create_deployment_status", "github_search_repo_content",
		"github_create_or_update_file", "github_get_branch",
		"github_create_branch", "github_create_pull_request",
		"github_merge_pull_request", "github_get_repo_file",
	}

	require.Len(t, tools, len(expectedNames))
	for i, tool := range tools {
		assert.Equal(t, expectedNames[i], tool.Name)
		assert.NotEmpty(t, tool.Description)
		assert.NotNil(t, tool.Handler)
		assert.NotNil(t, tool.Parameters)

		var params map[string]interface{}
		require.NoError(t, json.Unmarshal(tool.Parameters, &params), "tool %s has invalid params JSON", tool.Name)
		assert.Equal(t, "object", params["type"])
	}
}

func TestGithubMergePullRequest_DefaultMethod(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/pulls/1/merge", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "merge", body["merge_method"])
		json.NewEncoder(w).Encode(map[string]interface{}{"merged": true})
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := MergePullRequestTool(gc)

	args, _ := json.Marshal(map[string]interface{}{"owner": "o", "repo": "r", "pr_number": 1})
	_, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
}

func TestGithubCreateDeploymentStatus_DeploymentFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/deployments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"message":"Conflict"}`)
	})

	gc, _ := setupGithubTestServer(t, mux)
	tool := CreateDeploymentStatusTool(gc)

	args, _ := json.Marshal(map[string]interface{}{
		"owner": "o", "repo": "r", "ref": "main",
		"environment": "prod", "state": "pending",
	})
	_, err := tool.Handler(context.Background(), args)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create deployment")
}
