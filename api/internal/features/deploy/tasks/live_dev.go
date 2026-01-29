package tasks

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/google/uuid"
	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/live/devrunner"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	ServiceNamePrefix          = "nixopus-dev-"
	DefaultHealthCheckTimeout  = 120 * time.Second
	DefaultHealthCheckInterval = 2 * time.Second
	DefaultMemoryLimit         = 2 * 1024 * 1024 * 1024
	DefaultCPULimit            = 2 * 1000000000
)

// StartLiveDevTask queues a live dev deployment task
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

// HandleLiveDevDeployment handles the live dev deployment task
func (s *TaskService) HandleLiveDevDeployment(ctx context.Context, config LiveDevConfig) error {
	// Create task context for logging (only if ApplicationID is provided)
	var taskCtx *LiveDevTaskContext
	if config.ApplicationID != uuid.Nil {
		var err error
		taskCtx, err = s.NewLiveDevTaskContext(config)
		if err != nil {
			s.Logger.Log(logger.Warning, "Failed to create live dev task context, continuing without deployment logs: "+err.Error(), "")
		}
	}

	// Log initial start
	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Starting live dev deployment for application %s", config.ApplicationID.String()))
	}

	strategy, err := s.getFrameworkStrategy(ctx, config, taskCtx)
	if err != nil {
		if taskCtx != nil {
			taskCtx.LogAndUpdateStatus(fmt.Sprintf("Failed to get framework strategy: %v", err), shared_types.Failed)
		}
		return err
	}

	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Detected framework: %s", strategy.Name()))
	}

	port, err := s.determinePort(ctx, config, strategy, taskCtx)
	if err != nil {
		if taskCtx != nil {
			taskCtx.LogAndUpdateStatus(fmt.Sprintf("Failed to determine port: %v", err), shared_types.Failed)
		}
		return err
	}

	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Using port %d for dev server", port))
		taskCtx.UpdateStatus(shared_types.Building)
	}

	service, err := s.ensureService(config, strategy, port, taskCtx)
	if err != nil {
		if taskCtx != nil {
			taskCtx.LogAndUpdateStatus(fmt.Sprintf("Failed to create/update service: %v", err), shared_types.Failed)
		}
		return err
	}

	if taskCtx != nil {
		taskCtx.AddLog("Service created, waiting for health check...")
		taskCtx.UpdateStatus(shared_types.Deploying)
	}

	// Add domain to Caddy if configured
	if config.Domain != "" {
		if err := s.addDomainToCaddy(ctx, config.Domain, port, config.OrganizationID, taskCtx); err != nil {
			// Log error but don't fail the deployment - domain can be added later
			if taskCtx != nil {
				taskCtx.AddLog(fmt.Sprintf("Warning: Failed to add domain to Caddy: %v", err))
			}
			s.Logger.Log(logger.Warning, "failed to add domain to caddy", fmt.Sprintf("domain=%s port=%d application=%s err=%v", config.Domain, port, config.ApplicationID, err))
		} else {
			if taskCtx != nil {
				taskCtx.AddLog(fmt.Sprintf("Domain %s added to Caddy successfully", config.Domain))
			}
		}
	}

	// Update deployment with container info
	if taskCtx != nil {
		taskCtx.UpdateDeployment(map[string]interface{}{
			"container_name":   service.Spec.Name,
			"container_image":  strategy.GetBaseImage(),
			"container_status": "running",
		})
		taskCtx.LogAndUpdateStatus(fmt.Sprintf("Dev server started successfully on port %d", port), shared_types.Deployed)
	}

	// we will perform healthcheck after caddy configuring etc since healthcheck may take time ..
	if err := s.waitForLiveDevServiceHealthy(ctx, *service, taskCtx); err != nil {
		if taskCtx != nil {
			taskCtx.LogAndUpdateStatus(fmt.Sprintf("Dev server failed to start: %v", err), shared_types.Failed)
		}
		return fmt.Errorf("dev server failed to start: %w", err)
	}

	return nil
}

