package git

import (
	"errors"
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/stretchr/testify/assert"
)

func newTestGit(runFn func(string) (string, error)) Git {
	if runFn == nil {
		runFn = func(string) (string, error) {
			return "", errors.New("unexpected run() in test (pass a runFn or use args that fail validation)")
		}
	}
	return &sshGit{
		logger: logger.NewLogger(),
		runCmd: runFn,
	}
}

func TestNewGit(t *testing.T) {
	c := NewGit(logger.NewLogger(), nil)
	assert.NotNil(t, c)
}

func TestRun_UsesSSHManagerWhenRunCmdNotProvided(t *testing.T) {
	manager := ssh.NewSSHManager()
	defer manager.Close()

	c := &sshGit{
		logger:  logger.NewLogger(),
		manager: manager,
	}

	_, err := c.run("echo hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get SSH session")
}

func TestClone_InvalidRepoURL(t *testing.T) {
	c := newTestGit(nil)
	err := c.Clone("https://bad;injection", "/tmp/dst")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git clone")
}

func TestClone_InvalidDestPath(t *testing.T) {
	c := newTestGit(nil)
	err := c.Clone("https://github.com/u/r.git", "dst; rm -rf /")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git clone")
}

func TestClone_RunError(t *testing.T) {
	c := newTestGit(func(cmd string) (string, error) {
		return "stderr output", errors.New("exit 128")
	})
	err := c.Clone("https://github.com/u/r.git", "/tmp/dst")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git clone failed")
}

func TestClone_Success(t *testing.T) {
	c := newTestGit(func(cmd string) (string, error) {
		return "", nil
	})
	err := c.Clone("https://github.com/u/r.git", "/tmp/dst")
	assert.NoError(t, err)
}

func TestPull_InvalidRepoURL(t *testing.T) {
	c := newTestGit(nil)
	err := c.Pull("https://bad;injection", "/tmp/dst")
	assert.Error(t, err)
}

func TestPull_InvalidDestPath(t *testing.T) {
	c := newTestGit(nil)
	err := c.Pull("https://github.com/u/r.git", "dst; rm -rf /")
	assert.Error(t, err)
}

func TestPull_RunError(t *testing.T) {
	c := newTestGit(func(cmd string) (string, error) {
		return "err", errors.New("exit 1")
	})
	err := c.Pull("https://github.com/u/r.git", "/tmp/dst")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git pull failed")
}

func TestPull_Success(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", nil })
	assert.NoError(t, c.Pull("https://github.com/u/r.git", "/tmp/dst"))
}

func TestSetHeadToCommitHash_InvalidPath(t *testing.T) {
	c := newTestGit(nil)
	err := c.SetHeadToCommitHash("url", "dst; bad", "abc123")
	assert.Error(t, err)
}

func TestSetHeadToCommitHash_InvalidRef(t *testing.T) {
	c := newTestGit(nil)
	err := c.SetHeadToCommitHash("url", "/tmp/dst", "bad ref with spaces")
	assert.Error(t, err)
}

func TestSetHeadToCommitHash_RunError(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", errors.New("exit 1") })
	err := c.SetHeadToCommitHash("url", "/tmp/dst", "abc123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git checkout failed")
}

func TestSetHeadToCommitHash_Success(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", nil })
	assert.NoError(t, c.SetHeadToCommitHash("url", "/tmp/dst", "abc123"))
}

func TestSwitchBranch_InvalidPath(t *testing.T) {
	c := newTestGit(nil)
	err := c.SwitchBranch("dst; bad", "main")
	assert.Error(t, err)
}

func TestSwitchBranch_InvalidRef(t *testing.T) {
	c := newTestGit(nil)
	err := c.SwitchBranch("/tmp/dst", "bad branch name")
	assert.Error(t, err)
}

func TestSwitchBranch_RunError(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", errors.New("exit 1") })
	err := c.SwitchBranch("/tmp/dst", "main")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git checkout branch failed")
}

