package tasks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	s3store "github.com/raghavyuva/nixopus-api/internal/features/deploy/s3"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	sshpkg "github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type ExportConfig struct {
	ImageTag     string
	OrgID        uuid.UUID
	AppID        uuid.UUID
	DeploymentID uuid.UUID
}

// ExportImageToS3 runs `docker save <tag> | gzip` on the remote server via SSH
// and streams the output directly to S3 as a multipart upload.
// Returns the S3 key and uploaded size in bytes.
func (s *TaskService) ExportImageToS3(ctx context.Context, cfg ExportConfig, taskCtx *TaskContext) (string, int64, error) {
	store, err := s3store.NewImageStore(config.AppConfig.S3)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create S3 image store: %w", err)
	}

	sshManager, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get SSH manager: %w", err)
	}

	clientConn, err := sshManager.Connect()
	if err != nil {
		return "", 0, fmt.Errorf("failed to connect via SSH: %w", err)
	}

	session, err := clientConn.NewSession()
	if err != nil {
		clientConn.Close()
		if sshpkg.IsClosedConnectionError(err) {
			sshManager.CloseConnection("")
		}
		return "", 0, fmt.Errorf("failed to create SSH session: %w", err)
	}

	escape := func(x string) string { return "'" + fmt.Sprintf("%s", x) + "'" }
	cmd := fmt.Sprintf("docker save %s | gzip", escape(cfg.ImageTag))

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		clientConn.Close()
		return "", 0, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := session.Start(cmd); err != nil {
		session.Close()
		clientConn.Close()
		return "", 0, fmt.Errorf("failed to start docker save: %w", err)
	}

	key := s3store.ImageS3Key(cfg.OrgID, cfg.AppID, cfg.DeploymentID)
	taskCtx.AddLog("Uploading image to S3: " + key)

	size, err := store.UploadImage(ctx, key, stdout)

	waitErr := session.Wait()
	session.Close()
	clientConn.Close()

	if err != nil {
		return "", 0, fmt.Errorf("failed to upload image to S3: %w", err)
	}
	if waitErr != nil {
		return "", 0, fmt.Errorf("docker save command failed: %w", waitErr)
	}

	taskCtx.AddLog(fmt.Sprintf("Image uploaded to S3 (%d bytes)", size))
	return key, size, nil
}

// ExportAndRecordImage exports the built image to S3 in the background so it
// does not block the deployment pipeline. The export is non-fatal: failures
// are logged but do not affect deployment success.
func (s *TaskService) ExportAndRecordImage(ctx context.Context, payload shared_types.TaskPayload, commitTag string, taskCtx *TaskContext) {
	if !s3store.IsConfigured(config.AppConfig.S3) {
		return
	}

	deploymentCopy := payload.ApplicationDeployment
	go func() {
		s.Logger.Log(logger.Info, "Starting async S3 image export", deploymentCopy.ID.String())
		key, size, err := s.ExportImageToS3(ctx, ExportConfig{
			ImageTag:     commitTag,
			OrgID:        payload.Application.OrganizationID,
			AppID:        payload.Application.ID,
			DeploymentID: deploymentCopy.ID,
		}, taskCtx)
		if err != nil {
			s.Logger.Log(logger.Warning, "Failed to export image to S3 (non-fatal): "+err.Error(), deploymentCopy.ID.String())
			return
		}

		deploymentCopy.ImageS3Key = key
		deploymentCopy.ImageSize = size
		if err := s.Storage.UpdateApplicationDeployment(&deploymentCopy); err != nil {
			s.Logger.Log(logger.Warning, "Failed to record S3 image metadata: "+err.Error(), deploymentCopy.ID.String())
		}
	}()
}

// LoadImageFromS3 downloads an image from S3 and loads it into Docker on the remote server.
// Runs `gunzip | docker load` via SSH with the S3 stream piped to stdin.
func (s *TaskService) LoadImageFromS3(ctx context.Context, s3Key string, taskCtx *TaskContext) error {
	store, err := s3store.NewImageStore(config.AppConfig.S3)
	if err != nil {
		return fmt.Errorf("failed to create S3 image store: %w", err)
	}

	body, err := store.DownloadImage(ctx, s3Key)
	if err != nil {
		return fmt.Errorf("failed to download image from S3: %w", err)
	}
	defer body.Close()

	sshManager, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get SSH manager: %w", err)
	}

	clientConn, err := sshManager.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}

	session, err := clientConn.NewSession()
	if err != nil {
		clientConn.Close()
		if sshpkg.IsClosedConnectionError(err) {
			sshManager.CloseConnection("")
		}
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func() {
		session.Close()
		clientConn.Close()
	}()

	session.Stdin = body

	taskCtx.AddLog("Loading image from S3 into Docker...")
	output, err := session.CombinedOutput("gunzip | docker load")
	if err != nil {
		return fmt.Errorf("docker load failed: %w (output: %s)", err, string(output))
	}

	taskCtx.AddLog("Image loaded from S3: " + string(output))
	return nil
}
