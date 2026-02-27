package utils

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

// GetOrganizationIDFromBetterAuth gets organization ID from Better Auth session
// Returns empty string if not found
func GetOrganizationIDFromBetterAuth(r *http.Request) (string, error) {
	sessionResp, err := auth.VerifySession(r)
	if err != nil {
		return "", fmt.Errorf("failed to verify session: %w", err)
	}

	if sessionResp.Session.ActiveOrganizationID != nil && *sessionResp.Session.ActiveOrganizationID != "" {
		return *sessionResp.Session.ActiveOrganizationID, nil
	}

	// Fallback to header
	orgID := r.Header.Get("X-Organization-Id")
	return orgID, nil
}

// GetOrCreateOrganizationID gets organization ID from context, Better Auth session, or creates one
// This ensures organization exists in local database for foreign key constraints
func GetOrCreateOrganizationID(ctx context.Context, r *http.Request, app *storage.App) (uuid.UUID, error) {
	// First try to get from context (set by auth middleware)
	orgIDAny := ctx.Value(types.OrganizationIDKey)
	if orgIDAny != nil {
		if strID, ok := orgIDAny.(string); ok {
			if id, err := uuid.Parse(strID); err == nil {
				return id, nil
			}
		}
		if id, ok := orgIDAny.(uuid.UUID); ok {
			return id, nil
		}
	}

	// Try to get from Better Auth session
	orgIDStr, err := GetOrganizationIDFromBetterAuth(r)
	if err != nil || orgIDStr == "" {
		return uuid.Nil, fmt.Errorf("organization ID not found: %w", err)
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid organization ID format: %w", err)
	}

	// Ensure organization exists in local database
	if err := ensureOrganizationExists(ctx, app, orgIDStr); err != nil {
		return uuid.Nil, fmt.Errorf("failed to ensure organization exists: %w", err)
	}

	return orgID, nil
}

// ensureOrganizationExists checks if organization exists in local database,
// and if not, creates a minimal record. Better Auth is the source of truth.
func ensureOrganizationExists(ctx context.Context, app *storage.App, organizationID string) error {
	var org types.Organization
	err := app.Store.DB.NewSelect().
		Model(&org).
		Where("id = ?", organizationID).
		Scan(ctx)

	// If organization exists, we're done
	if err == nil && org.ID != uuid.Nil {
		return nil
	}

	// Organization doesn't exist locally, create minimal record
	// Better Auth is the source of truth, we just need a local record for foreign key constraints
	orgUUID, err := uuid.Parse(organizationID)
	if err != nil {
		return fmt.Errorf("invalid organization ID format: %w", err)
	}

	newOrg := types.Organization{
		ID:        orgUUID,
		Name:      "Organization " + organizationID[:8], // Minimal name, Better Auth has the real name
		Slug:      organizationID[:8],                   // Use first 8 chars as slug
		CreatedAt: time.Now(),
	}

	_, err = app.Store.DB.NewInsert().
		Model(&newOrg).
		On("CONFLICT (id) DO NOTHING").
		Exec(ctx)
	if err != nil {
		// Check again in case it was created concurrently
		var checkOrg types.Organization
		checkErr := app.Store.DB.NewSelect().
			Model(&checkOrg).
			Where("id = ?", organizationID).
			Scan(ctx)
		if checkErr != nil || checkOrg.ID == uuid.Nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}
	}

	return nil
}

// GetOrganizationIDString gets organization ID as string from context or Better Auth
func GetOrganizationIDString(ctx context.Context, r *http.Request, app *storage.App) (string, error) {
	// First try context
	orgIDAny := ctx.Value(types.OrganizationIDKey)
	if orgIDAny != nil {
		if strID, ok := orgIDAny.(string); ok && strID != "" {
			return strID, nil
		}
		if id, ok := orgIDAny.(uuid.UUID); ok && id != uuid.Nil {
			return id.String(), nil
		}
	}

	// Try Better Auth
	orgIDStr, err := GetOrganizationIDFromBetterAuth(r)
	if err != nil || orgIDStr == "" {
		return "", fmt.Errorf("organization ID not found: %w", err)
	}

	// Ensure organization exists
	if app != nil {
		if err := ensureOrganizationExists(ctx, app, orgIDStr); err != nil {
			return "", fmt.Errorf("failed to ensure organization exists: %w", err)
		}
	}

	return orgIDStr, nil
}

