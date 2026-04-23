package tasks

import (
	"context"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (t *TaskService) composeStageResolveRepo(ctx context.Context, payload shared_types.TaskPayload, deploymentType string, taskCtx *TaskContext) (string, error) {
	return t.cloneRepositoryForCompose(ctx, payload, deploymentType, taskCtx)
}

func (t *TaskService) composeStagePrepareProject(orgCtx context.Context, payload shared_types.TaskPayload, repoPath string, taskCtx *TaskContext) (composeFilePath string, overrideFiles []string, envVars map[string]string, err error) {
	composeFilePath = t.buildComposeFilePath(payload, repoPath, taskCtx)
	envVars = GetMapFromString(payload.Application.EnvironmentVariables)

	if err := t.discoverAndPersistComposeServices(orgCtx, composeFilePath, payload, taskCtx, envVars); err != nil {
		taskCtx.AddLog("Warning: failed to discover compose services: " + err.Error())
	}

	overrideFile, werr := t.writeComposeLabelsOverride(orgCtx, composeFilePath, payload, taskCtx)
	if werr != nil {
		taskCtx.AddLog("Warning: failed to write labels override: " + werr.Error())
	}
	if overrideFile != "" {
		overrideFiles = append(overrideFiles, overrideFile)
	}
	return composeFilePath, overrideFiles, envVars, nil
}

func (t *TaskService) composeStageExecuteStack(orgCtx context.Context, deploymentTypeEnum shared_types.DeploymentType, composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext, overrideFiles []string) error {
	return t.executeComposeDeployment(orgCtx, deploymentTypeEnum, composeFilePath, envVars, outputCallback, taskCtx, overrideFiles)
}

func (t *TaskService) composeStagePublishArtifacts(orgCtx context.Context, payload shared_types.TaskPayload, composeFilePath string, envVars map[string]string, taskCtx *TaskContext) {
	if payload.Application.Source != shared_types.SourceTemplate {
		t.ExportComposeImagesToS3(orgCtx, payload, composeFilePath, envVars, taskCtx)
	}
}

func (t *TaskService) composeStageConfigureProxy(orgCtx context.Context, payload shared_types.TaskPayload, taskCtx *TaskContext) error {
	if err := t.addDomainsForCompose(orgCtx, payload, taskCtx); err != nil {
		return err
	}
	taskCtx.LogAndUpdateStatus("Docker Compose deployment completed successfully", shared_types.Deployed)
	return nil
}
