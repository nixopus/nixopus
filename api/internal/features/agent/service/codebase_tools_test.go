package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeRepositoryHandler_Success(t *testing.T) {
	treeResponse := map[string]interface{}{
		"sha": "abc123def",
		"tree": []map[string]interface{}{
			{"path": "package.json", "type": "blob", "content": `{"name":"test","scripts":{"start":"node index.js"}}`},
			{"path": "index.js", "type": "blob", "content": `const http = require("http"); http.createServer((req,res) => res.end("ok")).listen(3000);`},
			{"path": "README.md", "type": "blob", "content": "# Test App"},
		},
	}

	connectorResponse := []map[string]interface{}{
		{"id": "conn-123"},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/github-connector", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(connectorResponse)
	})
	mux.HandleFunc("/api/v1/github-connector/repository/tree", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "testowner", r.URL.Query().Get("owner"))
		assert.Equal(t, "testrepo", r.URL.Query().Get("repo"))
		json.NewEncoder(w).Encode(treeResponse)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	svc := &AgentService{}
	ctx := context.WithValue(context.Background(), ctxKeyAuthToken, "test-token")
	ctx = context.WithValue(ctx, ctxKeyOrgID, "org-1")
	ctx = context.WithValue(ctx, ctxKeyBaseURL, server.URL)

	args, _ := json.Marshal(map[string]string{
		"owner": "testowner",
		"repo":  "testrepo",
	})

	result, err := svc.analyzeRepositoryHandler(ctx, args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))

	assert.Equal(t, float64(3), resp["file_count"])
	assert.Equal(t, "abc123def", resp["commit_sha"])
	assert.NotNil(t, resp["hints"])
	assert.NotEmpty(t, resp["message"])
}

