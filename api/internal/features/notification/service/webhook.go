package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *NotificationService) CreateWebhookConfig(ctx context.Context, req *notification.CreateWebhookConfigRequest, userID uuid.UUID, organizationID uuid.UUID) (*shared_types.WebhookConfig, error) {
	config := &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           req.Type,
		WebhookURL:     req.WebhookURL,
		IsActive:       true,
		UserID:         userID,
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.storage.CreateWebhookConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook config: %w", err)
	}

	return config, nil
}

func (s *NotificationService) UpdateWebhookConfig(ctx context.Context, req *notification.UpdateWebhookConfigRequest, organizationID uuid.UUID) (*shared_types.WebhookConfig, error) {
	config, err := s.storage.GetWebhookConfig(ctx, req.Type, organizationID)
	if err != nil {
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
		return nil, fmt.Errorf("failed to update webhook config: %w", err)
	}

	return config, nil
}

func (s *NotificationService) DeleteWebhookConfig(ctx context.Context, req *notification.DeleteWebhookConfigRequest, organizationID uuid.UUID) error {
	err := s.storage.DeleteWebhookConfig(ctx, req.Type, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook config: %w", err)
	}
	return nil
}

func (s *NotificationService) GetWebhookConfig(ctx context.Context, req *notification.GetWebhookConfigRequest, organizationID uuid.UUID) (*shared_types.WebhookConfig, error) {
	config, err := s.storage.GetWebhookConfig(ctx, req.Type, organizationID)
	if err != nil {
		return nil, fmt.Errorf("webhook config not found: %w", err)
	}
	return config, nil
}
