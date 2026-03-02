package tasks

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/caddy"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (t *TaskService) deployDockerCompose(ctx context.Context, TaskPayload shared_types.TaskPayload, deploymentType string) error {
	taskCtx := t.NewTaskContext(TaskPayload)

	repoPath, err := t.cloneRepositoryForCompose(ctx, TaskPayload, deploymentType, taskCtx)
	if err != nil {
		return err
	}

	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, TaskPayload.Application.OrganizationID.String())

	composeFilePath := t.buildComposeFilePath(TaskPayload, repoPath, taskCtx)

	if err := t.discoverAndPersistComposeServices(orgCtx, composeFilePath, TaskPayload, taskCtx); err != nil {
		taskCtx.AddLog("Warning: failed to discover compose services: " + err.Error())
	}

	envVars := GetMapFromString(TaskPayload.Application.EnvironmentVariables)
	outputCallback := t.createOutputCallback(taskCtx)

	deploymentTypeEnum := shared_types.DeploymentType(deploymentType)
	if err := t.executeComposeDeployment(orgCtx, deploymentTypeEnum, composeFilePath, envVars, outputCallback, taskCtx); err != nil {
		return err
	}

	if err := t.addDomainsForCompose(orgCtx, TaskPayload, taskCtx); err != nil {
		return err
	}

	taskCtx.LogAndUpdateStatus("Docker Compose deployment completed successfully", shared_types.Deployed)
	return nil
}

func (t *TaskService) cloneRepositoryForCompose(ctx context.Context, TaskPayload shared_types.TaskPayload, deploymentType string, taskCtx *TaskContext) (string, error) {
	taskCtx.LogAndUpdateStatus("Starting deployment process", shared_types.Cloning)

	resolver := t.GetSourceResolver(TaskPayload.Application.Source)
	repoPath, err := resolver.Resolve(ctx, SourceResolveConfig{
		TaskPayload:    TaskPayload,
		DeploymentType: deploymentType,
		TaskContext:    taskCtx,
	})
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to resolve source: "+err.Error(), shared_types.Failed)
		return "", err
	}

	taskCtx.LogAndUpdateStatus("Source resolved successfully", shared_types.Deploying)
	taskCtx.AddLog("Source resolved to: " + repoPath)
	return repoPath, nil
}

func (t *TaskService) buildComposeFilePath(TaskPayload shared_types.TaskPayload, repoPath string, taskCtx *TaskContext) string {
	basePath := TaskPayload.Application.BasePath
	if basePath == "" || basePath == "/" {
		basePath = "."
	}

	composeFileName := "docker-compose.yml"
	if TaskPayload.Application.DockerfilePath != "" && TaskPayload.Application.DockerfilePath != "Dockerfile" {
		composeFileName = TaskPayload.Application.DockerfilePath
	}

	composeFilePath := filepath.Join(repoPath, basePath, composeFileName)
	taskCtx.AddLog("Starting Docker Compose services from: " + composeFilePath)
	return composeFilePath
}

func (t *TaskService) createOutputCallback(taskCtx *TaskContext) func(string) {
	return func(line string) {
		if strings.TrimSpace(line) != "" {
			taskCtx.AddLog("Docker Compose: " + line)
		}
	}
}

func (t *TaskService) executeComposeDeployment(ctx context.Context, deploymentType shared_types.DeploymentType, composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext) error {
	switch deploymentType {
	case shared_types.DeploymentTypeCreate:
		return t.composeUp(ctx, composeFilePath, envVars, outputCallback, taskCtx, "Starting Docker Compose services", "Docker Compose services started successfully")

	case shared_types.DeploymentTypeReDeploy, shared_types.DeploymentTypeUpdate, shared_types.DeploymentTypeRollback:
		if err := t.composeDown(ctx, composeFilePath, outputCallback, taskCtx); err != nil {
			return err
		}
		taskCtx.AddLog("Existing services stopped, starting with new code")
		return t.composeUp(ctx, composeFilePath, envVars, outputCallback, taskCtx, "Starting Docker Compose services", "Docker Compose services restarted successfully")

	case shared_types.DeploymentTypeRestart:
		return t.composeRestart(ctx, composeFilePath, envVars, outputCallback, taskCtx)

	default:
		taskCtx.LogAndUpdateStatus("Unknown deployment type: "+string(deploymentType), shared_types.Failed)
		return fmt.Errorf("unknown deployment type: %s", deploymentType)
	}
}

