package git

import (
	"fmt"
	"strings"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// Git is the interface for remote git operations executed over SSH on the tenant server.
type Git interface {
	Clone(repoURL, destinationPath string) error
	Pull(repoURL, destinationPath string) error
	SetHeadToCommitHash(repoURL, destinationPath, commitHash string) error
	SwitchBranch(destinationPath, branch string) error
	HasUncommittedChanges(destinationPath string) (bool, error)
	Stash(destinationPath string) (string, error)
	ApplyStash(destinationPath, stashID string) error
	ResetHard(destinationPath string) error
	RemoveRepository(repoPath string) error
}

// sshGit uses SSHManager for pooled connections — git commands share one TCP connection.
type sshGit struct {
	logger  logger.Logger
	manager *ssh.SSHManager
	// runCmd substitutes manager.RunCommand in tests when non-nil.
	runCmd func(string) (string, error)
}

// NewGit creates a Git implementation backed by the org SSHManager.
func NewGit(logger logger.Logger, manager *ssh.SSHManager) Git {
	return &sshGit{
		logger:  logger,
		manager: manager,
	}
}

func (g *sshGit) run(cmd string) (string, error) {
	if g.runCmd != nil {
		return g.runCmd(cmd)
	}
	return g.manager.RunCommand(cmd)
}

// Clone clones a git repository to the specified path.
func (g *sshGit) Clone(repoURL, destinationPath string) error {
	if err := utils.ValidateShellArg(repoURL, "repoURL"); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	if err := utils.ValidatePath(destinationPath, "destinationPath"); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	cmd := fmt.Sprintf("git clone %s %s", utils.ShellQuote(repoURL), utils.ShellQuote(destinationPath))
	output, err := g.run(cmd)
	if err != nil {
		return fmt.Errorf("git clone failed: %s, output: %s", err.Error(), output)
	}
	g.logger.Log(logger.Info, fmt.Sprintf("github connector git: cloned to %s", destinationPath), "")
	return nil
}

// Pull updates a git repository from remote.
func (g *sshGit) Pull(repoURL, destinationPath string) error {
	if err := utils.ValidateShellArg(repoURL, "repoURL"); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	if err := utils.ValidatePath(destinationPath, "destinationPath"); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	cmd := fmt.Sprintf("cd %s && git pull %s", utils.ShellQuote(destinationPath), utils.ShellQuote(repoURL))
	output, err := g.run(cmd)
	if err != nil {
		return fmt.Errorf("git pull failed: %s, output: %s", err.Error(), output)
	}
	g.logger.Log(logger.Info, fmt.Sprintf("github connector git: pulled at %s", destinationPath), "")
	return nil
}

// SetHeadToCommitHash sets HEAD to a specific commit hash.
func (g *sshGit) SetHeadToCommitHash(repoURL, destinationPath, commitHash string) error {
	if err := utils.ValidatePath(destinationPath, "destinationPath"); err != nil {
		return fmt.Errorf("git checkout: %w", err)
	}
	if err := utils.ValidateGitRef(commitHash, "commitHash"); err != nil {
		return fmt.Errorf("git checkout: %w", err)
	}
	cmd := fmt.Sprintf("cd %s && git checkout %s", utils.ShellQuote(destinationPath), utils.ShellQuote(commitHash))
	output, err := g.run(cmd)
	if err != nil {
		return fmt.Errorf("git checkout failed: %s, output: %s", err.Error(), output)
	}
	g.logger.Log(logger.Info, fmt.Sprintf("github connector git: checked out commit %s at %s", commitHash, destinationPath), "")
	return nil
}

// SwitchBranch switches to the specified branch.
func (g *sshGit) SwitchBranch(destinationPath, branch string) error {
	if err := utils.ValidatePath(destinationPath, "destinationPath"); err != nil {
		return fmt.Errorf("git switch branch: %w", err)
	}
	if err := utils.ValidateGitRef(branch, "branch"); err != nil {
		return fmt.Errorf("git switch branch: %w", err)
	}
	cmd := fmt.Sprintf("cd %s && git checkout %s", utils.ShellQuote(destinationPath), utils.ShellQuote(branch))
	output, err := g.run(cmd)
	if err != nil {
		return fmt.Errorf("git checkout branch failed: %s, output: %s", err.Error(), output)
	}
	g.logger.Log(logger.Info, fmt.Sprintf("github connector git: switched to branch %s at %s", branch, destinationPath), "")
	return nil
}

func (g *sshGit) HasUncommittedChanges(destinationPath string) (bool, error) {
	if err := utils.ValidatePath(destinationPath, "destinationPath"); err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	cmd := fmt.Sprintf("cd %s && git status --porcelain", utils.ShellQuote(destinationPath))
	output, err := g.run(cmd)
	if err != nil {
		return false, fmt.Errorf("git status failed: %s, output: %s", err.Error(), output)
	}
	return strings.TrimSpace(output) != "", nil
}

func (g *sshGit) Stash(destinationPath string) (string, error) {
	if err := utils.ValidatePath(destinationPath, "destinationPath"); err != nil {
		return "", fmt.Errorf("git stash: %w", err)
	}
	quoted := utils.ShellQuote(destinationPath)
	cmd := fmt.Sprintf("cd %s && git stash push -m 'nixopus-auto-stash'", quoted)
	output, err := g.run(cmd)
	if err != nil {
		return "", fmt.Errorf("git stash push failed: %s, output: %s", err.Error(), output)
	}

	cmd = fmt.Sprintf("cd %s && git stash list --format='%%H' -n 1", quoted)
	stashOutput, err := g.run(cmd)
	if err != nil {
		return "", fmt.Errorf("git stash list failed: %s, output: %s", err.Error(), stashOutput)
	}

	stashID := strings.TrimSpace(stashOutput)
	if stashID == "" {
		return "", fmt.Errorf("no stash created")
	}
	g.logger.Log(logger.Info, fmt.Sprintf("github connector git: stashed at %s id=%s", destinationPath, stashID), "")
	return stashID, nil
}

func (g *sshGit) ApplyStash(destinationPath, stashID string) error {
	if err := utils.ValidatePath(destinationPath, "destinationPath"); err != nil {
		return fmt.Errorf("git stash apply: %w", err)
	}
	if err := utils.ValidateGitRef(stashID, "stashID"); err != nil {
		return fmt.Errorf("git stash apply: %w", err)
	}
	cmd := fmt.Sprintf("cd %s && git stash apply %s", utils.ShellQuote(destinationPath), utils.ShellQuote(stashID))
	output, err := g.run(cmd)
	if err != nil {
		return fmt.Errorf("git stash apply failed: %s, output: %s", err.Error(), output)
	}
	g.logger.Log(logger.Info, fmt.Sprintf("github connector git: applied stash %s at %s", stashID, destinationPath), "")
	return nil
}

func (g *sshGit) ResetHard(destinationPath string) error {
	if err := utils.ValidatePath(destinationPath, "destinationPath"); err != nil {
		return fmt.Errorf("git reset: %w", err)
	}
	cmd := fmt.Sprintf("cd %s && git reset --hard", utils.ShellQuote(destinationPath))
	output, err := g.run(cmd)
	if err != nil {
		return fmt.Errorf("git reset --hard failed: %s, output: %s", err.Error(), output)
	}
	g.logger.Log(logger.Info, fmt.Sprintf("github connector git: reset hard at %s", destinationPath), "")
	return nil
}

func (g *sshGit) RemoveRepository(repoPath string) error {
	if err := utils.ValidatePath(repoPath, "repoPath"); err != nil {
		return fmt.Errorf("remove repository: %w", err)
	}
	cmd := fmt.Sprintf("rm -rf %s", utils.ShellQuote(repoPath))
	output, err := g.run(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove repository directory: %s, output: %s", err.Error(), output)
	}
	return nil
}
