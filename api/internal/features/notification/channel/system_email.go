package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SystemEmailChannel struct {
	apiKey    string
	fromEmail string
	client    *http.Client
}

func NewSystemEmailChannel(apiKey, fromEmail string) *SystemEmailChannel {
	return &SystemEmailChannel{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SystemEmailChannel) Type() string { return "system_email" }

func (s *SystemEmailChannel) Send(ctx context.Context, msg Message) error {
	if s.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY not configured")
	}

	templateID, ok := msg.Metadata["resend_template_id"]
	if !ok || templateID == "" {
		return fmt.Errorf("resend_template_id not set in message metadata")
	}

	recipient := msg.To
	if recipient == "" {
		return fmt.Errorf("recipient address is required")
	}

	tmpl := map[string]interface{}{"id": templateID}
	if len(msg.TemplateData) > 0 {
		tmpl["variables"] = msg.TemplateData
	}

	payload := map[string]interface{}{
		"from":     s.fromEmail,
		"to":       recipient,
		"template": tmpl,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal resend payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create resend request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