func (t *TaskService) composeUp(ctx context.Context, composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext, startMsg, successMsg string) error {
	taskCtx.AddLog(startMsg)

	dockerSvc, err := t.getDockerService(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get docker service: "+err.Error(), shared_types.Failed)
		return err
	}

	if ds, ok := dockerSvc.(*docker.DockerService); ok {
		_, err := ds.ComposeUpWithCallback(composeFilePath, envVars, outputCallback)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to start docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	} else {
		output, err := dockerSvc.ComposeUp(composeFilePath, envVars)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to start docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
		if output != "" {
			taskCtx.AddLog("Docker Compose output: " + output)
		}
	}

	taskCtx.AddLog(successMsg)
	return nil
}

func (t *TaskService) composeDown(ctx context.Context, composeFilePath string, outputCallback func(string), taskCtx *TaskContext) error {
	taskCtx.AddLog("Stopping existing Docker Compose services")

	dockerSvc, err := t.getDockerService(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get docker service: "+err.Error(), shared_types.Failed)
		return err
	}

	if ds, ok := dockerSvc.(*docker.DockerService); ok {
		err := ds.ComposeDownWithCallback(composeFilePath, outputCallback)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to stop docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	} else {
		err := dockerSvc.ComposeDown(composeFilePath)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to stop docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	}

	return nil
}

func (t *TaskService) composeRestart(ctx context.Context, composeFilePath string, envVars map[string]string, outputCallback func(string), taskCtx *TaskContext) error {
	taskCtx.AddLog("Restarting Docker Compose services")

	dockerSvc, err := t.getDockerService(ctx)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get docker service: "+err.Error(), shared_types.Failed)
		return err
	}

	if ds, ok := dockerSvc.(*docker.DockerService); ok {
		err := ds.ComposeRestart(composeFilePath, envVars, outputCallback)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to restart docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
	} else {
		if err := t.composeDown(ctx, composeFilePath, outputCallback, taskCtx); err != nil {
			return err
		}
		output, err := dockerSvc.ComposeUp(composeFilePath, envVars)
		if err != nil {
			taskCtx.LogAndUpdateStatus("Failed to start docker compose services: "+err.Error(), shared_types.Failed)
			return err
		}
		if output != "" {
			taskCtx.AddLog("Docker Compose output: " + output)
		}
	}

	taskCtx.AddLog("Docker Compose services restarted successfully")
	return nil
}

func (t *TaskService) discoverAndPersistComposeServices(ctx context.Context, composeFilePath string, TaskPayload shared_types.TaskPayload, taskCtx *TaskContext) error {
	parsed, err := ParseComposeFile(composeFilePath)
	if err != nil {
		return err
	}

	appID := TaskPayload.Application.ID

	oldServices, err := t.Storage.GetComposeServices(appID)
	if err != nil {
		taskCtx.AddLog("Warning: failed to load existing compose services: " + err.Error())
	}
	oldDomains, err := t.Storage.GetApplicationDomains(appID)
	if err != nil {
		taskCtx.AddLog("Warning: failed to load existing domains: " + err.Error())
	}

	services := buildRoutableServices(parsed, taskCtx)
	if err := t.Storage.UpsertComposeServices(appID, services); err != nil {
		return fmt.Errorf("failed to persist compose services: %w", err)
	}

	if len(services) == 0 {
		taskCtx.AddLog("No routable services found in compose file (none expose host ports)")
	}
	for _, svc := range services {
		taskCtx.AddLog(fmt.Sprintf("Discovered compose service: %s (port %d)", svc.ServiceName, svc.Port))
	}

	newServiceNames := make(map[string]bool, len(services))
	for _, svc := range services {
		newServiceNames[svc.ServiceName] = true
	}
	t.cleanupRemovedServices(ctx, oldServices, newServiceNames, oldDomains, taskCtx)

	return nil
}

