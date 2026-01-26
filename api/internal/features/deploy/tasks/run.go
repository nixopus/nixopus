package tasks

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/swarm"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type AtomicUpdateContainerResult struct {
	ContainerID     string
	ContainerName   string
	ContainerImage  string
	ContainerStatus string
	UpdatedAt       time.Time
	AvailablePort   string
}

func (s *TaskService) formatLog(
	taskContext *TaskContext,
	message string,
	args ...interface{},
) {
	if len(args) > 0 {
		formattedMessage := fmt.Sprintf(message, args...)
		s.Logger.Log(logger.Info, formattedMessage, taskContext.GetDeploymentID().String())
		taskContext.AddLog(formattedMessage)
	} else {
		s.Logger.Log(logger.Info, message, taskContext.GetDeploymentID().String())
		taskContext.AddLog(message)
	}
}

// sanitizeEnvVars masks sensitive environment variables for logging
func (s *TaskService) sanitizeEnvVars(envVars map[string]string) []string {
	logEnvVars := make([]string, 0, len(envVars))

	for k, v := range envVars {
		if containsSensitiveKeyword(k) {
			logEnvVars = append(logEnvVars, fmt.Sprintf("%s=********", k))
		} else {
			logEnvVars = append(logEnvVars, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return logEnvVars
}

func (s *TaskService) getAvailablePort() (string, error) {
	manager := ssh.GetSSHManager()
	client, err := manager.Connect()
	if err != nil {
		return "", err
	}
	defer client.Close()

	generatePorts := "seq 49152 65535"

	getUsedPorts := "command -v ss >/dev/null 2>&1 && ss -tan | awk '{print $4}' | cut -d':' -f2 | grep '[0-9]\\{1,5\\}' | sort -u || netstat -tan | awk '{print $4}' | grep ':[0-9]' | cut -d':' -f2 | sort -u"

	cmd := fmt.Sprintf("comm -23 <(%s) <(%s) | sort -R | head -n 1 | tr -d '\\n'", generatePorts, getUsedPorts)

	output, err := client.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to find available port: %w", err)
	}

	port := string(output)
	if port == "" {
		return "", fmt.Errorf("no available ports found in range 49152-65535")
	}

	return port, nil
}

// AtomicUpdateContainer performs a zero-downtime update of a running container
func (s *TaskService) AtomicUpdateContainer(r shared_types.TaskPayload, taskContext *TaskContext) (AtomicUpdateContainerResult, error) {
	if r.Application.Name == "" {
		return AtomicUpdateContainerResult{}, types.ErrMissingImageName
	}

	taskContext.LogAndUpdateStatus("Starting container update", shared_types.Deploying)

	s.Logger.Log(logger.Info, types.LogUpdatingContainer, r.Application.Name)
	s.formatLog(taskContext, types.LogPreparingToUpdateContainer, r.Application.Name)

	// Check if service already exists
	existingService, err := s.getExistingService(r, taskContext)
	if err != nil {
		s.formatLog(taskContext, "No existing service found, creating new service", "")
	}

	// Create service spec
	serviceSpec, availablePort := s.createServiceSpec(r, taskContext)
	if availablePort == "" {
		taskContext.LogAndUpdateStatus("Failed to get available port", shared_types.Failed)
		return AtomicUpdateContainerResult{}, types.ErrFailedToGetAvailablePort
	}

	// Create or update service using shared utility
	serviceID, err := CreateOrUpdateService(s.DockerRepo, serviceSpec, existingService)
	if err != nil {
		taskContext.LogAndUpdateStatus("Failed to create/update service: "+err.Error(), shared_types.Failed)
		return AtomicUpdateContainerResult{}, err
	}
	if existingService != nil {
		s.formatLog(taskContext, "Service updated successfully: %s", serviceID)
	} else {
		s.formatLog(taskContext, "Service created successfully: %s", serviceID)
	}

	// Wait for service to be ready with retries
	serviceInfo, err := s.waitForServiceHealthy(r, taskContext, 120*time.Second, 5*time.Second)
	if err != nil {
		taskContext.LogAndUpdateStatus("Service health check failed: "+err.Error(), shared_types.Failed)
		return AtomicUpdateContainerResult{}, types.ErrFailedToUpdateContainer
	}

	taskContext.LogAndUpdateStatus("Service update completed successfully", shared_types.Deployed)

	// Update deployment record
	r.ApplicationDeployment.ContainerID = serviceInfo.ID
	r.ApplicationDeployment.ContainerName = serviceInfo.Spec.Annotations.Name
	r.ApplicationDeployment.ContainerImage = serviceInfo.Spec.TaskTemplate.ContainerSpec.Image
	r.ApplicationDeployment.ContainerStatus = "running"
	r.ApplicationDeployment.UpdatedAt = time.Now()

	taskContext.UpdateDeployment(&r.ApplicationDeployment)

	return AtomicUpdateContainerResult{
		ContainerID:     serviceInfo.ID,
		ContainerName:   serviceInfo.Spec.Annotations.Name,
		ContainerImage:  serviceInfo.Spec.TaskTemplate.ContainerSpec.Image,
		ContainerStatus: "running",
		UpdatedAt:       time.Now(),
		AvailablePort:   availablePort,
	}, nil
}

// getExistingService finds an existing swarm service for the application
func (s *TaskService) getExistingService(r shared_types.TaskPayload, taskContext *TaskContext) (*swarm.Service, error) {
	return FindServiceByName(s.DockerRepo, r.Application.Name)
}

// FindServiceByName finds a service by name (shared utility for both deploy and devrunner)
func FindServiceByName(dockerRepo docker.DockerRepository, serviceName string) (*swarm.Service, error) {
	services, err := dockerRepo.GetClusterServices()
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if service.Spec.Annotations.Name == serviceName {
			return &service, nil
		}
	}
	return nil, nil
}

// FindServiceByLabel finds a service by label key-value pair (shared utility)
func FindServiceByLabel(dockerRepo docker.DockerRepository, labelKey, labelValue string) (*swarm.Service, error) {
	services, err := dockerRepo.GetClusterServices()
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if service.Spec.Labels != nil && service.Spec.Labels[labelKey] == labelValue {
			return &service, nil
		}
	}
	return nil, nil
}