func TestAnalyzeRepositoryHandler_WithBranch(t *testing.T) {
	treeResponse := map[string]interface{}{
		"sha":  "def456",
		"tree": []map[string]interface{}{{"path": "main.go", "type": "blob", "content": "package main"}},
	}

	var capturedBranch string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/github-connector", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]interface{}{{"id": "c1"}})
	})
	mux.HandleFunc("/api/v1/github-connector/repository/tree", func(w http.ResponseWriter, r *http.Request) {
		capturedBranch = r.URL.Query().Get("branch")
		json.NewEncoder(w).Encode(treeResponse)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	svc := &AgentService{}
	ctx := context.WithValue(context.Background(), ctxKeyBaseURL, server.URL)

	args, _ := json.Marshal(map[string]string{"owner": "o", "repo": "r", "branch": "develop"})
	_, err := svc.analyzeRepositoryHandler(ctx, args)
	require.NoError(t, err)
	assert.Equal(t, "develop", capturedBranch)
}

func TestAnalyzeRepositoryHandler_MissingOwnerRepo(t *testing.T) {
	svc := &AgentService{}
	ctx := context.Background()

	args, _ := json.Marshal(map[string]string{"owner": "", "repo": ""})
	_, err := svc.analyzeRepositoryHandler(ctx, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestAnalyzeRepositoryHandler_NoConnectors(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/github-connector", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]interface{}{})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	svc := &AgentService{}
	ctx := context.WithValue(context.Background(), ctxKeyBaseURL, server.URL)
	args, _ := json.Marshal(map[string]string{"owner": "o", "repo": "r"})

	_, err := svc.analyzeRepositoryHandler(ctx, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no GitHub connectors")
}

func TestAnalyzeRepositoryHandler_TreeAPIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/github-connector", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]interface{}{{"id": "c1"}})
	})
	mux.HandleFunc("/api/v1/github-connector/repository/tree", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	svc := &AgentService{}
	ctx := context.WithValue(context.Background(), ctxKeyBaseURL, server.URL)
	args, _ := json.Marshal(map[string]string{"owner": "o", "repo": "r"})

	result, err := svc.analyzeRepositoryHandler(ctx, args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Contains(t, resp["error"], "404")
}

func TestAnalyzeRepositoryHandler_NoFiles(t *testing.T) {
	treeResponse := map[string]interface{}{
		"sha":  "abc",
		"tree": []map[string]interface{}{{"path": "images/logo.png", "type": "blob", "content": ""}},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/github-connector", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]interface{}{{"id": "c1"}})
	})
	mux.HandleFunc("/api/v1/github-connector/repository/tree", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(treeResponse)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	svc := &AgentService{}
	ctx := context.WithValue(context.Background(), ctxKeyBaseURL, server.URL)
	args, _ := json.Marshal(map[string]string{"owner": "o", "repo": "r"})

	result, err := svc.analyzeRepositoryHandler(ctx, args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Contains(t, resp["error"], "no analyzable files")
}

func TestLoadRemoteRepositoryHandler_Success(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := createTestGitRepo(t, map[string]string{
		"package.json": `{"name":"test","scripts":{"start":"node index.js"}}`,
		"index.js":     `console.log("hello")`,
	})

	args, _ := json.Marshal(map[string]string{
		"repo_url": "file://" + repoDir,
	})

	result, err := loadRemoteRepositoryHandler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))

	assert.Equal(t, float64(2), resp["file_count"])
	assert.NotEmpty(t, resp["commit"])
	assert.NotEmpty(t, resp["branch"])
	assert.NotNil(t, resp["hints"])
	assert.NotEmpty(t, resp["message"])
}

func TestLoadRemoteRepositoryHandler_WithBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := createTestGitRepo(t, map[string]string{"main.py": "print('hi')"})

	cmd := exec.Command("git", "-C", repoDir, "checkout", "-b", "feature")
	require.NoError(t, cmd.Run())

	os.WriteFile(filepath.Join(repoDir, "extra.py"), []byte("x = 1"), 0644)
	exec.Command("git", "-C", repoDir, "add", ".").Run()
	exec.Command("git", "-C", repoDir, "commit", "-m", "feature").Run()

	args, _ := json.Marshal(map[string]string{
		"repo_url": "file://" + repoDir,
		"branch":   "feature",
	})

	result, err := loadRemoteRepositoryHandler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Equal(t, float64(2), resp["file_count"])
}

func TestLoadRemoteRepositoryHandler_MissingURL(t *testing.T) {
	args, _ := json.Marshal(map[string]string{"repo_url": ""})
	_, err := loadRemoteRepositoryHandler(context.Background(), args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestLoadRemoteRepositoryHandler_NonHTTPS(t *testing.T) {
	args, _ := json.Marshal(map[string]string{"repo_url": "git@github.com:owner/repo.git"})
	_, err := loadRemoteRepositoryHandler(context.Background(), args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTPS")
}

func TestLoadRemoteRepositoryHandler_InvalidURL(t *testing.T) {
	args, _ := json.Marshal(map[string]string{"repo_url": "https://example.com/nonexistent.git"})
	_, err := loadRemoteRepositoryHandler(context.Background(), args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git clone failed")
}

func TestLoadRemoteRepositoryHandler_BinaryOnlyRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	exec.Command("git", "-C", repoDir, "init").Run()
	exec.Command("git", "-C", repoDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", repoDir, "config", "user.name", "Test").Run()

	os.WriteFile(filepath.Join(repoDir, "image.png"), []byte{0x89, 0x50, 0x4E, 0x47, 0x00}, 0644)
	exec.Command("git", "-C", repoDir, "add", ".").Run()
	exec.Command("git", "-C", repoDir, "commit", "-m", "binary").Run()

	args, _ := json.Marshal(map[string]string{"repo_url": "file://" + repoDir})
	result, err := loadRemoteRepositoryHandler(context.Background(), args)
	require.NoError(t, err)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &resp))
	assert.Contains(t, resp["error"], "no analyzable")
}

func TestCollectFiles_SkipsDirs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0755)
	os.MkdirAll(filepath.Join(dir, "src"), 0755)

	os.WriteFile(filepath.Join(dir, "node_modules", "pkg", "index.js"), []byte("skip"), 0644)
	os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main"), 0644)

	files, err := collectFiles(dir)
	require.NoError(t, err)

	assert.Len(t, files, 1)
	assert.Equal(t, filepath.Join("src", "main.go"), files[0].Path)
}

func TestCollectFiles_SkipsBinaryExtensions(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log('hi')"), 0644)
	os.WriteFile(filepath.Join(dir, "logo.png"), []byte("fake png"), 0644)
	os.WriteFile(filepath.Join(dir, "bundle.wasm"), []byte("fake wasm"), 0644)

	files, err := collectFiles(dir)
	require.NoError(t, err)

	assert.Len(t, files, 1)
	assert.Equal(t, "app.js", files[0].Path)
}

func TestCollectFiles_SkipsBinaryContent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "text.txt"), []byte("hello world"), 0644)
	os.WriteFile(filepath.Join(dir, "binary.dat"), []byte{0x00, 0x01, 0x02}, 0644)

	files, err := collectFiles(dir)
	require.NoError(t, err)

	assert.Len(t, files, 1)
	assert.Equal(t, "text.txt", files[0].Path)
}