// getFrameworkStrategy retrieves or detects the framework strategy
func (s *TaskService) getFrameworkStrategy(ctx context.Context, config LiveDevConfig, taskCtx *LiveDevTaskContext) (devrunner.FrameworkStrategy, error) {
	if config.Framework != "" {
		if taskCtx != nil {
			taskCtx.AddLog(fmt.Sprintf("Using specified framework: %s", config.Framework))
		}
		strategy, err := devrunner.GetStrategyByNameWithPath(ctx, config.Framework, config.StagingPath)
		if err != nil {
			s.Logger.Log(logger.Error, fmt.Sprintf("[LiveDev] [%s] Failed to get strategy by name '%s': %v", config.ApplicationID, config.Framework, err), "")
			return nil, fmt.Errorf("failed to get framework strategy '%s': %w", config.Framework, err)
		}
		return strategy, nil
	}

	if taskCtx != nil {
		taskCtx.AddLog("Auto-detecting framework from project files...")
	}

	strategy, err := devrunner.DetectFramework(ctx, config.StagingPath)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("[LiveDev] [%s] Framework detection failed: %v", config.ApplicationID, err), "")
		return nil, fmt.Errorf("failed to detect framework at %s: %w", config.StagingPath, err)
	}

	return strategy, nil
}

// determinePort determines the port to use for the service
func (s *TaskService) determinePort(ctx context.Context, config LiveDevConfig, strategy devrunner.FrameworkStrategy, taskCtx *LiveDevTaskContext) (int, error) {
	if config.Port > 0 {
		if taskCtx != nil {
			taskCtx.AddLog(fmt.Sprintf("Using specified port: %d", config.Port))
		}
		return config.Port, nil
	}

	if taskCtx != nil {
		taskCtx.AddLog("Auto-allocating available port...")
	}

	availablePort, err := s.getAvailablePort(ctx)
	if err != nil {
		defaultPort := strategy.GetDefaultPort()
		if taskCtx != nil {
			taskCtx.AddLog(fmt.Sprintf("Failed to get available port, using default: %d", defaultPort))
		}
		return defaultPort, nil
	}

	port, err := strconv.Atoi(availablePort)
	if err != nil {
		defaultPort := strategy.GetDefaultPort()
		if taskCtx != nil {
			taskCtx.AddLog(fmt.Sprintf("Invalid port value, using default: %d", defaultPort))
		}
		return defaultPort, nil
	}

	if port == 0 {
		return 0, fmt.Errorf("invalid port: port must be greater than 0")
	}

	return port, nil
}

// ensureService creates or updates the service and returns it
// For live dev, we match by application ID to update existing services for the same project
func (s *TaskService) ensureService(config LiveDevConfig, strategy devrunner.FrameworkStrategy, port int, taskCtx *LiveDevTaskContext) (*swarm.Service, error) {
	if taskCtx != nil {
		taskCtx.AddLog("Checking for existing service...")
	}

	// Find existing service by application ID
	if config.ApplicationID == uuid.Nil {
		return nil, fmt.Errorf("application ID is required to find or create service")
	}

	existingService, err := FindServiceByLabel(s.DockerRepo, "com.application.id", config.ApplicationID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check existing service: %w", err)
	}

	if existingService != nil {
		if taskCtx != nil {
			taskCtx.AddLog(fmt.Sprintf("Found existing service: %s, updating...", existingService.Spec.Name))
		}
	} else {
		if taskCtx != nil {
			taskCtx.AddLog("No existing service found, creating new service...")
		}
	}

	serviceSpec := s.createLiveDevServiceSpec(config, strategy, port)

	// Preserve existing service name when updating (Docker Swarm doesn't support renaming)
	if existingService != nil {
		serviceSpec.Annotations.Name = existingService.Spec.Annotations.Name
		// Add timestamp label to force Docker Swarm to recognize the update and refresh bind mount
		if serviceSpec.Annotations.Labels == nil {
			serviceSpec.Annotations.Labels = make(map[string]string)
		}
		serviceSpec.Annotations.Labels["nixopus.last_update"] = fmt.Sprintf("%d", time.Now().Unix())
	}

	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Creating service with image: %s", strategy.GetBaseImage()))
		taskCtx.AddLog(fmt.Sprintf("Mount: %s -> %s", config.StagingPath, strategy.GetWorkdir()))
	}

	_, err = CreateOrUpdateService(s.DockerRepo, serviceSpec, existingService)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update service: %w", err)
	}

	if taskCtx != nil {
		taskCtx.AddLog("Service created/updated successfully, verifying...")
	}

	// Verify service by application ID
	service, err := FindServiceByLabel(s.DockerRepo, "com.application.id", config.ApplicationID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	if service == nil {
		return nil, fmt.Errorf("service not found after creation")
	}

	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Service verified: %s", service.Spec.Name))
	}

	return service, nil
}

