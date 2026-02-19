package loader

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/extension/parser"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type ExtensionLoader struct {
	db *bun.DB
}

func NewExtensionLoader(db *bun.DB) *ExtensionLoader {
	return &ExtensionLoader{
		db: db,
	}
}

func (l *ExtensionLoader) LoadExtensionsFromDirectory(ctx context.Context, dirPath string) error {
	parser := parser.NewParser()

	extensions, allVariables, err := parser.LoadExtensionsFromDirectory(dirPath)
	if err != nil {
		return fmt.Errorf("failed to load extensions from directory: %w", err)
	}

	log.Printf("Found %d extension files in %s", len(extensions), dirPath)

	if len(extensions) == 0 {
		return nil
	}

	// Collect extension IDs found in templates directory
	foundExtensionIDs := make(map[string]bool)
	for _, extension := range extensions {
		foundExtensionIDs[extension.ExtensionID] = true
	}

	// Batch fetch all existing extensions in one query
	extensionIDs := make([]string, 0, len(extensions))
	for _, ext := range extensions {
		extensionIDs = append(extensionIDs, ext.ExtensionID)
	}

	var existingExtensions []types.Extension
	err = l.db.NewSelect().
		Model(&existingExtensions).
		Where("extension_id IN (?) AND deleted_at IS NULL", bun.In(extensionIDs)).
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch existing extensions: %w", err)
	}

	// Create a map of existing extensions by extension_id
	existingMap := make(map[string]*types.Extension)
	for i := range existingExtensions {
		existingMap[existingExtensions[i].ExtensionID] = &existingExtensions[i]
	}

	// Separate extensions into insert, update, and skip batches
	var toInsert []*types.Extension
	var insertVariables [][]types.ExtensionVariable
	var toUpdate []*types.Extension
	var updateVariables [][]types.ExtensionVariable
	skippedCount := 0

	for i, extension := range extensions {
		variables := allVariables[i]
		existing, exists := existingMap[extension.ExtensionID]

		if exists {
			// Check if content hash changed
			if existing.ContentHash == extension.ContentHash {
				skippedCount++
				continue
			}
			// Extension exists but changed - needs update
			extension.ID = existing.ID
			extension.CreatedAt = existing.CreatedAt
			toUpdate = append(toUpdate, extension)
			updateVariables = append(updateVariables, variables)
		} else {
			// New extension - needs insert
			extension.ID = uuid.New()
			toInsert = append(toInsert, extension)
			insertVariables = append(insertVariables, variables)
		}
	}

	log.Printf("Processing: %d new, %d updated, %d unchanged", len(toInsert), len(toUpdate), skippedCount)

	// Batch insert new extensions
	if len(toInsert) > 0 {
		if err := l.batchInsertExtensions(ctx, toInsert, insertVariables); err != nil {
			return fmt.Errorf("failed to batch insert extensions: %w", err)
		}
	}

	// Batch update changed extensions
	if len(toUpdate) > 0 {
		if err := l.batchUpdateExtensions(ctx, toUpdate, updateVariables); err != nil {
			return fmt.Errorf("failed to batch update extensions: %w", err)
		}
	}

	// // Remove extensions from database that are no longer in templates directory
	// if err := l.removeDeletedExtensions(ctx, foundExtensionIDs); err != nil {
	// 	log.Printf("Warning: Failed to remove deleted extensions: %v", err)
	// 	// Don't return error here as the main loading succeeded
	// }

	return nil
}

// batchInsertExtensions inserts multiple extensions and their variables in a single transaction
func (l *ExtensionLoader) batchInsertExtensions(ctx context.Context, extensions []*types.Extension, allVariables [][]types.ExtensionVariable) error {
	if len(extensions) == 0 {
		return nil
	}

	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Bulk insert extensions
	if _, err := tx.NewInsert().Model(&extensions).Exec(ctx); err != nil {
		return fmt.Errorf("failed to batch insert extensions: %w", err)
	}

	// Collect all variables for bulk insert
	var allVars []types.ExtensionVariable
	for i, extension := range extensions {
		variables := allVariables[i]
		for j := range variables {
			variables[j].ID = uuid.New()
			variables[j].ExtensionID = extension.ID
			allVars = append(allVars, variables[j])
		}
	}

	// Bulk insert all variables
	if len(allVars) > 0 {
		if _, err := tx.NewInsert().Model(&allVars).Exec(ctx); err != nil {
			return fmt.Errorf("failed to batch insert variables: %w", err)
		}
	}

	return tx.Commit()
}

