package service

import (
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *DeployService) formatLog(
	applicationID uuid.UUID,
	deploymentID uuid.UUID,
	message string,
	args ...interface{},
) {
	if len(args) > 0 {
		s.addLog(applicationID, fmt.Sprintf(message, args...), deploymentID)
	} else {
		s.addLog(applicationID, message, deploymentID)
	}
}

// sanitizeEnvVars masks sensitive environment variables for logging
func (s *DeployService) sanitizeEnvVars(envVars map[string]string) []string {
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

// prepareContainerConfig creates Docker container configuration
func (s *DeployService) prepareContainerConfig(
	imageName string,
	port nat.Port,
	envVars []string,
	applicationID string,
) container.Config {
	return container.Config{
		Image:    imageName,
		Hostname: "nixopus",
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
		Env: envVars,
		Labels: map[string]string{
			"com.docker.compose.project": "nixopus",
			"com.docker.compose.version": "0.0.1",
			"com.project.name":           imageName,
			"com.application.id":         applicationID,
		},
	}
}

// prepareHostConfig creates Docker host configuration with port bindings
func (s *DeployService) prepareHostConfig(port nat.Port, availablePort string) container.HostConfig {
	return container.HostConfig{
		NetworkMode: "bridge",
		PortBindings: map[nat.Port][]nat.PortBinding{
			port: {
				{
					HostIP:   "0.0.0.0",
					HostPort: availablePort,
				},
			},
		},
		PublishAllPorts: true,
	}
}

func (s *DeployService) getAvailablePort() (string, error) {
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

// prepareNetworkConfig creates Docker network configuration
func (s *DeployService) prepareNetworkConfig() network.NetworkingConfig {
	return network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"bridge": {},
		},
	}
}

func (s *DeployService) getRunningContainers(r DeployerConfig) ([]container.Summary, error) {
	all_containers, err := s.dockerRepo.ListAllContainers()
	if err != nil {
		return nil, types.ErrFailedToListContainers
	}

	var currentContainers []container.Summary
	for _, ctr := range all_containers {
		if ctr.Labels["com.application.id"] == r.application.ID.String() {
			currentContainers = append(currentContainers, ctr)
		}
	}

	s.formatLog(r.application.ID, r.deployment_config.ID, "Found %d running containers", len(currentContainers))
	return currentContainers, nil
}

func (s *DeployService) createContainerConfigs(r DeployerConfig) (container.Config, container.HostConfig, network.NetworkingConfig, string) {
	port_str := fmt.Sprintf("%d", r.application.Port)
	port, _ := nat.NewPort("tcp", port_str)

	var env_vars []string
	for k, v := range GetMapFromString(r.application.EnvironmentVariables) {
		env_vars = append(env_vars, fmt.Sprintf("%s=%s", k, v))
	}

	logEnvVars := s.sanitizeEnvVars(GetMapFromString(r.application.EnvironmentVariables))
	s.formatLog(r.application.ID, r.deployment_config.ID, types.LogEnvironmentVariables, logEnvVars)
	s.formatLog(r.application.ID, r.deployment_config.ID, types.LogContainerExposingPort, port_str)

	container_config := s.prepareContainerConfig(
		fmt.Sprintf("%s:latest", r.application.Name),
		port,
		env_vars,
		r.application.ID.String(),
	)
	availablePort, err := s.getAvailablePort()
	if err != nil {
		s.logger.Log(logger.Error, types.ErrFailedToGetAvailablePort.Error(), err.Error())
		return container.Config{}, container.HostConfig{}, network.NetworkingConfig{}, ""
	}
	host_config := s.prepareHostConfig(port, availablePort)
	network_config := s.prepareNetworkConfig()

	return container_config, host_config, network_config, availablePort
}

