package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/audit/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestAuditService(t *testing.T) {
	setup := testutils.NewTestSetup()
	auditService := service.NewAuditService(setup.DB, setup.Ctx, logger.NewLogger())

	user := &types.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := setup.DB.NewInsert().Model(user).Exec(setup.Ctx)
	assert.NoError(t, err)

	org := &types.Organization{
		ID:        uuid.New(),
		Name:      "Test Org",
		CreatedAt: time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(org).Exec(setup.Ctx)
	assert.NoError(t, err)

	resourceID := uuid.New()
	requestID := uuid.New()

	t.Run("LogAction", func(t *testing.T) {
		req := &service.AuditLogRequest{
			UserID:         user.ID,
			OrganizationID: org.ID,
			Action:         types.AuditActionCreate,
			ResourceType:   types.AuditResourceUser,
			ResourceID:     resourceID,
			OldValues:      map[string]any{"name": "old"},
			NewValues:      map[string]any{"name": "new"},
			Metadata:       map[string]any{"key": "value"},
			IPAddress:      "127.0.0.1",
			UserAgent:      "test-agent",
			RequestID:      requestID,
		}

		err := auditService.LogAction(req)
		assert.NoError(t, err)
	})

	t.Run("GetAuditLogs", func(t *testing.T) {
		filters := map[string]interface{}{
			"user_id": user.ID,
		}
		logs, total, err := auditService.GetAuditLogs(filters, 1, 10)
		assert.NoError(t, err)
		assert.Greater(t, total, 0)
		assert.NotEmpty(t, logs)
	})

	t.Run("GetAuditLogsByResource", func(t *testing.T) {
		logs, total, err := auditService.GetAuditLogsByResource(types.AuditResourceUser, resourceID, 1, 10)
		assert.NoError(t, err)
		assert.Greater(t, total, 0)
		assert.NotEmpty(t, logs)
	})

	t.Run("GetAuditLogsByUser", func(t *testing.T) {
		logs, total, err := auditService.GetAuditLogsByUser(user.ID, 1, 10)
		assert.NoError(t, err)
		assert.Greater(t, total, 0)
		assert.NotEmpty(t, logs)
	})

	t.Run("GetAuditLogsByOrganization", func(t *testing.T) {
		logs, total, err := auditService.GetAuditLogsByOrganization(org.ID, 1, 10)
		assert.NoError(t, err)
		assert.Greater(t, total, 0)
		assert.NotEmpty(t, logs)
	})

	t.Run("Pagination", func(t *testing.T) {
		req := &service.AuditLogRequest{
			UserID:         user.ID,
			OrganizationID: org.ID,
			Action:         types.AuditActionUpdate,
			ResourceType:   types.AuditResourceUser,
			ResourceID:     resourceID,
			OldValues:      map[string]any{"name": "old"},
			NewValues:      map[string]any{"name": "new"},
			Metadata:       map[string]any{"key": "value"},
			IPAddress:      "127.0.0.1",
			UserAgent:      "test-agent",
			RequestID:      uuid.New(),
		}
		err := auditService.LogAction(req)
		assert.NoError(t, err)

		page1, total1, err := auditService.GetAuditLogs(map[string]interface{}{}, 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(page1))
		assert.Greater(t, total1, 1)

		page2, total2, err := auditService.GetAuditLogs(map[string]interface{}{}, 2, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(page2))
		assert.Equal(t, total1, total2)
		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})

	t.Run("InvalidFilters", func(t *testing.T) {
		invalidFilters := map[string]interface{}{
			"user_id": uuid.New(),
		}
		logs, total, err := auditService.GetAuditLogs(invalidFilters, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, logs)
	})
}
