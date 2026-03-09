package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type SlackChannel struct {
	db  *bun.DB
	ctx context.Context
}

func NewSlackChannel(db *bun.DB, ctx context.Context) *SlackChannel {
	return &SlackChannel{db: db, ctx: ctx}
}

func (s *SlackChannel) Type() string { return "slack" }

func (s *SlackChannel) Send(ctx context.Context, msg Message) error {
	orgID, ok := msg.Metadata["organization_id"]
	if !ok {
		return fmt.Errorf("organization_id required in message metadata for slack channel")
	}

	webhookURL, err := s.getWebhookURL(orgID)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(struct {
		Text string `json:"text"`
	}{Text: msg.Body})
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *SlackChannel) getWebhookURL(organizationID string) (string, error) {
	var config shared_types.WebhookConfig
	err := s.db.NewSelect().
		Model(&config).
		Where("type = ?", "slack").
		Where("organization_id = ?", organizationID).
		Where("is_active = ?", true).
		Scan(s.ctx)
	if err != nil {
		return "", fmt.Errorf("no active slack webhook for org %s: %w", organizationID, err)
	}
	return config.WebhookURL, nil
}
