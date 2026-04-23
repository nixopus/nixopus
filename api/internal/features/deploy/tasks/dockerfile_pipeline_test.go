package tasks

import (
	"testing"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func TestDockerfilePipelineMode_deploymentType(t *testing.T) {
	tests := []struct {
		mode dockerfilePipelineMode
		want shared_types.DeploymentType
	}{
		{dockerfilePipelineCreate, shared_types.DeploymentTypeCreate},
		{dockerfilePipelineReDeploy, shared_types.DeploymentTypeReDeploy},
		{dockerfilePipelineUpdate, shared_types.DeploymentTypeUpdate},
	}
	for _, tt := range tests {
		if got := tt.mode.deploymentType(); got != tt.want {
			t.Fatalf("mode %v: got %v want %v", tt.mode, got, tt.want)
		}
	}
}

func TestDockerfilePipeline_buildFlags_fromPayload(t *testing.T) {
	payload := shared_types.TaskPayload{
		UpdateOptions: shared_types.UpdateOptions{
			Force:             true,
			ForceWithoutCache: true,
		},
	}
	force, noCache := dockerfileBuildFlagsForMode(dockerfilePipelineCreate, payload)
	if force || noCache {
		t.Fatalf("create mode must ignore UpdateOptions: force=%v noCache=%v", force, noCache)
	}
	force, noCache = dockerfileBuildFlagsForMode(dockerfilePipelineReDeploy, payload)
	if !force || !noCache {
		t.Fatalf("redeploy must pass UpdateOptions through: force=%v noCache=%v", force, noCache)
	}
	force, noCache = dockerfileBuildFlagsForMode(dockerfilePipelineUpdate, payload)
	if !force || !noCache {
		t.Fatalf("update must pass UpdateOptions through: force=%v noCache=%v", force, noCache)
	}
}
