package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

type NotificationStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

type NotificationRepository interface {
	AddSmtp(config *shared_types.SMTPConfigs) error
	UpdateSmtp(config *notification.UpdateSMTPConfigRequest) error
	DeleteSmtp(ID string) error
	GetSmtp(ID string) (*shared_types.SMTPConfigs, error)
	GetOrganizationsSmtp(organizationID string) ([]shared_types.SMTPConfigs, error)
	UpdatePreference(ctx context.Context, req notification.UpdatePreferenceRequest, userID uuid.UUID) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (*notification.GetPreferencesResponse, error)
	CreateWebhookConfig(ctx context.Context, config *shared_types.WebhookConfig) error
	UpdateWebhookConfig(ctx context.Context, config *shared_types.WebhookConfig) error
	DeleteWebhookConfig(ctx context.Context, webhookType string, organizationID uuid.UUID) error
	GetWebhookConfig(ctx context.Context, webhookType string, organizationID uuid.UUID) (*shared_types.WebhookConfig, error)
}

// AddSmtp adds a new SMTP configuration to the database.
//
// It takes a shared_types.SMTPConfigs as a parameter and inserts it into the database.
// It returns an error if the database operation fails.
func (s NotificationStorage) AddSmtp(config *shared_types.SMTPConfigs) error {
	_, err := s.DB.NewInsert().Model(config).Exec(s.Ctx)
	return err
}

// UpdateSmtp updates an existing SMTP configuration in the database.
//
// It takes a notification.UpdateSMTPConfigRequest as a parameter and updates the
// corresponding SMTP configuration in the database.
// It returns an error if the database operation fails.
func (s NotificationStorage) UpdateSmtp(config *notification.UpdateSMTPConfigRequest) error {
	var smtp *shared_types.SMTPConfigs
	_, err := s.DB.NewUpdate().Model(smtp).
		Set("host = ?", config.Host).
		Set("port = ?", strconv.Itoa(*config.Port)).
		Set("username = ?", config.Username).
		Set("password = ?", config.Password).
		Set("from_name = ?", config.FromName).
		Set("from_email = ?", config.FromEmail).
		Where("id = ?", config.ID).Exec(s.Ctx)
	return err
}

// DeleteSmtp deletes a SMTP configuration associated with the given ID.
//
// It takes an ID as a parameter, deletes the corresponding SMTP configuration
// from the database, and returns an error if the database operation fails.
func (s NotificationStorage) DeleteSmtp(ID string) error {
	var config shared_types.SMTPConfigs
	_, err := s.DB.NewDelete().Model(config).Where("id = ?", ID).Exec(s.Ctx)
	return err
}

// GetSmtp returns the SMTP configuration associated with the given ID.
//
// It takes an ID as a parameter, queries the database for the corresponding
// SMTP configuration, and returns it. It returns an error if the database
// operation fails.
func (s NotificationStorage) GetSmtp(ID string) (*shared_types.SMTPConfigs, error) {
	config := &shared_types.SMTPConfigs{}
	err := s.DB.NewSelect().Model(config).Where("user_id = ?", ID).Scan(s.Ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return config, nil
}

// GetOrganizationsSmtp returns the SMTP configurations associated with the given organization ID.
//
// It takes an organization ID as a parameter, queries the database for the
// corresponding SMTP configurations, and returns them. It returns an error
// if the database operation fails.
func (s NotificationStorage) GetOrganizationsSmtp(organizationID string) ([]shared_types.SMTPConfigs, error) {
	configs := []shared_types.SMTPConfigs{}
	err := s.DB.NewSelect().Model(&configs).Where("organization_id = ?", organizationID).Scan(s.Ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return []shared_types.SMTPConfigs{}, nil
		}
		return nil, err
	}
	return configs, nil
}

