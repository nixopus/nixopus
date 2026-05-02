package types_test

import (
	"encoding/json"
	"testing"

	audittypes "github.com/nixopus/nixopus/api/internal/features/audit/types"
	apitypes "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
)

func TestActivityMessage_JSON_roundTrip(t *testing.T) {
	t.Parallel()
	in := &audittypes.ActivityMessage{
		ID:         "1",
		Message:    "m",
		Action:     apitypes.AuditActionCreate,
		Actor:      "a",
		Resource:   "r",
		ResourceID: "rid",
		Timestamp:  "t",
		Metadata:   map[string]interface{}{"k": 1},
	}
	data, err := json.Marshal(in)
	require.NoError(t, err)
	var out audittypes.ActivityMessage
	require.NoError(t, json.Unmarshal(data, &out))
	require.Equal(t, in.ID, out.ID)
	require.Equal(t, in.Action, out.Action)
}

func TestGetActivitiesResponse_JSON(t *testing.T) {
	t.Parallel()
	in := audittypes.GetActivitiesResponse{
		Status:  "success",
		Message: "ok",
		Data: audittypes.GetActivitiesResponseData{
			Pagination: audittypes.PaginationInfo{
				CurrentPage: 1,
				PageSize:    10,
				TotalCount:  0,
				TotalPages:  0,
			},
		},
	}
	_, err := json.Marshal(in)
	require.NoError(t, err)
}

func TestErrorResponse_JSON(t *testing.T) {
	t.Parallel()
	b, err := json.Marshal(audittypes.ErrorResponse{Status: "error", Error: "e"})
	require.NoError(t, err)
	require.Contains(t, string(b), "error")
}
