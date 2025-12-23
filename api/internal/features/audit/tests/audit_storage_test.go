package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/audit/storage"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestAuditStorage(t *testing.T) {
	setup := testutils.NewTestSetup()
	auditStorage := storage.NewAuditStorage(setup.DB, setup.Ctx)

	user := &types.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "hashedpassword",
		Type:      "viewer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := setup.DB.NewInsert().Model(user).Exec(setup.Ctx)
	assert.NoError(t, err)

	org := &types.Organization{
		ID:          uuid.New(),
		Name:        "Test Org",
		Description: "Test Organization",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(org).Exec(setup.Ctx)
	assert.NoError(t, err)

	t.Run("CreateAuditLog", func(t *testing.T) {
		log := &types.AuditLog{
			ID:             uuid.New(),
			UserID:         user.ID,
			OrganizationID: org.ID,
			Action:         types.AuditActionCreate,
			ResourceType:   types.AuditResourceUser,
			ResourceID:     uuid.New(),
			OldValues:      map[string]any{"name": "old"},
			NewValues:      map[string]any{"name": "new"},
			Metadata:       map[string]any{"key": "value"},
			IPAddress:      "127.0.0.1",
			UserAgent:      "test-agent",
			CreatedAt:      time.Now(),
			RequestID:      uuid.New(),
		}

		err := auditStorage.CreateAuditLog(log)
		assert.NoError(t, err)
	})

	t.Run("GetAuditLogs", func(t *testing.T) {
		filters := map[string]interface{}{
			"user_id": user.ID,
		}
		logs, total, err := auditStorage.GetAuditLogs(filters, 1, 10)
		assert.NoError(t, err)
		assert.Greater(t, total, 0)
		assert.NotEmpty(t, logs)

		for _, log := range logs {
			assert.Equal(t, user.ID, log.UserID)
			assert.NotNil(t, log.User)
			assert.NotNil(t, log.Organization)
			assert.Equal(t, user.Email, log.User.Email)
		}
	})

	t.Run("GetAuditLogs_Pagination", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			log := &types.AuditLog{
				ID:             uuid.New(),
				UserID:         user.ID,
				OrganizationID: org.ID,
				Action:         types.AuditActionUpdate,
				ResourceType:   types.AuditResourceUser,
				ResourceID:     uuid.New(),
				CreatedAt:      time.Now(),
				RequestID:      uuid.New(),
			}
			err := auditStorage.CreateAuditLog(log)
			assert.NoError(t, err)
		}

		page1, total1, err := auditStorage.GetAuditLogs(nil, 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(page1))
		assert.Greater(t, total1, 2)

		page2, total2, err := auditStorage.GetAuditLogs(nil, 2, 2)
		assert.NoError(t, err)
		assert.NotEmpty(t, page2)
		assert.Equal(t, total1, total2)
		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})

	t.Run("GetAuditLogs_Ordering", func(t *testing.T) {
		logs, _, err := auditStorage.GetAuditLogs(nil, 1, 10)
		assert.NoError(t, err)
		assert.NotEmpty(t, logs)

		for i := 1; i < len(logs); i++ {
			assert.True(t, logs[i-1].CreatedAt.After(logs[i].CreatedAt) ||
				logs[i-1].CreatedAt.Equal(logs[i].CreatedAt))
		}
	})

	t.Run("GetAuditLogs_NoResults", func(t *testing.T) {
		filters := map[string]interface{}{
			"user_id": uuid.New(),
		}
		logs, total, err := auditStorage.GetAuditLogs(filters, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, logs)
	})
}