func TestCollectFiles_SkipsLargeFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "small.txt"), []byte("ok"), 0644)

	large := make([]byte, maxFileSize+1)
	for i := range large {
		large[i] = 'x'
	}
	os.WriteFile(filepath.Join(dir, "huge.txt"), large, 0644)

	files, err := collectFiles(dir)
	require.NoError(t, err)

	assert.Len(t, files, 1)
	assert.Equal(t, "small.txt", files[0].Path)
}

func TestCollectFiles_SkipsEmptyFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "empty.txt"), []byte{}, 0644)
	os.WriteFile(filepath.Join(dir, "real.txt"), []byte("content"), 0644)

	files, err := collectFiles(dir)
	require.NoError(t, err)

	assert.Len(t, files, 1)
	assert.Equal(t, "real.txt", files[0].Path)
}

func TestCollectFiles_RespectsMaxFileCount(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < maxFileCount+100; i++ {
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("file_%05d.txt", i)), []byte("data"), 0644)
	}

	files, err := collectFiles(dir)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(files), maxFileCount)
}

func TestIsBinaryContent(t *testing.T) {
	assert.False(t, isBinaryContent([]byte("hello world")))
	assert.True(t, isBinaryContent([]byte{0x00, 0x01, 0x02}))
	assert.False(t, isBinaryContent([]byte{}))
	assert.True(t, isBinaryContent([]byte("text\x00more")))
}

func TestShouldSkipPath(t *testing.T) {
	assert.True(t, shouldSkipPath("node_modules/express/index.js"))
	assert.True(t, shouldSkipPath("src/assets/logo.png"))
	assert.True(t, shouldSkipPath(".git/config"))
	assert.True(t, shouldSkipPath("__pycache__/mod.pyc"))
	assert.False(t, shouldSkipPath("src/main.go"))
	assert.False(t, shouldSkipPath("Dockerfile"))
	assert.False(t, shouldSkipPath("docker-compose.yml"))
}

func TestConfidenceMessage(t *testing.T) {
	assert.Contains(t, confidenceMessage("high"), "high confidence")
	assert.Contains(t, confidenceMessage("medium"), "medium confidence")
	assert.Contains(t, confidenceMessage("low"), "low confidence")
	assert.Contains(t, confidenceMessage("unknown"), "low confidence")
}

func TestToolDefinitions(t *testing.T) {
	svc := &AgentService{}

	analyzeTool := svc.analyzeRepositoryTool()
	assert.Equal(t, "analyze_repository", analyzeTool.Name)
	assert.NotNil(t, analyzeTool.Handler)
	assert.NotEmpty(t, analyzeTool.Description)

	var params map[string]interface{}
	require.NoError(t, json.Unmarshal(analyzeTool.Parameters, &params))
	props := params["properties"].(map[string]interface{})
	assert.Contains(t, props, "owner")
	assert.Contains(t, props, "repo")

	loadTool := loadRemoteRepositoryTool()
	assert.Equal(t, "load_remote_repository", loadTool.Name)
	assert.NotNil(t, loadTool.Handler)
	assert.NotEmpty(t, loadTool.Description)

	require.NoError(t, json.Unmarshal(loadTool.Parameters, &params))
	props = params["properties"].(map[string]interface{})
	assert.Contains(t, props, "repo_url")
}

func TestAnalyzeRepositoryHandler_InvalidJSON(t *testing.T) {
	svc := &AgentService{}
	_, err := svc.analyzeRepositoryHandler(context.Background(), json.RawMessage(`{bad`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestLoadRemoteRepositoryHandler_InvalidJSON(t *testing.T) {
	_, err := loadRemoteRepositoryHandler(context.Background(), json.RawMessage(`not json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// createTestGitRepo initializes a bare-compatible git repo with the given files
// and returns the path. The repo is cleaned up automatically by t.Cleanup.
func createTestGitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v failed: %s", args, string(out))
	}

	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	for name, content := range files {
		p := filepath.Join(dir, name)
		os.MkdirAll(filepath.Dir(p), 0755)
		require.NoError(t, os.WriteFile(p, []byte(content), 0644))
	}

	run("add", ".")
	run("commit", "-m", "initial")

	return dir
}