// batchUpdateExtensions updates multiple extensions and their variables in batches
func (l *ExtensionLoader) batchUpdateExtensions(ctx context.Context, extensions []*types.Extension, allVariables [][]types.ExtensionVariable) error {
	if len(extensions) == 0 {
		return nil
	}

	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Collect extension IDs for bulk variable deletion
	extensionIDs := make([]uuid.UUID, len(extensions))
	for i, ext := range extensions {
		extensionIDs[i] = ext.ID
	}

	// Bulk delete old variables
	if _, err := tx.NewDelete().
		Model((*types.ExtensionVariable)(nil)).
		Where("extension_id IN (?)", bun.In(extensionIDs)).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to bulk delete variables: %w", err)
	}

	// Update each extension individually (bun doesn't support bulk update with different values)
	for _, extension := range extensions {
		if _, err := tx.NewUpdate().
			Model(extension).
			Column("name", "description", "author", "icon", "category", "extension_type", "version", "is_verified", "featured", "yaml_content", "parsed_content", "content_hash", "validation_status", "updated_at").
			Where("id = ?", extension.ID).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to update extension %s: %w", extension.ExtensionID, err)
		}
	}

	// Collect all variables for bulk insert
	var allVars []types.ExtensionVariable
	for i, extension := range extensions {
		variables := allVariables[i]
		for j := range variables {
			variables[j].ID = uuid.New()
			variables[j].ExtensionID = extension.ID
			allVars = append(allVars, variables[j])
		}
	}

	// Bulk insert new variables
	if len(allVars) > 0 {
		if _, err := tx.NewInsert().Model(&allVars).Exec(ctx); err != nil {
			return fmt.Errorf("failed to batch insert variables: %w", err)
		}
	}

	return tx.Commit()
}

// upsertExtension is kept for backward compatibility but is no longer used in batch operations
func (l *ExtensionLoader) upsertExtension(ctx context.Context, extension *types.Extension, variables []types.ExtensionVariable) error {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var existingExtension types.Extension
	err = tx.NewSelect().
		Model(&existingExtension).
		Where("extension_id = ?", extension.ExtensionID).
		Scan(ctx)

	if err != nil && err.Error() != "sql: no rows in result set" {
		return fmt.Errorf("failed to check existing extension: %w", err)
	}

	if err == nil {
		if existingExtension.ContentHash == extension.ContentHash {
			// log.Printf("Extension %s unchanged, skipping", extension.ExtensionID)
			return tx.Commit()
		}

		extension.ID = existingExtension.ID
		extension.CreatedAt = existingExtension.CreatedAt

		if _, err := tx.NewUpdate().
			Model(extension).
			Where("id = ?", extension.ID).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to update extension: %w", err)
		}

		if err := l.deleteExtensionVariables(ctx, tx, extension.ID); err != nil {
			return fmt.Errorf("failed to delete old variables: %w", err)
		}
	} else {
		extension.ID = uuid.New()

		if _, err := tx.NewInsert().
			Model(extension).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert extension: %w", err)
		}
	}

	for i := range variables {
		variables[i].ID = uuid.New()
		variables[i].ExtensionID = extension.ID
	}

	if len(variables) > 0 {
		if _, err := tx.NewInsert().
			Model(&variables).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert variables: %w", err)
		}
	}

	return tx.Commit()
}

func (l *ExtensionLoader) deleteExtensionVariables(ctx context.Context, tx bun.Tx, extensionID uuid.UUID) error {
	_, err := tx.NewDelete().
		Model((*types.ExtensionVariable)(nil)).
		Where("extension_id = ?", extensionID).
		Exec(ctx)
	return err
}

func (l *ExtensionLoader) removeDeletedExtensions(ctx context.Context, foundExtensionIDs map[string]bool) error {
	// Get all extensions from database that are not deleted
	var allExtensions []types.Extension
	err := l.db.NewSelect().
		Model(&allExtensions).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to query extensions: %w", err)
	}

	// Find extensions that exist in database but not in templates directory
	var extensionsToDelete []string
	for _, ext := range allExtensions {
		if !foundExtensionIDs[ext.ExtensionID] {
			extensionsToDelete = append(extensionsToDelete, ext.ExtensionID)
		}
	}

	if len(extensionsToDelete) == 0 {
		return nil
	}

	log.Printf("Removing %d extensions that are no longer in templates directory", len(extensionsToDelete))

	// Soft delete extensions that are no longer in templates
	_, err = l.db.NewUpdate().
		Model((*types.Extension)(nil)).
		Set("deleted_at = NOW()").
		Where("extension_id IN (?) AND deleted_at IS NULL", bun.In(extensionsToDelete)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete removed extensions: %w", err)
	}

	return nil
}

func (l *ExtensionLoader) LoadExtensionsFromTemplates(ctx context.Context) error {
	templatesPath := filepath.Join(".", "templates")
	return l.LoadExtensionsFromDirectory(ctx, templatesPath)
}

func (l *ExtensionLoader) GetExtensionByID(ctx context.Context, extensionID string) (*types.Extension, error) {
	var extension types.Extension

	err := l.db.NewSelect().
		Model(&extension).
		Relation("Variables").
		Where("extension_id = ? AND deleted_at IS NULL", extensionID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get extension: %w", err)
	}

	return &extension, nil
}

func (l *ExtensionLoader) ListExtensions(ctx context.Context, category *types.ExtensionCategory) ([]types.Extension, error) {
	var extensions []types.Extension

	query := l.db.NewSelect().
		Model(&extensions).
		Relation("Variables").
		Where("deleted_at IS NULL")

	if category != nil {
		query = query.Where("category = ?", *category)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list extensions: %w", err)
	}

	return extensions, nil
}
