package tasks

import shared_types "github.com/nixopus/nixopus/api/internal/types"

type dockerfilePipelineMode byte

const (
	dockerfilePipelineCreate dockerfilePipelineMode = iota
	dockerfilePipelineReDeploy
	dockerfilePipelineUpdate
)

func (m dockerfilePipelineMode) deploymentType() shared_types.DeploymentType {
	switch m {
	case dockerfilePipelineReDeploy:
		return shared_types.DeploymentTypeReDeploy
	case dockerfilePipelineUpdate:
		return shared_types.DeploymentTypeUpdate
	default:
		return shared_types.DeploymentTypeCreate
	}
}

func dockerfileBuildFlagsForMode(mode dockerfilePipelineMode, payload shared_types.TaskPayload) (force, forceWithoutCache bool) {
	if mode == dockerfilePipelineCreate {
		return false, false
	}
	return payload.UpdateOptions.Force, payload.UpdateOptions.ForceWithoutCache
}