// UpdatePreference updates a user's notification preference.
//
// It will update the corresponding preference item in the database.
//
// If the user has no preferences, this will create a new preference item
// in the database.
//
// The function will log an info message with the details of the preference update.
func (s *NotificationStorage) UpdatePreference(ctx context.Context, req notification.UpdatePreferenceRequest, userID uuid.UUID) error {
	var preferenceID uuid.UUID
	err := s.DB.NewSelect().
		Model((*shared_types.NotificationPreferences)(nil)).
		Column("id").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Scan(ctx, &preferenceID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.initUserPreferences(ctx, userID, req)
		}
		return fmt.Errorf("failed to fetch user preferences: %w", err)
	}

	_, err = s.DB.NewUpdate().
		Model((*shared_types.PreferenceItem)(nil)).
		Set("enabled = ?", req.Enabled).
		Set("updated_at = ?", time.Now()).
		Where("preference_id = ?", preferenceID).
		Where("category = ?", req.Category).
		Where("type = ?", req.Type).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update preference: %w", err)
	}

	return nil
}

// GetPreferences retrieves the notification preferences for a specified user.
//
// The function first attempts to fetch the existing preferences for the user
// based on the user ID. If no preferences are found, it initializes default
// preferences for the user and retrieves them.
//
// It returns a GetPreferencesResponse containing the user's preferences or an
// error if any database operation fails.
func (s *NotificationStorage) GetPreferences(ctx context.Context, userID uuid.UUID) (*notification.GetPreferencesResponse, error) {
	var preferenceID uuid.UUID
	err := s.DB.NewSelect().
		Model((*shared_types.NotificationPreferences)(nil)).
		Column("id").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Scan(ctx, &preferenceID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := s.initDefaultPreferences(ctx, userID); err != nil {
				return nil, fmt.Errorf("failed to initialize preferences: %w", err)
			}

			err = s.DB.NewSelect().
				Model((*shared_types.NotificationPreferences)(nil)).
				Column("id").
				Where("user_id = ?", userID).
				Where("deleted_at IS NULL").
				Scan(ctx, &preferenceID)

			if err != nil {
				return nil, fmt.Errorf("failed to fetch newly created preferences: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to fetch user preferences: %w", err)
		}
	}

	var items []shared_types.PreferenceItem
	err = s.DB.NewSelect().
		Model((*shared_types.PreferenceItem)(nil)).
		Where("preference_id = ?", preferenceID).
		Scan(ctx, &items)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch preference items: %w", err)
	}

	response := MapToResponse(items)
	return &response, nil
}

// initUserPreferences initializes a user's notification preferences in the database.
//
// This function begins a new database transaction to create a new set of notification
// preferences for the given user ID. It generates a new UUID for the preference ID
// and sets the current time for both the creation and update timestamps.
//
// Default preference items are created and customized based on the provided
// UpdatePreferenceRequest. If the category and type of a default preference item
// match those in the request, the item's enabled status is updated accordingly.
//
// If any database operation fails, the transaction is rolled back and an error
// is returned. Otherwise, the transaction is committed, and nil is returned
// to indicate success.
func (s *NotificationStorage) initUserPreferences(ctx context.Context, userID uuid.UUID, update notification.UpdatePreferenceRequest) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	prefID := uuid.New()
	now := time.Now()
	preferences := &shared_types.NotificationPreferences{
		ID:        prefID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = tx.NewInsert().Model(preferences).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert preferences: %w", err)
	}

	items := createDefaultPreferenceItems(prefID)

	for i, item := range items {
		if item.Category == update.Category && item.Type == update.Type {
			items[i].Enabled = update.Enabled
		}
	}

	_, err = tx.NewInsert().Model(&items).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert preference items: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// initDefaultPreferences initializes the default notification preferences for a user in the database.
//
// If the user has no preferences, this will create a new preference item in the database.
//
// The function will log an info message with the details of the preference initialization.
func (s *NotificationStorage) initDefaultPreferences(ctx context.Context, userID uuid.UUID) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	prefID := uuid.New()
	now := time.Now()
	preferences := &shared_types.NotificationPreferences{
		ID:        prefID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = tx.NewInsert().Model(preferences).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert preferences: %w", err)
	}

	items := createDefaultPreferenceItems(prefID)

	_, err = tx.NewInsert().Model(&items).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert preference items: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// createDefaultPreferenceItems returns a slice of default preference items for a user.
