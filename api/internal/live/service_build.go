package live

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// BuildStatusFunc is called at key points during the build lifecycle.
// phase is one of: "starting", "generating_dockerfile", "dockerfile_ready", "building_container", "error".
type BuildStatusFunc func(applicationID uuid.UUID, phase, message, errMsg string)

type BuildFirstManager struct {
	stagingManager       *StagingManager
	taskService          *tasks.TaskService
	injector             *FileInjector
	logger               logger.Logger
	sessionEnvResolver   SessionEnvResolver
	pipelineProgressFunc PipelineProgressFunc
	buildStatusFunc      BuildStatusFunc
	codebaseIndexedFunc  CodebaseIndexedFunc

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

// PipelineProgressFunc is called for each progress event during Dockerfile generation.
// applicationID identifies which app the progress is for. stageId and message describe the event.
type PipelineProgressFunc func(applicationID uuid.UUID, stageId, message string)

// CodebaseIndexedFunc is called when the CLI should run the deployment workflow (e.g. after sync_complete or dependency change).
// The gateway sends codebase_indexed to the client.
type CodebaseIndexedFunc func(appCtx *ApplicationContext)

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
		// When not deployed: only recover or resume. Build is triggered by sync_complete, not by individual files.
		if service, err := tasks.FindServiceByLabel(ctx, "com.application.id", appCtx.ApplicationID.String()); err == nil && service != nil {
			if bfm.injector.IsContainerRunning(ctx, appCtx) {
				workdir := tasks.GetWorkdirFromService(service)
				bfm.MarkDeployed(appCtx.ApplicationID, workdir)
				bfm.logger.Log(logger.Info, "recovered deployed state from running container", appCtx.ApplicationID.String())
			} else if tasks.IsLiveDevServicePaused(service) {
				workdir, err := bfm.taskService.ResumeLiveDevService(ctx, appCtx.ApplicationID, nil)
				if err != nil {
					bfm.logger.Log(logger.Warning, "resume failed, waiting for sync_complete to build", err.Error())
					return
				}
				bfm.MarkDeployed(appCtx.ApplicationID, workdir)
				bfm.logger.Log(logger.Info, "resumed paused live dev service", appCtx.ApplicationID.String())
			}
			// else: service exists but not running/paused; wait for sync_complete
		}
		// No service or not recoverable: wait for sync_complete to trigger build
		return
	}

	if IsDependencyFile(filePath) {
		bfm.logger.Log(logger.Info, "dependency file changed, triggering rebuild", fmt.Sprintf("path=%s app=%s", filePath, appCtx.ApplicationID))
		bfm.markUndeployed(appCtx.ApplicationID)
		bfm.triggerBuild(ctx, appCtx)
		return
	}

	// .env is excluded from sync; values are sent via env_vars message from client
	var lastErr error
	for attempt := 0; attempt < fileInjectRetries(); attempt++ {
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
	appCtxCopy := appCtx
	bfm.buildTimers[appCtx.ApplicationID] = time.AfterFunc(buildDebounce(), func() {
		if bfm.codebaseIndexedFunc != nil {
			bfm.codebaseIndexedFunc(appCtxCopy)
		}
	})
	bfm.buildDebounceMu.Unlock()
}

// TryRecoverFromSyncComplete attempts to recover (resume, mark deployed) when sync_complete is received.
// Returns true if recovery was done, false if a new build is needed (caller should send codebase_indexed).
func (bfm *BuildFirstManager) TryRecoverFromSyncComplete(ctx context.Context, appCtx *ApplicationContext) bool {
	if bfm.IsDeployed(appCtx.ApplicationID) {
		return true
	}
	if service, err := tasks.FindServiceByLabel(ctx, "com.application.id", appCtx.ApplicationID.String()); err == nil && service != nil {
		if bfm.injector.IsContainerRunning(ctx, appCtx) {
			workdir := tasks.GetWorkdirFromService(service)
			bfm.MarkDeployed(appCtx.ApplicationID, workdir)
			bfm.logger.Log(logger.Info, "recovered deployed state from sync_complete", appCtx.ApplicationID.String())
			return true
		}
		if tasks.IsLiveDevServicePaused(service) {
			workdir, err := bfm.taskService.ResumeLiveDevService(ctx, appCtx.ApplicationID, nil)
			if err != nil {
				bfm.logger.Log(logger.Warning, "resume failed on sync_complete", err.Error())
				return false
			}
			bfm.MarkDeployed(appCtx.ApplicationID, workdir)
			bfm.logger.Log(logger.Info, "resumed paused service on sync_complete", appCtx.ApplicationID.String())
			return true
		}
	}
	return false
}

// HandleSyncComplete is called when the client sends sync_complete (all files synced).
// For recovery cases, performs recovery. For build-needed cases, caller should send codebase_indexed.
// Deprecated: Use TryRecoverFromSyncComplete; gateway sends codebase_indexed when it returns false.
func (bfm *BuildFirstManager) HandleSyncComplete(ctx context.Context, appCtx *ApplicationContext) {
	if bfm.TryRecoverFromSyncComplete(ctx, appCtx) {
		return
	}
	// Need new build - gateway sends codebase_indexed; do not call pipeline
}

