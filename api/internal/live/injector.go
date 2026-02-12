package live

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

type FileInjector struct {
	logger logger.Logger
}

func NewFileInjector(logger logger.Logger) *FileInjector {
	return &FileInjector{logger: logger}
}

func (fi *FileInjector) InjectFile(ctx context.Context, appCtx *ApplicationContext, relativePath string, content []byte, workdir string) error {
	containerID, err := fi.getContainerID(ctx, appCtx)
	if err != nil {
		return fmt.Errorf("failed to find running container: %w", err)
	}

	dockerSvc, err := docker.GetDockerServiceDirect(ctx)
	if err != nil {
		return fmt.Errorf("failed to get docker service: %w", err)
	}

	if workdir == "" {
		workdir = "/app"
	}
	targetPath := filepath.Join(workdir, relativePath)
	targetDir := filepath.Dir(targetPath)

	mkdirCmd := []string{"sh", "-c", fmt.Sprintf("mkdir -p %s", targetDir)}
	if err := dockerSvc.ExecInContainer(containerID, mkdirCmd, nil); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	writeCmd := []string{"sh", "-c", fmt.Sprintf("cat > %s", targetPath)}
	reader := bytes.NewReader(content)
	if err := dockerSvc.ExecInContainer(containerID, writeCmd, reader); err != nil {
		return fmt.Errorf("failed to inject file %s: %w", relativePath, err)
	}

	fi.logger.Log(logger.Info, "file injected into container", fmt.Sprintf("path=%s container=%s", relativePath, containerID[:12]))
	return nil
}

func (fi *FileInjector) DeleteFile(ctx context.Context, appCtx *ApplicationContext, relativePath string, workdir string) error {
	containerID, err := fi.getContainerID(ctx, appCtx)
	if err != nil {
		return fmt.Errorf("failed to find running container: %w", err)
	}

	dockerSvc, err := docker.GetDockerServiceDirect(ctx)
	if err != nil {
		return fmt.Errorf("failed to get docker service: %w", err)
	}

	if workdir == "" {
		workdir = "/app"
	}
	targetPath := filepath.Join(workdir, relativePath)

	rmCmd := []string{"sh", "-c", fmt.Sprintf("rm -f %s", targetPath)}
	if err := dockerSvc.ExecInContainer(containerID, rmCmd, nil); err != nil {
		fi.logger.Log(logger.Warning, "failed to delete file from container", fmt.Sprintf("path=%s err=%v", relativePath, err))
		return nil
	}

	fi.logger.Log(logger.Info, "file deleted from container", fmt.Sprintf("path=%s container=%s", relativePath, containerID[:12]))
	return nil
}

func (fi *FileInjector) getContainerID(ctx context.Context, appCtx *ApplicationContext) (string, error) {
	service, err := tasks.FindServiceByLabel(ctx, "com.application.id", appCtx.ApplicationID.String())
	if err != nil {
		return "", fmt.Errorf("failed to find service: %w", err)
	}
	if service == nil {
		return "", fmt.Errorf("no service found for application %s", appCtx.ApplicationID)
	}

	dockerSvc, err := docker.GetDockerServiceDirect(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get docker service: %w", err)
	}

	return dockerSvc.GetRunningTaskContainerID(service.ID)
}

func (fi *FileInjector) IsContainerRunning(ctx context.Context, appCtx *ApplicationContext) bool {
	_, err := fi.getContainerID(ctx, appCtx)
	return err == nil
}