func buildRoutableServices(parsed []ParsedComposeService, taskCtx *TaskContext) []shared_types.ComposeService {
	var services []shared_types.ComposeService
	for _, p := range parsed {
		port := 0
		if len(p.Ports) > 0 {
			port = p.Ports[0]
		}
		if port == 0 {
			taskCtx.AddLog(fmt.Sprintf("Skipping compose service %q: no host port exposed", p.ServiceName))
			continue
		}
		services = append(services, shared_types.ComposeService{
			ServiceName: p.ServiceName,
			Port:        port,
		})
	}
	return services
}

func collectOrphanedDomains(svc shared_types.ComposeService, domains []shared_types.ApplicationDomain) []string {
	var orphaned []string
	for _, d := range domains {
		if d.ComposeServiceID != nil && *d.ComposeServiceID == svc.ID && d.Domain != "" {
			orphaned = append(orphaned, d.Domain)
		}
	}
	return orphaned
}

func (t *TaskService) cleanupRemovedServices(ctx context.Context, oldServices []shared_types.ComposeService, newServiceNames map[string]bool, oldDomains []shared_types.ApplicationDomain, taskCtx *TaskContext) {
	for _, oldSvc := range oldServices {
		if newServiceNames[oldSvc.ServiceName] {
			continue
		}

		orphanedDomains := collectOrphanedDomains(oldSvc, oldDomains)
		if len(orphanedDomains) == 0 {
			taskCtx.AddLog(fmt.Sprintf("Compose service %q removed (no domains were linked)", oldSvc.ServiceName))
			continue
		}

		taskCtx.AddLog(fmt.Sprintf(
			"Warning: compose service %q was removed. Unlinking %d domain(s) from Caddy: %s",
			oldSvc.ServiceName, len(orphanedDomains), strings.Join(orphanedDomains, ", ")))

		if err := caddy.RemoveDomainsWithRetry(ctx, nil, &t.Logger, orphanedDomains); err != nil {
			taskCtx.AddLog("Warning: failed to remove orphaned domains from proxy: " + err.Error())
		}
	}
}

func (t *TaskService) addDomainsForCompose(ctx context.Context, TaskPayload shared_types.TaskPayload, taskCtx *TaskContext) error {
	domains, err := t.Storage.GetApplicationDomains(TaskPayload.Application.ID)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to load domains: "+err.Error(), shared_types.Failed)
		return err
	}
	if len(domains) == 0 {
		return nil
	}

	upstreamHost, err := GetSSHHostForOrganization(ctx, TaskPayload.Application.OrganizationID)
	if err != nil {
		taskCtx.LogAndUpdateStatus("Failed to get SSH host: "+err.Error(), shared_types.Failed)
		return err
	}

	var routes []caddy.DomainRoute
	for i := range domains {
		d := &domains[i]
		if d.Domain == "" {
			continue
		}
		port := d.ResolvePort()
		if port == 0 {
			taskCtx.AddLog(fmt.Sprintf("Skipping domain %s: no service linked and no port override", d.Domain))
			continue
		}
		routes = append(routes, caddy.DomainRoute{
			Domain:       d.Domain,
			UpstreamDial: caddy.FormatDial(upstreamHost, port),
		})
	}
	if len(routes) == 0 {
		return nil
	}

	if err := caddy.AddDomainsAtomic(ctx, nil, &t.Logger, routes); err != nil {
		taskCtx.LogAndUpdateStatus("Failed to configure proxy: "+err.Error(), shared_types.Failed)
		return err
	}
	for _, r := range routes {
		taskCtx.AddLog("Domain " + r.Domain + " -> " + r.UpstreamDial + " added successfully with TLS")
	}
	return nil
}