// StartBuildFromDockerfile writes the Dockerfile to staging and starts the live dev build.
// Used when the client sends trigger_build with a generated Dockerfile from the Mastra workflow.
func (bfm *BuildFirstManager) StartBuildFromDockerfile(ctx context.Context, appCtx *ApplicationContext, dockerfile, dockerignore string, port int, workdir string) error {
	appID := appCtx.ApplicationID
	bfm.logger.Log(logger.Info, "starting build from provided Dockerfile", appID.String())
	bfm.emitBuildStatus(appID, "starting", "Writing Dockerfile and starting container build...", "")

	if port <= 0 {
		port = 3000
	}
	if workdir == "" {
		workdir = "/app"
	}

	dockerfilePath, err := WriteDockerfileToStaging(ctx, appCtx.StagingPath, dockerfile)
	if err != nil {
		errMsg := fmt.Sprintf("app=%s err=%v", appID, err)
		bfm.logger.Log(logger.Error, "failed to write dockerfile", errMsg)
		bfm.emitBuildStatus(appID, "error", "Failed to write Dockerfile", err.Error())
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	if dockerignore != "" {
		if _, err := WriteDockerignoreToStaging(ctx, appCtx.StagingPath, dockerignore); err != nil {
			bfm.logger.Log(logger.Warning, "failed to write .dockerignore", fmt.Sprintf("app=%s err=%v", appID, err))
		}
	}

	var sessionEnv map[string]string
	if bfm.sessionEnvResolver != nil {
		sessionEnv = bfm.sessionEnvResolver(appID)
	}
	envVars := mergeEnvVars(appCtx.EnvironmentVariables, sessionEnv)
	cfg := tasks.LiveDevConfig{
		ApplicationID:  appID,
		OrganizationID: appCtx.OrganizationID,
		StagingPath:    appCtx.StagingPath,
		EnvVars:        envVars,
		DockerfilePath: dockerfilePath,
		InternalPort:   port,
		Workdir:        workdir,
	}
	if appCtx.Domain != "" {
		cfg.Domain = appCtx.Domain
	}

	bfm.emitBuildStatus(appID, "building_container", "Starting container build...", "")

	if err := bfm.taskService.StartLiveDevTask(ctx, cfg); err != nil {
		errMsg := fmt.Sprintf("app=%s err=%v", appID, err)
		bfm.logger.Log(logger.Error, "build-first deployment failed", errMsg)
		bfm.emitBuildStatus(appID, "error", "Container build failed", err.Error())
		return fmt.Errorf("container build failed: %w", err)
	}

	bfm.logger.Log(logger.Info, "build-first deployment queued", appCtx.ApplicationID.String())
	return nil
}

func (bfm *BuildFirstManager) markUndeployed(appID uuid.UUID) {
	bfm.deployedMu.Lock()
	delete(bfm.deployed, appID)
	bfm.deployedMu.Unlock()
}

// SetPipelineProgressFunc sets the callback for real-time pipeline progress events.
// Must be called before any builds start, typically right after creating the manager.
func (bfm *BuildFirstManager) SetPipelineProgressFunc(fn PipelineProgressFunc) {
	bfm.pipelineProgressFunc = fn
}

// SetCodebaseIndexedFunc sets the callback for when the CLI should run the deployment workflow.
// When set, triggerBuild will call this instead of the pipeline.
func (bfm *BuildFirstManager) SetCodebaseIndexedFunc(fn CodebaseIndexedFunc) {
	bfm.codebaseIndexedFunc = fn
}

// SetBuildStatusFunc sets the callback for build lifecycle events.
func (bfm *BuildFirstManager) SetBuildStatusFunc(fn BuildStatusFunc) {
	bfm.buildStatusFunc = fn
}

// emitBuildStatus sends a build lifecycle event if the callback is set.
func (bfm *BuildFirstManager) emitBuildStatus(appID uuid.UUID, phase, message, errMsg string) {
	if bfm.buildStatusFunc != nil {
		bfm.buildStatusFunc(appID, phase, message, errMsg)
	}
}

// MarkDeployed is called when a live dev build completes and the container is healthy.
// Enables file injection instead of full rebuilds.
func (bfm *BuildFirstManager) MarkDeployed(appID uuid.UUID, workdir string) {
	if workdir == "" {
		workdir = "/app"
	}
	bfm.deployedMu.Lock()
	bfm.deployed[appID] = &deployedApp{workdir: workdir}
	bfm.deployedMu.Unlock()
	bfm.logger.Log(logger.Info, "live dev container ready, inject mode enabled", appID.String())
}

func WriteDockerfileToStaging(ctx context.Context, stagingPath string, content string) (string, error) {
	fullPath := filepath.Join(stagingPath, generatedDockerfileName())

	if isLocalStagingPath(stagingPath) {
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("failed to write Dockerfile: %w", err)
		}
		return generatedDockerfileName(), nil
	}

	err := utils.WithSFTPClientFromPool(ctx, func(sftpClient *sftp.Client) error {
		file, err := sftpClient.Create(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create Dockerfile: %w", err)
		}
		defer file.Close()
		_, err = file.Write([]byte(content))
		return err
	})
	if err != nil {
		return "", err
	}
	return generatedDockerfileName(), nil
}

func WriteDockerignoreToStaging(ctx context.Context, stagingPath string, content string) (string, error) {
	fullPath := filepath.Join(stagingPath, ".dockerignore")

	if isLocalStagingPath(stagingPath) {
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("failed to write .dockerignore: %w", err)
		}
		return ".dockerignore", nil
	}

	err := utils.WithSFTPClientFromPool(ctx, func(sftpClient *sftp.Client) error {
		file, err := sftpClient.Create(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create .dockerignore: %w", err)
		}
		defer file.Close()
		_, err = file.Write([]byte(content))
		return err
	})
	if err != nil {
		return "", err
	}
	return ".dockerignore", nil
}

func getDependencyFiles() []string {
	return []string{
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
		generatedDockerfileName(),
		".dockerignore",
	}
}

func IsDependencyFile(filePath string) bool {
	baseName := filepath.Base(filePath)
	for _, df := range getDependencyFiles() {
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
