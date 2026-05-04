package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nixopus/nixopus/api/internal/features/github-connector/service/git"
	gh "github.com/nixopus/nixopus/api/internal/features/github-connector/service/github"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/testutil"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// cloneGitClient is a minimal git.Git for CloneRepository tests.
type cloneGitClient struct {
	cloneErr      error
	pullErr       error
	switchErr     error
	setHeadErr    error
	pullHook      *bool
	cloneCalled   bool
	switchCalled  bool
	setHeadCalled bool
	resetHardErr  error
	hasChanges    bool
	hasChangesErr error
}

func (c *cloneGitClient) Clone(_, _ string) error {
	c.cloneCalled = true
	return c.cloneErr
}
func (c *cloneGitClient) Pull(_, dest string) error {
	if c.pullHook != nil {
		*c.pullHook = true
	}
	return c.pullErr
}
func (c *cloneGitClient) SetHeadToCommitHash(_, _, _ string) error {
	c.setHeadCalled = true
	return c.setHeadErr
}
func (c *cloneGitClient) SwitchBranch(_, _ string) error {
	c.switchCalled = true
	return c.switchErr
}
func (c *cloneGitClient) HasUncommittedChanges(_ string) (bool, error) {
	return c.hasChanges, c.hasChangesErr
}
func (c *cloneGitClient) Stash(_ string) (string, error)  { return "", errors.New("not used") }
func (c *cloneGitClient) ApplyStash(_, _ string) error    { return errors.New("not used") }
func (c *cloneGitClient) ResetHard(_ string) error        { return c.resetHardErr }
func (c *cloneGitClient) RemoveRepository(_ string) error { return nil }

func makeConnector(userID uuid.UUID) shared_types.GithubConnector {
	return testutil.MakeTestConnector(userID)
}

func testCloneHTTPServer(t *testing.T, repo shared_types.GithubRepository, installID string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/app/installations/"+installID+"/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"token": "tok123"})
	})
	mux.HandleFunc("/repositories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(repo)
	})
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/test/commits/HEAD" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"sha": "deadbeef"})
			return
		}
		http.NotFound(w, r)
	})
	return httptest.NewServer(mux)
}

func TestCloneRepository_GetRepositoryByIDError(t *testing.T) {
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", "user-1").Return([]shared_types.GithubConnector{}, nil)

	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        1,
		UserID:        "user-1",
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
}

func TestCloneRepository_NoConnectorsForToken(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		Name:     "test",
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := testCloneHTTPServer(t, repo, "67890")
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Once()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{}, nil).Once()

	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no connectors found")
}

func TestCloneRepository_InstallationTokenError(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repositories/9" {
			json.NewEncoder(w).Encode(repo)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
}

func TestCloneRepository_AuthURLInvalid(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "ftp://bad", // CreateAuthenticatedRepoURL rejects
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
}

func TestCloneRepository_GitClientError(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) {
		return nil, errors.New("no ssh")
	}
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", false, nil
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no ssh")
}

func TestCloneRepository_LatestCommitError(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	mux := http.NewServeMux()
	mux.HandleFunc("/app/installations/67890/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"token": "tok123"})
	})
	mux.HandleFunc("/repositories/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(repo)
	})
	// no commits/HEAD route -> 404 from default
	srv := httptest.NewServer(mux)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) { return &cloneGitClient{}, nil }
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", false, nil
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
}

func TestCloneRepository_GetClonePathError(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) {
		return &cloneGitClient{}, nil
	}
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "", false, errors.New("path err")
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
}

func TestCloneRepository_RollbackBranch(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"
	hash := "rollbacksha"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	mockGit := &cloneGitClient{}
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) { return mockGit, nil }
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", false, nil
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:         9,
		UserID:         userID,
		Environment:    "prod",
		ApplicationID:  "app",
		DeploymentType: shared_types.DeploymentTypeRollback,
	}, &hash)
	require.NoError(t, err)
	assert.True(t, mockGit.setHeadCalled)
}

func TestCloneRepository_RollbackSetHeadError(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"
	hash := "rollbacksha"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	mockGit := &cloneGitClient{setHeadErr: errors.New("checkout failed")}
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) { return mockGit, nil }
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", false, nil
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:         9,
		UserID:         userID,
		Environment:    "prod",
		ApplicationID:  "app",
		DeploymentType: shared_types.DeploymentTypeRollback,
	}, &hash)
	require.Error(t, err)
}

func TestCloneRepository_CloneFreshAndBranch(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	mockGit := &cloneGitClient{}
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) { return mockGit, nil }
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", false, nil
	}

	path, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
		Branch:        "main",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/r", path)
	assert.True(t, mockGit.cloneCalled)
	assert.True(t, mockGit.switchCalled)
}

func TestCloneRepository_CloneError(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	mockGit := &cloneGitClient{cloneErr: errors.New("clone failed")}
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) { return mockGit, nil }
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", false, nil
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
}

func TestCloneRepository_PullPath(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	var pulled bool
	mockGit := &cloneGitClient{pullHook: &pulled}
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) { return mockGit, nil }
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", true, nil
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.NoError(t, err)
	assert.True(t, pulled)
}

func TestCloneRepository_PullPathError(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	mockGit := &cloneGitClient{pullErr: errors.New("pull failed")}
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) { return mockGit, nil }
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", true, nil
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
}

func TestCloneRepository_SwitchBranchError(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	mockGit := &cloneGitClient{switchErr: errors.New("switch failed")}
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) { return mockGit, nil }
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", false, nil
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
		Branch:        "main",
	}, nil)
	require.Error(t, err)
}

func TestCloneRepository_WithPinnedCommitHash(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}
	connector := makeConnector(uuid.MustParse(userID))
	connector.InstallationID = "67890"
	pinned := "aaa111"

	srv := testCloneHTTPServer(t, repo, connector.InstallationID)
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return([]shared_types.GithubConnector{connector}, nil).Twice()

	mockGit := &cloneGitClient{}
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	svc.gitClientProvider = func(context.Context) (git.Git, error) { return mockGit, nil }
	svc.clonePathProvider = func(context.Context, string, string, string) (string, bool, error) {
		return "/tmp/r", false, nil
	}

	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, &pinned)
	require.NoError(t, err)
	assert.True(t, mockGit.cloneCalled)
}

func TestCloneRepository_StorageErrorOnConnectors(t *testing.T) {
	userID := uuid.New().String()
	repo := shared_types.GithubRepository{
		ID:       9,
		CloneURL: "https://github.com/owner/test.git",
		FullName: "owner/test",
	}

	srv := testCloneHTTPServer(t, repo, "67890")
	defer srv.Close()

	prev := gh.APIBaseURL
	gh.SetAPIBaseURL(srv.URL)
	defer gh.SetAPIBaseURL(prev)

	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return(nil, errors.New("db down")).Once()

	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockSt)
	_, err := svc.CloneRepository(context.Background(), CloneRepositoryConfig{
		RepoID:        9,
		UserID:        userID,
		Environment:   "prod",
		ApplicationID: "app",
	}, nil)
	require.Error(t, err)
}