//
// The preference items are of the following types:
// activity: team-updates
// security: login-alerts, password-changes, security-alerts
// update: product-updates, newsletter, marketing
//
// The enabled status for each item is as follows:
// activity: team-updates (true)
// security: login-alerts (true), password-changes (true), security-alerts (true)
// update: product-updates (true), newsletter (false), marketing (false)
func createDefaultPreferenceItems(preferenceID uuid.UUID) []shared_types.PreferenceItem {
	return []shared_types.PreferenceItem{
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     "activity",
			Type:         "team-updates",
			Enabled:      true,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     "security",
			Type:         "login-alerts",
			Enabled:      true,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     "security",
			Type:         "password-changes",
			Enabled:      true,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     "security",
			Type:         "security-alerts",
			Enabled:      true,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     "update",
			Type:         "product-updates",
			Enabled:      true,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     "update",
			Type:         "newsletter",
			Enabled:      false,
		},
		{
			ID:           uuid.New(),
			PreferenceID: preferenceID,
			Category:     "update",
			Type:         "marketing",
			Enabled:      false,
		},
	}
}

// MapToResponse converts a slice of PreferenceItem into a GetPreferencesResponse.
//
// It maps each PreferenceItem to a PreferenceType with corresponding labels and
// descriptions based on its category and type. The function returns a structured
// GetPreferencesResponse containing categorized notification preferences for activity,
// security, and update. If a preference type is not recognized, it is ignored.
func MapToResponse(items []shared_types.PreferenceItem) notification.GetPreferencesResponse {
	response := notification.GetPreferencesResponse{
		Activity: []notification.PreferenceType{},
		Security: []notification.PreferenceType{},
		Update:   []notification.PreferenceType{},
	}

	typeInfo := map[string]map[string]struct {
		Label       string
		Description string
	}{
		"activity": {
			"team-updates": {
				Label:       "Team Updates",
				Description: "When team members join or leave your team",
			},
		},
		"security": {
			"login-alerts": {
				Label:       "Login Alerts",
				Description: "When a new device logs into your account",
			},
			"password-changes": {
				Label:       "Password Changes",
				Description: "When your password is changed",
			},
			"security-alerts": {
				Label:       "Security Alerts",
				Description: "Important security notifications",
			},
		},
		"update": {
			"product-updates": {
				Label:       "Product Updates",
				Description: "New features and improvements",
			},
			"newsletter": {
				Label:       "Newsletter",
				Description: "Our monthly newsletter with tips and updates",
			},
			"marketing": {
				Label:       "Marketing",
				Description: "Promotions and special offers",
			},
		},
	}

	for _, item := range items {
		info, exists := typeInfo[item.Category][item.Type]
		if !exists {
			continue
		}

		pref := notification.PreferenceType{
			ID:          item.Type,
			Label:       info.Label,
			Description: info.Description,
			Enabled:     item.Enabled,
		}

		switch item.Category {
		case "activity":
			response.Activity = append(response.Activity, pref)
		case "security":
			response.Security = append(response.Security, pref)
		case "update":
			response.Update = append(response.Update, pref)
		}
	}

	return response
}

// CreateWebhookConfig creates a new webhook configuration in the database.
func (s NotificationStorage) CreateWebhookConfig(ctx context.Context, config *shared_types.WebhookConfig) error {
	_, err := s.DB.NewInsert().Model(config).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}
	return nil
}

// UpdateWebhookConfig updates an existing webhook configuration in the database.
func (s NotificationStorage) UpdateWebhookConfig(ctx context.Context, config *shared_types.WebhookConfig) error {
	_, err := s.DB.NewUpdate().Model(config).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update webhook config: %w", err)
	}
	return nil
}

// DeleteWebhookConfig deletes a webhook configuration from the database.
func (s NotificationStorage) DeleteWebhookConfig(ctx context.Context, webhookType string, organizationID uuid.UUID) error {
	_, err := s.DB.NewDelete().Model((*shared_types.WebhookConfig)(nil)).
		Where("type = ? AND organization_id = ?", webhookType, organizationID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete webhook config: %w", err)
	}
	return nil
}

// GetWebhookConfig retrieves a webhook configuration from the database.
func (s NotificationStorage) GetWebhookConfig(ctx context.Context, webhookType string, organizationID uuid.UUID) (*shared_types.WebhookConfig, error) {
	config := &shared_types.WebhookConfig{}
	err := s.DB.NewSelect().Model(config).
		Where("type = ? AND organization_id = ?", webhookType, organizationID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("webhook config not found: %w", err)
	}
	return config, nil
}
