package live

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/dockerfile_generator"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

type BuildFirstManager struct {
	stagingManager     *StagingManager
	taskService        *tasks.TaskService
	injector           *FileInjector
	logger             logger.Logger
	sessionEnvResolver SessionEnvResolver

	deployedMu sync.RWMutex
	deployed   map[uuid.UUID]*deployedApp

	buildDebounceMu sync.Mutex
	buildTimers     map[uuid.UUID]*time.Timer

	buildingMu sync.Mutex
	building   map[uuid.UUID]bool
}

type deployedApp struct {
	workdir string
}

// SessionEnvResolver returns session env vars (from client's set-env file) for an application.
// Session env overrides DB env when building. Nil = no session env.
type SessionEnvResolver func(applicationID uuid.UUID) map[string]string

func NewBuildFirstManager(stagingManager *StagingManager, taskService *tasks.TaskService, logger logger.Logger, sessionEnvResolver SessionEnvResolver) *BuildFirstManager {
	return &BuildFirstManager{
		stagingManager:     stagingManager,
		taskService:        taskService,
		injector:           NewFileInjector(logger),
		logger:             logger,
		deployed:           make(map[uuid.UUID]*deployedApp),
		buildTimers:        make(map[uuid.UUID]*time.Timer),
		building:           make(map[uuid.UUID]bool),
		sessionEnvResolver: sessionEnvResolver,
	}
}

func (bfm *BuildFirstManager) IsDeployed(appID uuid.UUID) bool {
	bfm.deployedMu.RLock()
	defer bfm.deployedMu.RUnlock()
	_, ok := bfm.deployed[appID]
	return ok
}

func (bfm *BuildFirstManager) GetWorkdir(appID uuid.UUID) string {
	bfm.deployedMu.RLock()
	defer bfm.deployedMu.RUnlock()
	if app, ok := bfm.deployed[appID]; ok {
		return app.workdir
	}
	return ""
}

func (bfm *BuildFirstManager) HandleFileWritten(ctx context.Context, appCtx *ApplicationContext, filePath string, content []byte) {
	if !bfm.IsDeployed(appCtx.ApplicationID) {
		if service, err := tasks.FindServiceByLabel(ctx, "com.application.id", appCtx.ApplicationID.String()); err == nil && service != nil {
			if bfm.injector.IsContainerRunning(ctx, appCtx) {
				workdir := tasks.GetWorkdirFromService(service)
				bfm.MarkDeployed(appCtx.ApplicationID, workdir)
				bfm.logger.Log(logger.Info, "recovered deployed state from running container", appCtx.ApplicationID.String())
			} else {
				bfm.triggerBuild(ctx, appCtx)
				return
			}
		} else {
			bfm.triggerBuild(ctx, appCtx)
			return
		}
	}

	if IsDependencyFile(filePath) {
		bfm.logger.Log(logger.Info, "dependency file changed, triggering rebuild", fmt.Sprintf("path=%s app=%s", filePath, appCtx.ApplicationID))
		bfm.markUndeployed(appCtx.ApplicationID)
		bfm.triggerBuild(ctx, appCtx)
		return
	}

	// .env is excluded from sync; values are sent via env_vars message from client
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		lastErr = bfm.injector.InjectFile(ctx, appCtx, filePath, content, bfm.GetWorkdir(appCtx.ApplicationID))
		if lastErr == nil {
			return
		}
	}
	bfm.logger.Log(logger.Warning, "file injection failed after retries, triggering rebuild", fmt.Sprintf("path=%s err=%v", filePath, lastErr))
	bfm.markUndeployed(appCtx.ApplicationID)
	bfm.triggerBuild(ctx, appCtx)
}

func (bfm *BuildFirstManager) HandleFileDeleted(ctx context.Context, appCtx *ApplicationContext, filePath string) {
	if !bfm.IsDeployed(appCtx.ApplicationID) {
		return
	}

	if IsDependencyFile(filePath) {
		bfm.logger.Log(logger.Info, "dependency file deleted, triggering rebuild", fmt.Sprintf("path=%s", filePath))
		bfm.markUndeployed(appCtx.ApplicationID)
		bfm.triggerBuild(ctx, appCtx)
		return
	}

	if err := bfm.injector.DeleteFile(ctx, appCtx, filePath, bfm.GetWorkdir(appCtx.ApplicationID)); err != nil {
		bfm.logger.Log(logger.Warning, "file delete from container failed", fmt.Sprintf("path=%s err=%v", filePath, err))
	}
}

func (bfm *BuildFirstManager) triggerBuild(ctx context.Context, appCtx *ApplicationContext) {
	bfm.buildDebounceMu.Lock()
	if existing, ok := bfm.buildTimers[appCtx.ApplicationID]; ok {
		existing.Stop()
	}
	ctxCopy := context.WithoutCancel(ctx)
	bfm.buildTimers[appCtx.ApplicationID] = time.AfterFunc(3*time.Second, func() {
		bfm.doBuild(ctxCopy, appCtx)
	})
	bfm.buildDebounceMu.Unlock()
}

