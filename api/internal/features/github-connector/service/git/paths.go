package git

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/pkg/sftp"
)

// ResolveClonePath computes the tenant clone path via SFTP and ensures the directory exists.
func ResolveClonePath(ctx context.Context, userID, environment, applicationID string) (string, bool, error) {
	repoBaseURL := "/var/nixopus/repos"
	clonePath := filepath.Join(repoBaseURL, userID, environment, applicationID)
	var shouldPull bool
	err := utils.WithSFTPClientFromPool(ctx, func(sftpClient *sftp.Client) error {
		info, statErr := sftpClient.Stat(clonePath)
		if statErr == nil && info.IsDir() {
			shouldPull = true
		}
		if !shouldPull {
			if mkErr := sftpClient.MkdirAll(clonePath); mkErr != nil {
				return fmt.Errorf("failed to create directory via SFTP: %w", mkErr)
			}
		}
		return nil
	})
	if err != nil {
		return "", false, err
	}
	return clonePath, shouldPull, nil
}
