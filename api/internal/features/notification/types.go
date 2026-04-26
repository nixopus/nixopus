package notification

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type CreateSMTPConfigRequest struct {
	Host           string    `json:"host" validate:"required" description:"SMTP server hostname" example:"smtp.gmail.com"`
	Port           int       `json:"port" validate:"required,gt=0" description:"SMTP server port" example:"587"`
	Username       string    `json:"username" validate:"required" description:"SMTP authentication username" example:"user@gmail.com"`
	Password       string    `json:"password" validate:"required" description:"SMTP authentication password" example:"app-password"`
	FromName       string    `json:"from_name" description:"Display name for outgoing emails" example:"Nixopus"`
	FromEmail      string    `json:"from_email" description:"Sender email address for outgoing emails" example:"noreply@example.com"`
	OrganizationID uuid.UUID `json:"organization_id" validate:"required" description:"Organization this SMTP config belongs to" example:"550e8400-e29b-41d4-a716-446655440000"`
}

func (r CreateSMTPConfigRequest) String() string {
	return fmt.Sprintf("{Host: %s, Port: %d, Username: %s, FromName: %s, FromEmail: %s, OrgID: %s}",
		r.Host, r.Port, r.Username, r.FromName, r.FromEmail, r.OrganizationID)
}

type UpdateSMTPConfigRequest struct {
	ID             uuid.UUID `json:"id" validate:"required" description:"SMTP configuration ID to update" example:"550e8400-e29b-41d4-a716-446655440000"`
	Host           *string   `json:"host,omitempty" description:"SMTP server hostname" example:"smtp.gmail.com"`
	Port           *int      `json:"port,omitempty" description:"SMTP server port" example:"587"`
	Username       *string   `json:"username,omitempty" description:"SMTP authentication username" example:"user@gmail.com"`
	Password       *string   `json:"password,omitempty" description:"SMTP authentication password" example:"app-password"`
	FromName       *string   `json:"from_name,omitempty" description:"Display name for outgoing emails" example:"Nixopus"`
	FromEmail      *string   `json:"from_email,omitempty" description:"Sender email address for outgoing emails" example:"noreply@example.com"`
	OrganizationID uuid.UUID `json:"organization_id" validate:"required" description:"Organization this SMTP config belongs to" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type DeleteSMTPConfigRequest struct {
	ID uuid.UUID `json:"id" validate:"required" description:"SMTP configuration ID to delete" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type GetSMTPConfigRequest struct {
	ID uuid.UUID `json:"id" validate:"required" description:"SMTP configuration ID to retrieve" example:"550e8400-e29b-41d4-a716-446655440000"`
}

func NewSMTPConfig(c *CreateSMTPConfigRequest, userID uuid.UUID) *shared_types.SMTPConfigs {
	if c.FromEmail == "" {
		c.FromEmail = c.Username
	}
	if c.FromName == "" {
		c.FromName = strings.Split(c.Username, "@")[0]
	}
	return &shared_types.SMTPConfigs{
		Host:           c.Host,
		Port:           c.Port,
		Username:       c.Username,
		Password:       c.Password,
		FromName:       c.FromName,
		FromEmail:      c.FromEmail,
		UserID:         userID,
		ID:             uuid.New(),
		OrganizationID: c.OrganizationID,
	}
}

type Category string

const (
	ActivityCategory Category = "activity"
	SecurityCategory Category = "security"
	UpdateCategory   Category = "update"
)

func (Category) OpenAPISchemaType() []string { return []string{"string"} }
func (Category) OpenAPISchemaEnum() []any {
	return []any{string(ActivityCategory), string(SecurityCategory), string(UpdateCategory)}
}
func (Category) Description() string {
	return "Notification category: activity, security, or update."
}

type PreferenceType struct {
	ID          string `json:"id" validate:"required" description:"Preference type identifier" example:"deployment_success"`
	Label       string `json:"label" validate:"required" description:"Display label for the preference" example:"Deployment Success"`
	Description string `json:"description" validate:"required" description:"Description of the preference type" example:"Notifies when a deployment succeeds"`
	Enabled     bool   `json:"enabled" description:"Whether this preference is enabled" example:"true"`
}

type CategoryPreferences struct {
	Category    Category         `json:"category" validate:"required" description:"Notification category" example:"activity"`
	Preferences []PreferenceType `json:"preferences" validate:"required" description:"List of preference types in this category"`
}

type UpdatePreferenceRequest struct {
	Category string `json:"category" validate:"required,oneof=activity security update" description:"Notification category" example:"activity"`
	Type     string `json:"type" validate:"required" description:"Preference type identifier" example:"deployment_success"`
	Enabled  bool   `json:"enabled" description:"Whether this notification preference is enabled" example:"true"`
}

type GetPreferencesResponse struct {
	Activity []PreferenceType `json:"activity"`
	Security []PreferenceType `json:"security"`
	Update   []PreferenceType `json:"update"`
}

type PreferenceItem struct {
	ID           uuid.UUID `json:"id"`
	PreferenceID uuid.UUID `json:"preference_id"`
	Category     string    `json:"category"`
	Type         string    `json:"type"`
	Enabled      bool      `json:"enabled"`
}

type CreateWebhookConfigRequest struct {
	Type       string `json:"type" validate:"required,oneof=slack discord" description:"Webhook integration type" example:"slack"`
	WebhookURL string `json:"webhook_url" validate:"required" description:"Webhook URL for the integration" example:"https://hooks.slack.com/services/..."`
}

type UpdateWebhookConfigRequest struct {
	Type       string  `json:"type" validate:"required,oneof=slack discord" description:"Webhook integration type" example:"slack"`
	WebhookURL *string `json:"webhook_url,omitempty" description:"Updated webhook URL" example:"https://hooks.slack.com/services/..."`
	IsActive   *bool   `json:"is_active,omitempty" description:"Whether the webhook integration is active" example:"true"`
}

type DeleteWebhookConfigRequest struct {
	Type string `json:"type" validate:"required,oneof=slack discord" description:"Webhook integration type to delete" example:"slack"`
}

type GetWebhookConfigRequest struct {
	Type string `json:"type" validate:"required,oneof=slack discord" description:"Webhook integration type to retrieve" example:"slack"`
}

type SendNotificationRequest struct {
	Channel  string            `json:"channel" validate:"required,oneof=slack discord email" description:"Notification delivery channel" example:"email"`
	Message  string            `json:"message" validate:"required" description:"Notification message body" example:"Deployment completed successfully"`
	Subject  string            `json:"subject,omitempty" description:"Email subject line (only for email channel)" example:"Deployment Update"`
	To       string            `json:"to,omitempty" description:"Recipient email address (only for email channel)" example:"user@example.com"`
	Metadata map[string]string `json:"metadata,omitempty" description:"Additional key-value metadata for the notification"`
}

type SendNotificationResponse struct {
	Channel string `json:"channel"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

var (
	ErrInvalidRequestType                      = errors.New("invalid request type")
	ErrMissingHost                             = errors.New("host is required")
	ErrMissingPort                             = errors.New("port is required")
	ErrMissingUsername                         = errors.New("username is required")
	ErrMissingPassword                         = errors.New("password is required")
	ErrMissingID                               = errors.New("id is required")
	ErrMissingCategory                         = errors.New("category is required")
	ErrMissingType                             = errors.New("type is required")
	ErrPermissionDenied                        = errors.New("permission denied")
	ErrSMTPConfigNotFound                      = errors.New("smtp config not found")
	ErrAccessDenied                            = errors.New("access denied")
	ErrUserDoesNotBelongToOrganization         = errors.New("user does not belong to organization")
	ErrUserDoesNotHavePermissionForTheResource = errors.New("user does not have permission for the resource")
	ErrInvalidResource                         = errors.New("invalid resource")
	ErrMissingOrganization                     = errors.New("organization is required")
	ErrSmtpAlreadyExists                       = errors.New("smtp already exists")
)