// AtomicUpdateContainer performs a zero-downtime update of a running container
func (s *DeployService) AtomicUpdateContainer(r DeployerConfig) (string, string, error) {
	if r.application.Name == "" {
		return "", "", types.ErrMissingImageName
	}

	s.logger.Log(logger.Info, types.LogUpdatingContainer, r.application.Name)
	s.formatLog(r.application.ID, r.deployment_config.ID, types.LogPreparingToUpdateContainer, r.application.Name)
	s.updateStatus(r.deployment_config.ID, shared_types.Deploying, r.appStatus.ID)

	currentContainers, err := s.getRunningContainers(r)
	if err != nil {
		return "", "", err
	}

	container_config, host_config, network_config, availablePort := s.createContainerConfigs(r)

	s.formatLog(r.application.ID, r.deployment_config.ID, types.LogCreatingNewContainer)
	resp, err := s.dockerRepo.CreateContainer(container_config, host_config, network_config, "")
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		return "", "", types.ErrFailedToCreateContainer
	}
	s.formatLog(r.application.ID, r.deployment_config.ID, types.LogNewContainerCreated+"%s", resp.ID)

	for _, ctr := range currentContainers {
		s.formatLog(r.application.ID, r.deployment_config.ID, types.LogStoppingOldContainer+"%s", ctr.ID)
		err = s.dockerRepo.StopContainer(ctr.ID, container.StopOptions{Timeout: intPtr(10)})
		if err != nil {
			s.formatLog(r.application.ID, r.deployment_config.ID, types.LogFailedToStopOldContainer, err.Error())
		}

		// s.formatLog(r.application.ID, r.deployment_config.ID, types.LogRemovingOldContainer+"%s", ctr.ID)
		// err = s.dockerRepo.RemoveContainer(ctr.ID, container.RemoveOptions{Force: true})
		// if err != nil {
		// 	s.formatLog(r.application.ID, r.deployment_config.ID, types.LogFailedToRemoveOldContainer, err.Error())
		// }
	}

	s.formatLog(r.application.ID, r.deployment_config.ID, types.LogStartingNewContainer)
	err = s.dockerRepo.StartContainer(resp.ID, container.StartOptions{})
	if err != nil {
		fmt.Printf("Failed to start container: %v\n", err)
		s.dockerRepo.RemoveContainer(resp.ID, container.RemoveOptions{Force: true})
		return "", "", types.ErrFailedToStartNewContainer
	}
	s.formatLog(r.application.ID, r.deployment_config.ID, types.LogNewContainerStartedSuccessfully)

	time.Sleep(time.Second * 5)

	containerInfo, err := s.dockerRepo.GetContainerById(resp.ID)
	if err != nil || containerInfo.State.Status != "running" {
		s.dockerRepo.StopContainer(resp.ID, container.StopOptions{})
		s.dockerRepo.RemoveContainer(resp.ID, container.RemoveOptions{Force: true})
		return "", "", types.ErrFailedToUpdateContainer
	}

	s.formatLog(r.application.ID, r.deployment_config.ID, types.LogContainerUpdateCompleted)
	s.updateStatus(r.deployment_config.ID, shared_types.Deployed, r.appStatus.ID)

	r.deployment_config.ContainerID = resp.ID
	r.deployment_config.ContainerName = containerInfo.Name
	r.deployment_config.ContainerImage = containerInfo.Image
	r.deployment_config.ContainerStatus = "running"
	r.deployment_config.UpdatedAt = time.Now()

	// here we update the deployment with the container informations
	s.updateDeployment(r.deployment_config)

	log_collection_config := ContainerLogCollection{
		r.application.ID,
		resp.ID,
		r.deployment_config,
	}

	go s.collectContainerLogs(log_collection_config)

	return resp.ID, availablePort, nil
}

// Helper function to create a pointer to an integer
func intPtr(i int) *int {
	return &i
}

func (s *DeployService) RestartContainer(r DeployerConfig) error {
	if r.application.Name == "" {
		return types.ErrMissingImageName
	}

	s.logger.Log(logger.Info, types.LogRestartingContainer, r.application.Name)

	containerInfo, err := s.dockerRepo.GetContainerById(r.deployment_config.ContainerID)
	if err != nil {
		return err
	}

	if containerInfo.State.Status != "running" {
		return types.ErrContainerNotRunning
	}

	err = s.dockerRepo.RestartContainer(r.deployment_config.ContainerID, container.StopOptions{})
	if err != nil {
		return err
	}

	return nil
}
