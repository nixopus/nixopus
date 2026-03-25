package tasks

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nixopus/nixopus/api/internal/config"
	s3store "github.com/nixopus/nixopus/api/internal/features/deploy/s3"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/pkg/sftp"
	"golang.org/x/sync/errgroup"
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

type ObjectStore interface {
	ListObjects(ctx context.Context, prefix string) ([]string, error)
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
}

type S3SourceResolver struct {
	task  *TaskService
	store ObjectStore
}

const (
	s3DownloadWorkers = 10
	copyBufSize       = 256 * 1024 // 256 KB
)

var copyBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, copyBufSize)
		return &b
	},
}

func (r *S3SourceResolver) Resolve(ctx context.Context, cfg SourceResolveConfig) (string, error) {
	app := cfg.Application
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, app.OrganizationID.String())

	store := r.store
	if store == nil {
		s, err := s3store.NewImageStore(config.AppConfig.S3)
		if err != nil {
			return "", fmt.Errorf("failed to create S3 client: %w", err)
		}
		store = s
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

		type fileEntry struct {
			key     string
			relPath string
		}

		var files []fileEntry
		for _, key := range keys {
			relPath := strings.TrimPrefix(key, prefix)
			if relPath == "" || strings.HasSuffix(relPath, "/") {
				continue
			}
			files = append(files, fileEntry{key: key, relPath: relPath})
		}

		dirs := make(map[string]struct{})
		for _, f := range files {
			destDir := filepath.Dir(filepath.Join(stagingPath, f.relPath))
			if _, ok := dirs[destDir]; ok {
				continue
			}
			if err := sftpClient.MkdirAll(destDir); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destDir, err)
			}
			dirs[destDir] = struct{}{}
		}

		g, gCtx := errgroup.WithContext(ctx)
		g.SetLimit(s3DownloadWorkers)

		for _, fe := range files {
			fe := fe
			g.Go(func() error {
				return downloadFile(gCtx, store, sftpClient, fe.key, filepath.Join(stagingPath, fe.relPath))
			})
		}
		return g.Wait()
	})
	if err != nil {
		return "", err
	}

	return stagingPath, nil
}

func downloadFile(ctx context.Context, store ObjectStore, sftpClient *sftp.Client, s3Key, destPath string) error {
	body, err := store.GetObject(ctx, s3Key)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", s3Key, err)
	}
	defer body.Close()

	f, err := sftpClient.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer f.Close()

	bufp := copyBufPool.Get().(*[]byte)
	defer copyBufPool.Put(bufp)

	if _, err := io.CopyBuffer(f, body, *bufp); err != nil {
		return fmt.Errorf("failed to write %s: %w", destPath, err)
	}
	return nil
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
