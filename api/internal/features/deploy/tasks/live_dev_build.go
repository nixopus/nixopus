package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/swarm"
	"github.com/google/uuid"
	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	ServiceNamePrefix          = "nixopus-dev-"
	DefaultHealthCheckTimeout  = 120 * time.Second
	DefaultHealthCheckInterval = 2 * time.Second
	DefaultMemoryLimit         = 2 * 1024 * 1024 * 1024
	DefaultCPULimit            = 2 * 1000000000
)

func (s *TaskService) StartLiveDevTask(ctx context.Context, config LiveDevConfig) error {
	if LiveDevQueue == nil {
		return fmt.Errorf("live dev queue not initialized - call SetupLiveDevQueue first")
	}
	err := LiveDevQueue.Add(TaskLiveDev.WithArgs(ctx, config))
	if err != nil {
		return fmt.Errorf("failed to enqueue live dev task: %w", err)
	}
	return nil
}

func (s *TaskService) SetupLiveDevQueue() {
	s.SetupCreateDeploymentQueue()
}

func LiveDevImageName(applicationID uuid.UUID) string {
	return fmt.Sprintf("nixopus-dev-%s", applicationID.String())
}

func workdirOrDefault(w string) string {
	if w != "" {
		return w
	}
	return "/app"
}

func (s *TaskService) HandleBuildFirstLiveDev(ctx context.Context, config LiveDevConfig) error {
	if config.OrganizationID != uuid.Nil {
		ctx = context.WithValue(ctx, shared_types.OrganizationIDKey, config.OrganizationID.String())
	}

	var taskCtx *LiveDevTaskContext
	if config.ApplicationID != uuid.Nil {
		var err error
		taskCtx, err = s.NewLiveDevTaskContext(config)
		if err != nil {
			s.Logger.Log(logger.Warning, "Failed to create task context: "+err.Error(), "")
		}
	}

	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Starting build-first live dev for application %s", config.ApplicationID))
	}

	dockerfilePath, err := s.resolveDockerfile(ctx, config, taskCtx)
	if err != nil {
		if taskCtx != nil {
			taskCtx.LogAndUpdateStatus(fmt.Sprintf("Failed to resolve Dockerfile: %v", err), shared_types.Failed)
		}
		return err
	}

	imageTag := LiveDevImageName(config.ApplicationID)
	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Building image %s from %s", imageTag, dockerfilePath))
		taskCtx.UpdateStatus(shared_types.Building)
	}

	if err := s.buildLiveDevImage(ctx, config, dockerfilePath, imageTag, taskCtx); err != nil {
		if taskCtx != nil {
			taskCtx.LogAndUpdateStatus(fmt.Sprintf("Build failed: %v", err), shared_types.Failed)
		}
		return err
	}

	if taskCtx != nil {
		taskCtx.AddLog("Image built successfully")
		taskCtx.UpdateStatus(shared_types.Deploying)
	}

	port, err := s.determineLiveDevPort(ctx, config, taskCtx)
	if err != nil {
		if taskCtx != nil {
			taskCtx.LogAndUpdateStatus(fmt.Sprintf("Failed to determine port: %v", err), shared_types.Failed)
		}
		return err
	}

	service, err := s.deployLiveDevService(ctx, config, imageTag, port, taskCtx)
	if err != nil {
		if taskCtx != nil {
			taskCtx.LogAndUpdateStatus(fmt.Sprintf("Deploy failed: %v", err), shared_types.Failed)
		}
		return err
	}

	if config.Domain != "" {
		if err := s.addDomainToCaddy(ctx, config.ApplicationID, config.Domain, port, config.OrganizationID, taskCtx); err != nil {
			if taskCtx != nil {
				taskCtx.AddLog(fmt.Sprintf("Warning: domain setup failed: %v", err))
			}
		}
	}

	if err := s.waitForLiveDevServiceHealthy(ctx, *service, taskCtx); err != nil {
		if taskCtx != nil {
			taskCtx.LogAndUpdateStatus(fmt.Sprintf("Health check failed: %v", err), shared_types.Failed)
		}
		return err
	}

	if taskCtx != nil {
		taskCtx.LogAndUpdateStatus(fmt.Sprintf("Dev server running on port %d", port), shared_types.Deployed)
	}

	// Mark deployed only after container is healthy, so file injection can succeed.
	// Previously we marked when the task was queued, causing injection failures and extra rebuilds.
	workdir := config.Workdir
	if workdir == "" {
		workdir = "/app"
	}
	if s.OnLiveDevDeployed != nil {
		s.OnLiveDevDeployed(config.ApplicationID, workdir)
	}

	return nil
}

