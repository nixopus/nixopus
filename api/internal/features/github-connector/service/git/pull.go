package git

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// HandlePullWithClient resets local changes if needed then pulls.
func HandlePullWithClient(log logger.Logger, client Git, authenticatedURL, clonePath, userID string) error {
	hasChanges, err := client.HasUncommittedChanges(clonePath)
	if err != nil {
		log.Log(logger.Error, fmt.Sprintf("github connector git: uncommitted changes check: %s", err.Error()), userID)
		return err
	}
	if hasChanges {
		log.Log(logger.Info, "github connector git: discarding local changes for clean pull", userID)
		if err := client.ResetHard(clonePath); err != nil {
			log.Log(logger.Error, fmt.Sprintf("github connector git: reset hard: %s", err.Error()), userID)
			return err
		}
	}
	log.Log(logger.Info, "github connector git: pulling latest", userID)
	if err := client.Pull(authenticatedURL, clonePath); err != nil {
		log.Log(logger.Error, fmt.Sprintf("github connector git: pull: %s", err.Error()), userID)
		return err
	}
	return nil
}
