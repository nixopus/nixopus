package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/helpers/discord"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/helpers/email"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/helpers/preferences"
	slackhelper "github.com/raghavyuva/nixopus-api/internal/features/notification/helpers/slack"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	slackgo "github.com/slack-go/slack"
	"github.com/uptrace/bun"
)

type NotificationChannels struct {
	Email   *Email   `json:"email"`
	Slack   *Slack   `json:"slack"`
	Discord *Discord `json:"discord"`
}

type Email struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromName  string `json:"from_name"`
	FromEmail string `json:"from_email"`
}

type Slack struct {
	SlackClient *slackgo.Client
	ChannelID   string
}

type Discord struct {
	WebhookUrl string `json:"webhook_url"`
}

type NotificationBaseData struct {
	IP      string
	Browser string
}

type NotificationAuthenticationData struct {
	NotificationBaseData
	Email    string
	UserName string
}

type NotificationOrganizationData struct {
	NotificationBaseData
	OrganizationID string
	UserID         string
}

type AddUserToOrganizationData struct {
	NotificationBaseData
	OrganizationName string
	UserName         string
	UserEmail        string
}

type RemoveUserFromOrganizationData struct {
	NotificationBaseData
	OrganizationName string
	UserName         string
	UserEmail        string
}

// type UpdateUserRoleData struct {
// 	OrganizationName string
// 	UserName         string
// 	NewRole          string
// }

type NotificationManager struct {
	sync.RWMutex
	Channels       *NotificationChannels
	PayloadChan    chan NotificationPayload
	ctx            context.Context
	cancel         context.CancelFunc
	db             *bun.DB
	prefManager    *preferences.PreferenceManager
	emailManager   *email.EmailManager
	slackManager   *slackhelper.SlackManager
	discordManager *discord.DiscordManager
}

type NotificationPasswordResetData struct {
	NotificationBaseData
	Email string
	Token string
}

type NotificationVerificationEmailData struct {
	NotificationBaseData
	Email string
	Token string
}

type NotificationPayloadType string

const (
	NotificationPayloadTypeRegister                   NotificationPayloadType = "register"
	NotificationPayloadTypeLogin                      NotificationPayloadType = "login"
	NotificationPayloadTypeLogout                     NotificationPayloadType = "logout"
	NotificationPayloadTypePasswordReset              NotificationPayloadType = "password_reset"
	NotificationPayloadTypeAddUserToOrganization      NotificationPayloadType = "add_user_to_organization"
	NotificationPayloadTypeRemoveUserFromOrganization NotificationPayloadType = "remove_user_from_organization"
	NotificationPayloadTypeVerificationEmail          NotificationPayloadType = "verification_email"
	// NotificationPayloadTypeUpdateUserRole             NotificationPayloadType = "update_user_role"
)

const (
	NortificationPayloadTypeCreateOrganization         NotificationPayloadType = "create_organization"
	NortificationPayloadTypeRemoveUserFromOrganization NotificationPayloadType = "remove_user_from_organization"
	NotificationPayloadTypeDeleteOrganization          NotificationPayloadType = "delete_organization"
	NotificationPayloadTypeUpdateOrganization          NotificationPayloadType = "update_organization"
)

type NotificationCategory string

const (
	NotificationCategoryAuthentication NotificationCategory = "authentication"
	NotificationCategoryOrganization   NotificationCategory = "organization"
)

type NotificationPayload struct {
	Category  NotificationCategory    `json:"category"`
	Type      NotificationPayloadType `json:"type"`
	UserID    string                  `json:"user_id"`
	Timestamp time.Time               `json:"timestamp"`
	Data      interface{}             `json:"data"`
}

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

type ResetEmailData struct {
	ResetURL string `json:"reset_url"`
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