func (s *TaskService) resolveDockerfile(ctx context.Context, config LiveDevConfig, taskCtx *LiveDevTaskContext) (string, error) {
	if config.DockerfilePath != "" {
		if taskCtx != nil {
			taskCtx.AddLog(fmt.Sprintf("Using Dockerfile: %s", config.DockerfilePath))
		}
		return config.DockerfilePath, nil
	}

	return "Dockerfile", fmt.Errorf("no Dockerfile resolved")
}

func (s *TaskService) buildLiveDevImage(ctx context.Context, config LiveDevConfig, dockerfilePath, imageTag string, taskCtx *LiveDevTaskContext) error {
	tc := taskCtx.toTaskContext()

	payload := shared_types.TaskPayload{
		Application: shared_types.Application{
			ID:             config.ApplicationID,
			Name:           imageTag,
			OrganizationID: config.OrganizationID,
			DockerfilePath: dockerfilePath,
		},
		ApplicationDeployment: shared_types.ApplicationDeployment{
			ID: tc.GetDeploymentID(),
		},
	}

	buildConfig := BuildConfig{
		TaskPayload:       payload,
		ContextPath:       config.StagingPath,
		Force:             false,
		ForceWithoutCache: false,
		TaskContext:       tc,
		Context:           ctx,
	}

	_, err := s.BuildImage(buildConfig)
	return err
}

func (s *TaskService) determineLiveDevPort(ctx context.Context, config LiveDevConfig, taskCtx *LiveDevTaskContext) (int, error) {
	if config.Port > 0 {
		return config.Port, nil
	}

	if taskCtx != nil {
		taskCtx.AddLog("Auto-allocating port")
	}

	portStr, err := s.getAvailablePort(ctx)
	if err != nil {
		if taskCtx != nil {
			taskCtx.AddLog("Port allocation failed, using default 3000")
		}
		return 3000, nil
	}

	port := 0
	fmt.Sscanf(portStr, "%d", &port)
	if port == 0 {
		return 3000, nil
	}

	return port, nil
}

func (s *TaskService) deployLiveDevService(ctx context.Context, config LiveDevConfig, imageTag string, port int, taskCtx *LiveDevTaskContext) (*swarm.Service, error) {
	if taskCtx != nil {
		taskCtx.AddLog("Deploying service from built image (no bind mount)")
	}

	existingService, err := FindServiceByLabel(ctx, "com.application.id", config.ApplicationID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check existing service: %w", err)
	}

	serviceSpec := s.createBuildFirstServiceSpec(config, imageTag, port)

	if existingService != nil {
		serviceSpec.Annotations.Name = existingService.Spec.Annotations.Name
		if serviceSpec.Annotations.Labels == nil {
			serviceSpec.Annotations.Labels = make(map[string]string)
		}
		serviceSpec.Annotations.Labels["nixopus.last_update"] = fmt.Sprintf("%d", time.Now().Unix())

		if taskCtx != nil {
			taskCtx.AddLog(fmt.Sprintf("Updating existing service: %s", existingService.Spec.Name))
		}
	}

	_, err = CreateOrUpdateService(ctx, serviceSpec, existingService)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update service: %w", err)
	}

	service, err := FindServiceByLabel(ctx, "com.application.id", config.ApplicationID.String())
	if err != nil || service == nil {
		return nil, fmt.Errorf("service not found after creation")
	}

	return service, nil
}

func (s *TaskService) createBuildFirstServiceSpec(config LiveDevConfig, imageName string, port int) swarm.ServiceSpec {
	serviceName := ServiceNamePrefix + config.ApplicationID.String()
	imageRef := imageName + ":latest"

	var envVars []string
	for k, v := range config.EnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	replicas := uint64(1)
	return swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: serviceName,
			Labels: map[string]string{
				"com.application.id": config.ApplicationID.String(),
				"nixopus.type":       "devrunner",
				"nixopus.build_mode": "build-first",
				"nixopus.workdir":    workdirOrDefault(config.Workdir),
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image: imageRef,
				Env:   envVars,
				Init:  func() *bool { v := true; return &v }(),
				DNSConfig: &swarm.DNSConfig{
					Nameservers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
				},
			},
			RestartPolicy: &swarm.RestartPolicy{
				Condition:   swarm.RestartPolicyConditionOnFailure,
				MaxAttempts: func() *uint64 { v := uint64(3); return &v }(),
				Delay:       func() *time.Duration { d := 5 * time.Second; return &d }(),
			},
			Resources: &swarm.ResourceRequirements{
				Limits: &swarm.Limit{
					MemoryBytes: DefaultMemoryLimit,
					NanoCPUs:    DefaultCPULimit,
				},
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Mode: swarm.ResolutionModeVIP,
			Ports: func() []swarm.PortConfig {
				if port <= 0 {
					return nil
				}
				internalPort := config.InternalPort
				if internalPort <= 0 {
					internalPort = 3000
				}
				return []swarm.PortConfig{
					{
						Protocol:      swarm.PortConfigProtocolTCP,
						TargetPort:    uint32(internalPort),
						PublishedPort: uint32(port),
						PublishMode:   swarm.PortConfigPublishModeHost,
					},
				}
			}(),
		},
	}
}

