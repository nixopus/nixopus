package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *NotificationService) CreateWebhookConfig(ctx context.Context, req *notification.CreateWebhookConfigRequest, userID uuid.UUID, organizationID uuid.UUID) (*shared_types.WebhookConfig, error) {
	config := &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           req.Type,
		WebhookURL:     req.WebhookURL,
		ChannelID:      req.Type + ":" + organizationID.String(),
		IsActive:       true,
		UserID:         userID,
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.storage.CreateWebhookConfig(ctx, config)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("notification service: CreateWebhookConfig: %v", err), fmt.Sprintf("org_id=%s user_id=%s type=%s", organizationID, userID, req.Type))
		return nil, fmt.Errorf("failed to create webhook config: %w", err)
	}

	return config, nil
}

func (s *NotificationService) UpdateWebhookConfig(ctx context.Context, req *notification.UpdateWebhookConfigRequest, organizationID uuid.UUID) (*shared_types.WebhookConfig, error) {
	config, err := s.storage.GetWebhookConfig(ctx, req.Type, organizationID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("notification service: UpdateWebhookConfig load: %v", err), fmt.Sprintf("org_id=%s type=%s", organizationID, req.Type))
		return nil, fmt.Errorf("webhook config not found: %w", err)
	}

	if req.WebhookURL != nil {
		config.WebhookURL = *req.WebhookURL
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}
	config.UpdatedAt = time.Now()

	err = s.storage.UpdateWebhookConfig(ctx, config)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("notification service: UpdateWebhookConfig: %v", err), fmt.Sprintf("org_id=%s type=%s webhook_id=%s", organizationID, req.Type, config.ID))
		return nil, fmt.Errorf("failed to update webhook config: %w", err)
	}

	return config, nil
}

func (s *NotificationService) DeleteWebhookConfig(ctx context.Context, req *notification.DeleteWebhookConfigRequest, organizationID uuid.UUID) error {
	err := s.storage.DeleteWebhookConfig(ctx, req.Type, organizationID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("notification service: DeleteWebhookConfig: %v", err), fmt.Sprintf("org_id=%s type=%s", organizationID, req.Type))
		return fmt.Errorf("failed to delete webhook config: %w", err)
	}
	return nil
}

func (s *NotificationService) GetWebhookConfig(ctx context.Context, req *notification.GetWebhookConfigRequest, organizationID uuid.UUID) (*shared_types.WebhookConfig, error) {
	config, err := s.storage.GetWebhookConfig(ctx, req.Type, organizationID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("notification service: GetWebhookConfig: %v", err), fmt.Sprintf("org_id=%s type=%s", organizationID, req.Type))
		return nil, fmt.Errorf("webhook config not found: %w", err)
	}
	return config, nil
}
