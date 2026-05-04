package notification

import (
	"testing"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestDefaultChannelsForEvent(t *testing.T) {
	tests := []struct {
		event shared_types.EventType
		want  []string
	}{
		{shared_types.EventLoginAlert, []string{"email"}},
		{shared_types.EventPasswordReset, []string{"email"}},
		{shared_types.EventVerificationEmail, []string{"email"}},
		{shared_types.EventUserAddedToOrg, []string{"email", "slack", "discord"}},
		{shared_types.EventUserRemovedFromOrg, []string{"email", "slack", "discord"}},
		{shared_types.EventDeploySuccess, []string{"slack", "discord"}},
		{shared_types.EventDeployFailed, []string{"slack", "discord", "agent"}},
		{shared_types.EventBuildFailed, []string{"email", "slack", "discord", "agent"}},
		{shared_types.EventHealthCheckCritical, []string{"email", "slack", "discord", "agent"}},
		{shared_types.EventTrialExpired, []string{"system_email"}},
		{shared_types.EventContainerCrashed, []string{"email"}},
	}
	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			assert.Equal(t, tt.want, defaultChannelsForEvent(tt.event))
		})
	}
}