func (s *TaskService) waitForLiveDevServiceHealthy(ctx context.Context, service swarm.Service, taskCtx *LiveDevTaskContext) error {
	deadline := time.Now().Add(DefaultHealthCheckTimeout)
	ticker := time.NewTicker(DefaultHealthCheckInterval)
	defer ticker.Stop()

	applicationID := ""
	if service.Spec.Annotations.Labels != nil {
		applicationID = service.Spec.Annotations.Labels["com.application.id"]
	}

	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Waiting for Docker service to be running (timeout: %v)", DefaultHealthCheckTimeout))
	}

	checkCount := 0
	for {
		select {
		case <-ctx.Done():
			if taskCtx != nil {
				taskCtx.AddLog("Health check cancelled: context done")
			}
			return ctx.Err()
		case <-ticker.C:
			checkCount++
			if time.Now().After(deadline) {
				var finalService *swarm.Service
				if applicationID != "" {
					if refreshedService, err := FindServiceByLabel(ctx, "com.application.id", applicationID); err == nil && refreshedService != nil {
						finalService = refreshedService
					}
				}
				if finalService == nil {
					finalService = &service
				}
				taskStates := s.getTaskStatesForService(ctx, *finalService)
				if taskCtx != nil {
					taskCtx.AddLog(fmt.Sprintf("Timeout waiting for service to start after %d attempts. Task states: %s", checkCount, taskStates))
				}
				return fmt.Errorf("timeout waiting for dev server to start. Task states: %s", taskStates)
			}

			var currentService *swarm.Service
			if applicationID != "" {
				if refreshedService, err := FindServiceByLabel(ctx, "com.application.id", applicationID); err == nil && refreshedService != nil {
					currentService = refreshedService
				}
			}
			if currentService == nil {
				currentService = &service
			}

			dockerService, err := docker.GetDockerServiceFromContext(ctx)
			if err != nil {
				if taskCtx != nil && checkCount%10 == 0 {
					taskCtx.AddLog(fmt.Sprintf("Failed to get docker service (attempt %d): %v", checkCount, err))
				}
				continue
			}
			running, desired, err := dockerService.GetServiceHealth(*currentService)
			if err != nil {
				if taskCtx != nil && checkCount%10 == 0 {
					taskCtx.AddLog(fmt.Sprintf("Failed to get Docker service health (attempt %d): %v", checkCount, err))
				}
				continue
			}

			if taskCtx != nil && checkCount%10 == 0 {
				taskStates := s.getTaskStatesForService(ctx, *currentService)
				taskCtx.AddLog(fmt.Sprintf("Docker health: %d/%d running (attempt %d). Task states: %s", running, desired, checkCount, taskStates))
			}

			if running >= desired && desired > 0 {
				if taskCtx != nil {
					taskCtx.AddLog(fmt.Sprintf("Docker service is running (%d/%d replicas) after %d attempts", running, desired, checkCount))
				}
				return nil
			}
		}
	}
}

// PauseLiveDevService scales the live dev service to 0 replicas, pausing it.
// Returns nil if service was paused or did not exist.
func PauseLiveDevService(ctx context.Context, applicationID uuid.UUID) error {
	existingService, err := FindServiceByLabel(ctx, "com.application.id", applicationID.String())
	if err != nil {
		return fmt.Errorf("failed to find live dev service: %w", err)
	}
	if existingService == nil {
		return nil // No service to pause
	}
	if !isLiveDevService(existingService) {
		return fmt.Errorf("service is not a live dev service")
	}
	dockerService, err := docker.GetDockerServiceFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get docker service: %w", err)
	}
	return dockerService.ScaleService(existingService.ID, 0, "")
}

// ResumeLiveDevService scales a paused live dev service from 0 to 1 and waits for it to be healthy.
// Returns the workdir on success.
func (s *TaskService) ResumeLiveDevService(ctx context.Context, applicationID uuid.UUID, taskCtx *LiveDevTaskContext) (string, error) {
	existingService, err := FindServiceByLabel(ctx, "com.application.id", applicationID.String())
	if err != nil {
		return "", fmt.Errorf("failed to find live dev service: %w", err)
	}
	if existingService == nil {
		return "", fmt.Errorf("no live dev service found for application %s", applicationID)
	}
	if !isLiveDevService(existingService) {
		return "", fmt.Errorf("service is not a live dev service")
	}
	replicas := uint64(0)
	if existingService.Spec.Mode.Replicated != nil && existingService.Spec.Mode.Replicated.Replicas != nil {
		replicas = *existingService.Spec.Mode.Replicated.Replicas
	}
	if replicas > 0 {
		return GetWorkdirFromService(existingService), nil
	}
	dockerService, err := s.getDockerService(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get docker service: %w", err)
	}
	if err := dockerService.ScaleService(existingService.ID, 1, ""); err != nil {
		return "", fmt.Errorf("failed to scale service: %w", err)
	}
	if taskCtx != nil {
		taskCtx.AddLog("Resuming paused live dev service")
	}
	if err := s.waitForLiveDevServiceHealthy(ctx, *existingService, taskCtx); err != nil {
		return "", fmt.Errorf("service failed to become healthy: %w", err)
	}
	refreshed, _ := FindServiceByLabel(ctx, "com.application.id", applicationID.String())
	if refreshed != nil {
		return GetWorkdirFromService(refreshed), nil
	}
	return GetWorkdirFromService(existingService), nil
}