// createLiveDevServiceSpec creates a swarm service specification for live dev mode
func (s *TaskService) createLiveDevServiceSpec(config LiveDevConfig, strategy devrunner.FrameworkStrategy, port int) swarm.ServiceSpec {
	serviceName := ServiceNamePrefix + config.ApplicationID.String()
	workdir := strategy.GetWorkdir()
	internalPort := strategy.GetDefaultPort()

	envVars := s.buildEnvVars(config, strategy)
	cmd := s.buildCommand(strategy)

	replicas := uint64(1)
	return swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name:   serviceName,
			Labels: s.buildServiceLabels(config, strategy),
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   strategy.GetBaseImage(),
				Command: cmd,
				Dir:     workdir,
				Env:     envVars,
				DNSConfig: &swarm.DNSConfig{
					Nameservers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
				},
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeBind,
						Source: config.StagingPath,
						Target: workdir,
					},
				},
			},
			RestartPolicy: &swarm.RestartPolicy{
				Condition: swarm.RestartPolicyConditionAny,
			},
			Resources: &swarm.ResourceRequirements{
				Limits: &swarm.Limit{
					MemoryBytes: DefaultMemoryLimit,
					NanoCPUs:    DefaultCPULimit,
				},
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Mode:  swarm.ResolutionModeVIP,
			Ports: s.buildPortConfig(port, internalPort),
		},
	}
}

// buildEnvVars builds environment variables from strategy and config
func (s *TaskService) buildEnvVars(config LiveDevConfig, strategy devrunner.FrameworkStrategy) []string {
	envVars := make([]string, 0)
	for k, v := range strategy.GetEnvVars() {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range config.EnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}
	return envVars
}

// buildCommand builds the container command from install and dev commands
func (s *TaskService) buildCommand(strategy devrunner.FrameworkStrategy) []string {
	installCmd := strategy.GetInstallCommand()
	devCmd := strategy.GetDevCommand()

	if len(installCmd) == 0 {
		return devCmd
	}

	installCommand := extractCommandFromShellWrapper(installCmd)
	devCommand := extractCommandFromShellWrapper(devCmd)
	combinedCommand := fmt.Sprintf("%s && %s", installCommand, devCommand)

	return []string{"sh", "-c", combinedCommand}
}

// extractCommandFromShellWrapper extracts the actual command from sh -c wrapper
func extractCommandFromShellWrapper(cmd []string) string {
	if len(cmd) >= 3 && cmd[0] == "sh" && cmd[1] == "-c" {
		return cmd[2]
	}
	return shellJoin(cmd)
}

// buildServiceLabels builds service labels
func (s *TaskService) buildServiceLabels(config LiveDevConfig, strategy devrunner.FrameworkStrategy) map[string]string {
	labels := map[string]string{
		"nixopus.framework": strategy.Name(),
		"nixopus.type":      "devrunner",
	}

	// Add application ID label if available
	if config.ApplicationID != uuid.Nil {
		labels["com.application.id"] = config.ApplicationID.String()
	}

	return labels
}