func (bfm *BuildFirstManager) doBuild(ctx context.Context, appCtx *ApplicationContext) {
	bfm.buildDebounceMu.Lock()
	delete(bfm.buildTimers, appCtx.ApplicationID)
	bfm.buildDebounceMu.Unlock()

	bfm.buildingMu.Lock()
	if bfm.building[appCtx.ApplicationID] {
		bfm.buildingMu.Unlock()
		return
	}
	bfm.building[appCtx.ApplicationID] = true
	bfm.buildingMu.Unlock()

	defer func() {
		bfm.buildingMu.Lock()
		delete(bfm.building, appCtx.ApplicationID)
		bfm.buildingMu.Unlock()
	}()

	bfm.logger.Log(logger.Info, "starting build-first deployment", appCtx.ApplicationID.String())

	source := appCtx.StagingPath
	auth := dockerfile_generator.AuthHeaders{
		Token:  appCtx.AuthToken,
		Cookie: appCtx.AuthCookie,
	}
	client := dockerfile_generator.NewClient(config.AppConfig.Agent.Endpoint)
	resp, err := client.GenerateWithMode(ctx, source, appCtx.OrganizationID.String(), "development", auth)
	if err != nil {
		bfm.logger.Log(logger.Error, "dockerfile generator failed", fmt.Sprintf("app=%s err=%v", appCtx.ApplicationID, err))
		return
	}

	dockerfilePath, err := WriteDockerfileToStaging(ctx, appCtx.StagingPath, resp.Dockerfile)
	if err != nil {
		bfm.logger.Log(logger.Error, "failed to write dockerfile", fmt.Sprintf("app=%s err=%v", appCtx.ApplicationID, err))
		return
	}

	if resp.Dockerignore != "" {
		if _, err := WriteDockerignoreToStaging(ctx, appCtx.StagingPath, resp.Dockerignore); err != nil {
			bfm.logger.Log(logger.Warning, "failed to write .dockerignore", fmt.Sprintf("app=%s err=%v", appCtx.ApplicationID, err))
		}
	}

	var sessionEnv map[string]string
	if bfm.sessionEnvResolver != nil {
		sessionEnv = bfm.sessionEnvResolver(appCtx.ApplicationID)
	}
	envVars := mergeEnvVars(appCtx.EnvironmentVariables, sessionEnv)
	cfg := tasks.LiveDevConfig{
		ApplicationID:  appCtx.ApplicationID,
		OrganizationID: appCtx.OrganizationID,
		StagingPath:    appCtx.StagingPath,
		EnvVars:        envVars,
		DockerfilePath: dockerfilePath,
		InternalPort:   resp.Port,
		Workdir:        resp.Workdir,
	}
	if appCtx.Domain != "" {
		cfg.Domain = appCtx.Domain
	}

	if err := bfm.taskService.StartLiveDevTask(ctx, cfg); err != nil {
		bfm.logger.Log(logger.Error, "build-first deployment failed", fmt.Sprintf("app=%s err=%v", appCtx.ApplicationID, err))
		return
	}

	// Deployed state is set by TaskService.OnLiveDevDeployed when the container is healthy,
	// not here. Previously marking immediately caused injection to fail (container not ready).
	bfm.logger.Log(logger.Info, "build-first deployment queued", appCtx.ApplicationID.String())
}

func (bfm *BuildFirstManager) markUndeployed(appID uuid.UUID) {
	bfm.deployedMu.Lock()
	delete(bfm.deployed, appID)
	bfm.deployedMu.Unlock()
}

// MarkDeployed is called when a live dev build completes and the container is healthy.
// Enables file injection instead of full rebuilds. Used by TaskService.OnLiveDevDeployed callback.
func (bfm *BuildFirstManager) MarkDeployed(appID uuid.UUID, workdir string) {
	if workdir == "" {
		workdir = "/app"
	}
	bfm.deployedMu.Lock()
	bfm.deployed[appID] = &deployedApp{workdir: workdir}
	bfm.deployedMu.Unlock()
	bfm.logger.Log(logger.Info, "live dev container ready, inject mode enabled", appID.String())
}

const generatedDockerfileName = "Dockerfile.nixopus.dev"

func WriteDockerfileToStaging(ctx context.Context, stagingPath string, content string) (string, error) {
	fullPath := filepath.Join(stagingPath, generatedDockerfileName)

	if isLocalStagingPath(stagingPath) {
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("failed to write Dockerfile: %w", err)
		}
		return generatedDockerfileName, nil
	}

	sftpClient, err := getSFTPClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}
	defer sftpClient.Close()

	file, err := sftpClient.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create Dockerfile: %w", err)
	}
	defer file.Close()

	if _, err := file.Write([]byte(content)); err != nil {
		return "", fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	return generatedDockerfileName, nil
}

func WriteDockerignoreToStaging(ctx context.Context, stagingPath string, content string) (string, error) {
	fullPath := filepath.Join(stagingPath, ".dockerignore")

	if isLocalStagingPath(stagingPath) {
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("failed to write .dockerignore: %w", err)
		}
		return ".dockerignore", nil
	}

	sftpClient, err := getSFTPClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}
	defer sftpClient.Close()

	file, err := sftpClient.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create .dockerignore: %w", err)
	}
	defer file.Close()

	if _, err := file.Write([]byte(content)); err != nil {
		return "", fmt.Errorf("failed to write .dockerignore: %w", err)
	}

	return ".dockerignore", nil
}

var dependencyFiles = []string{
	"package.json",
	"package-lock.json",
	"yarn.lock",
	"pnpm-lock.yaml",
	"bun.lockb",
	".npmrc",
	".yarnrc",
	".yarnrc.yml",
	"requirements.txt",
	"requirements-dev.txt",
	"pyproject.toml",
	"setup.py",
	"setup.cfg",
	"Pipfile",
	"Pipfile.lock",
	"poetry.lock",
	"uv.lock",
	"Dockerfile",
	"Dockerfile.dev",
	"dockerfile",
	generatedDockerfileName,
	".dockerignore",
}

func IsDependencyFile(filePath string) bool {
	baseName := filepath.Base(filePath)
	for _, df := range dependencyFiles {
		if baseName == df {
			return true
		}
	}
	return false
}

// mergeEnvVars merges base and override; override takes precedence.
func mergeEnvVars(base, override map[string]string) map[string]string {
	if len(override) == 0 {
		return base
	}
	out := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}