func isLiveDevService(service *swarm.Service) bool {
	if service == nil {
		return false
	}
	if service.Spec.Annotations.Labels != nil && service.Spec.Annotations.Labels["nixopus.type"] == "devrunner" {
		return true
	}
	return false
}

// IsLiveDevServicePaused returns true if the service exists and is scaled to 0 replicas.
func IsLiveDevServicePaused(service *swarm.Service) bool {
	if service == nil || !isLiveDevService(service) {
		return false
	}
	if service.Spec.Mode.Replicated == nil || service.Spec.Mode.Replicated.Replicas == nil {
		return false
	}
	return *service.Spec.Mode.Replicated.Replicas == 0
}

// UpdateLiveDevServiceEnv updates the environment variables of a running live dev service.
// Docker Swarm will roll out new tasks with the updated env, so the process picks up the new values.
func UpdateLiveDevServiceEnv(ctx context.Context, applicationID uuid.UUID, envVars map[string]string) error {
	existingService, err := FindServiceByLabel(ctx, "com.application.id", applicationID.String())
	if err != nil {
		return fmt.Errorf("failed to find service: %w", err)
	}
	if existingService == nil {
		return fmt.Errorf("no live dev service found for application %s", applicationID)
	}

	// Ensure it's a devrunner service
	if existingService.Spec.Annotations.Labels == nil ||
		existingService.Spec.Annotations.Labels["nixopus.type"] != "devrunner" {
		return fmt.Errorf("service is not a live dev service")
	}

	// Build new env slice
	var envSlice []string
	for k, v := range envVars {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	// Create updated spec: clone current spec and replace Env
	updatedSpec := existingService.Spec
	if updatedSpec.TaskTemplate.ContainerSpec == nil {
		return fmt.Errorf("service spec has no container spec")
	}
	// Clone ContainerSpec to avoid mutating the original
	containerSpec := *updatedSpec.TaskTemplate.ContainerSpec
	containerSpec.Env = envSlice
	updatedSpec.TaskTemplate.ContainerSpec = &containerSpec

	// Preserve service name for update
	updatedSpec.Annotations.Name = existingService.Spec.Annotations.Name
	if updatedSpec.Annotations.Labels == nil {
		updatedSpec.Annotations.Labels = make(map[string]string)
	}
	for k, v := range existingService.Spec.Annotations.Labels {
		updatedSpec.Annotations.Labels[k] = v
	}
	updatedSpec.Annotations.Labels["nixopus.last_update"] = fmt.Sprintf("%d", time.Now().Unix())

	_, err = CreateOrUpdateService(ctx, updatedSpec, existingService)
	return err
}

func (s *TaskService) addDomainToCaddy(ctx context.Context, applicationID uuid.UUID, domain string, port int, organizationID uuid.UUID, taskCtx *LiveDevTaskContext) error {
	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Adding domain %s to Caddy proxy...", domain))
	}

	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, organizationID.String())
	client, err := GetCaddyClient(orgCtx, nil, &s.Logger)
	if err != nil {
		return fmt.Errorf("failed to get Caddy client: %w", err)
	}
	upstreamHost, err := GetSSHHostForOrganization(ctx, organizationID)
	if err != nil {
		return err
	}

	if err := client.AddDomainWithAutoTLS(domain, upstreamHost, port, caddygo.DomainOptions{}); err != nil {
		return fmt.Errorf("failed to add domain to caddy: %w", err)
	}

	// Store domain in application_domains for lookup (e.g. auth, domain resolution)
	// Idempotent: skip if this application already has this domain (e.g. on rebuild)
	existingDomains, err := s.Storage.GetApplicationDomains(applicationID)
	if err == nil {
		for _, d := range existingDomains {
			if d.Domain == domain {
				client.Reload()
				return nil // already stored
			}
		}
	}
	if err := s.Storage.AddApplicationDomains(applicationID, []string{domain}); err != nil {
		return fmt.Errorf("failed to store domain: %w", err)
	}

	client.Reload()
	return nil
}
