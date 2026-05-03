package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewNotificationService
// ---------------------------------------------------------------------------

func TestNewNotificationService_ReturnsWiredService(t *testing.T) {
	repo := &mockNotificationRepo{}
	l := logger.NewLogger()
	ctx := context.Background()

	svc := NewNotificationService(nil, ctx, l, repo)

	require.NotNil(t, svc)
	assert.Equal(t, ctx, svc.Ctx)
}

// ---------------------------------------------------------------------------
// GetPreferences
// ---------------------------------------------------------------------------

func TestGetPreferences_Success(t *testing.T) {
	userID := uuid.New()
	expected := &notification.GetPreferencesResponse{
		Activity: []notification.PreferenceType{{ID: "team-updates"}},
	}
	repo := &mockNotificationRepo{
		getPreferences: func(ctx context.Context, id uuid.UUID) (*notification.GetPreferencesResponse, error) {
			assert.Equal(t, userID, id)
			return expected, nil
		},
	}
	svc := newTestService(repo)

	got, err := svc.GetPreferences(userID)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestGetPreferences_StorageError(t *testing.T) {
	storageErr := errors.New("db unavailable")
	repo := &mockNotificationRepo{
		getPreferences: func(_ context.Context, _ uuid.UUID) (*notification.GetPreferencesResponse, error) {
			return nil, storageErr
		},
	}
	svc := newTestService(repo)

	_, err := svc.GetPreferences(uuid.New())

	assert.ErrorIs(t, err, storageErr)
}

// ---------------------------------------------------------------------------
// AddSmtp
// ---------------------------------------------------------------------------

func TestAddSmtp_ExistingConfig_ReturnsAlreadyExistsError(t *testing.T) {
	existing := &shared_types.SMTPConfigs{ID: uuid.New()}
	repo := &mockNotificationRepo{
		getSmtp: func(_ string) (*shared_types.SMTPConfigs, error) {
			return existing, nil
		},
	}
	svc := newTestService(repo)

	err := svc.AddSmtp(notification.CreateSMTPConfigRequest{
		Host: "smtp.example.com", Port: 587, Username: "user@example.com", Password: "pass",
		OrganizationID: uuid.New(),
	}, uuid.New())

	assert.ErrorIs(t, err, notification.ErrSmtpAlreadyExists)
}

func TestAddSmtp_GetSmtpErrorIgnored_ProceedsToAdd(t *testing.T) {
	repo := &mockNotificationRepo{
		getSmtp: func(_ string) (*shared_types.SMTPConfigs, error) {
			return nil, errors.New("lookup failed")
		},
		addSmtp: func(_ *shared_types.SMTPConfigs) error {
			return nil
		},
	}
	svc := newTestService(repo)

	err := svc.AddSmtp(notification.CreateSMTPConfigRequest{
		Host: "smtp.example.com", Port: 587, Username: "user@example.com", Password: "pass",
		OrganizationID: uuid.New(),
	}, uuid.New())

	require.NoError(t, err)
}

func TestAddSmtp_Success(t *testing.T) {
	userID := uuid.New()
	var capturedConfig *shared_types.SMTPConfigs
	repo := &mockNotificationRepo{
		getSmtp: func(_ string) (*shared_types.SMTPConfigs, error) {
			return nil, nil
		},
		addSmtp: func(c *shared_types.SMTPConfigs) error {
			capturedConfig = c
			return nil
		},
	}
	svc := newTestService(repo)

	err := svc.AddSmtp(notification.CreateSMTPConfigRequest{
		Host: "smtp.example.com", Port: 587, Username: "user@example.com", Password: "pass",
		FromName: "Nixopus", FromEmail: "noreply@example.com",
		OrganizationID: uuid.New(),
	}, userID)

	require.NoError(t, err)
	require.NotNil(t, capturedConfig)
	assert.Equal(t, userID, capturedConfig.UserID)
}

func TestAddSmtp_StorageAddError(t *testing.T) {
	storageErr := errors.New("insert failed")
	repo := &mockNotificationRepo{
		getSmtp: func(_ string) (*shared_types.SMTPConfigs, error) {
			return nil, nil
		},
		addSmtp: func(_ *shared_types.SMTPConfigs) error {
			return storageErr
		},
	}
	svc := newTestService(repo)

	err := svc.AddSmtp(notification.CreateSMTPConfigRequest{
		Host: "smtp.example.com", Port: 587, Username: "user@example.com", Password: "pass",
		OrganizationID: uuid.New(),
	}, uuid.New())

	assert.ErrorIs(t, err, storageErr)
}

// ---------------------------------------------------------------------------
// DeleteSmtp
// ---------------------------------------------------------------------------

func TestDeleteSmtp_Success(t *testing.T) {
	const smtpID = "some-id"
	var called string
	repo := &mockNotificationRepo{
		deleteSmtp: func(id string) error {
			called = id
			return nil
		},
	}
	svc := newTestService(repo)

	err := svc.DeleteSmtp(smtpID)

	require.NoError(t, err)
	assert.Equal(t, smtpID, called)
}

func TestDeleteSmtp_StorageError(t *testing.T) {
	storageErr := errors.New("delete failed")
	repo := &mockNotificationRepo{
		deleteSmtp: func(_ string) error { return storageErr },
	}
	svc := newTestService(repo)

	err := svc.DeleteSmtp("id")

	assert.ErrorIs(t, err, storageErr)
}

// ---------------------------------------------------------------------------
// GetSmtp
// ---------------------------------------------------------------------------

func TestGetSmtp_GetSmtpReturnsError(t *testing.T) {
	storageErr := errors.New("query failed")
	repo := &mockNotificationRepo{
		getSmtp: func(_ string) (*shared_types.SMTPConfigs, error) {
			return nil, storageErr
		},
	}
	svc := newTestService(repo)

	_, err := svc.GetSmtp("user-id", "org-id")

	assert.ErrorIs(t, err, storageErr)
}

func TestGetSmtp_GetSmtpReturnsConfig_EarlyReturn(t *testing.T) {
	expected := &shared_types.SMTPConfigs{ID: uuid.New(), Host: "smtp.example.com"}
	repo := &mockNotificationRepo{
		getSmtp: func(_ string) (*shared_types.SMTPConfigs, error) {
			return expected, nil
		},
	}
	svc := newTestService(repo)

	got, err := svc.GetSmtp("user-id", "org-id")

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestGetSmtp_FallsBackToOrgSmtp_OrgError_ReturnsSMTPConfigNotFound(t *testing.T) {
	repo := &mockNotificationRepo{
		getSmtp: func(_ string) (*shared_types.SMTPConfigs, error) {
			return nil, nil
		},
		getOrganizationsSmtp: func(_ string) ([]shared_types.SMTPConfigs, error) {
			return nil, errors.New("org lookup failed")
		},
	}
	svc := newTestService(repo)

	_, err := svc.GetSmtp("user-id", "org-id")

	assert.ErrorIs(t, err, notification.ErrSMTPConfigNotFound)
}

func TestGetSmtp_FallsBackToOrgSmtp_EmptySlice_ReturnsSMTPConfigNotFound(t *testing.T) {
	repo := &mockNotificationRepo{
		getSmtp: func(_ string) (*shared_types.SMTPConfigs, error) {
			return nil, nil
		},
		getOrganizationsSmtp: func(_ string) ([]shared_types.SMTPConfigs, error) {
			return []shared_types.SMTPConfigs{}, nil
		},
	}
	svc := newTestService(repo)

	_, err := svc.GetSmtp("user-id", "org-id")

	assert.ErrorIs(t, err, notification.ErrSMTPConfigNotFound)
}

func TestGetSmtp_FallsBackToOrgSmtp_ReturnsFirstConfig(t *testing.T) {
	first := shared_types.SMTPConfigs{ID: uuid.New(), Host: "org-smtp.example.com"}
	second := shared_types.SMTPConfigs{ID: uuid.New(), Host: "second.example.com"}
	repo := &mockNotificationRepo{
		getSmtp: func(_ string) (*shared_types.SMTPConfigs, error) {
			return nil, nil
		},
		getOrganizationsSmtp: func(_ string) ([]shared_types.SMTPConfigs, error) {
			return []shared_types.SMTPConfigs{first, second}, nil
		},
	}
	svc := newTestService(repo)

	got, err := svc.GetSmtp("user-id", "org-id")

	require.NoError(t, err)
	assert.Equal(t, first.ID, got.ID)
}

// ---------------------------------------------------------------------------
// UpdateSmtp
// ---------------------------------------------------------------------------

func TestUpdateSmtp_NilStorage_ReturnsError(t *testing.T) {
	l := logger.NewLogger()
	svc := &NotificationService{
		storage: nil,
		Ctx:     context.Background(),
		logger:  l,
	}

	err := svc.UpdateSmtp(notification.UpdateSMTPConfigRequest{ID: uuid.New()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "storage layer not initialized")
}

func TestUpdateSmtp_HostNil_Success(t *testing.T) {
	repo := &mockNotificationRepo{
		updateSmtp: func(_ *notification.UpdateSMTPConfigRequest) error { return nil },
	}
	svc := newTestService(repo)

	err := svc.UpdateSmtp(notification.UpdateSMTPConfigRequest{ID: uuid.New()})

	require.NoError(t, err)
}

func TestUpdateSmtp_HostSet_Success(t *testing.T) {
	host := "new-smtp.example.com"
	repo := &mockNotificationRepo{
		updateSmtp: func(c *notification.UpdateSMTPConfigRequest) error {
			assert.Equal(t, host, *c.Host)
			return nil
		},
	}
	svc := newTestService(repo)

	err := svc.UpdateSmtp(notification.UpdateSMTPConfigRequest{ID: uuid.New(), Host: &host})

	require.NoError(t, err)
}

func TestUpdateSmtp_StorageError_WrapsError(t *testing.T) {
	storageErr := errors.New("update failed")
	repo := &mockNotificationRepo{
		updateSmtp: func(_ *notification.UpdateSMTPConfigRequest) error { return storageErr },
	}
	svc := newTestService(repo)

	err := svc.UpdateSmtp(notification.UpdateSMTPConfigRequest{ID: uuid.New()})

	require.Error(t, err)
	assert.ErrorIs(t, err, storageErr)
	assert.Contains(t, err.Error(), "failed to update SMTP configuration")
}

// ---------------------------------------------------------------------------
// UpdatePreference
// ---------------------------------------------------------------------------

func TestUpdatePreference_EmptyCategory_ReturnsError(t *testing.T) {
	svc := newTestService(&mockNotificationRepo{})

	err := svc.UpdatePreference(notification.UpdatePreferenceRequest{
		Category: "", Type: "login-alerts", Enabled: true,
	}, uuid.New())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "category cannot be empty")
}

func TestUpdatePreference_EmptyType_ReturnsError(t *testing.T) {
	svc := newTestService(&mockNotificationRepo{})

	err := svc.UpdatePreference(notification.UpdatePreferenceRequest{
		Category: "security", Type: "", Enabled: true,
	}, uuid.New())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "type cannot be empty")
}

func TestUpdatePreference_NilUserID_ReturnsError(t *testing.T) {
	svc := newTestService(&mockNotificationRepo{})

	err := svc.UpdatePreference(notification.UpdatePreferenceRequest{
		Category: "security", Type: "login-alerts", Enabled: true,
	}, uuid.Nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "userID cannot be empty")
}

func TestUpdatePreference_StorageError_WrapsError(t *testing.T) {
	storageErr := errors.New("update failed")
	repo := &mockNotificationRepo{
		updatePreference: func(_ context.Context, _ notification.UpdatePreferenceRequest, _ uuid.UUID) error {
			return storageErr
		},
	}
	svc := newTestService(repo)

	err := svc.UpdatePreference(notification.UpdatePreferenceRequest{
		Category: "security", Type: "login-alerts", Enabled: true,
	}, uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, storageErr)
	assert.Contains(t, err.Error(), "failed to update preference")
}

func TestUpdatePreference_Success(t *testing.T) {
	userID := uuid.New()
	var capturedReq notification.UpdatePreferenceRequest
	var capturedUserID uuid.UUID
	repo := &mockNotificationRepo{
		updatePreference: func(_ context.Context, req notification.UpdatePreferenceRequest, id uuid.UUID) error {
			capturedReq = req
			capturedUserID = id
			return nil
		},
	}
	svc := newTestService(repo)

	err := svc.UpdatePreference(notification.UpdatePreferenceRequest{
		Category: "activity", Type: "team-updates", Enabled: false,
	}, userID)

	require.NoError(t, err)
	assert.Equal(t, "activity", capturedReq.Category)
	assert.Equal(t, userID, capturedUserID)
}

// ---------------------------------------------------------------------------
// CreateWebhookConfig
// ---------------------------------------------------------------------------

func TestCreateWebhookConfig_Success(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	var capturedConfig *shared_types.WebhookConfig
	repo := &mockNotificationRepo{
		createWebhookConfig: func(_ context.Context, c *shared_types.WebhookConfig) error {
			capturedConfig = c
			return nil
		},
	}
	svc := newTestService(repo)

	got, err := svc.CreateWebhookConfig(context.Background(), &notification.CreateWebhookConfigRequest{
		Type: "slack", WebhookURL: "https://hooks.slack.com/test",
	}, userID, orgID)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "slack", got.Type)
	assert.Equal(t, "https://hooks.slack.com/test", got.WebhookURL)
	assert.True(t, got.IsActive)
	assert.Equal(t, fmt.Sprintf("slack:%s", orgID.String()), got.ChannelID)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, orgID, got.OrganizationID)
	assert.NotEqual(t, uuid.Nil, got.ID)
	assert.Equal(t, capturedConfig, got)
}

