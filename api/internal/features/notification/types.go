package notification

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type CreateSMTPConfigRequest struct {
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	FromName       string    `json:"from_name"`
	FromEmail      string    `json:"from_email"`
	OrganizationID uuid.UUID `json:"organization_id"`
}

func (r CreateSMTPConfigRequest) String() string {
	return fmt.Sprintf("{Host: %s, Port: %d, Username: %s, FromName: %s, FromEmail: %s, OrgID: %s}",
		r.Host, r.Port, r.Username, r.FromName, r.FromEmail, r.OrganizationID)
}

type UpdateSMTPConfigRequest struct {
	ID             uuid.UUID `json:"id"`
	Host           *string   `json:"host,omitempty"`
	Port           *int      `json:"port,omitempty"`
	Username       *string   `json:"username,omitempty"`
	Password       *string   `json:"password,omitempty"`
	FromName       *string   `json:"from_name,omitempty"`
	FromEmail      *string   `json:"from_email,omitempty"`
	OrganizationID uuid.UUID `json:"organization_id"`
}

type DeleteSMTPConfigRequest struct {
	ID uuid.UUID `json:"id"`
}

type GetSMTPConfigRequest struct {
	ID uuid.UUID `json:"id"`
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

type PreferenceType struct {
	ID          string `json:"id" validate:"required"`
	Label       string `json:"label" validate:"required"`
	Description string `json:"description" validate:"required"`
	Enabled     bool   `json:"enabled"`
}

type CategoryPreferences struct {
	Category    Category         `json:"category" validate:"required"`
	Preferences []PreferenceType `json:"preferences" validate:"required"`
}

type UpdatePreferenceRequest struct {
	Category string `json:"category" validate:"required,oneof=activity security update"`
	Type     string `json:"type" validate:"required"`
	Enabled  bool   `json:"enabled"`
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
	Type       string `json:"type" validate:"required,oneof=slack discord"`
	WebhookURL string `json:"webhook_url"`
}

type UpdateWebhookConfigRequest struct {
	Type       string  `json:"type" validate:"required,oneof=slack discord"`
	WebhookURL *string `json:"webhook_url,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
}

type DeleteWebhookConfigRequest struct {
	Type string `json:"type" validate:"required,oneof=slack discord"`
}

type GetWebhookConfigRequest struct {
	Type string `json:"type" validate:"required,oneof=slack discord"`
}

type SendNotificationRequest struct {
	Channel  string            `json:"channel" validate:"required,oneof=slack discord email"`
	Message  string            `json:"message" validate:"required"`
	Subject  string            `json:"subject,omitempty"`
	To       string            `json:"to,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
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
