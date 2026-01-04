package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	container_tools "github.com/raghavyuva/nixopus-api/internal/features/container/tools"
	dashboard_tools "github.com/raghavyuva/nixopus-api/internal/features/dashboard/tools"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	deploy_service "github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	deploy_storage "github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	deploy_tasks "github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	deploy_tools "github.com/raghavyuva/nixopus-api/internal/features/deploy/tools"
	extension_tools "github.com/raghavyuva/nixopus-api/internal/features/extension/tools"
	file_manager_tools "github.com/raghavyuva/nixopus-api/internal/features/file-manager/tools"
	github_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	github_storage "github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	ssh_tools "github.com/raghavyuva/nixopus-api/internal/features/ssh/tools"
	mcp_middleware "github.com/raghavyuva/nixopus-api/internal/mcp/middleware"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// RegisterTool registers an MCP tool with automatic authentication middleware.
// Organization ID is automatically extracted from the API key and set in context.
func RegisterTool[Input any, Output any](
	server *mcp.Server,
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	tool *mcp.Tool,
	handler func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, Output, error),
) {
	// Automatically wrap handler with auth middleware
	wrappedHandler := mcp_middleware.WithAuth(store, l, handler)
	mcp.AddTool(server, tool, wrappedHandler)
}

// RegisterTools registers all MCP tools with the server
func RegisterTools(
	server *mcp.Server,
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) {
	dockerService, err := docker.GetDockerManager().GetDefaultService()
	if err != nil {
		l.Log(logger.Error, fmt.Sprintf("failed to get docker service: %v", err), "")
		return
	}
	if dockerService == nil {
		l.Log(logger.Error, "docker service is nil", "")
		return
	}

	// Create deploy services for TaskService
	deployStorage := deploy_storage.DeployStorage{DB: store.DB, Ctx: ctx}
	githubConnectorStorage := github_storage.GithubConnectorStorage{DB: store.DB, Ctx: ctx}
	githubConnectorService := github_service.NewGithubConnectorService(store, ctx, l, &githubConnectorStorage)
	taskService := deploy_tasks.NewTaskService(&deployStorage, l, dockerService, githubConnectorService, store)
	taskService.SetupCreateDeploymentQueue()
	taskService.StartConsumers(ctx)

	// Create deploy service for read operations
	deployService := deploy_service.NewDeployService(store, ctx, l, &deployStorage)

	containerLogsHandler := container_tools.GetContainerLogsHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_container_logs",
		Description: "Get logs from a Docker container. Requires container ID.",
	}, containerLogsHandler)

	containerHandler := container_tools.GetContainerHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_container",
		Description: "Get detailed information about a Docker container. Requires container ID.",
	}, containerHandler)

	listContainersHandler := container_tools.ListContainersHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "list_containers",
		Description: "List Docker containers with pagination, filtering, and sorting.",
	}, listContainersHandler)

	listImagesHandler := container_tools.ListImagesHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "list_images",
		Description: "List Docker images with optional filtering by container ID or image prefix.",
	}, listImagesHandler)

	pruneImagesHandler := container_tools.PruneImagesHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "prune_images",
		Description: "Prune Docker images with optional filtering by until time, label, or dangling status.",
	}, pruneImagesHandler)

	pruneBuildCacheHandler := container_tools.PruneBuildCacheHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "prune_build_cache",
		Description: "Prune Docker build cache. Optionally prune all cache entries.",
	}, pruneBuildCacheHandler)

	removeContainerHandler := container_tools.RemoveContainerHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "remove_container",
		Description: "Remove a Docker container. Requires container ID. Optionally force removal.",
	}, removeContainerHandler)

	restartContainerHandler := container_tools.RestartContainerHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "restart_container",
		Description: "Restart a Docker container. Requires container ID. Optionally specify timeout in seconds.",
	}, restartContainerHandler)

	startContainerHandler := container_tools.StartContainerHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "start_container",
		Description: "Start a Docker container. Requires container ID.",
	}, startContainerHandler)

	stopContainerHandler := container_tools.StopContainerHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "stop_container",
		Description: "Stop a Docker container. Requires container ID. Optionally specify timeout in seconds.",
	}, stopContainerHandler)

	updateContainerResourcesHandler := container_tools.UpdateContainerResourcesHandler(store, ctx, l, dockerService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "update_container_resources",
		Description: "Update resource limits (memory, memory swap, CPU shares) of a running Docker container. Requires container ID. Optionally specify memory (bytes), memory_swap (bytes), and cpu_shares.",
	}, updateContainerResourcesHandler)

	// File Manager Tools
	listFilesHandler := file_manager_tools.ListFilesHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "list_files",
		Description: "List files and directories in a given path. Requires path.",
	}, listFilesHandler)

	createDirectoryHandler := file_manager_tools.CreateDirectoryHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "create_directory",
		Description: "Create a new directory at the given path. Requires path.",
	}, createDirectoryHandler)

	deleteFileHandler := file_manager_tools.DeleteFileHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "delete_file",
		Description: "Delete a file or directory at the given path. Requires path.",
	}, deleteFileHandler)

	moveFileHandler := file_manager_tools.MoveFileHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "move_file",
		Description: "Move or rename a file or directory from one path to another. Requires from_path and to_path.",
	}, moveFileHandler)

	copyDirectoryHandler := file_manager_tools.CopyDirectoryHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "copy_directory",
		Description: "Copy a file or directory from one path to another. Requires from_path and to_path.",
	}, copyDirectoryHandler)

	// Dashboard Tools
	getSystemStatsHandler := dashboard_tools.GetSystemStatsHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_system_stats",
		Description: "Get system statistics including CPU, memory, disk, network, and load information.",
	}, getSystemStatsHandler)

	// SSH Tools
	runCommandHandler := ssh_tools.RunCommandHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "run_command",
		Description: "Run a command on a remote server via SSH. Requires command. Optionally specify client_id for multi-client support.",
	}, runCommandHandler)

	// Extension Tools
	listExtensionsHandler := extension_tools.ListExtensionsHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "list_extensions",
		Description: "List extensions with pagination, filtering, and sorting. Optionally specify category, type, search, sort_by, sort_dir, page, and page_size.",
	}, listExtensionsHandler)

	getExtensionHandler := extension_tools.GetExtensionHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_extension",
		Description: "Get detailed information about an extension. Requires extension ID.",
	}, getExtensionHandler)

	runExtensionHandler := extension_tools.RunExtensionHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "run_extension",
		Description: "Run an extension with variable values. Requires extension_id. Optionally specify variables map.",
	}, runExtensionHandler)

	getExecutionHandler := extension_tools.GetExecutionHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_execution",
		Description: "Get execution status and details. Requires execution_id.",
	}, getExecutionHandler)

	listExecutionLogsHandler := extension_tools.ListExecutionLogsHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "list_execution_logs",
		Description: "List execution logs with pagination. Requires execution_id. Optionally specify after_seq and limit.",
	}, listExecutionLogsHandler)

	cancelExecutionHandler := extension_tools.CancelExecutionHandler(store, ctx, l)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "cancel_execution",
		Description: "Cancel a running extension execution. Requires execution_id.",
	}, cancelExecutionHandler)

	// Deploy Tools
	deleteApplicationHandler := deploy_tools.DeleteApplicationHandler(store, ctx, l, taskService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "delete_application",
		Description: "Delete a deployed application. Requires application ID.",
	}, deleteApplicationHandler)

	getApplicationDeploymentsHandler := deploy_tools.GetApplicationDeploymentsHandler(store, ctx, l, deployService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_application_deployments",
		Description: "Get deployments for an application with pagination. Requires application ID. Optionally specify page and page_size.",
	}, getApplicationDeploymentsHandler)

	getApplicationHandler := deploy_tools.GetApplicationHandler(store, ctx, l, deployService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_application",
		Description: "Get a single application by ID. Requires application ID.",
	}, getApplicationHandler)

	getApplicationsHandler := deploy_tools.GetApplicationsHandler(store, ctx, l, deployService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_applications",
		Description: "Get all applications with pagination. Optionally specify page and page_size.",
	}, getApplicationsHandler)

	getDeploymentByIdHandler := deploy_tools.GetDeploymentByIdHandler(store, ctx, l, deployService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_deployment_by_id",
		Description: "Get a single deployment by ID. Requires deployment ID.",
	}, getDeploymentByIdHandler)

	getDeploymentLogsHandler := deploy_tools.GetDeploymentLogsHandler(store, ctx, l, deployService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "get_deployment_logs",
		Description: "Get logs for a deployment with pagination and filtering. Requires deployment ID. Optionally specify page, page_size, level, start_time (RFC3339), end_time (RFC3339), and search_term.",
	}, getDeploymentLogsHandler)

	createProjectHandler := deploy_tools.CreateProjectHandler(store, ctx, l, deployService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "create_project",
		Description: "Create a new project (application) without triggering deployment. Requires name, domain, and repository. Optionally specify environment, build_pack, branch, port, dockerfile_path, base_path, pre_run_command, post_run_command, build_variables, and environment_variables.",
	}, createProjectHandler)

	deployProjectHandler := deploy_tools.DeployProjectHandler(store, ctx, l, taskService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "deploy_project",
		Description: "Deploy an existing project (application) that was saved as a draft. Requires project ID.",
	}, deployProjectHandler)

	duplicateProjectHandler := deploy_tools.DuplicateProjectHandler(store, ctx, l, deployService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "duplicate_project",
		Description: "Duplicate an existing project with a different environment. Requires source_project_id, domain, and environment. Optionally specify branch.",
	}, duplicateProjectHandler)

	restartDeploymentHandler := deploy_tools.RestartDeploymentHandler(store, ctx, l, taskService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "restart_deployment",
		Description: "Restart a deployment. Requires deployment ID.",
	}, restartDeploymentHandler)

	rollbackDeploymentHandler := deploy_tools.RollbackDeploymentHandler(store, ctx, l, taskService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "rollback_deployment",
		Description: "Rollback a deployment to a previous version. Requires deployment ID.",
	}, rollbackDeploymentHandler)

	redeployApplicationHandler := deploy_tools.RedeployApplicationHandler(store, ctx, l, taskService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "redeploy_application",
		Description: "Redeploy an application. Requires application ID. Optionally specify force and force_without_cache.",
	}, redeployApplicationHandler)

	updateProjectHandler := deploy_tools.UpdateProjectHandler(store, ctx, l, taskService)
	RegisterTool(server, store, ctx, l, &mcp.Tool{
		Name:        "update_project",
		Description: "Update a project configuration without triggering deployment. Requires application ID. Optionally specify name, environment, pre_run_command, post_run_command, build_variables, environment_variables, port, force, dockerfile_path, and base_path.",
	}, updateProjectHandler)
}
