package docker

import (
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	deploy_types "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// buildImageFromDockerfile builds a Docker image from a Dockerfile in the
// contextPath directory. The Dockerfile is specified by the dockerfile
// parameter, which is relative to the contextPath directory. If force is true,
// the build will be forced, even if the image already exists. The buildArgs
// parameter is a map of build arguments, which are passed to the Dockerfile as
// environment variables. The labels parameter is a map of labels, which are
// applied to the built image. The image_name parameter is the name of the
// built image. The function returns the ID of the built image, or an error if
// the build fails.
func (s *DockerService) buildImageFromDockerfile(contextPath string, dockerfile string, force bool, buildArgs map[string]*string, labels map[string]string, image_name string) (string, error) {
	buildContextTar, err := archive.TarWithOptions(contextPath, &archive.TarOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create tar from build context: %w", err)
	}

	relativeDockerfilePath := filepath.Base(dockerfile)

	s.logger.Log(logger.Info, "Using relative Dockerfile path", relativeDockerfilePath)

	buildOptions := types.ImageBuildOptions{
		Dockerfile:  relativeDockerfilePath,
		Remove:      true,
		Tags:        []string{fmt.Sprintf("%s:latest", image_name)},
		NoCache:     force,
		ForceRemove: force,
		BuildArgs:   buildArgs,
		Labels:      labels,
		BuildID:     uuid.New().String(),
	}

	return s.BuildImage(buildOptions, buildContextTar)
}

// RunImage runs a Docker container from the specified image, maps the
// specified port from the container to the host, and sets the specified
// environment variables. The function returns an error if the container
// cannot be started.
func (s *DockerService) RunImage(imageName string, environment_variables map[string]string, port_str string) error {
	if imageName == "" {
		return fmt.Errorf("image name is empty")
	}

	port, _ := nat.NewPort("tcp", port_str)
	var env_vars []string
	for k, v := range environment_variables {
		env_vars = append(env_vars, fmt.Sprintf("%s=%s", k, v))
	}

	container_config := container.Config{
		Image:    imageName,
		Hostname: "nixopus",
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
		Env: env_vars,
		Labels: map[string]string{
			"com.docker.compose.project": "nixopus",
			"com.docker.compose.version": "0.0.1",
		},
	}

	host_config := container.HostConfig{
		NetworkMode: "bridge",
		PortBindings: map[nat.Port][]nat.PortBinding{
			port: {
				{
					HostIP:   "0.0.0.0",
					HostPort: port_str,
				},
			},
		},
	}

	network_config := network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"bridge": {}, 
		},
	}

	resp, err := s.Cli.ContainerCreate(s.Ctx, &container_config, &host_config, &network_config, nil, imageName)
	if err != nil {
		return err
	}

	err = s.Cli.ContainerStart(s.Ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	return nil
}

// CreateDeployment creates a new Docker deployment for the given
// deployment request. It creates a new Docker image based on the specified
// build pack and environment variables, and then starts a new container from
// the image. The container is configured to listen on the specified port and
// to use the specified domain. The function returns an error if the
// deployment cannot be created.
func (s *DockerService) CreateDeployment(deployment *deploy_types.CreateDeploymentRequest, userID uuid.UUID, contextPath string) error {
	s.logger.Log(logger.Info, "Creating deployment", contextPath)

	switch deployment.BuildPack {
	case shared_types.DockerFile:
		s.logger.Log(logger.Info, "Dockerfile building", "")

		buildArgs := make(map[string]*string)
		for k, v := range deployment.BuildVariables {
			value := v
			buildArgs[k] = &value
		}

		labels := make(map[string]string)
		for k, v := range deployment.EnvironmentVariables {
			labels[k] = v
		}

		dockerfilePath := "Dockerfile"

		s.logger.Log(logger.Info, "Build context path", contextPath)
		s.logger.Log(logger.Info, "Using Dockerfile", dockerfilePath)

		_, err := s.buildImageFromDockerfile(
			contextPath,
			dockerfilePath,
			false,
			buildArgs,
			labels,
			deployment.Name,
		)
		if err != nil {
			return fmt.Errorf("failed to build Docker image: %w", err)
		}

		s.logger.Log(logger.Info, "Dockerfile built successfully", "")

		err = s.RunImage(deployment.Name, deployment.EnvironmentVariables, fmt.Sprintf("%d", deployment.Port))
		if err != nil {
			return fmt.Errorf("failed to run Docker image: %w", err)
		}

	case shared_types.DockerCompose:
		s.logger.Log(logger.Info, "Docker compose building", "")
		return nil
	}

	return nil
}
