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
		t.Run(string(tt.want), func(t *testing.T) {
			if got := tt.mode.deploymentType(); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestDockerfilePipeline_buildFlags_fromPayload(t *testing.T) {
	payload := shared_types.TaskPayload{
		UpdateOptions: shared_types.UpdateOptions{
			Force:             true,
			ForceWithoutCache: true,
		},
	}
	t.Run("create_ignores_update_options", func(t *testing.T) {
		force, noCache := dockerfileBuildFlagsForMode(dockerfilePipelineCreate, payload)
		if force || noCache {
			t.Fatalf("force=%v noCache=%v", force, noCache)
		}
	})
	t.Run("redeploy_uses_update_options", func(t *testing.T) {
		force, noCache := dockerfileBuildFlagsForMode(dockerfilePipelineReDeploy, payload)
		if !force || !noCache {
			t.Fatalf("force=%v noCache=%v", force, noCache)
		}
	})
	t.Run("update_uses_update_options", func(t *testing.T) {
		force, noCache := dockerfileBuildFlagsForMode(dockerfilePipelineUpdate, payload)
		if !force || !noCache {
			t.Fatalf("force=%v noCache=%v", force, noCache)
		}
	})
}