func TestCreateWebhookConfig_StorageError(t *testing.T) {
	storageErr := errors.New("insert failed")
	repo := &mockNotificationRepo{
		createWebhookConfig: func(_ context.Context, _ *shared_types.WebhookConfig) error {
			return storageErr
		},
	}
	svc := newTestService(repo)

	_, err := svc.CreateWebhookConfig(context.Background(), &notification.CreateWebhookConfigRequest{
		Type: "slack", WebhookURL: "https://hooks.slack.com/test",
	}, uuid.New(), uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, storageErr)
	assert.Contains(t, err.Error(), "failed to create webhook config")
}

// ---------------------------------------------------------------------------
// UpdateWebhookConfig
// ---------------------------------------------------------------------------

func TestUpdateWebhookConfig_GetError(t *testing.T) {
	storageErr := errors.New("not found")
	repo := &mockNotificationRepo{
		getWebhookConfig: func(_ context.Context, _ string, _ uuid.UUID) (*shared_types.WebhookConfig, error) {
			return nil, storageErr
		},
	}
	svc := newTestService(repo)

	_, err := svc.UpdateWebhookConfig(context.Background(), &notification.UpdateWebhookConfigRequest{
		Type: "slack",
	}, uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, storageErr)
	assert.Contains(t, err.Error(), "webhook config not found")
}

