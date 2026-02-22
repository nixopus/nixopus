package addcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// getGitBranch gets the current git branch
func getGitBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "main"
	}
	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "main"
	}
	return branch
}

// getGitInfo gets the git repository remote URL
func getGitInfo() (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	repoName := filepath.Base(cwd)

	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return repoName, repoName, nil
	}

	repoURL := strings.TrimSpace(string(output))
	if repoURL == "" {
		return repoName, repoName, nil
	}

	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		lastPart := strings.TrimSuffix(parts[len(parts)-1], ".git")
		if lastPart != "" {
			repoName = lastPart
		}
	}

	return repoName, repoURL, nil
}

// normalizeBasePath normalizes the base path
// Removes leading "./" and ensures it's a clean relative path
func normalizeBasePath(path string) string {
	// Remove leading "./"
	path = strings.TrimPrefix(path, "./")

	// Clean the path
	path = filepath.Clean(path)

	// If it's ".", return "/" (root)
	if path == "." || path == "" {
		return "/"
	}

	// Ensure it doesn't start with "/" (should be relative)
	path = strings.TrimPrefix(path, "/")

	return path
}
