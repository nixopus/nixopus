package service

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type RunImageConfig struct {
	applicationID         uuid.UUID
	imageName             string
	environment_variables map[string]string
	port_str              string
	statusID              uuid.UUID
	deployment_config     *shared_types.ApplicationDeployment
}

// Generic helper functions for error handling and logging
func (s *DeployService) logAndReturnError(
	applicationID uuid.UUID,
	deploymentID uuid.UUID,
	message string,
	err error,
) (string, error) {
	errMsg := fmt.Sprintf(message, err.Error())
	s.addLog(applicationID, errMsg, deploymentID)
	return "", errors.New(errMsg)
}

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
			"application.id":             applicationID,
		},
	}
}

// prepareHostConfig creates Docker host configuration with port bindings
func (s *DeployService) prepareHostConfig(port nat.Port, portStr string) container.HostConfig {
	return container.HostConfig{
		NetworkMode: "bridge",
		PortBindings: map[nat.Port][]nat.PortBinding{
			port: {
				{
					HostIP:   "0.0.0.0",
					HostPort: portStr,
				},
			},
		},
	}
}

// prepareNetworkConfig creates Docker network configuration
func (s *DeployService) prepareNetworkConfig() network.NetworkingConfig {
	return network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"bridge": {},
		},
	}
}

// RunImage runs a Docker container from the specified image
func (s *DeployService) RunImage(r RunImageConfig) (string, error) {
	if r.imageName == "" {
		return "", types.ErrMissingImageName
	}

	s.logger.Log(logger.Info, types.LogRunningContainerFromImage, r.imageName)
	s.formatLog(r.applicationID, r.deployment_config.ID, types.LogPreparingToRunContainer, r.imageName)
	s.updateStatus(r.applicationID, shared_types.Deploying, r.statusID)

	port, _ := nat.NewPort("tcp", r.port_str)

	var env_vars []string
	for k, v := range r.environment_variables {
		env_vars = append(env_vars, fmt.Sprintf("%s=%s", k, v))
	}

	logEnvVars := s.sanitizeEnvVars(r.environment_variables)
	s.formatLog(r.applicationID, r.deployment_config.ID, types.LogEnvironmentVariables, logEnvVars)
	s.formatLog(r.applicationID, r.deployment_config.ID, types.LogContainerExposingPort, r.port_str)

	container_config := s.prepareContainerConfig(
		r.imageName,
		port,
		env_vars,
		r.applicationID.String(),
	)
	host_config := s.prepareHostConfig(port, r.port_str)
	network_config := s.prepareNetworkConfig()

	s.formatLog(r.applicationID, r.deployment_config.ID, types.LogCreatingContainer)
	resp, err := s.dockerRepo.CreateContainer(container_config, host_config, network_config, r.imageName)
	if err != nil {
		return s.logAndReturnError(r.applicationID, r.deployment_config.ID, types.LogFailedToCreateContainer, err)
	}
	s.formatLog(r.applicationID, r.deployment_config.ID, types.LogContainerCreated, resp.ID)

	s.formatLog(r.applicationID, r.deployment_config.ID, types.LogStartingContainer)
	err = s.dockerRepo.StartContainer(resp.ID, container.StartOptions{})
	if err != nil {
		return s.logAndReturnError(r.applicationID, r.deployment_config.ID, types.LogFailedToStartContainer, err)
	}
	s.formatLog(r.applicationID, r.deployment_config.ID, types.LogContainerStartedSuccessfully)

	log_collection_config := ContainerLogCollection{
		r.applicationID,
		resp.ID,
		r.deployment_config,
	}

	go s.collectContainerLogs(log_collection_config)

	return resp.ID, nil
}
