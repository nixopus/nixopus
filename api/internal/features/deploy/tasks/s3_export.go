package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	s3store "github.com/nixopus/nixopus/api/internal/features/deploy/s3"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	sshpkg "github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type ExportConfig struct {
	ImageTag     string
	ImageTags    []string
	OrgID        uuid.UUID
	AppID        uuid.UUID
	DeploymentID uuid.UUID
}

// ExportImageToS3 runs `docker save <tag(s)> | gzip` on the remote server via SSH
// and streams the output directly to S3 as a multipart upload.
// Supports single image (ImageTag) or multiple images (ImageTags) for compose.
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

	clientConn, release, err := sshManager.Borrow("")
	if err != nil {
		return "", 0, fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer release()

	session, err := clientConn.NewSession()
	if err != nil {
		if sshpkg.IsClosedConnectionError(err) {
			sshManager.CloseConnection("")
		}
		return "", 0, fmt.Errorf("failed to create SSH session: %w", err)
	}

	var saveArgs string
	if len(cfg.ImageTags) > 0 {
		quoted := make([]string, len(cfg.ImageTags))
		for i, tag := range cfg.ImageTags {
			quoted[i] = utils.ShellQuote(tag)
		}
		saveArgs = strings.Join(quoted, " ")
	} else {
		saveArgs = utils.ShellQuote(cfg.ImageTag)
	}
	cmd := fmt.Sprintf("docker save %s | gzip", saveArgs)

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return "", 0, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := session.Start(cmd); err != nil {
		session.Close()
		return "", 0, fmt.Errorf("failed to start docker save: %w", err)
	}

	key := s3store.ImageS3Key(cfg.OrgID, cfg.AppID, cfg.DeploymentID)
	taskCtx.AddLog("Uploading image to S3: " + key)

	size, err := store.UploadImage(ctx, key, stdout)

	waitErr := session.Wait()
	session.Close()

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

	orgSettings, err := utils.GetOrganizationSettings(ctx, s.Store.DB, payload.Application.OrganizationID)
	if err != nil {
		s.Logger.Log(logger.Warning, "Failed to get organization settings for S3 export: "+err.Error(), payload.ApplicationDeployment.ID.String())
		return
	}
	if orgSettings.S3ArtifactUploadEnabled == nil || !*orgSettings.S3ArtifactUploadEnabled {
		return
	}

	deploymentCopy := payload.ApplicationDeployment
	go func() {
		bgCtx := context.Background()
		bgCtx = context.WithValue(bgCtx, shared_types.OrganizationIDKey, payload.Application.OrganizationID.String())
		if serverID, ok := ctx.Value(shared_types.ServerIDKey).(string); ok {
			bgCtx = context.WithValue(bgCtx, shared_types.ServerIDKey, serverID)
		}

		s.Logger.Log(logger.Info, "Starting async S3 image export", deploymentCopy.ID.String())
		key, size, err := s.ExportImageToS3(bgCtx, ExportConfig{
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
		taskCtx.FlushLogs()
	}()
}

// ExportComposeImagesToS3 lists images used by a compose project and exports them
// to S3 as a single tarball. Runs asynchronously and is non-fatal.
func (s *TaskService) ExportComposeImagesToS3(ctx context.Context, payload shared_types.TaskPayload, composeFilePath string, envVars map[string]string, taskCtx *TaskContext) {
	if !s3store.IsConfigured(config.AppConfig.S3) {
		return
	}

	orgSettings, err := utils.GetOrganizationSettings(ctx, s.Store.DB, payload.Application.OrganizationID)
	if err != nil {
		s.Logger.Log(logger.Warning, "Failed to get organization settings for S3 export: "+err.Error(), payload.ApplicationDeployment.ID.String())
		return
	}
	if orgSettings.S3ArtifactUploadEnabled == nil || !*orgSettings.S3ArtifactUploadEnabled {
		return
	}

	deploymentCopy := payload.ApplicationDeployment
	// List images synchronously while the SSH context is still alive,
	// then run the S3 upload in a background goroutine with a detached context.
	imageTags, err := s.listComposeImages(ctx, composeFilePath, envVars)
	if err != nil {
		s.Logger.Log(logger.Warning, "Failed to list compose images (non-fatal): "+err.Error(), deploymentCopy.ID.String())
		return
	}
	if len(imageTags) == 0 {
		s.Logger.Log(logger.Warning, "No compose images found to export", deploymentCopy.ID.String())
		return
	}

	s.Logger.Log(logger.Info, fmt.Sprintf("Found %d compose image(s) to export: %s", len(imageTags), strings.Join(imageTags, ", ")), deploymentCopy.ID.String())

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.Logger.Log(logger.Error, fmt.Sprintf("panic in compose S3 export: %v", r), deploymentCopy.ID.String())
			}
		}()
		bgCtx := context.Background()
		bgCtx = context.WithValue(bgCtx, shared_types.OrganizationIDKey, payload.Application.OrganizationID.String())
		if serverID, ok := ctx.Value(shared_types.ServerIDKey).(string); ok {
			bgCtx = context.WithValue(bgCtx, shared_types.ServerIDKey, serverID)
		}

		taskCtx.AddLog(fmt.Sprintf("Exporting %d compose image(s) to S3: %s", len(imageTags), strings.Join(imageTags, ", ")))

		key, size, err := s.ExportImageToS3(bgCtx, ExportConfig{
			ImageTags:    imageTags,
			OrgID:        payload.Application.OrganizationID,
			AppID:        payload.Application.ID,
			DeploymentID: deploymentCopy.ID,
		}, taskCtx)
		if err != nil {
			s.Logger.Log(logger.Warning, "Failed to export compose images to S3 (non-fatal): "+err.Error(), deploymentCopy.ID.String())
			return
		}

		deploymentCopy.ImageS3Key = key
		deploymentCopy.ImageSize = size
		if err := s.Storage.UpdateApplicationDeployment(&deploymentCopy); err != nil {
			s.Logger.Log(logger.Warning, "Failed to record S3 image metadata: "+err.Error(), deploymentCopy.ID.String())
		}
		taskCtx.FlushLogs()
	}()
}

