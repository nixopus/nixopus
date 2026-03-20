package channel

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type EmailChannel struct {
	db  *bun.DB
	ctx context.Context
}

func NewEmailChannel(db *bun.DB, ctx context.Context) *EmailChannel {
	return &EmailChannel{db: db, ctx: ctx}
}

func (e *EmailChannel) Type() string { return "email" }

func (e *EmailChannel) Send(ctx context.Context, msg Message) error {
	orgID, ok := msg.Metadata["organization_id"]
	if !ok {
		return fmt.Errorf("organization_id required in message metadata for email channel")
	}

	smtpConfig, err := e.getSmtpByOrg(orgID)
	if err != nil {
		return fmt.Errorf("failed to resolve SMTP config: %w", err)
	}

	recipient := msg.To
	if recipient == "" {
		return fmt.Errorf("recipient address is required")
	}

	body := msg.Body
	if msg.TemplateName != "" {
		rendered, err := e.renderTemplate(msg.TemplateName, msg.TemplateData)
		if err != nil {
			return fmt.Errorf("template render failed: %w", err)
		}
		body = rendered
	}

	contentType := "text/plain; charset=UTF-8"
	if msg.TemplateName != "" || msg.HTMLBody != "" {
		contentType = "text/html; charset=UTF-8"
		if msg.HTMLBody != "" {
			body = msg.HTMLBody
		}
	}

	from := smtpConfig.Username
	subject := msg.Subject
	if subject == "" {
		subject = "Notification from Nixopus"
	}

	raw := []byte(fmt.Sprintf(
		"Subject: %s\r\nFrom: %s <%s>\r\nTo: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s\r\n\r\n%s",
		subject, smtpConfig.FromName, from, recipient, contentType, body,
	))

	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)
	addr := fmt.Sprintf("%s:%d", smtpConfig.Host, smtpConfig.Port)

	if err := smtp.SendMail(addr, auth, from, []string{recipient}, raw); err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}

	return nil
}

func (e *EmailChannel) getSmtpByOrg(organizationID string) (*shared_types.SMTPConfigs, error) {
	config := &shared_types.SMTPConfigs{}
	err := e.db.NewSelect().
		Model(config).
		Where("organization_id = ?", organizationID).
		Where("is_active = ?", true).
		Limit(1).
		Scan(e.ctx)
	if err != nil {
		return nil, fmt.Errorf("no active SMTP config for org %s: %w", organizationID, err)
	}
	return config, nil
}

func (e *EmailChannel) renderTemplate(name string, data map[string]interface{}) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working directory: %w", err)
	}

	tmpl, err := template.ParseFiles(filepath.Join(wd, "internal/features/notification/templates/"+name))
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	return buf.String(), nil
}
