package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/service/git"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/testutil"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGitClient struct {
	removeErr error
}

func (m *mockGitClient) Clone(string, string) error                       { return nil }
func (m *mockGitClient) Pull(string, string) error                        { return nil }
func (m *mockGitClient) SetHeadToCommitHash(string, string, string) error { return nil }
func (m *mockGitClient) SwitchBranch(string, string) error                { return nil }
func (m *mockGitClient) HasUncommittedChanges(string) (bool, error)       { return false, nil }
func (m *mockGitClient) Stash(string) (string, error)                     { return "", nil }
func (m *mockGitClient) ApplyStash(string, string) error                  { return nil }
func (m *mockGitClient) ResetHard(string) error                           { return nil }
func (m *mockGitClient) RemoveRepository(string) error                    { return m.removeErr }

// ---------------------------------------------------------------------------
// NewGithubConnectorService / init wiring
// ---------------------------------------------------------------------------

func TestNewGithubConnectorService(t *testing.T) {
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), testutil.NewMockGithubConnectorStorage())
	assert.NotNil(t, svc)
}

func TestGetSSHManager_ErrorWithoutContext(t *testing.T) {
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), testutil.NewMockGithubConnectorStorage())
	_, err := svc.getSSHManager(context.Background())
	assert.Error(t, err)
}

func TestGetGitClient_ErrorWithoutContext(t *testing.T) {
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), testutil.NewMockGithubConnectorStorage())
	_, err := svc.getGitClient(context.Background())
	assert.Error(t, err)
}

func TestRemoveRepository_ErrorWithoutContext(t *testing.T) {
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), testutil.NewMockGithubConnectorStorage())
	err := svc.RemoveRepository(context.Background(), "/some/path")
	assert.Error(t, err)
}

func TestRemoveRepository_UsesProvidedGitClient(t *testing.T) {
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), testutil.NewMockGithubConnectorStorage())
	svc.gitClientProvider = func(context.Context) (git.Git, error) {
		return &mockGitClient{}, nil
	}

	err := svc.RemoveRepository(context.Background(), "/some/path")
	require.NoError(t, err)
}

func TestCreateAuthenticatedRepoURL_DelegatesToGithubPackage(t *testing.T) {
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), testutil.NewMockGithubConnectorStorage())

	got, err := svc.CreateAuthenticatedRepoURL("https://github.com/nixopus/nixopus.git", "token-123")
	require.NoError(t, err)
	assert.Equal(t, "https://oauth2:token-123@github.com/nixopus/nixopus.git", got)
}

// ---------------------------------------------------------------------------
// GetClonePath
// ---------------------------------------------------------------------------

func TestGetClonePath_NoOrgInContext(t *testing.T) {
	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), testutil.NewMockGithubConnectorStorage())
	_, _, err := svc.GetClonePath(context.Background(), "uid", "prod", "app-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "organization ID required for SFTP pool")
}

func TestGetClonePath_SFTPClientError(t *testing.T) {
	org := uuid.New().String()
	pool := utils.NewSFTPPool(5*time.Minute, func(string, *ssh.SSHManager) (*sftp.Client, error) {
		return nil, errors.New("no client")
	})
	sshMgr := ssh.NewSSHManager()
	ctx := utils.WithTestSFTPPool(context.Background(), org, pool, sshMgr)

	svc := NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), testutil.NewMockGithubConnectorStorage())
	_, _, err := svc.GetClonePath(ctx, "user1", "prod", "app1")
	require.Error(t, err)
}
