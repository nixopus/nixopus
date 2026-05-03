package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// mockNotificationRepo is a hand-written test double for storage.NotificationRepository.
// Each method delegates to an optional func field; nil fields return benign zero values.
type mockNotificationRepo struct {
	addSmtp              func(config *shared_types.SMTPConfigs) error
	updateSmtp           func(config *notification.UpdateSMTPConfigRequest) error
	deleteSmtp           func(ID string) error
	getSmtp              func(ID string) (*shared_types.SMTPConfigs, error)
	getOrganizationsSmtp func(organizationID string) ([]shared_types.SMTPConfigs, error)
	updatePreference     func(ctx context.Context, req notification.UpdatePreferenceRequest, userID uuid.UUID) error
	getPreferences       func(ctx context.Context, userID uuid.UUID) (*notification.GetPreferencesResponse, error)
	createWebhookConfig  func(ctx context.Context, config *shared_types.WebhookConfig) error
	updateWebhookConfig  func(ctx context.Context, config *shared_types.WebhookConfig) error
	deleteWebhookConfig  func(ctx context.Context, webhookType string, organizationID uuid.UUID) error
	getWebhookConfig     func(ctx context.Context, webhookType string, organizationID uuid.UUID) (*shared_types.WebhookConfig, error)
}

func (m *mockNotificationRepo) AddSmtp(config *shared_types.SMTPConfigs) error {
	if m.addSmtp != nil {
		return m.addSmtp(config)
	}
	return nil
}

func (m *mockNotificationRepo) UpdateSmtp(config *notification.UpdateSMTPConfigRequest) error {
	if m.updateSmtp != nil {
		return m.updateSmtp(config)
	}
	return nil
}

func (m *mockNotificationRepo) DeleteSmtp(ID string) error {
	if m.deleteSmtp != nil {
		return m.deleteSmtp(ID)
	}
	return nil
}

func (m *mockNotificationRepo) GetSmtp(ID string) (*shared_types.SMTPConfigs, error) {
	if m.getSmtp != nil {
		return m.getSmtp(ID)
	}
	return nil, nil
}

func (m *mockNotificationRepo) GetOrganizationsSmtp(organizationID string) ([]shared_types.SMTPConfigs, error) {
	if m.getOrganizationsSmtp != nil {
		return m.getOrganizationsSmtp(organizationID)
	}
	return nil, nil
}

func (m *mockNotificationRepo) UpdatePreference(ctx context.Context, req notification.UpdatePreferenceRequest, userID uuid.UUID) error {
	if m.updatePreference != nil {
		return m.updatePreference(ctx, req, userID)
	}
	return nil
}

func (m *mockNotificationRepo) GetPreferences(ctx context.Context, userID uuid.UUID) (*notification.GetPreferencesResponse, error) {
	if m.getPreferences != nil {
		return m.getPreferences(ctx, userID)
	}
	return &notification.GetPreferencesResponse{}, nil
}

func (m *mockNotificationRepo) CreateWebhookConfig(ctx context.Context, config *shared_types.WebhookConfig) error {
	if m.createWebhookConfig != nil {
		return m.createWebhookConfig(ctx, config)
	}
	return nil
}

func (m *mockNotificationRepo) UpdateWebhookConfig(ctx context.Context, config *shared_types.WebhookConfig) error {
	if m.updateWebhookConfig != nil {
		return m.updateWebhookConfig(ctx, config)
	}
	return nil
}

func (m *mockNotificationRepo) DeleteWebhookConfig(ctx context.Context, webhookType string, organizationID uuid.UUID) error {
	if m.deleteWebhookConfig != nil {
		return m.deleteWebhookConfig(ctx, webhookType, organizationID)
	}
	return nil
}

func (m *mockNotificationRepo) GetWebhookConfig(ctx context.Context, webhookType string, organizationID uuid.UUID) (*shared_types.WebhookConfig, error) {
	if m.getWebhookConfig != nil {
		return m.getWebhookConfig(ctx, webhookType, organizationID)
	}
	return &shared_types.WebhookConfig{}, nil
}

// newTestService creates a NotificationService wired with the provided mock and sane defaults.
func newTestService(repo *mockNotificationRepo) *NotificationService {
	l := logger.NewLogger()
	return &NotificationService{
		storage: repo,
		Ctx:     context.Background(),
		logger:  l,
	}
}
