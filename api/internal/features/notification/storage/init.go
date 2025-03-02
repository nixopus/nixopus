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

func (s NotificationStorage) AddSmtp(config *shared_types.SMTPConfigs) error {
	_, err := s.DB.NewInsert().Model(config).Exec(s.Ctx)
	return err
}

func (s NotificationStorage) UpdateSmtp(config *notification.UpdateSMTPConfigRequest) error {
	var smtp *shared_types.SMTPConfigs
	_, err := s.DB.NewUpdate().Model(smtp).
		Set("host = ?", config.Host).
		Set("port = ?", strconv.Itoa(config.Port)).
		Set("username = ?", config.Username).
		Set("password = ?", config.Password).
		Set("from_name = ?", config.FromName).
		Set("from_email = ?", config.FromEmail).
		Where("id = ?", config.ID).Exec(s.Ctx)
	return err
}

func (s NotificationStorage) DeleteSmtp(ID string) error {
	var config shared_types.SMTPConfigs
	_, err := s.DB.NewDelete().Model(config).Where("id = ?", ID).Exec(s.Ctx)
	return err
}

func (s NotificationStorage) GetSmtp(ID string) (*shared_types.SMTPConfigs, error) {
	config := &shared_types.SMTPConfigs{}
	err := s.DB.NewSelect().Model(config).Where("user_id = ?", ID).Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return config, nil
}

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
