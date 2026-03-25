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

type DiscordChannel struct {
	db  *bun.DB
	ctx context.Context
}

func NewDiscordChannel(db *bun.DB, ctx context.Context) *DiscordChannel {
	return &DiscordChannel{db: db, ctx: ctx}
}

func (d *DiscordChannel) Type() string { return "discord" }

func (d *DiscordChannel) Send(ctx context.Context, msg Message) error {
	orgID, ok := msg.Metadata["organization_id"]
	if !ok {
		return fmt.Errorf("organization_id required in message metadata for discord channel")
	}

	webhookURL, err := d.getWebhookURL(orgID)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(struct {
		Content string `json:"content"`
	}{Content: msg.Body})
	if err != nil {
		return fmt.Errorf("failed to marshal discord message: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send discord notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func (d *DiscordChannel) getWebhookURL(organizationID string) (string, error) {
	var config shared_types.WebhookConfig
	err := d.db.NewSelect().
		Model(&config).
		Where("type = ?", "discord").
		Where("organization_id = ?", organizationID).
		Where("is_active = ?", true).
		Scan(d.ctx)
	if err != nil {
		return "", fmt.Errorf("no active discord webhook for org %s: %w", organizationID, err)
	}
	return config.WebhookURL, nil
}