// CreateOrUpdateService creates or updates a service (shared utility)
func CreateOrUpdateService(dockerRepo docker.DockerRepository, serviceSpec swarm.ServiceSpec, existingService *swarm.Service) (string, error) {
	if existingService != nil {
		// Update existing service
		err := dockerRepo.UpdateService(existingService.ID, serviceSpec, "")
		if err != nil {
			return "", fmt.Errorf("failed to update service: %w", err)
		}
		return existingService.ID, nil
	} else {
		// Create new service
		err := dockerRepo.CreateService(swarm.Service{
			Spec: serviceSpec,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create service: %w", err)
		}
		// Get created service ID by name
		createdService, err := FindServiceByName(dockerRepo, serviceSpec.Annotations.Name)
		if err != nil || createdService == nil {
			return "", fmt.Errorf("failed to get created service: %w", err)
		}
		return createdService.ID, nil
	}
}

// createServiceSpec creates a swarm service specification
func (s *TaskService) createServiceSpec(r shared_types.TaskPayload, taskContext *TaskContext) (swarm.ServiceSpec, string) {
	availablePort, err := s.getAvailablePort()
	if err != nil {
		taskContext.LogAndUpdateStatus("Failed to get available port: "+err.Error(), shared_types.Failed)
		return swarm.ServiceSpec{}, ""
	}

	var env_vars []string
	for k, v := range GetMapFromString(r.Application.EnvironmentVariables) {
		env_vars = append(env_vars, fmt.Sprintf("%s=%s", k, v))
	}

	replicas := uint64(1)
	port, _ := strconv.Atoi(availablePort)

	serviceSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: r.Application.Name,
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image: fmt.Sprintf("%s:latest", r.Application.Name),
				Env:   env_vars,
				Labels: map[string]string{
					"com.application.id": r.Application.ID.String(),
				},
			},
			RestartPolicy: &swarm.RestartPolicy{
				Condition: swarm.RestartPolicyConditionAny,
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Mode: swarm.ResolutionModeVIP,
			Ports: []swarm.PortConfig{
				{
					Protocol:      swarm.PortConfigProtocolTCP,
					TargetPort:    uint32(r.Application.Port),
					PublishedPort: uint32(port),
					PublishMode:   swarm.PortConfigPublishModeHost,
				},
			},
		},
	}

	return serviceSpec, availablePort
}

