package tasks

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/config"
	s3store "github.com/raghavyuva/nixopus-api/internal/features/deploy/s3"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type SourceResolver interface {
	Resolve(ctx context.Context, config SourceResolveConfig) (string, error)
}

type SourceResolveConfig struct {
	shared_types.TaskPayload
	DeploymentType string
	TaskContext    *TaskContext
}

type GithubSourceResolver struct {
	task *TaskService
}

func (r *GithubSourceResolver) Resolve(ctx context.Context, config SourceResolveConfig) (string, error) {
	return r.task.Clone(ctx, CloneConfig{
		TaskPayload:    config.TaskPayload,
		DeploymentType: config.DeploymentType,
		TaskContext:    config.TaskContext,
	})
}

type StagingSourceResolver struct{}

func (r *StagingSourceResolver) Resolve(ctx context.Context, config SourceResolveConfig) (string, error) {
	app := config.Application
	stagingPath := filepath.Join(
		"/var/nixopus/repos",
		app.UserID.String(),
		string(app.Environment),
		app.ID.String(),
	)
	return stagingPath, nil
}

type S3SourceResolver struct {
	task *TaskService
}

func (r *S3SourceResolver) Resolve(ctx context.Context, cfg SourceResolveConfig) (string, error) {
	app := cfg.Application
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, app.OrganizationID.String())

	store, err := s3store.NewImageStore(config.AppConfig.S3)
	if err != nil {
		return "", fmt.Errorf("failed to create S3 client: %w", err)
	}

	prefix := s3store.WorkspaceS3Prefix(app.ID)
	keys, err := store.ListObjects(ctx, prefix)
	if err != nil {
		return "", fmt.Errorf("failed to list workspace files in S3: %w", err)
	}
	if len(keys) == 0 {
		return "", fmt.Errorf("no files found in S3 workspace for application %s", app.ID)
	}

	stagingPath := filepath.Join(
		"/var/nixopus/repos",
		app.UserID.String(),
		string(app.Environment),
		app.ID.String(),
	)

	err = utils.WithSFTPClientFromPool(orgCtx, func(sftpClient *sftp.Client) error {
		if err := sftpClient.MkdirAll(stagingPath); err != nil {
			return fmt.Errorf("failed to create staging directory: %w", err)
		}

		for _, key := range keys {
			relPath := strings.TrimPrefix(key, prefix)
			if relPath == "" || strings.HasSuffix(relPath, "/") {
				continue
			}

			destPath := filepath.Join(stagingPath, relPath)
			destDir := filepath.Dir(destPath)
			if err := sftpClient.MkdirAll(destDir); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destDir, err)
			}

			body, err := store.GetObject(ctx, key)
			if err != nil {
				return fmt.Errorf("failed to download %s: %w", key, err)
			}

			f, err := sftpClient.Create(destPath)
			if err != nil {
				body.Close()
				return fmt.Errorf("failed to create file %s: %w", destPath, err)
			}

			_, copyErr := io.Copy(f, body)
			f.Close()
			body.Close()
			if copyErr != nil {
				return fmt.Errorf("failed to write %s: %w", destPath, copyErr)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return stagingPath, nil
}

type ZipSourceResolver struct {
	task *TaskService
}

func (r *ZipSourceResolver) Resolve(ctx context.Context, config SourceResolveConfig) (string, error) {
	return "", fmt.Errorf("ZIP source resolver not yet implemented")
}

func (t *TaskService) GetSourceResolver(source shared_types.Source) SourceResolver {
	switch source {
	case shared_types.SourceS3:
		return &S3SourceResolver{task: t}
	case shared_types.SourceZip:
		return &ZipSourceResolver{task: t}
	case shared_types.SourceStaging:
		return &StagingSourceResolver{}
	default:
		return &GithubSourceResolver{task: t}
	}
}
