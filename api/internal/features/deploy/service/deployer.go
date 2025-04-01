package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type DeployerConfig struct {
	application       shared_types.Application
	deployment        *shared_types.DeploymentRequestConfig
	userID            uuid.UUID
	contextPath       string
	appStatus         shared_types.ApplicationDeploymentStatus
	deployment_config *shared_types.ApplicationDeployment
}

func (s *DeployService) Deployer(d DeployerConfig) error {
	s.addLog(d.application.ID, fmt.Sprintf(types.LogDeploymentBuildPack, d.application.BuildPack), d.deployment_config.ID)
	s.PrerunCommands(d)

	var err error
	switch d.application.BuildPack {
	case shared_types.DockerFile:
		err = s.handleDockerfileDeployment(d)
	case shared_types.DockerCompose:
		err = s.handleDockerComposeDeployment(d)
	case shared_types.Static:
		err = s.handleStaticDeployment(d)
	default:
		return types.ErrInvalidBuildPack
	}

	if err != nil {
		return err
	}

	return s.PostRunCommands(d)
}
