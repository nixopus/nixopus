package git

import (
	"errors"
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/assert"
)

type mockGit struct {
	hasChanges    bool
	hasChangesErr error
	resetHardErr  error
	pullErr       error
}

func (m *mockGit) Clone(_, _ string) error                      { return nil }
func (m *mockGit) Pull(_, _ string) error                       { return m.pullErr }
func (m *mockGit) SetHeadToCommitHash(_, _, _ string) error     { return nil }
func (m *mockGit) SwitchBranch(_, _ string) error               { return nil }
func (m *mockGit) HasUncommittedChanges(_ string) (bool, error) { return m.hasChanges, m.hasChangesErr }
func (m *mockGit) Stash(_ string) (string, error)               { return "", nil }
func (m *mockGit) ApplyStash(_, _ string) error                 { return nil }
func (m *mockGit) ResetHard(_ string) error                     { return m.resetHardErr }
func (m *mockGit) RemoveRepository(_ string) error              { return nil }

func TestHandlePullWithClient_HasChangesError(t *testing.T) {
	client := &mockGit{hasChangesErr: errors.New("git status error")}
	err := HandlePullWithClient(logger.NewLogger(), client, "url", "/path", "u1")
	assert.EqualError(t, err, "git status error")
}

func TestHandlePullWithClient_ResetHardError(t *testing.T) {
	client := &mockGit{hasChanges: true, resetHardErr: errors.New("reset error")}
	err := HandlePullWithClient(logger.NewLogger(), client, "url", "/path", "u1")
	assert.EqualError(t, err, "reset error")
}

func TestHandlePullWithClient_PullError(t *testing.T) {
	client := &mockGit{hasChanges: false, pullErr: errors.New("pull error")}
	err := HandlePullWithClient(logger.NewLogger(), client, "url", "/path", "u1")
	assert.EqualError(t, err, "pull error")
}

func TestHandlePullWithClient_SuccessNoChanges(t *testing.T) {
	client := &mockGit{hasChanges: false}
	assert.NoError(t, HandlePullWithClient(logger.NewLogger(), client, "url", "/path", "u1"))
}

func TestHandlePullWithClient_SuccessWithChanges(t *testing.T) {
	client := &mockGit{hasChanges: true}
	assert.NoError(t, HandlePullWithClient(logger.NewLogger(), client, "url", "/path", "u1"))
}
