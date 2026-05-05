package queue

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMachineLifecyclePayload_JSON(t *testing.T) {
	payload := MachineLifecyclePayload{
		RequestID:    "abc-123",
		InstanceName: "trail-xyz",
		Action:       "status",
		ServerID:     "srv-1",
		ExpiresAt:    1700000000,
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded MachineLifecyclePayload
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, payload.RequestID, decoded.RequestID)
	assert.Equal(t, payload.InstanceName, decoded.InstanceName)
	assert.Equal(t, payload.Action, decoded.Action)
	assert.Equal(t, payload.ServerID, decoded.ServerID)
	assert.Equal(t, payload.ExpiresAt, decoded.ExpiresAt)
}

func TestMachineLifecycleResult_JSON(t *testing.T) {
	result := MachineLifecycleResult{
		RequestID: "abc-123",
		Success:   true,
		Action:    "status",
		Data:      json.RawMessage(`{"state":"Running"}`),
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded MachineLifecycleResult
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.True(t, decoded.Success)
	assert.Equal(t, "status", decoded.Action)
	assert.JSONEq(t, `{"state":"Running"}`, string(decoded.Data))
}

func TestMachineLifecycleResult_ErrorJSON(t *testing.T) {
	result := MachineLifecycleResult{
		RequestID: "abc-456",
		Success:   false,
		Action:    "pause",
		Error:     "already paused",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded MachineLifecycleResult
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.False(t, decoded.Success)
	assert.Equal(t, "already paused", decoded.Error)
	assert.Nil(t, decoded.Data)
}