// listComposeImages runs `docker compose images` via SSH and returns the image tags.
func (s *TaskService) listComposeImages(ctx context.Context, composeFilePath string, envVars map[string]string) ([]string, error) {
	sshManager, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH manager: %w", err)
	}

	var envPrefix string
	if len(envVars) > 0 {
		validKeyRegex := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
		var parts []string
		for k, v := range envVars {
			if !validKeyRegex.MatchString(k) {
				return nil, fmt.Errorf("invalid environment variable key: %s", k)
			}
			parts = append(parts, fmt.Sprintf("export %s=%s", k, utils.ShellQuote(v)))
		}
		envPrefix = strings.Join(parts, " && ") + " && "
	}

	cmd := fmt.Sprintf("%sdocker compose -f %s images --format json 2>&1",
		envPrefix, utils.ShellQuote(composeFilePath))

	output, err := sshManager.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("docker compose images failed (output: %s): %w", output, err)
	}

	return parseComposeImagesOutput(output)
}

type composeImageEntry struct {
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
}

func parseComposeImagesOutput(output string) ([]string, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	// Try JSON array first
	var entries []composeImageEntry
	if err := json.Unmarshal([]byte(output), &entries); err == nil {
		return extractImageTags(entries), nil
	}

	// Try JSON-lines (one JSON object per line)
	var tags []string
	seen := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry composeImageEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		tag := entry.Repository + ":" + entry.Tag
		if tag != ":" && !seen[tag] {
			seen[tag] = true
			tags = append(tags, tag)
		}
	}
	return tags, nil
}

func extractImageTags(entries []composeImageEntry) []string {
	var tags []string
	seen := make(map[string]bool)
	for _, e := range entries {
		tag := e.Repository + ":" + e.Tag
		if tag != ":" && !seen[tag] {
			seen[tag] = true
			tags = append(tags, tag)
		}
	}
	return tags
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

	clientConn, release, err := sshManager.Borrow("")
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer release()

	session, err := clientConn.NewSession()
	if err != nil {
		if sshpkg.IsClosedConnectionError(err) {
			sshManager.CloseConnection("")
		}
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	session.Stdin = body

	taskCtx.AddLog("Loading image from S3 into Docker...")
	output, err := session.CombinedOutput("gunzip | docker load")
	if err != nil {
		return fmt.Errorf("docker load failed: %w (output: %s)", err, string(output))
	}

	taskCtx.AddLog("Image loaded from S3: " + string(output))
	return nil
}