func TestSwitchBranch_Success(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", nil })
	assert.NoError(t, c.SwitchBranch("/tmp/dst", "main"))
}

func TestHasUncommittedChanges_InvalidPath(t *testing.T) {
	c := newTestGit(nil)
	_, err := c.HasUncommittedChanges("dst; bad")
	assert.Error(t, err)
}

func TestHasUncommittedChanges_RunError(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", errors.New("exit 1") })
	_, err := c.HasUncommittedChanges("/tmp/dst")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git status failed")
}

func TestHasUncommittedChanges_Clean(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "  ", nil })
	has, err := c.HasUncommittedChanges("/tmp/dst")
	assert.NoError(t, err)
	assert.False(t, has)
}

func TestHasUncommittedChanges_Dirty(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return " M file.go\n", nil })
	has, err := c.HasUncommittedChanges("/tmp/dst")
	assert.NoError(t, err)
	assert.True(t, has)
}

func TestStash_InvalidPath(t *testing.T) {
	c := newTestGit(nil)
	_, err := c.Stash("dst; bad")
	assert.Error(t, err)
}

func TestStash_PushRunError(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", errors.New("push failed") })
	_, err := c.Stash("/tmp/dst")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git stash push failed")
}

func TestStash_ListRunError(t *testing.T) {
	callCount := 0
	c := newTestGit(func(cmd string) (string, error) {
		callCount++
		if callCount == 1 {
			return "", nil
		}
		return "", errors.New("list failed")
	})
	_, err := c.Stash("/tmp/dst")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git stash list failed")
}

func TestStash_EmptyStashID(t *testing.T) {
	callCount := 0
	c := newTestGit(func(_ string) (string, error) {
		callCount++
		return "", nil
	})
	_, err := c.Stash("/tmp/dst")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no stash created")
}

func TestStash_Success(t *testing.T) {
	callCount := 0
	c := newTestGit(func(_ string) (string, error) {
		callCount++
		if callCount == 2 {
			return "stash-abc123", nil
		}
		return "", nil
	})
	id, err := c.Stash("/tmp/dst")
	assert.NoError(t, err)
	assert.Equal(t, "stash-abc123", id)
}

func TestApplyStash_InvalidPath(t *testing.T) {
	c := newTestGit(nil)
	err := c.ApplyStash("dst; bad", "stash-id")
	assert.Error(t, err)
}

func TestApplyStash_InvalidRef(t *testing.T) {
	c := newTestGit(nil)
	err := c.ApplyStash("/tmp/dst", "bad stash id")
	assert.Error(t, err)
}

func TestApplyStash_RunError(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", errors.New("exit 1") })
	err := c.ApplyStash("/tmp/dst", "stash-abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git stash apply failed")
}

func TestApplyStash_Success(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", nil })
	assert.NoError(t, c.ApplyStash("/tmp/dst", "stash-abc"))
}

func TestResetHard_InvalidPath(t *testing.T) {
	c := newTestGit(nil)
	err := c.ResetHard("dst; bad")
	assert.Error(t, err)
}

func TestResetHard_RunError(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", errors.New("exit 1") })
	err := c.ResetHard("/tmp/dst")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git reset --hard failed")
}

func TestResetHard_Success(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", nil })
	assert.NoError(t, c.ResetHard("/tmp/dst"))
}

func TestRemoveRepository_InvalidPath(t *testing.T) {
	c := newTestGit(nil)
	err := c.RemoveRepository("dst; bad")
	assert.Error(t, err)
}

func TestRemoveRepository_RunError(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", errors.New("exit 1") })
	err := c.RemoveRepository("/tmp/dst")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove repository directory")
}

func TestRemoveRepository_Success(t *testing.T) {
	c := newTestGit(func(_ string) (string, error) { return "", nil })
	assert.NoError(t, c.RemoveRepository("/tmp/dst"))
}
