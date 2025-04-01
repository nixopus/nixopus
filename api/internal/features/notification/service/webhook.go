package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bytes"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *NotificationService) CreateWebhookConfig(ctx context.Context, req *notification.CreateWebhookConfigRequest, userID uuid.UUID, organizationID uuid.UUID) (*shared_types.WebhookConfig, error) {
	config := &shared_types.WebhookConfig{
		ID:             uuid.New(),
		Type:           req.Type,
		WebhookURL:     req.WebhookURL,
		WebhookSecret:  req.WebhookSecret,
		ChannelID:      req.ChannelID,
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
	if req.WebhookSecret != nil {
		config.WebhookSecret = req.WebhookSecret
	}
	if req.ChannelID != nil {
		config.ChannelID = *req.ChannelID
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

func (s *NotificationService) SendSlackMessage(webhookURL string, message string, secret *string) error {
	payload := map[string]string{
		"text": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if secret != nil {
		timestamp := time.Now().Unix()
		signature := fmt.Sprintf("v0:%d:%s", timestamp, *secret)
		req.Header.Set("X-Slack-Request-Timestamp", fmt.Sprintf("%d", timestamp))
		req.Header.Set("X-Slack-Signature", signature)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *NotificationService) SendDiscordMessage(webhookURL string, message string, secret *string) error {
	payload := map[string]interface{}{
		"content": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal discord payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if secret != nil {
		req.Header.Set("X-Webhook-Secret", *secret)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send discord message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord webhook returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