// GetOrganizationSettings gets organization settings, creating default if not exists
func GetOrganizationSettings(ctx context.Context, db *bun.DB, orgID uuid.UUID) (types.OrganizationSettingsData, error) {
	var settings types.OrganizationSettings
	err := db.NewSelect().
		Model(&settings).
		Where("organization_id = ?", orgID.String()).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			// Create default settings if they don't exist
			defaultSettings := &types.OrganizationSettings{
				ID:             uuid.New(),
				OrganizationID: orgID,
				Settings:       types.DefaultOrganizationSettingsData(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			_, insertErr := db.NewInsert().
				Model(defaultSettings).
				On("CONFLICT (organization_id) DO NOTHING").
				Exec(ctx)
			if insertErr == nil {
				return defaultSettings.Settings, nil
			}
			// If insert failed, try to read again (might have been created concurrently)
			err = db.NewSelect().
				Model(&settings).
				Where("organization_id = ?", orgID.String()).
				Scan(ctx)
			if err != nil {
				return types.DefaultOrganizationSettingsData(), nil
			}
		} else {
			return types.DefaultOrganizationSettingsData(), nil
		}
	}

	// Merge with defaults to ensure all fields are set
	defaults := types.DefaultOrganizationSettingsData()
	result := types.OrganizationSettingsData{
		WebsocketReconnectAttempts:       settings.Settings.WebsocketReconnectAttempts,
		WebsocketReconnectInterval:       settings.Settings.WebsocketReconnectInterval,
		ApiRetryAttempts:                 settings.Settings.ApiRetryAttempts,
		DisableApiCache:                  settings.Settings.DisableApiCache,
		ContainerLogTailLines:            defaults.ContainerLogTailLines,
		ContainerDefaultRestartPolicy:    defaults.ContainerDefaultRestartPolicy,
		ContainerStopTimeout:             defaults.ContainerStopTimeout,
		ContainerAutoPruneDanglingImages: defaults.ContainerAutoPruneDanglingImages,
		ContainerAutoPruneBuildCache:     defaults.ContainerAutoPruneBuildCache,
		DeploymentLogsCleanupEnabled:     defaults.DeploymentLogsCleanupEnabled,
		DeploymentLogsRetentionDays:      defaults.DeploymentLogsRetentionDays,
		AuditLogsCleanupEnabled:          defaults.AuditLogsCleanupEnabled,
		AuditLogsRetentionDays:           defaults.AuditLogsRetentionDays,
		ExtensionLogsCleanupEnabled:      defaults.ExtensionLogsCleanupEnabled,
		ExtensionLogsRetentionDays:       defaults.ExtensionLogsRetentionDays,
	}

	if settings.Settings.ContainerLogTailLines != nil {
		result.ContainerLogTailLines = settings.Settings.ContainerLogTailLines
	}
	if settings.Settings.ContainerDefaultRestartPolicy != nil {
		result.ContainerDefaultRestartPolicy = settings.Settings.ContainerDefaultRestartPolicy
	}
	if settings.Settings.ContainerStopTimeout != nil {
		result.ContainerStopTimeout = settings.Settings.ContainerStopTimeout
	}
	if settings.Settings.ContainerAutoPruneDanglingImages != nil {
		result.ContainerAutoPruneDanglingImages = settings.Settings.ContainerAutoPruneDanglingImages
	}
	if settings.Settings.ContainerAutoPruneBuildCache != nil {
		result.ContainerAutoPruneBuildCache = settings.Settings.ContainerAutoPruneBuildCache
	}
	if settings.Settings.DeploymentLogsCleanupEnabled != nil {
		result.DeploymentLogsCleanupEnabled = settings.Settings.DeploymentLogsCleanupEnabled
	}
	if settings.Settings.DeploymentLogsRetentionDays != nil {
		result.DeploymentLogsRetentionDays = settings.Settings.DeploymentLogsRetentionDays
	}
	if settings.Settings.AuditLogsCleanupEnabled != nil {
		result.AuditLogsCleanupEnabled = settings.Settings.AuditLogsCleanupEnabled
	}
	if settings.Settings.AuditLogsRetentionDays != nil {
		result.AuditLogsRetentionDays = settings.Settings.AuditLogsRetentionDays
	}
	if settings.Settings.ExtensionLogsCleanupEnabled != nil {
		result.ExtensionLogsCleanupEnabled = settings.Settings.ExtensionLogsCleanupEnabled
	}
	if settings.Settings.ExtensionLogsRetentionDays != nil {
		result.ExtensionLogsRetentionDays = settings.Settings.ExtensionLogsRetentionDays
	}

	return result, nil
}