func TestUpdateWebhookConfig_UpdatesURLAndIsActive(t *testing.T) {
	existing := &shared_types.WebhookConfig{ID: uuid.New(), Type: "slack", IsActive: false}
	newURL := "https://hooks.slack.com/new"
	active := true
	repo := &mockNotificationRepo{
		getWebhookConfig: func(_ context.Context, _ string, _ uuid.UUID) (*shared_types.WebhookConfig, error) {
			return existing, nil
		},
		updateWebhookConfig: func(_ context.Context, c *shared_types.WebhookConfig) error {
			assert.Equal(t, newURL, c.WebhookURL)
			assert.True(t, c.IsActive)
			return nil
		},
	}
	svc := newTestService(repo)

	got, err := svc.UpdateWebhookConfig(context.Background(), &notification.UpdateWebhookConfigRequest{
		Type: "slack", WebhookURL: &newURL, IsActive: &active,
	}, uuid.New())

	require.NoError(t, err)
	assert.Equal(t, newURL, got.WebhookURL)
	assert.True(t, got.IsActive)
}

func TestUpdateWebhookConfig_NilOptionalFields_NoChange(t *testing.T) {
	existing := &shared_types.WebhookConfig{
		ID: uuid.New(), Type: "discord", WebhookURL: "https://discord.com/original", IsActive: true,
	}
	repo := &mockNotificationRepo{
		getWebhookConfig: func(_ context.Context, _ string, _ uuid.UUID) (*shared_types.WebhookConfig, error) {
			return existing, nil
		},
	}
	svc := newTestService(repo)

	got, err := svc.UpdateWebhookConfig(context.Background(), &notification.UpdateWebhookConfigRequest{
		Type: "discord",
	}, uuid.New())

	require.NoError(t, err)
	assert.Equal(t, "https://discord.com/original", got.WebhookURL)
	assert.True(t, got.IsActive)
}

