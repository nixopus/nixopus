package service

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	// "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
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

// RunImage runs a Docker container from the specified image, maps the
// specified port from the container to the host, and sets the specified
// environment variables. The function returns an error if the container
// cannot be started.
func (s *DeployService) RunImage(r RunImageConfig) (string, error) {
	if r.imageName == "" {
		return "", fmt.Errorf("image name is empty")
	}
	s.logger.Log(logger.Info, "Running container from image", r.imageName)
	s.addLog(r.applicationID, fmt.Sprintf("Preparing to run container from image %s", r.imageName), r.deployment_config.ID)
	s.updateStatus(r.applicationID, shared_types.Deploying, r.statusID)

	port, _ := nat.NewPort("tcp", r.port_str)
	var env_vars []string
	for k, v := range r.environment_variables {
		env_vars = append(env_vars, fmt.Sprintf("%s=%s", k, v))
	}

	logEnvVars := make([]string, 0)
	for k, v := range r.environment_variables {
		if containsSensitiveKeyword(k) {
			logEnvVars = append(logEnvVars, fmt.Sprintf("%s=********", k))
		} else {
			logEnvVars = append(logEnvVars, fmt.Sprintf("%s=%s", k, v))
		}
	}
	s.addLog(r.applicationID, fmt.Sprintf("Environment variables: %v", logEnvVars), r.deployment_config.ID)
	s.addLog(r.applicationID, fmt.Sprintf("Container will expose port %s", r.port_str), r.deployment_config.ID)

	// images := s.dockerRepo.ListAllImages(image.ListOptions{})
	// var targetImage string
	// for _, image := range images {
	// 	if image.RepoTags[0] == imageName {
	// 		s.logger.Log(logger.Info, "Image already exists",image.ID)
	// 		targetImage = image.ID
	// 		break
	// 	}
	// }

	container_config := container.Config{
		Image:    r.imageName,
		Hostname: "nixopus",
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
		Env: env_vars,
		Labels: map[string]string{
			"com.docker.compose.project": "nixopus",
			"com.docker.compose.version": "0.0.1",
			"com.project.name":           r.imageName,
			"application.id":             r.applicationID.String(),
		},
	}

	host_config := container.HostConfig{
		NetworkMode: "bridge",
		PortBindings: map[nat.Port][]nat.PortBinding{
			port: {
				{
					HostIP:   "0.0.0.0",
					HostPort: r.port_str,
				},
			},
		},
	}

	network_config := network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"bridge": {},
		},
	}

	s.addLog(r.applicationID, "Creating container...", r.deployment_config.ID)
	resp, err := s.dockerRepo.CreateContainer(container_config, host_config, network_config, r.imageName)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create container: %s", err.Error())
		s.addLog(r.applicationID, errMsg, r.deployment_config.ID)
		return "", errors.New(errMsg)
	}
	s.addLog(r.applicationID, fmt.Sprintf("Container created with ID: %s", resp.ID), r.deployment_config.ID)

	s.addLog(r.applicationID, "Starting container...", r.deployment_config.ID)
	err = s.dockerRepo.StartContainer(resp.ID, container.StartOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to start container: %s", err.Error())
		s.addLog(r.applicationID, errMsg, r.deployment_config.ID)
		return "", errors.New(errMsg)
	}
	s.addLog(r.applicationID, "Container started successfully", r.deployment_config.ID)

	log_collection_config := ContainerLogCollection{
		r.applicationID,
		resp.ID,
		r.deployment_config,
	}

	go s.collectContainerLogs(log_collection_config)

	return resp.ID, nil
}
