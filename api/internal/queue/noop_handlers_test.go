package queue

import (
	"context"
	"testing"

	trail_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/stretchr/testify/require"
)

func TestNoopTaskHandlers(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, handleProvisionTrailTask(ctx, trail_types.ProvisionPayload{}))
	require.NoError(t, handleRegisterCustomDomainTask(ctx, CustomDomainPayload{}))
	require.NoError(t, handleRemoveCustomDomainTask(ctx, RemoveCustomDomainPayload{}))
	require.NoError(t, handleResourceUpdateTask(ctx, ResourceUpdatePayload{}))
	require.NoError(t, handleVMDeleteTask(ctx, VMDeletePayload{}))
	require.NoError(t, handleMachineLifecycleTask(ctx, MachineLifecyclePayload{}))
}