func TestUpdateWebhookConfig_UpdateStorageError(t *testing.T) {
	storageErr := errors.New("update failed")
	repo := &mockNotificationRepo{
		getWebhookConfig: func(_ context.Context, _ string, _ uuid.UUID) (*shared_types.WebhookConfig, error) {
			return &shared_types.WebhookConfig{ID: uuid.New()}, nil
		},
		updateWebhookConfig: func(_ context.Context, _ *shared_types.WebhookConfig) error {
			return storageErr
		},
	}
	svc := newTestService(repo)

	_, err := svc.UpdateWebhookConfig(context.Background(), &notification.UpdateWebhookConfigRequest{
		Type: "slack",
	}, uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, storageErr)
	assert.Contains(t, err.Error(), "failed to update webhook config")
}

// ---------------------------------------------------------------------------
// DeleteWebhookConfig
// ---------------------------------------------------------------------------

func TestDeleteWebhookConfig_Success(t *testing.T) {
	orgID := uuid.New()
	var capturedType string
	var capturedOrgID uuid.UUID
	repo := &mockNotificationRepo{
		deleteWebhookConfig: func(_ context.Context, wt string, o uuid.UUID) error {
			capturedType = wt
			capturedOrgID = o
			return nil
		},
	}
	svc := newTestService(repo)

	err := svc.DeleteWebhookConfig(context.Background(), &notification.DeleteWebhookConfigRequest{
		Type: "slack",
	}, orgID)

	require.NoError(t, err)
	assert.Equal(t, "slack", capturedType)
	assert.Equal(t, orgID, capturedOrgID)
}

