package tests

import (
	"encoding/json"
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMachineStateResponse_JSON(t *testing.T) {
	resp := types.MachineStateResponse{
		Status:  "success",
		Message: "Machine status retrieved",
		Data: &types.MachineState{
			Active:    true,
			State:     "running",
			PID:       1234,
			UptimeSec: 86400,
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded types.MachineStateResponse
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "success", decoded.Status)
	assert.True(t, decoded.Data.Active)
	assert.Equal(t, "running", decoded.Data.State)
	assert.Equal(t, 1234, decoded.Data.PID)
	assert.Equal(t, int64(86400), decoded.Data.UptimeSec)
}

func TestMachineActionResponse_JSON(t *testing.T) {
	resp := types.MachineActionResponse{
		Status:  "success",
		Message: "Machine restarted",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded types.MachineActionResponse
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "success", decoded.Status)
	assert.Equal(t, "Machine restarted", decoded.Message)
}

func TestMachineErrors_Defined(t *testing.T) {
	assert.NotNil(t, types.ErrMachineNotProvisioned)
	assert.NotNil(t, types.ErrMachineOperationTimeout)
	assert.NotNil(t, types.ErrMachineOperationLocked)
	assert.NotNil(t, types.ErrMachineNotRunning)
	assert.NotNil(t, types.ErrMachineAlreadyPaused)
	assert.NotNil(t, types.ErrMachineNotPaused)

	assert.Contains(t, types.ErrMachineNotProvisioned.Error(), "provisioned")
	assert.Contains(t, types.ErrMachineOperationTimeout.Error(), "timed out")
}