// getServiceInfo retrieves service information
func (s *TaskService) getServiceInfo(r shared_types.TaskPayload, taskContext *TaskContext) (swarm.Service, error) {
	service, err := FindServiceByName(s.DockerRepo, r.Application.Name)
	if err != nil {
		return swarm.Service{}, err
	}
	if service == nil {
		return swarm.Service{}, fmt.Errorf("service not found: %s", r.Application.Name)
	}
	return *service, nil
}

// waitForServiceHealthy polls the service health until all replicas are running or timeout
func (s *TaskService) waitForServiceHealthy(r shared_types.TaskPayload, taskContext *TaskContext, timeout, pollInterval time.Duration) (swarm.Service, error) {
	deadline := time.Now().Add(timeout)

	s.formatLog(taskContext, "Waiting for service to become healthy (timeout: %s)", timeout)

	for time.Now().Before(deadline) {
		serviceInfo, err := s.getServiceInfo(r, taskContext)
		if err != nil {
			s.formatLog(taskContext, "Failed to get service info, retrying: %s", err.Error())
			time.Sleep(pollInterval)
			continue
		}

		// Check if service has replicated mode configured
		if serviceInfo.Spec.Mode.Replicated == nil || serviceInfo.Spec.Mode.Replicated.Replicas == nil {
			s.formatLog(taskContext, "Service is not in replicated mode, skipping health check")
			return serviceInfo, nil
		}

		desiredReplicas := int(*serviceInfo.Spec.Mode.Replicated.Replicas)
		running, _, err := s.DockerRepo.GetServiceHealth(serviceInfo)
		if err != nil {
			s.formatLog(taskContext, "Failed to get service health, retrying: %s", err.Error())
			time.Sleep(pollInterval)
			continue
		}

		// Get detailed task states for debugging
		taskStates := s.getTaskStatesForService(serviceInfo)
		s.formatLog(taskContext, "Service health: %d/%d running, task states: %s", running, desiredReplicas, taskStates)

		if running >= desiredReplicas {
			s.formatLog(taskContext, "Service is healthy: %d/%d replicas running", running, desiredReplicas)
			return serviceInfo, nil
		}

		time.Sleep(pollInterval)
	}

	// Final attempt to get detailed error info
	serviceInfo, _ := s.getServiceInfo(r, taskContext)
	taskStates := s.getTaskStatesForService(serviceInfo)
	return swarm.Service{}, fmt.Errorf("timeout waiting for service to become healthy, task states: %s", taskStates)
}

// getTaskStatesForService returns a summary of task states for debugging
func (s *TaskService) getTaskStatesForService(service swarm.Service) string {
	tasks, err := s.DockerRepo.GetClusterTasks()
	if err != nil {
		return fmt.Sprintf("failed to get tasks: %s", err.Error())
	}

	stateCounts := make(map[string]int)
	var latestError string

	for _, task := range tasks {
		if task.ServiceID != service.ID {
			continue
		}
		state := string(task.Status.State)
		stateCounts[state]++

		// Capture the latest error message from failed tasks
		if task.Status.State == swarm.TaskStateFailed ||
			task.Status.State == swarm.TaskStateRejected ||
			task.Status.State == swarm.TaskStateShutdown {
			if task.Status.Err != "" {
				latestError = task.Status.Err
			}
		}
	}

	if len(stateCounts) == 0 {
		return "no tasks found"
	}

	var states []string
	for state, count := range stateCounts {
		states = append(states, fmt.Sprintf("%s=%d", state, count))
	}

	result := strings.Join(states, ", ")
	if latestError != "" {
		result += fmt.Sprintf(" (latest error: %s)", latestError)
	}

	return result
}

// containsSensitiveKeyword checks if a key likely contains sensitive information
func containsSensitiveKeyword(key string) bool {
	sensitiveKeywords := []string{
		"password", "secret", "token", "key", "auth", "credential", "private",
	}

	lowerKey := strings.ToLower(key)
	for _, word := range sensitiveKeywords {
		if strings.Contains(lowerKey, word) {
			return true
		}
	}

	return false
}
