package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/queue"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	webhookDedupPrefix = "webhook:dedup:"
	webhookDedupTTL    = 30 * time.Second
)

// webhookDedupKey returns a Redis key scoped to app + commit to prevent
// the same push from triggering duplicate builds.
func webhookDedupKey(appID, commitHash string) string {
	return webhookDedupPrefix + appID + ":" + commitHash
}

func (t *TaskService) EnqueueWebhookTask(payload shared_types.WebhookPayload) error {
	parts := strings.Split(payload.Repository.FullName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository name format")
	}
	repositoryID := payload.Repository.ID

	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	applications, err := t.Storage.GetApplicationByRepositoryIDAndBranch(repositoryID, branch)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}

	if len(applications) == 0 {
		return fmt.Errorf("application not found")
	}

	commitHash := payload.After

	for _, application := range applications {
		if application.Branch != branch {
			continue
		}

		if commitHash != "" {
			if isDup, _ := t.isWebhookDuplicate(application.ID.String(), commitHash); isDup {
				t.Logger.Log(logger.Info, "skipping duplicate webhook for app "+application.Name+" commit "+commitHash, "")
				continue
			}
		}

		deployment := &types.UpdateDeploymentRequest{
			ID:                   application.ID,
			Force:                true,
			PreRunCommand:        application.PreRunCommand,
			PostRunCommand:       application.PostRunCommand,
			BuildVariables:       GetMapFromString(application.BuildVariables),
			EnvironmentVariables: GetMapFromString(application.EnvironmentVariables),
			Port:                 application.Port,
			DockerfilePath:       application.DockerfilePath,
			BasePath:             application.BasePath,
		}

		_, err := t.UpdateDeploymentWithTrigger(deployment, application.UserID, application.OrganizationID)
		if err != nil {
			t.Logger.Log(logger.Error, "failed to update deployment for webhook", err.Error())
			continue
		}

		t.Logger.Log(logger.Info, types.LogDeploymentStarted, "")
	}

	return nil
}

// isWebhookDuplicate uses Redis SET NX with a TTL to atomically check and
// claim a dedup slot. Returns true if this app+commit was already processed.
func (t *TaskService) isWebhookDuplicate(appID, commitHash string) (bool, error) {
	rc := queue.RedisClient()
	if rc == nil {
		return false, nil
	}
	key := webhookDedupKey(appID, commitHash)
	set, err := rc.SetNX(context.Background(), key, "1", webhookDedupTTL).Result()
	if err != nil {
		return false, err
	}
	// SetNX returns true if the key was set (first time), false if it already existed
	return !set, nil
}
