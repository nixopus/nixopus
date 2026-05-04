package git

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/pkg/sftp"
)

// ResolveClonePath computes the tenant clone path via SFTP and ensures the directory exists.
func ResolveClonePath(ctx context.Context, l logger.Logger, userID, environment, applicationID string) (string, bool, error) {
	ctxStr := fmt.Sprintf("user_id=%s env=%s application_id=%s", userID, environment, applicationID)
	l.Log(logger.Debug, "github connector git: ResolveClonePath", ctxStr)

	repoBaseURL := "/var/nixopus/repos"
	clonePath := filepath.Join(repoBaseURL, userID, environment, applicationID)
	var shouldPull bool
	err := utils.WithSFTPClientFromPool(ctx, func(sftpClient *sftp.Client) error {
		info, statErr := sftpClient.Stat(clonePath)
		if statErr == nil && info.IsDir() {
			shouldPull = true
			l.Log(logger.Debug, "github connector git: ResolveClonePath existing repo dir", fmt.Sprintf("%s path=%s", ctxStr, clonePath))
		}
		if !shouldPull {
			if mkErr := sftpClient.MkdirAll(clonePath); mkErr != nil {
				l.Log(logger.Error, fmt.Sprintf("github connector git: ResolveClonePath MkdirAll: %v", mkErr), fmt.Sprintf("%s path=%s", ctxStr, clonePath))
				return fmt.Errorf("failed to create directory via SFTP: %w", mkErr)
			}
			l.Log(logger.Debug, "github connector git: ResolveClonePath created dirs", fmt.Sprintf("%s path=%s", ctxStr, clonePath))
		}
		return nil
	})
	if err != nil {
		l.Log(logger.Error, fmt.Sprintf("github connector git: ResolveClonePath SFTP pool: %v", err), ctxStr)
		return "", false, err
	}
	l.Log(logger.Debug, "github connector git: ResolveClonePath ok", fmt.Sprintf("%s path=%s should_pull=%t", ctxStr, clonePath, shouldPull))
	return clonePath, shouldPull, nil
}
