package channel

import "context"

// Channel is the delivery backend interface. Each notification channel
// (email, slack, discord, etc.) implements this to handle its own transport.
type Channel interface {
	Type() string
	Send(ctx context.Context, msg Message) error
}

// Message carries the delivery payload after event resolution.
// TemplateName + TemplateData are used by channels that support templates (email).
// Body is used by plain-text channels (slack, discord).
type Message struct {
	To           string                 `json:"to"`
	Subject      string                 `json:"subject,omitempty"`
	Body         string                 `json:"body"`
	HTMLBody     string                 `json:"html_body,omitempty"`
	TemplateName string                 `json:"template_name,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
}

// DeliveryPayload is the serializable struct enqueued into the Redis taskq.
// The worker deserializes this and dispatches to the correct Channel adapter.
type DeliveryPayload struct {
	Channel        string  `json:"channel"`
	OrganizationID string  `json:"organization_id"`
	UserID         string  `json:"user_id"`
	Message        Message `json:"message"`
}