func TestDeleteWebhookConfig_StorageError(t *testing.T) {
	storageErr := errors.New("delete failed")
	repo := &mockNotificationRepo{
		deleteWebhookConfig: func(_ context.Context, _ string, _ uuid.UUID) error {
			return storageErr
		},
	}
	svc := newTestService(repo)

	err := svc.DeleteWebhookConfig(context.Background(), &notification.DeleteWebhookConfigRequest{
		Type: "slack",
	}, uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, storageErr)
	assert.Contains(t, err.Error(), "failed to delete webhook config")
}

// ---------------------------------------------------------------------------
// GetWebhookConfig
// ---------------------------------------------------------------------------

func TestGetWebhookConfig_Success(t *testing.T) {
	expected := &shared_types.WebhookConfig{ID: uuid.New(), Type: "discord"}
	repo := &mockNotificationRepo{
		getWebhookConfig: func(_ context.Context, _ string, _ uuid.UUID) (*shared_types.WebhookConfig, error) {
			return expected, nil
		},
	}
	svc := newTestService(repo)

	got, err := svc.GetWebhookConfig(context.Background(), &notification.GetWebhookConfigRequest{
		Type: "discord",
	}, uuid.New())

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestGetWebhookConfig_StorageError(t *testing.T) {
	storageErr := errors.New("query failed")
	repo := &mockNotificationRepo{
		getWebhookConfig: func(_ context.Context, _ string, _ uuid.UUID) (*shared_types.WebhookConfig, error) {
			return nil, storageErr
		},
	}
	svc := newTestService(repo)

	_, err := svc.GetWebhookConfig(context.Background(), &notification.GetWebhookConfigRequest{
		Type: "discord",
	}, uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, storageErr)
	assert.Contains(t, err.Error(), "webhook config not found")
}