// buildPortConfig builds port configuration
func (s *TaskService) buildPortConfig(publishedPort, targetPort int) []swarm.PortConfig {
	if publishedPort <= 0 {
		return nil
	}

	return []swarm.PortConfig{
		{
			Protocol:      swarm.PortConfigProtocolTCP,
			TargetPort:    uint32(targetPort),
			PublishedPort: uint32(publishedPort),
			PublishMode:   swarm.PortConfigPublishModeHost,
		},
	}
}

// waitForLiveDevServiceHealthy waits for Docker service to be running
func (s *TaskService) waitForLiveDevServiceHealthy(ctx context.Context, service swarm.Service, taskCtx *LiveDevTaskContext) error {
	deadline := time.Now().Add(DefaultHealthCheckTimeout)
	ticker := time.NewTicker(DefaultHealthCheckInterval)
	defer ticker.Stop()

	// Extract application ID from service labels for refreshing service
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
				// Get final service state for detailed error
				var finalService *swarm.Service
				if applicationID != "" {
					if refreshedService, err := FindServiceByLabel(s.DockerRepo, "com.application.id", applicationID); err == nil && refreshedService != nil {
						finalService = refreshedService
					}
				}
				if finalService == nil {
					finalService = &service
				}
				taskStates := s.getTaskStatesForService(*finalService)
				if taskCtx != nil {
					taskCtx.AddLog(fmt.Sprintf("Timeout waiting for service to start after %d attempts. Task states: %s", checkCount, taskStates))
				}
				return fmt.Errorf("timeout waiting for dev server to start. Task states: %s", taskStates)
			}

			// Refresh service object to get latest task states
			var currentService *swarm.Service
			if applicationID != "" {
				if refreshedService, err := FindServiceByLabel(s.DockerRepo, "com.application.id", applicationID); err == nil && refreshedService != nil {
					currentService = refreshedService
				}
			}
			if currentService == nil {
				currentService = &service
			}

			// Check Docker service health
			running, desired, err := s.DockerRepo.GetServiceHealth(*currentService)
			if err != nil {
				if taskCtx != nil && checkCount%10 == 0 {
					taskCtx.AddLog(fmt.Sprintf("Failed to get Docker service health (attempt %d): %v", checkCount, err))
				}
				continue
			}

			// Log task states periodically for debugging
			if taskCtx != nil && checkCount%10 == 0 {
				taskStates := s.getTaskStatesForService(*currentService)
				taskCtx.AddLog(fmt.Sprintf("Docker health: %d/%d running (attempt %d). Task states: %s", running, desired, checkCount, taskStates))
			}

			// Service is healthy when all desired replicas are running
			if running >= desired && desired > 0 {
				if taskCtx != nil {
					taskCtx.AddLog(fmt.Sprintf("Docker service is running (%d/%d replicas) after %d attempts", running, desired, checkCount))
				}
				return nil
			}
		}
	}
}

// shellJoin joins command parts into a shell-safe string
func shellJoin(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " " + parts[i]
	}
	return result
}

// addDomainToCaddy adds a domain to Caddy proxy
func (s *TaskService) addDomainToCaddy(ctx context.Context, domain string, port int, organizationID uuid.UUID, taskCtx *LiveDevTaskContext) error {
	if taskCtx != nil {
		taskCtx.AddLog(fmt.Sprintf("Adding domain %s to Caddy proxy...", domain))
	}

	client := caddygo.NewClient(config.AppConfig.Proxy.CaddyEndpoint)

	// Get SSH host from organization-specific SSH manager
	upstreamHost, err := GetSSHHostForOrganization(ctx, organizationID)
	if err != nil {
		return err
	}

	if err := client.AddDomainWithAutoTLS(domain, upstreamHost, port, caddygo.DomainOptions{}); err != nil {
		return fmt.Errorf("failed to add domain to caddy: %w", err)
	}

	client.Reload()
	return nil
}

// SetupLiveDevQueue sets up the live dev queue (should be called during initialization)
func (s *TaskService) SetupLiveDevQueue() {
	s.SetupCreateDeploymentQueue()
}
