package git

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// HandlePullWithClient resets local changes if needed then pulls.
func HandlePullWithClient(log logger.Logger, client Git, authenticatedURL, clonePath, userID string) error {
	hasChanges, err := client.HasUncommittedChanges(clonePath)
	if err != nil {
		log.Log(logger.Error, fmt.Sprintf("Failed to check for uncommitted changes: %s", err.Error()), userID)
		return err
	}
	if hasChanges {
		log.Log(logger.Info, "Discarding local changes for clean state", userID)
		if err := client.ResetHard(clonePath); err != nil {
			log.Log(logger.Error, fmt.Sprintf("Failed to reset repository: %s", err.Error()), userID)
			return err
		}
	}
	log.Log(logger.Info, "Pulling latest changes", userID)
	if err := client.Pull(authenticatedURL, clonePath); err != nil {
		log.Log(logger.Error, fmt.Sprintf("Failed to pull repository: %s", err.Error()), userID)
		return err
	}
	return nil
}
